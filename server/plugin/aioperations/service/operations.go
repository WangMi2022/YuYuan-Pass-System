package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvocationSearch struct {
	commonRequest.PageInfo
	Status   string `form:"status"`
	Module   string `form:"module"`
	Provider string `form:"provider"`
	UserID   uint   `form:"userId"`
}

type PromptInput struct {
	PromptKey    string `json:"promptKey"`
	Content      string `json:"content"`
	OutputSchema string `json:"outputSchema"`
}

type PromptActivation struct {
	PromptKey string `json:"promptKey"`
	Version   int    `json:"version"`
}

type UsageSummary struct {
	TodayRequests int64 `json:"todayRequests"`
	TodayTokens   int64 `json:"todayTokens"`
	MonthCost     int64 `json:"monthCostMicros"`
	TotalRequests int64 `json:"totalRequests"`
}

type operationsService struct{}

func (operationsService) Providers() config.AI {
	global.GVA_CONFIG_LOCK.Lock()
	configuration := global.GVA_CONFIG.AI
	global.GVA_CONFIG_LOCK.Unlock()
	configuration.Invoice = config.InvoiceRecognition{}
	configuration.Normalize()
	return configuration.Redacted()
}

func (operationsService) UpdateProviders(ctx context.Context, incoming config.AI) (config.AI, error) {
	global.GVA_CONFIG_LOCK.Lock()
	defer global.GVA_CONFIG_LOCK.Unlock()
	current := global.GVA_CONFIG.AI
	incoming.Invoice = current.Invoice
	prepared := incoming.MergeSecrets(current, true)
	prepared.Normalize()
	if err := prepared.Validate(); err != nil {
		return config.AI{}, err
	}
	copy := global.GVA_CONFIG
	copy.AI = prepared
	for key, value := range map[string]any{
		"ai": copy.AI,
	} {
		global.GVA_VP.Set(key, value)
	}
	if err := global.GVA_VP.WriteConfig(); err != nil {
		return config.AI{}, err
	}
	global.GVA_CONFIG.AI = prepared
	prepared.Invoice = config.InvoiceRecognition{}
	return prepared.Redacted(), nil
}

func (operationsService) UsageSummary(ctx context.Context, userID uint) (UsageSummary, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&ai.ModelInvocation{}).Where("user_id = ? AND status = ?", userID, ai.InvocationStatusSuccess)
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	summary := UsageSummary{}
	if err := query.Count(&summary.TotalRequests).Error; err != nil {
		return UsageSummary{}, err
	}
	if err := query.Select("COUNT(*) AS today_requests, COALESCE(SUM(input_tokens + output_tokens), 0) AS today_tokens").Where("created_at >= ?", dayStart).Scan(&summary).Error; err != nil {
		return UsageSummary{}, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&ai.ModelInvocation{}).Where("user_id = ? AND status = ? AND created_at >= ?", userID, ai.InvocationStatusSuccess, monthStart).Select("COALESCE(SUM(estimated_cost_micros), 0)").Scan(&summary.MonthCost).Error; err != nil {
		return UsageSummary{}, err
	}
	return summary, nil
}

