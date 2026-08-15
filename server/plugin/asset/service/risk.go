package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	announcementService "github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/service"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	assetResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Risk = new(riskService)

type riskService struct{}

const riskWriteTimeout = 5 * time.Second

var errRiskWriteBusy = errors.New("风险数据正在被扫描更新，请稍后重试")

func withRiskWriteTransaction(ctx context.Context, transaction func(*gorm.DB) error) error {
	return withRiskWriteTransactionTimeout(ctx, riskWriteTimeout, transaction)
}

func withRiskWriteTransactionTimeout(ctx context.Context, timeout time.Duration, transaction func(*gorm.DB) error) error {
	writeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := global.GVA_DB.WithContext(writeCtx).Transaction(transaction)
	if err != nil && ctx.Err() == nil && errors.Is(writeCtx.Err(), context.DeadlineExceeded) {
		return errRiskWriteBusy
	}
	return err
}

func (s *riskService) SeedRules(ctx context.Context) {
	for _, rule := range defaultRiskRules() {
		if err := global.GVA_DB.WithContext(ctx).Where("code = ?", rule.Code).FirstOrCreate(&rule).Error; err != nil && global.GVA_LOG != nil {
			global.GVA_LOG.Warn("预置资产风险规则失败", zap.String("ruleCode", rule.Code), zap.Error(err))
		}
	}
}

func (s *riskService) Rules(ctx context.Context) ([]model.AssetRiskRule, error) {
	var rules []model.AssetRiskRule
	return rules, global.GVA_DB.WithContext(ctx).Order("category ASC, severity DESC, code ASC").Find(&rules).Error
}

func (s *riskService) UpdateRule(ctx context.Context, input assetRequest.RiskRuleUpdate) (model.AssetRiskRule, error) {
	if input.ID == 0 {
		return model.AssetRiskRule{}, errors.New("风险规则 ID 不正确")
	}
	if !validRiskSeverity(input.Severity) {
		return model.AssetRiskRule{}, errors.New("风险等级不正确")
	}
	if input.Parameters == nil {
		input.Parameters = map[string]any{}
	}
	var rule model.AssetRiskRule
	err := withRiskWriteTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&rule, input.ID).Error; err != nil {
			return errors.New("风险规则不存在")
		}
		if err := ValidateRiskRuleParameters(rule.Code, input.Parameters); err != nil {
			return err
		}
		rule.Severity = input.Severity
		rule.Parameters = input.Parameters
		rule.Enabled = input.Enabled
		rule.Version++
		return tx.Model(&rule).Select("severity", "parameters", "enabled", "version", "updated_at").Updates(&rule).Error
	})
	return rule, err
}

func (s *riskService) List(ctx context.Context, search assetRequest.RiskSearch) ([]model.AssetRiskEvent, int64, error) {
	var list []model.AssetRiskEvent
	var total int64
	db := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEvent{})
	if status := strings.TrimSpace(search.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if severity := strings.TrimSpace(search.Severity); severity != "" {
		db = db.Where("severity = ?", severity)
	}
	if category := strings.TrimSpace(search.Category); category != "" {
		db = db.Where("category = ?", category)
	}
	if ruleCode := strings.TrimSpace(search.RuleCode); ruleCode != "" {
		db = db.Where("rule_code = ?", ruleCode)
	}
	if search.AssetID != 0 {
		db = db.Where("asset_id = ?", search.AssetID)
	}
	if search.AssignedTo != 0 {
		db = db.Where("assigned_to = ?", search.AssignedTo)
	}
	if keyword := strings.ToLower(strings.TrimSpace(search.Keyword)); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(`LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(rule_code) LIKE ? OR EXISTS (
			SELECT 1 FROM assets a WHERE a.id = asset_risk_events.asset_id AND a.deleted_at IS NULL
			AND (LOWER(a.asset_code) LIKE ? OR LOWER(a.name) LIKE ? OR LOWER(a.custodian) LIKE ?)
		)`, like, like, like, like, like, like)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Asset.Category").Order("CASE severity WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, last_detected_at DESC, id DESC").
		Scopes(search.Paginate()).Find(&list).Error
	return list, total, err
}

