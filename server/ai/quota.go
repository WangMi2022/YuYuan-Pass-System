package ai

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/gorm"
)

type quotaEstimate struct {
	Requests   int64
	Tokens     int64
	CostMicros int64
}

type quotaPending struct {
	quotaEstimate
	Concurrency int
}

type quotaReservation struct {
	key      string
	estimate quotaEstimate
}

type quotaLease struct {
	reservations []quotaReservation
}

var quotaState = struct {
	sync.Mutex
	pending map[string]quotaPending
}{pending: make(map[string]quotaPending)}

func (l quotaLease) releaseLocked() {
	for _, reservation := range l.reservations {
		pending := quotaState.pending[reservation.key]
		pending.Requests -= reservation.estimate.Requests
		pending.Tokens -= reservation.estimate.Tokens
		pending.CostMicros -= reservation.estimate.CostMicros
		pending.Concurrency--
		if pending.Requests <= 0 && pending.Tokens <= 0 && pending.CostMicros <= 0 && pending.Concurrency <= 0 {
			delete(quotaState.pending, reservation.key)
			continue
		}
		quotaState.pending[reservation.key] = pending
	}
}

func acquireQuota(ctx context.Context, request CompletionRequest, estimate quotaEstimate) (quotaLease, error) {
	if global.GVA_DB == nil {
		return quotaLease{}, nil
	}
	quotaState.Lock()
	defer quotaState.Unlock()
	quotas, err := matchedQuotas(ctx, request)
	if err != nil {
		return quotaLease{}, &Error{Type: ErrorTypeProvider, Message: "读取 AI 配额失败", Cause: err}
	}
	lease := quotaLease{reservations: make([]quotaReservation, 0, len(quotas))}
	for _, quota := range quotas {
		key := quota.ScopeType + ":" + quota.ScopeID
		pending := quotaState.pending[key]
		if err := verifyQuotaUsage(ctx, quota, request, pending, estimate); err != nil {
			return quotaLease{}, err
		}
		if quota.MaxConcurrency > 0 && pending.Concurrency >= quota.MaxConcurrency {
			return quotaLease{}, &Error{Type: ErrorTypeQuota, Message: "AI 并发调用已达到配额上限"}
		}
		lease.reservations = append(lease.reservations, quotaReservation{key: key, estimate: estimate})
	}
	for _, reservation := range lease.reservations {
		pending := quotaState.pending[reservation.key]
		pending.Requests += estimate.Requests
		pending.Tokens += estimate.Tokens
		pending.CostMicros += estimate.CostMicros
		pending.Concurrency++
		quotaState.pending[reservation.key] = pending
	}
	return lease, nil
}

func matchedQuotas(ctx context.Context, request CompletionRequest) ([]UsageQuota, error) {
	targets := [][2]string{
		{QuotaScopeGlobal, "global"}, {QuotaScopeModule, request.Module},
		{QuotaScopeAuthority, strconv.FormatUint(uint64(request.AuthorityID), 10)},
		{QuotaScopeUser, strconv.FormatUint(uint64(request.UserID), 10)},
	}
	quotas := make([]UsageQuota, 0, len(targets))
	for _, target := range targets {
		var quota UsageQuota
		err := global.GVA_DB.WithContext(ctx).Where("scope_type = ? AND scope_id = ? AND enabled = ?", target[0], target[1], true).First(&quota).Error
		if err == nil {
			quotas = append(quotas, quota)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return quotas, nil
}

func verifyQuotaUsage(ctx context.Context, quota UsageQuota, request CompletionRequest, pending quotaPending, estimate quotaEstimate) error {
	if quota.DailyRequests == 0 && quota.DailyTokens == 0 && quota.MonthlyCostMicros == 0 {
		return nil
	}
	query := global.GVA_DB.WithContext(ctx).Model(&ModelInvocation{}).Where("status <> ?", InvocationStatusBlocked)
	switch quota.ScopeType {
	case QuotaScopeUser:
		query = query.Where("user_id = ?", request.UserID)
	case QuotaScopeAuthority:
		query = query.Where("authority_id = ?", request.AuthorityID)
	case QuotaScopeModule:
		query = query.Where("module = ?", request.Module)
	}
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if quota.DailyRequests > 0 || quota.DailyTokens > 0 {
		var today struct {
			Requests int64
			Tokens   int64
		}
		if err := query.Select("COUNT(*) AS requests, COALESCE(SUM(input_tokens + output_tokens), 0) AS tokens").Where("created_at >= ?", dayStart).Scan(&today).Error; err != nil {
			return &Error{Type: ErrorTypeProvider, Message: "统计 AI 日配额失败", Cause: err}
		}
		if quota.DailyRequests > 0 && today.Requests+pending.Requests+estimate.Requests > quota.DailyRequests {
			return &Error{Type: ErrorTypeQuota, Message: "AI 每日调用次数已达到配额上限"}
		}
		if quota.DailyTokens > 0 && today.Tokens+pending.Tokens+estimate.Tokens > quota.DailyTokens {
			return &Error{Type: ErrorTypeQuota, Message: "AI 每日 Token 用量将超过配额上限"}
		}
	}
	if quota.MonthlyCostMicros > 0 {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		var cost int64
		if err := query.Select("COALESCE(SUM(estimated_cost_micros), 0)").Where("created_at >= ?", monthStart).Scan(&cost).Error; err != nil {
			return &Error{Type: ErrorTypeProvider, Message: "统计 AI 月预算失败", Cause: err}
		}
		if cost+pending.CostMicros+estimate.CostMicros > quota.MonthlyCostMicros {
			return &Error{Type: ErrorTypeQuota, Message: "AI 月度预算将超过配额上限"}
		}
	}
	return nil
}