func (operationsService) Invocations(ctx context.Context, search InvocationSearch) ([]ai.ModelInvocation, int64, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&ai.ModelInvocation{}).Order("created_at DESC")
	if value := strings.TrimSpace(search.Status); value != "" {
		query = query.Where("status = ?", value)
	}
	if value := strings.TrimSpace(search.Module); value != "" {
		query = query.Where("module = ?", value)
	}
	if value := strings.TrimSpace(search.Provider); value != "" {
		query = query.Where("provider = ?", value)
	}
	if search.UserID != 0 {
		query = query.Where("user_id = ?", search.UserID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []ai.ModelInvocation
	if err := query.Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (operationsService) Quotas(ctx context.Context) ([]ai.UsageQuota, error) {
	var list []ai.UsageQuota
	return list, global.GVA_DB.WithContext(ctx).Order("scope_type ASC, scope_id ASC").Find(&list).Error
}

func (operationsService) QuotaPage(ctx context.Context, pageInfo *commonRequest.PageInfo) ([]ai.UsageQuota, int64, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&ai.UsageQuota{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []ai.UsageQuota
	err := query.Order("scope_type ASC, scope_id ASC").Scopes(pageInfo.Paginate()).Find(&list).Error
	return list, total, err
}

func (operationsService) SaveQuota(ctx context.Context, quota ai.UsageQuota) (ai.UsageQuota, error) {
	quota.ScopeType = strings.TrimSpace(quota.ScopeType)
	quota.ScopeID = strings.TrimSpace(quota.ScopeID)
	if !validQuotaScope(quota.ScopeType) || quota.ScopeID == "" || len(quota.ScopeID) > 120 {
		return ai.UsageQuota{}, errors.New("AI 配额范围不正确")
	}
	if quota.DailyRequests < 0 || quota.DailyTokens < 0 || quota.MonthlyCostMicros < 0 || quota.MaxConcurrency < 0 {
		return ai.UsageQuota{}, errors.New("AI 配额不能为负数")
	}
	if quota.ID == 0 {
		var existing ai.UsageQuota
		err := global.GVA_DB.WithContext(ctx).Where("scope_type = ? AND scope_id = ?", quota.ScopeType, quota.ScopeID).First(&existing).Error
		if err == nil {
			quota.ID = existing.ID
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ai.UsageQuota{}, err
		}
	}
	if quota.ID == 0 {
		return quota, global.GVA_DB.WithContext(ctx).Create(&quota).Error
	}
	return quota, global.GVA_DB.WithContext(ctx).Model(&ai.UsageQuota{}).Where("id = ?", quota.ID).Updates(map[string]any{
		"scope_type": quota.ScopeType, "scope_id": quota.ScopeID, "daily_requests": quota.DailyRequests, "daily_tokens": quota.DailyTokens,
		"monthly_cost_micros": quota.MonthlyCostMicros, "max_concurrency": quota.MaxConcurrency, "enabled": quota.Enabled,
	}).Error
}

func (operationsService) Prompts(ctx context.Context) ([]ai.PromptTemplate, error) {
	var list []ai.PromptTemplate
	return list, global.GVA_DB.WithContext(ctx).Order("prompt_key ASC, version DESC").Find(&list).Error
}

func (operationsService) PromptPage(ctx context.Context, pageInfo *commonRequest.PageInfo) ([]ai.PromptTemplate, int64, error) {
	query := global.GVA_DB.WithContext(ctx).Model(&ai.PromptTemplate{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []ai.PromptTemplate
	err := query.Order("prompt_key ASC, version DESC").Scopes(pageInfo.Paginate()).Find(&list).Error
	return list, total, err
}

func (operationsService) CreatePrompt(ctx context.Context, input PromptInput, userID uint) (ai.PromptTemplate, error) {
	input.PromptKey = strings.TrimSpace(input.PromptKey)
	input.Content = strings.TrimSpace(input.Content)
	input.OutputSchema = strings.TrimSpace(input.OutputSchema)
	if input.PromptKey == "" || len(input.PromptKey) > 120 || input.Content == "" || len(input.Content) > 128*1024 || len(input.OutputSchema) > 128*1024 {
		return ai.PromptTemplate{}, errors.New("Prompt 模板参数不正确")
	}
	if err := ai.ValidateOutputSchema(input.OutputSchema); err != nil {
		return ai.PromptTemplate{}, err
	}
	for attempt := 0; attempt < 8; attempt++ {
		var latestVersion int
		if err := global.GVA_DB.WithContext(ctx).Model(&ai.PromptTemplate{}).Where("prompt_key = ?", input.PromptKey).Select("COALESCE(MAX(version), 0)").Scan(&latestVersion).Error; err != nil {
			return ai.PromptTemplate{}, err
		}
		prompt := ai.PromptTemplate{PromptKey: input.PromptKey, Version: latestVersion + 1, Content: input.Content, OutputSchema: input.OutputSchema, Status: ai.PromptStatusDraft, CreatedBy: userID}
		result := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&prompt)
		if result.Error != nil {
			return ai.PromptTemplate{}, result.Error
		}
		if result.RowsAffected == 1 {
			return prompt, nil
		}
	}
	return ai.PromptTemplate{}, errors.New("Prompt 版本创建冲突，请重试")
}

func (operationsService) ActivatePrompt(ctx context.Context, input PromptActivation) error {
	input.PromptKey = strings.TrimSpace(input.PromptKey)
	if input.PromptKey == "" || input.Version <= 0 {
		return errors.New("Prompt 模板版本不正确")
	}
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&ai.PromptTemplate{}).Where("prompt_key = ? AND version = ?", input.PromptKey, input.Version).Updates(map[string]any{"status": ai.PromptStatusActive, "activated_at": time.Now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("Prompt 模板不存在")
		}
		return tx.Model(&ai.PromptTemplate{}).Where("prompt_key = ? AND version <> ? AND status = ?", input.PromptKey, input.Version, ai.PromptStatusActive).Update("status", ai.PromptStatusRetired).Error
	})
}

func validQuotaScope(scope string) bool {
	switch scope {
	case ai.QuotaScopeGlobal, ai.QuotaScopeModule, ai.QuotaScopeAuthority, ai.QuotaScopeUser:
		return true
	default:
		return false
	}
}