func (s *riskService) Detail(ctx context.Context, id uint) (assetResponse.RiskDetail, error) {
	if id == 0 {
		return assetResponse.RiskDetail{}, errors.New("风险事件 ID 不正确")
	}
	var detail assetResponse.RiskDetail
	if err := global.GVA_DB.WithContext(ctx).Preload("Asset.Category").First(&detail.Event, id).Error; err != nil {
		return assetResponse.RiskDetail{}, errors.New("风险事件不存在")
	}
	if err := global.GVA_DB.WithContext(ctx).Where("event_id = ?", id).Order("created_at DESC, id DESC").Find(&detail.Logs).Error; err != nil {
		return assetResponse.RiskDetail{}, err
	}
	return detail, nil
}

func (s *riskService) ScanRuns(ctx context.Context, search assetRequest.RiskScanSearch) ([]model.AssetRiskScanRun, int64, error) {
	var list []model.AssetRiskScanRun
	var total int64
	db := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskScanRun{})
	if status := strings.TrimSpace(search.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if triggerType := strings.TrimSpace(search.TriggerType); triggerType != "" {
		db = db.Where("trigger_type = ?", triggerType)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("started_at DESC, id DESC").Scopes(search.Paginate()).Find(&list).Error
	return list, total, err
}

func (s *riskService) Dashboard(ctx context.Context) (assetResponse.RiskDashboard, error) {
	now := time.Now()
	activeStatuses := []string{model.RiskStatusOpen, model.RiskStatusAcknowledged}
	result := assetResponse.RiskDashboard{GeneratedAt: now}
	active := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEvent{}).Where("status IN ?", activeStatuses)
	if err := active.Count(&result.TotalOpen).Error; err != nil {
		return result, err
	}
	if err := active.Where("severity IN ?", []string{model.RiskSeverityHigh, model.RiskSeverityCritical}).Count(&result.HighOpen).Error; err != nil {
		return result, err
	}
	today := startOfRiskDay(now)
	if err := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEvent{}).Where("first_detected_at >= ?", today).Count(&result.TodayNew).Error; err != nil {
		return result, err
	}
	if err := active.Where("first_detected_at < ?", now.AddDate(0, 0, -7)).Count(&result.Overdue).Error; err != nil {
		return result, err
	}
	var err error
	if result.ByCategory, err = s.groupRiskMetrics(ctx, "category", activeStatuses); err != nil {
		return result, err
	}
	if result.BySeverity, err = s.groupRiskMetrics(ctx, "severity", activeStatuses); err != nil {
		return result, err
	}
	if result.ByStatus, err = s.groupRiskMetrics(ctx, "status", nil); err != nil {
		return result, err
	}
	if result.ByCustodian, err = s.custodianRiskMetrics(ctx, activeStatuses); err != nil {
		return result, err
	}
	if result.Trend, err = s.riskTrend(ctx, now); err != nil {
		return result, err
	}
	if err = global.GVA_DB.WithContext(ctx).Preload("Asset.Category").Order("last_detected_at DESC, id DESC").Limit(8).Find(&result.RecentEvents).Error; err != nil {
		return result, err
	}
	var latest model.AssetRiskScanRun
	if err = global.GVA_DB.WithContext(ctx).Order("started_at DESC, id DESC").First(&latest).Error; err == nil {
		result.LatestScan = &latest
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return result, err
	}
	return result, nil
}

func (s *riskService) groupRiskMetrics(ctx context.Context, column string, statuses []string) ([]assetResponse.RiskMetric, error) {
	if column != "category" && column != "severity" && column != "status" {
		return nil, errors.New("风险统计维度不正确")
	}
	query := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEvent{})
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	var metrics []assetResponse.RiskMetric
	if err := query.Select(column + " AS key, " + column + " AS label, COUNT(*) AS count").Group(column).Order("count DESC").Scan(&metrics).Error; err != nil {
		return nil, err
	}
	for index := range metrics {
		metrics[index].Label = riskMetricLabel(column, metrics[index].Key)
	}
	return metrics, nil
}

func (s *riskService) custodianRiskMetrics(ctx context.Context, statuses []string) ([]assetResponse.RiskMetric, error) {
	var metrics []assetResponse.RiskMetric
	err := global.GVA_DB.WithContext(ctx).Table("asset_risk_events AS e").
		Select("COALESCE(NULLIF(TRIM(a.custodian), ''), '未登记') AS key, COALESCE(NULLIF(TRIM(a.custodian), ''), '未登记') AS label, COUNT(*) AS count").
		Joins("JOIN assets a ON a.id = e.asset_id AND a.deleted_at IS NULL").
		Where("e.deleted_at IS NULL AND e.status IN ?", statuses).
		Group("COALESCE(NULLIF(TRIM(a.custodian), ''), '未登记')").Order("count DESC").Limit(10).Scan(&metrics).Error
	return metrics, err
}

func (s *riskService) riskTrend(ctx context.Context, now time.Time) ([]assetResponse.RiskTrendPoint, error) {
	start := startOfRiskDay(now).AddDate(0, 0, -29)
	points := make([]assetResponse.RiskTrendPoint, 30)
	indexByDate := make(map[string]int, 30)
	for index := range points {
		date := start.AddDate(0, 0, index).Format("2006-01-02")
		points[index] = assetResponse.RiskTrendPoint{Date: date}
		indexByDate[date] = index
	}
	var created []struct{ FirstDetectedAt time.Time }
	if err := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEvent{}).Select("first_detected_at").Where("first_detected_at >= ?", start).Find(&created).Error; err != nil {
		return nil, err
	}
	for _, item := range created {
		if index, ok := indexByDate[item.FirstDetectedAt.In(time.Local).Format("2006-01-02")]; ok {
			points[index].New++
		}
	}
	var resolved []struct{ CreatedAt time.Time }
	if err := global.GVA_DB.WithContext(ctx).Model(&model.AssetRiskEventLog{}).Select("created_at").
		Where("created_at >= ? AND action IN ?", start, []string{"resolve", "ignore", "auto_resolve"}).Find(&resolved).Error; err != nil {
		return nil, err
	}
	for _, item := range resolved {
		if index, ok := indexByDate[item.CreatedAt.In(time.Local).Format("2006-01-02")]; ok {
			points[index].Resolved++
		}
	}
	return points, nil
}

func (s *riskService) Acknowledge(ctx context.Context, input assetRequest.RiskAction, actorID uint, actorName string) error {
	return s.transition(ctx, input, actorID, actorName, "acknowledge")
}

func (s *riskService) Resolve(ctx context.Context, input assetRequest.RiskAction, actorID uint, actorName string) error {
	return s.transition(ctx, input, actorID, actorName, "resolve")
}

func (s *riskService) Ignore(ctx context.Context, input assetRequest.RiskAction, actorID uint, actorName string) error {
	if len(input.IDs) > 1 {
		return errors.New("风险忽略必须逐条确认，不能批量操作")
	}
	return s.transition(ctx, input, actorID, actorName, "ignore")
}

func (s *riskService) Reopen(ctx context.Context, input assetRequest.RiskAction, actorID uint, actorName string) error {
	return s.transition(ctx, input, actorID, actorName, "reopen")
}

func (s *riskService) transition(ctx context.Context, input assetRequest.RiskAction, actorID uint, actorName, action string) error {
	ids, err := normalizedRiskIDs(input.IDs)
	if err != nil {
		return err
	}
	note := strings.TrimSpace(input.Note)
	if (action == "resolve" || action == "ignore" || action == "reopen") && note == "" {
		return errors.New("请填写风险处理说明")
	}
	return withRiskWriteTransaction(ctx, func(tx *gorm.DB) error {
		var events []model.AssetRiskEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&events).Error; err != nil {
			return err
		}
		if len(events) != len(ids) {
			return errors.New("部分风险事件不存在")
		}
		for _, event := range events {
			toState, err := transitionRiskStatus(event.Status, action)
			if err != nil {
				return fmt.Errorf("风险 %d：%w", event.ID, err)
			}
			updates := map[string]any{"status": toState}
			if toState == model.RiskStatusResolved || toState == model.RiskStatusIgnored {
				now := time.Now()
				updates["handled_by"] = actorID
				updates["handled_by_name"] = strings.TrimSpace(actorName)
				updates["handled_at"] = now
				updates["resolution_note"] = note
			} else if toState == model.RiskStatusOpen {
				updates["handled_by"] = 0
				updates["handled_by_name"] = ""
				updates["handled_at"] = nil
				updates["resolution_note"] = ""
			}
			if err := tx.Model(&event).Updates(updates).Error; err != nil {
				return err
			}
			if err := createRiskLog(tx, event.ID, action, event.Status, toState, actorID, actorName, note); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *riskService) Assign(ctx context.Context, input assetRequest.RiskAssignment, actorID uint, actorName string) error {
	ids, err := normalizedRiskIDs(input.IDs)
	if err != nil {
		return err
	}
	assignedName := ""
	if input.AssignedTo != 0 {
		var user system.SysUser
		if err := global.GVA_DB.WithContext(ctx).Select("id", "username", "nick_name", "enable").First(&user, input.AssignedTo).Error; err != nil || user.Enable != 1 {
			return errors.New("指定处理人不存在或已停用")
		}
		assignedName = strings.TrimSpace(user.NickName)
		if assignedName == "" {
			assignedName = user.Username
		}
	}
	var notifyEvents []model.AssetRiskEvent
	err = withRiskWriteTransaction(ctx, func(tx *gorm.DB) error {
		var events []model.AssetRiskEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&events).Error; err != nil {
			return err
		}
		if len(events) != len(ids) {
			return errors.New("部分风险事件不存在")
		}
		for _, event := range events {
			if err := tx.Model(&event).Updates(map[string]any{"assigned_to": input.AssignedTo, "assigned_to_name": assignedName}).Error; err != nil {
				return err
			}
			note := "取消分配"
			if input.AssignedTo != 0 {
				note = "分配给 " + assignedName
			}
			if err := createRiskLog(tx, event.ID, "assign", event.Status, event.Status, actorID, actorName, note); err != nil {
				return err
			}
			if input.AssignedTo != 0 && isHighRisk(event.Severity) {
				notifyEvents = append(notifyEvents, event)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, event := range notifyEvents {
		announcementService.NotificationHub.PublishToUser(input.AssignedTo, announcementService.NotificationEvent{ID: event.ID, Title: event.Title, Kind: "asset-risk"})
	}
	return nil
}

func transitionRiskStatus(fromState, action string) (string, error) {
	switch action {
	case "acknowledge":
		if fromState != model.RiskStatusOpen {
			return "", errors.New("只有待处理风险可以确认")
		}
		return model.RiskStatusAcknowledged, nil
	case "resolve":
		if fromState != model.RiskStatusOpen && fromState != model.RiskStatusAcknowledged {
			return "", errors.New("当前风险状态不能标记解决")
		}
		return model.RiskStatusResolved, nil
	case "ignore":
		if fromState != model.RiskStatusOpen && fromState != model.RiskStatusAcknowledged {
			return "", errors.New("当前风险状态不能忽略")
		}
		return model.RiskStatusIgnored, nil
	case "reopen":
		if fromState != model.RiskStatusResolved && fromState != model.RiskStatusIgnored {
			return "", errors.New("只有已解决或已忽略风险可以重新打开")
		}
		return model.RiskStatusOpen, nil
	default:
		return "", errors.New("风险处理动作不正确")
	}
}

func normalizedRiskIDs(input []uint) ([]uint, error) {
	if len(input) == 0 || len(input) > 100 {
		return nil, errors.New("请选择 1 到 100 条风险事件")
	}
	seen := make(map[uint]struct{}, len(input))
	result := make([]uint, 0, len(input))
	for _, id := range input {
		if id == 0 {
			return nil, errors.New("风险事件 ID 不正确")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

func validRiskSeverity(severity string) bool {
	switch severity {
	case model.RiskSeverityLow, model.RiskSeverityMedium, model.RiskSeverityHigh, model.RiskSeverityCritical:
		return true
	default:
		return false
	}
}

func riskMetricLabel(dimension, key string) string {
	labels := map[string]map[string]string{
		"category": {"status": "状态异常", "value": "价值异常", "return": "归还与调拨", "maintenance": "维修异常", "warranty": "质保风险", "duplicate": "重复数据"},
		"severity": {model.RiskSeverityLow: "低风险", model.RiskSeverityMedium: "中风险", model.RiskSeverityHigh: "高风险", model.RiskSeverityCritical: "严重风险"},
		"status":   {model.RiskStatusOpen: "待处理", model.RiskStatusAcknowledged: "已确认", model.RiskStatusResolved: "已解决", model.RiskStatusIgnored: "已忽略"},
	}
	if label := labels[dimension][key]; label != "" {
		return label
	}
	return key
}
