package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/ai"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	announcementService "github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/service"
	assetModel "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	assetService "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/service"
	invoiceModel "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model/request"
	invoiceService "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/service"
	scheduleModel "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model"
	scheduleRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/model/request"
	scheduleService "github.com/flipped-aurora/gin-vue-admin/server/plugin/schedule/service"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/smart/model"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var toolDefinitions = []ToolDefinition{
	{Name: "asset.search", Description: "按编号、名称、品牌、型号或序列号查询资产", ReadOnly: true},
	{Name: "asset.detail", Description: "查询单项资产详情", ReadOnly: true},
	{Name: "asset.risk.list", Description: "查询开放资产风险和异常", ReadOnly: true},
	{Name: "asset.warranty.expiring", Description: "查询即将到期的资产质保", ReadOnly: true},
	{Name: "asset.custodian.summary", Description: "按保管人汇总资产数量和价值", ReadOnly: true},
	{Name: "asset.operation.summary", Description: "汇总资产流转单据状态", ReadOnly: true},
	{Name: "invoice.summary", Description: "查询当前权限范围内发票汇总", ReadOnly: true},
	{Name: "invoice.pending_reviews", Description: "查询待复核发票", ReadOnly: true},
	{Name: "invoice.failed_recognitions", Description: "查询识别失败发票", ReadOnly: true},
	{Name: "invoice.provider_quality", Description: "查询发票识别 Provider 质量", ReadOnly: true},
	{Name: "schedule.today", Description: "查询本人今日及近期日程", ReadOnly: true},
	{Name: "announcement.unread", Description: "查询本人未读公告", ReadOnly: true},
}

func (s *smartService) Tools(authorityID uint) []ToolDefinition {
	available := make([]ToolDefinition, 0, len(toolDefinitions))
	for _, definition := range toolDefinitions {
		if toolPermissionAllowed(authorityID, definition.Name) {
			available = append(available, definition)
		}
	}
	return available
}

type toolResult struct {
	Data      any
	Answer    string
	Citations []Citation
}

type invoiceProviderQualityRow struct {
	Provider    string  `json:"provider"`
	Total       int64   `json:"total"`
	Failed      int64   `json:"failed"`
	AvgDuration float64 `json:"avgDurationMs"`
}

func classify(question string) (intent, tool string) {
	q := strings.ToLower(strings.TrimSpace(question))
	switch {
	case strings.Contains(q, "质保") || strings.Contains(q, "保修"):
		return "warranty", "asset.warranty.expiring"
	case strings.Contains(q, "风险") || strings.Contains(q, "异常"):
		return "risk", "asset.risk.list"
	case strings.Contains(q, "保管人") || strings.Contains(q, "负责人"):
		return "custodian", "asset.custodian.summary"
	case strings.Contains(q, "流转") || strings.Contains(q, "维修") || strings.Contains(q, "领用") || strings.Contains(q, "报废"):
		return "operation", "asset.operation.summary"
	case strings.Contains(q, "资产") || strings.Contains(q, "设备") || strings.Contains(q, "电脑") || strings.Contains(q, "打印机"):
		if strings.Contains(q, "详情") || strings.Contains(q, "明细") {
			return "asset_detail", "asset.detail"
		}
		return "asset", "asset.search"
	case strings.Contains(q, "发票") || strings.Contains(q, "金额") || strings.Contains(q, "报销"):
		if strings.Contains(q, "质量") || strings.Contains(q, "provider") || strings.Contains(q, "供应商") {
			return "invoice_quality", "invoice.provider_quality"
		}
		if strings.Contains(q, "待复核") || strings.Contains(q, "待审核") {
			return "invoice_review", "invoice.pending_reviews"
		}
		if strings.Contains(q, "失败") || strings.Contains(q, "识别") {
			return "invoice_failed", "invoice.failed_recognitions"
		}
		return "invoice", "invoice.summary"
	case strings.Contains(q, "日程") || strings.Contains(q, "安排") || strings.Contains(q, "今天") || strings.Contains(q, "明天"):
		return "schedule", "schedule.today"
	case strings.Contains(q, "公告") || strings.Contains(q, "通知") || strings.Contains(q, "未读"):
		return "announcement", "announcement.unread"
	default:
		return "asset", "asset.search"
	}
}

func extractKeyword(question string) string {
	value := strings.TrimSpace(question)
	for _, token := range []string{"请问", "帮我", "查询", "查找", "搜索", "列出", "查看", "有哪些", "所有", "资产", "设备", "信息", "情况", "吗", "？", "?"} {
		value = strings.ReplaceAll(value, token, " ")
	}
	return strings.TrimSpace(value)
}

func isWriteIntent(question string) bool {
	for _, marker := range []string{"查询", "查找", "搜索", "查看", "列出", "统计", "汇总", "多少", "哪些", "明细", "详情", "情况"} {
		if strings.Contains(question, marker) {
			return false
		}
	}
	for _, action := range []string{"创建", "新增", "修改", "删除", "提交", "确认", "审核", "领用", "报废", "调拨", "归还", "维修", "入库", "出库"} {
		if strings.Contains(question, action) {
			return true
		}
	}
	return false
}

func (s *smartService) Query(ctx context.Context, userID, authorityID uint, question string, sessionID uint) (CopilotResult, error) {
	question = strings.TrimSpace(question)
	if userID == 0 || question == "" {
		return CopilotResult{}, errors.New("问题不能为空且必须先登录")
	}
	if len([]rune(question)) > 2000 {
		return CopilotResult{}, errors.New("问题长度不能超过 2000 个字符")
	}
	if isWriteIntent(question) {
		return CopilotResult{}, errors.New("业务助手仅支持只读查询，不能创建、修改、提交或删除业务数据")
	}
	intent, tool := classify(question)
	if !toolPermissionAllowed(authorityID, tool) {
		return CopilotResult{}, errors.New("当前角色没有该业务数据的读取权限")
	}
	result, err := s.executeTool(ctx, userID, authorityID, tool, question)
	if err != nil {
		return CopilotResult{}, err
	}
	answer, modelUsed := s.tryModelSummary(ctx, userID, authorityID, question, tool, result.Answer, result.Data)
	if answer == "" {
		answer = result.Answer
	}
	session, err := s.ensureSession(userID, authorityID, question, sessionID)
	if err != nil {
		return CopilotResult{}, err
	}
	citationsJSON := common.JSONMap{}
	for index, item := range result.Citations {
		citationsJSON[strconv.Itoa(index)] = item
	}
	userMessage := model.CopilotMessage{SessionID: session.ID, UserID: userID, AuthorityID: authorityID, Role: model.MessageRoleUser, Content: question, Intent: intent, Tool: tool}
	assistantMessage := model.CopilotMessage{SessionID: session.ID, UserID: userID, AuthorityID: authorityID, Role: model.MessageRoleAssistant, Content: answer, Intent: intent, Tool: tool, Citations: citationsJSON}
	now := time.Now()
	if err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&userMessage).Error; err != nil {
			return err
		}
		if err := tx.Create(&assistantMessage).Error; err != nil {
			return err
		}
		return tx.Model(&model.CopilotSession{}).Where("id = ? AND user_id = ? AND authority_id = ?", session.ID, userID, authorityID).Update("last_message_at", now).Error
	}); err != nil {
		return CopilotResult{}, err
	}
	return CopilotResult{SessionID: session.ID, Question: question, Intent: intent, Tool: tool, Scope: "当前用户可访问的数据范围", Answer: answer, Data: result.Data, Citations: result.Citations, GeneratedAt: now.Format(time.RFC3339), ReadOnly: true, ModelUsed: modelUsed}, nil
}

func (s *smartService) ensureSession(userID, authorityID uint, question string, sessionID uint) (model.CopilotSession, error) {
	var session model.CopilotSession
	if sessionID > 0 {
		if err := global.GVA_DB.Where("id = ? AND user_id = ? AND authority_id = ?", sessionID, userID, authorityID).First(&session).Error; err != nil {
			return session, errors.New("业务助手会话不存在或无权访问")
		}
		return session, nil
	}
	title := strings.TrimSpace(question)
	if len([]rune(title)) > 60 {
		title = string([]rune(title)[:60])
	}
	session = model.CopilotSession{UserID: userID, AuthorityID: authorityID, Title: title, LastMessageAt: time.Now()}
	return session, global.GVA_DB.Create(&session).Error
}

func (s *smartService) executeTool(ctx context.Context, userID, authorityID uint, tool, question string) (toolResult, error) {
	result := toolResult{}
	if !toolPermissionAllowed(authorityID, tool) {
		return result, errors.New("当前角色没有该业务数据的读取权限")
	}
	keyword := extractKeyword(question)
	switch tool {
	case "asset.search":
		list, total, err := assetService.Asset.List(assetRequest.AssetSearch{PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 20, Keyword: keyword}})
		if err != nil {
			return result, err
		}
		result.Data = map[string]any{"list": list, "total": total, "keyword": keyword}
		result.Answer = fmt.Sprintf("资产查询命中 %d 项。", total)
		for _, item := range list {
			result.Citations = append(result.Citations, Citation{Type: "asset", ID: item.ID, Label: item.AssetCode + " " + item.Name, Path: "/assetInventory", Params: "id=" + strconv.Itoa(int(item.ID))})
		}
	case "asset.detail":
		idPattern := regexp.MustCompile(`(?i)(?:资产|asset)?\s*(?:id|编号)?\s*[:：#]?\s*(\d+)`)
		match := idPattern.FindStringSubmatch(question)
		if len(match) != 2 {
			return result, errors.New("查询资产详情时请提供资产 ID，例如“查看资产 ID 12 详情”")
		}
		id, _ := strconv.ParseUint(match[1], 10, 64)
		item, err := assetService.Asset.Get(uint(id))
		if err != nil {
			return result, err
		}
		result.Data = item
		result.Answer = fmt.Sprintf("资产 %s 的详情已找到。", item.AssetCode)
		result.Citations = append(result.Citations, Citation{Type: "asset", ID: item.ID, Label: item.AssetCode + " " + item.Name, Path: "/assetInventory", Params: "id=" + strconv.Itoa(int(item.ID))})
	case "asset.risk.list":
		list, total, err := assetService.Risk.List(ctx, assetRequest.RiskSearch{PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 20}})
		if err != nil {
			return result, err
		}
		result.Data = map[string]any{"list": list, "total": total}
		result.Answer = fmt.Sprintf("当前风险中心共有 %d 条风险记录。", total)
		for _, item := range list {
			result.Citations = append(result.Citations, Citation{Type: "risk", ID: item.ID, Label: item.Title, Path: "/assetRiskCenter", Params: "id=" + strconv.Itoa(int(item.ID))})
		}
	case "asset.warranty.expiring":
		end := time.Now().AddDate(0, 0, 30)
		var list []assetModel.Asset
		err := global.GVA_DB.Preload("Category").Where("warranty_end_date IS NOT NULL AND warranty_end_date >= ? AND warranty_end_date <= ?", time.Now(), end).Order("warranty_end_date ASC").Limit(50).Find(&list).Error
		if err != nil {
			return result, err
		}
		result.Data = map[string]any{"list": list, "until": end.Format("2006-01-02")}
		result.Answer = fmt.Sprintf("未来 30 天内有 %d 项资产质保到期。", len(list))
		for _, item := range list {
			result.Citations = append(result.Citations, Citation{Type: "asset", ID: item.ID, Label: item.AssetCode + " " + item.Name, Path: "/assetInventory"})
		}
	case "asset.custodian.summary":
		var rows []struct {
			Custodian string  `json:"custodian"`
			Quantity  int64   `json:"quantity"`
			Value     float64 `json:"value"`
		}
		err := global.GVA_DB.Model(&assetModel.Asset{}).Select("custodian, COALESCE(SUM(quantity),0) AS quantity, COALESCE(SUM(current_value),0) AS value").Where("custodian <> ''").Group("custodian").Order("quantity DESC").Limit(30).Scan(&rows).Error
		if err != nil {
			return result, err
		}
		result.Data = rows
		result.Answer = fmt.Sprintf("已按保管人汇总 %d 组资产。", len(rows))
		result.Citations = append(result.Citations, Citation{Type: "asset", Label: "资产保管人汇总", Path: "/assetInventory", Params: "view=summary"})
	case "asset.operation.summary":
		var rows []struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			Count  int64  `json:"count"`
		}
		err := global.GVA_DB.Table("asset_operation_orders").Select("type, status, COUNT(*) AS count").Group("type, status").Order("count DESC").Scan(&rows).Error
		if err != nil {
			return result, err
		}
		result.Data = rows
		result.Answer = fmt.Sprintf("已汇总 %d 组资产流转单据状态。", len(rows))
		result.Citations = append(result.Citations, Citation{Type: "asset_operation", Label: "资产流转单据汇总", Path: "/assetOperation", Params: "view=list"})
	case "invoice.summary", "invoice.pending_reviews", "invoice.failed_recognitions":
		scope := invoiceService.ResolveAccessScope(userID, authorityID)
		status := ""
		if tool == "invoice.pending_reviews" {
			status = invoiceModel.InvoiceStatusPendingReview
		}
		if tool == "invoice.failed_recognitions" {
			status = invoiceModel.InvoiceStatusRecognitionFailed
		}
		list, total, err := invoiceService.InvoiceService{}.List(invoiceRequest.InvoiceSearch{PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 20}, Status: status}, scope)
		if err != nil {
			return result, err
		}
		if tool == "invoice.summary" {
			dashboard, dashboardErr := invoiceService.InvoiceService{}.Dashboard(scope)
			if dashboardErr != nil {
				return result, dashboardErr
			}
			result.Data = dashboard
			result.Answer = fmt.Sprintf("当前权限范围内确认发票 %d 张，待复核 %d 张，识别失败 %d 张。", dashboard.ConfirmedCount, dashboard.PendingCount, dashboard.FailedCount)
		} else {
			result.Data = map[string]any{"list": list, "total": total}
			result.Answer = fmt.Sprintf("当前权限范围内命中 %d 张发票。", total)
		}
		for _, item := range list {
			result.Citations = append(result.Citations, Citation{Type: "invoice", ID: item.ID, Label: item.InvoiceNumber + " " + item.SellerName, Path: "/invoiceLedger", Params: "id=" + strconv.Itoa(int(item.ID))})
		}
	case "invoice.provider_quality":
		rows, err := loadInvoiceProviderQuality(userID, authorityID)
		if err != nil {
			return result, err
		}
		result.Data = rows
		result.Answer = fmt.Sprintf("已汇总 %d 个发票识别 Provider。", len(rows))
		result.Citations = append(result.Citations, Citation{Type: "invoice_quality", Label: "发票识别质量", Path: "/invoiceQuality", Params: "view=provider"})
	case "schedule.today":
		items, err := scheduleService.WorkSchedule.List(ctx, userID)
		if err != nil {
			return result, err
		}
		today := time.Now().Format("2006-01-02")
		filtered := make([]any, 0)
		for _, item := range items {
			if item.Date >= today {
				filtered = append(filtered, item)
				result.Citations = append(result.Citations, Citation{Type: "schedule", ID: item.ID, Label: item.Date + " " + item.Time + " " + item.Title, Path: "/workCalendar/schedule", Params: "date=" + item.Date})
			}
		}
		result.Data = map[string]any{"list": filtered, "from": today}
		result.Answer = fmt.Sprintf("从今天开始共有 %d 项个人日程。", len(filtered))
	case "announcement.unread":
		items, err := announcementService.Info.Notifications(userID, 20)
		if err != nil {
			return result, err
		}
		result.Data = items
		result.Answer = fmt.Sprintf("你有 %d 条未读公告。", items.UnreadCount)
		for _, item := range items.List {
			if !item.IsRead {
				result.Citations = append(result.Citations, Citation{Type: "announcement", ID: item.ID, Label: item.Title, Path: "/info", Params: "id=" + strconv.Itoa(int(item.ID))})
			}
		}
	default:
		return result, errors.New("未注册的业务助手 Tool")
	}
	return result, nil
}

func loadInvoiceProviderQuality(userID, authorityID uint) ([]invoiceProviderQualityRow, error) {
	db := global.GVA_DB.Table("invoices").Where("deleted_at IS NULL")
	scope := invoiceService.ResolveAccessScope(userID, authorityID)
	if !scope.All {
		if scope.RoleWide && len(scope.AuthorityIDs) > 0 {
			db = db.Where("authority_id IN ?", scope.AuthorityIDs)
		} else {
			db = db.Where("created_by = ?", userID)
		}
	}
	var rows []invoiceProviderQualityRow
	err := db.Select(
		"recognition_provider AS provider, COUNT(*) AS total, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS failed, COALESCE(AVG(recognition_duration_ms), 0) AS avg_duration",
		invoiceModel.InvoiceStatusRecognitionFailed,
	).Group("recognition_provider").Order("total DESC, provider ASC").Scan(&rows).Error
	return rows, err
}

func toolPermissionAllowed(authorityID uint, tool string) bool {
	path, registered := toolPermissionPath(tool)
	if !registered {
		return false
	}
	return permissionAllowed(authorityID, path, "GET")
}

func permissionAllowed(authorityID uint, path, method string) bool {
	if authorityID == 0 || path == "" || method == "" {
		return false
	}
	enforcer := utils.GetCasbin()
	if enforcer == nil {
		return false
	}
	allowed, _ := enforcer.Enforce(strconv.Itoa(int(authorityID)), path, method)
	return allowed
}

func toolPermissionPath(tool string) (string, bool) {
	switch tool {
	case "asset.search", "asset.warranty.expiring", "asset.custodian.summary":
		return "/asset/list", true
	case "asset.detail":
		return "/asset/detail", true
	case "asset.risk.list":
		return "/assetRisk/list", true
	case "asset.operation.summary":
		return "/assetOperation/list", true
	case "invoice.summary", "invoice.pending_reviews", "invoice.failed_recognitions":
		return "/invoice/list", true
	case "invoice.provider_quality":
		return "/invoiceQuality/providerMetrics", true
	case "schedule.today":
		return "/workSchedule/list", true
	case "announcement.unread":
		return "/info/notifications", true
	default:
		return "", false
	}
}

func (s *smartService) tryModelSummary(ctx context.Context, userID, authorityID uint, question, tool, deterministic string, data any) (string, bool) {
	if !global.GVA_CONFIG.AI.Enabled {
		return "", false
	}
	permissionPath := "/smart/copilot/query"
	operation := "copilot-query"
	if tool == "smart.daily-report" {
		permissionPath = "/smartReport/generate"
		operation = "daily-report-summary"
	}
	payload := common.JSONMap{"question": question, "tool": tool, "deterministicAnswer": deterministic, "data": data}
	response, err := ai.Default.Complete(ctx, ai.CompletionRequest{UserID: userID, AuthorityID: authorityID, Module: "smart", Operation: operation, Prompt: "请基于给定的只读业务查询结果，用简洁中文回答用户问题。不得编造数据，不得建议执行写操作。", Payload: payload, MaxOutputTokens: 500, PermissionPath: permissionPath, PermissionMethod: "POST"})
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(response.Content), strings.TrimSpace(response.Content) != ""
}

func (s *smartService) Sessions(userID, authorityID uint) ([]model.CopilotSession, error) {
	var list []model.CopilotSession
	err := global.GVA_DB.Where("user_id = ? AND authority_id = ?", userID, authorityID).Order("last_message_at DESC, id DESC").Limit(50).Find(&list).Error
	return list, err
}

func (s *smartService) Session(userID, authorityID, id uint) (model.CopilotSession, []model.CopilotMessage, error) {
	var session model.CopilotSession
	if err := global.GVA_DB.Where("id = ? AND user_id = ? AND authority_id = ?", id, userID, authorityID).First(&session).Error; err != nil {
		return session, nil, err
	}
	var messages []model.CopilotMessage
	err := global.GVA_DB.Where("session_id = ? AND user_id = ? AND authority_id = ?", id, userID, authorityID).Order("created_at ASC, id ASC").Find(&messages).Error
	return session, messages, err
}

func (s *smartService) DeleteSession(userID, authorityID, id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if result := tx.Where("id = ? AND user_id = ? AND authority_id = ?", id, userID, authorityID).Delete(&model.CopilotSession{}); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("session_id = ? AND user_id = ? AND authority_id = ?", id, userID, authorityID).Delete(&model.CopilotMessage{}).Error
	})
}

func todayRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 0, 1)
}

func (s *smartService) buildReport(ctx context.Context, userID, authorityID uint) (common.JSONMap, string, error) {
	start, end := todayRange()
	metrics := common.JSONMap{}
	var count int64
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error; err != nil {
		return nil, "", err
	}
	assets := common.JSONMap{"created": count}
	for _, window := range []struct {
		days int
		key  string
	}{{30, "warrantyExpiring30d"}, {60, "warrantyExpiring60d"}, {90, "warrantyExpiring90d"}} {
		if err := global.GVA_DB.Model(&assetModel.Asset{}).
			Where("warranty_end_date >= ? AND warranty_end_date < ?", start, start.AddDate(0, 0, window.days+1)).
			Count(&count).Error; err != nil {
			return nil, "", err
		}
		assets[window.key] = count
	}
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("status = ?", assetModel.AssetStatusPendingInbound).Count(&count).Error; err != nil {
		return nil, "", err
	}
	assets["pendingInbound"] = count
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("status = ? AND updated_at < ?", assetModel.AssetStatusInUse, start.AddDate(0, 0, -30)).Count(&count).Error; err != nil {
		return nil, "", err
	}
	assets["longTermInUse"] = count
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("status = ? AND updated_at < ?", assetModel.AssetStatusMaintenance, start.AddDate(0, 0, -7)).Count(&count).Error; err != nil {
		return nil, "", err
	}
	assets["maintenanceOverdue"] = count
	operationCounts := common.JSONMap{"inbound": int64(0), "issue": int64(0), "transfer": int64(0), "return": int64(0), "maintenance": int64(0), "scrap": int64(0)}
	var operationRows []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := global.GVA_DB.Model(&assetModel.AssetOperationOrder{}).
		Select("type, COUNT(*) AS count").
		Where("status = ? AND business_date >= ? AND business_date < ?", assetModel.OperationStatusCompleted, start, end).
		Group("type").Scan(&operationRows).Error; err != nil {
		return nil, "", err
	}
	var operationTotal int64
	for _, row := range operationRows {
		operationCounts[row.Type] = row.Count
		operationTotal += row.Count
	}
	assets["todayOperations"] = operationCounts
	assets["todayOperationTotal"] = operationTotal
	metrics["assets"] = assets
	var open, newRisk, resolved int64
	if err := global.GVA_DB.Model(&assetModel.AssetRiskEvent{}).Where("status IN ?", []string{assetModel.RiskStatusOpen, assetModel.RiskStatusAcknowledged}).Count(&open).Error; err != nil {
		return nil, "", err
	}
	if err := global.GVA_DB.Model(&assetModel.AssetRiskEvent{}).Where("first_detected_at >= ? AND first_detected_at < ?", start, end).Count(&newRisk).Error; err != nil {
		return nil, "", err
	}
	if err := global.GVA_DB.Model(&assetModel.AssetRiskEvent{}).Where("handled_at >= ? AND handled_at < ? AND status IN ?", start, end, []string{assetModel.RiskStatusResolved, assetModel.RiskStatusIgnored}).Count(&resolved).Error; err != nil {
		return nil, "", err
	}
	metrics["risks"] = common.JSONMap{"open": open, "new": newRisk, "resolved": resolved}
	scope := invoiceService.ResolveAccessScope(userID, authorityID)
	invoiceDB := func() *gorm.DB {
		db := global.GVA_DB.Model(&invoiceModel.Invoice{})
		if !scope.All {
			if scope.RoleWide && len(scope.AuthorityIDs) > 0 {
				return db.Where("authority_id IN ?", scope.AuthorityIDs)
			}
			return db.Where("created_by = ?", userID)
		}
		return db
	}
	var invoiceToday, recognizedToday, reviewedToday, confirmedToday, pending, lowConfidence, failed, backlog int64
	if err := invoiceDB().Where("created_at >= ? AND created_at < ?", start, end).Count(&invoiceToday).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("recognition_provider <> '' AND updated_at >= ? AND updated_at < ?", start, end).Count(&recognizedToday).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("review_captured_at >= ? AND review_captured_at < ?", start, end).Count(&reviewedToday).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("confirmed_at >= ? AND confirmed_at < ?", start, end).Count(&confirmedToday).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("status = ?", invoiceModel.InvoiceStatusPendingReview).Count(&pending).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("status = ? AND recognition_confidence > 0 AND recognition_confidence < ?", invoiceModel.InvoiceStatusPendingReview, 0.7).Count(&lowConfidence).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("status = ?", invoiceModel.InvoiceStatusRecognitionFailed).Count(&failed).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("status IN ?", []string{invoiceModel.InvoiceStatusUploaded, invoiceModel.InvoiceStatusRecognizing}).Count(&backlog).Error; err != nil {
		return nil, "", err
	}
	weekday := int(start.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := start.AddDate(0, 0, -(weekday - 1))
	monthStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	confirmedAmount := func(from time.Time) (int64, error) {
		var amount int64
		err := invoiceDB().Where("status = ? AND confirmed_at >= ? AND confirmed_at < ?", invoiceModel.InvoiceStatusConfirmed, from, end).
			Select("COALESCE(SUM(total_cents), 0)").Scan(&amount).Error
		return amount, err
	}
	confirmedTodayCents, err := confirmedAmount(start)
	if err != nil {
		return nil, "", err
	}
	confirmedWeekCents, err := confirmedAmount(weekStart)
	if err != nil {
		return nil, "", err
	}
	confirmedMonthCents, err := confirmedAmount(monthStart)
	if err != nil {
		return nil, "", err
	}
	var providerTotal, providerFailed int64
	if err := invoiceDB().Where("recognition_provider <> ''").Count(&providerTotal).Error; err != nil {
		return nil, "", err
	}
	if err := invoiceDB().Where("recognition_provider <> '' AND status = ?", invoiceModel.InvoiceStatusRecognitionFailed).Count(&providerFailed).Error; err != nil {
		return nil, "", err
	}
	providerFailureRate := float64(0)
	if providerTotal > 0 {
		providerFailureRate = float64(providerFailed) * 100 / float64(providerTotal)
	}
	metrics["invoices"] = common.JSONMap{
		"todayUploaded": invoiceToday, "todayRecognized": recognizedToday, "todayReviewed": reviewedToday,
		"todayConfirmed": confirmedToday, "pendingReview": pending, "lowConfidence": lowConfidence,
		"recognitionFailed": failed, "recognitionBacklog": backlog, "providerFailureRate": providerFailureRate,
		"confirmedTodayCents": confirmedTodayCents, "confirmedWeekCents": confirmedWeekCents, "confirmedMonthCents": confirmedMonthCents,
	}
	schedules, err := scheduleService.WorkSchedule.List(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	today := time.Now().Format("2006-01-02")
	scheduleCount := 0
	for _, item := range schedules {
		if item.Date == today {
			scheduleCount++
		}
	}
	announcements, err := announcementService.Info.Notifications(userID, 20)
	if err != nil {
		return nil, "", err
	}
	metrics["collaboration"] = common.JSONMap{"todaySchedules": scheduleCount, "unreadAnnouncements": announcements.UnreadCount}
	var aiCalls, aiFailures, aiDurationMS, aiCostMicros int64
	aiDB := func() *gorm.DB {
		db := global.GVA_DB.Table("ai_model_invocations").Where("created_at >= ? AND created_at < ?", start, end)
		if !scope.All {
			if scope.RoleWide && len(scope.AuthorityIDs) > 0 {
				return db.Where("authority_id IN ?", scope.AuthorityIDs)
			}
			return db.Where("user_id = ?", userID)
		}
		return db
	}
	if err := aiDB().Count(&aiCalls).Error; err != nil {
		return nil, "", err
	}
	if err := aiDB().Where("status IN ?", []string{"failed", "blocked"}).Count(&aiFailures).Error; err != nil {
		return nil, "", err
	}
	if err := aiDB().Select("COALESCE(SUM(duration_ms), 0)").Scan(&aiDurationMS).Error; err != nil {
		return nil, "", err
	}
	if err := aiDB().Select("COALESCE(SUM(estimated_cost_micros), 0)").Scan(&aiCostMicros).Error; err != nil {
		return nil, "", err
	}
	aiAverageDurationMS := int64(0)
	aiFailureRate := float64(0)
	if aiCalls > 0 {
		aiAverageDurationMS = aiDurationMS / aiCalls
		aiFailureRate = float64(aiFailures) * 100 / float64(aiCalls)
	}
	metrics["system"] = common.JSONMap{"aiCalls": aiCalls, "aiFailures": aiFailures, "aiFailureRate": aiFailureRate, "aiAverageDurationMs": aiAverageDurationMS, "aiEstimatedCostMicros": aiCostMicros}
	summary := fmt.Sprintf("今日资产新增 %d 项、完成流转 %d 单，当前开放风险 %d 条（新增 %d、处理 %d），发票上传 %d 张、确认 %d 张、待复核 %d 张，今日有 %d 项日程，未读公告 %d 条。", assets["created"], operationTotal, open, newRisk, resolved, invoiceToday, confirmedToday, pending, scheduleCount, announcements.UnreadCount)
	return metrics, summary, nil
}

func (s *smartService) GenerateReport(ctx context.Context, userID, authorityID uint) (model.SmartDailyReport, error) {
	if userID == 0 {
		return model.SmartDailyReport{}, errors.New("用户未登录")
	}
	metrics, summary, err := s.buildReport(ctx, userID, authorityID)
	if err != nil {
		return model.SmartDailyReport{}, err
	}
	generatedBy := "deterministic"
	if global.GVA_CONFIG.AI.Enabled {
		if text, ok := s.tryModelSummary(ctx, userID, authorityID, "请生成今日智能日报", "smart.daily-report", summary, metrics); ok {
			summary = text
			generatedBy = "deterministic+model"
		}
	}
	start, _ := todayRange()
	now := time.Now()
	report := model.SmartDailyReport{UserID: userID, AuthorityID: authorityID, ReportDate: start, Metrics: metrics, Summary: summary, GeneratedBy: generatedBy, GeneratedAt: now}
	err = global.GVA_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "authority_id"}, {Name: "report_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"metrics", "summary", "generated_by", "generated_at", "updated_at"}),
	}).Create(&report).Error
	if err != nil {
		return model.SmartDailyReport{}, err
	}
	if err = global.GVA_DB.Where("user_id = ? AND authority_id = ? AND report_date = ?", userID, authorityID, start).First(&report).Error; err != nil {
		return model.SmartDailyReport{}, err
	}
	return report, nil
}

func (s *smartService) TodayReport(ctx context.Context, userID, authorityID uint) (model.SmartDailyReport, error) {
	return s.GenerateReport(ctx, userID, authorityID)
}

func (s *smartService) Reports(userID, authorityID uint, input ReportListInput) ([]model.SmartDailyReport, int64, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 || input.PageSize > 100 {
		input.PageSize = 20
	}
	var total int64
	db := global.GVA_DB.Model(&model.SmartDailyReport{}).Where("user_id = ? AND authority_id = ?", userID, authorityID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SmartDailyReport
	err := db.Order("report_date DESC").Offset((input.Page - 1) * input.PageSize).Limit(input.PageSize).Find(&list).Error
	return list, total, err
}

func (s *smartService) Report(userID, authorityID, id uint) (model.SmartDailyReport, error) {
	var report model.SmartDailyReport
	err := global.GVA_DB.Where("id = ? AND user_id = ? AND authority_id = ?", id, userID, authorityID).First(&report).Error
	return report, err
}

func (s *smartService) Subscription(userID uint) (model.SmartReportSubscription, error) {
	var item model.SmartReportSubscription
	err := global.GVA_DB.Where("user_id = ?", userID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = model.SmartReportSubscription{UserID: userID, Enabled: true, DeliveryTime: "09:00", Channels: "in_app"}
		err = global.GVA_DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoNothing: true}).Create(&item).Error
		if err == nil && item.ID == 0 {
			err = global.GVA_DB.Where("user_id = ?", userID).First(&item).Error
		}
	}
	return item, err
}

func (s *smartService) SaveSubscription(userID uint, input model.SmartReportSubscription) (model.SmartReportSubscription, error) {
	if input.DeliveryTime == "" {
		input.DeliveryTime = "09:00"
	}
	if !regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`).MatchString(input.DeliveryTime) {
		return input, errors.New("日报发送时间必须是 HH:MM")
	}
	channels, err := normalizeReportChannels(input.Channels)
	if err != nil {
		return input, err
	}
	input.Channels = channels
	input.ID = 0
	input.UserID = userID
	err = global.GVA_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "delivery_time", "channels", "updated_at"}),
	}).Create(&input).Error
	if err != nil {
		return input, err
	}
	var current model.SmartReportSubscription
	err = global.GVA_DB.Where("user_id = ?", userID).First(&current).Error
	return current, err
}

func normalizeReportChannels(raw string) (string, error) {
	seen := map[string]bool{}
	channels := make([]string, 0, 2)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" || seen[item] {
			continue
		}
		if item != "in_app" && item != "email" {
			return "", errors.New("日报渠道仅支持 in_app 或 email")
		}
		seen[item] = true
		channels = append(channels, item)
	}
	if len(channels) == 0 {
		return "", errors.New("至少选择一个日报渠道")
	}
	return strings.Join(channels, ","), nil
}

func (s *smartService) Deliveries(userID uint, limit int) ([]model.SmartReportDelivery, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var list []model.SmartReportDelivery
	err := global.GVA_DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

var datePattern = regexp.MustCompile(`(?m)(20\d{2})[-/年](\d{1,2})[-/月](\d{1,2})日?`)
var shortDatePattern = regexp.MustCompile(`(?m)(\d{1,2})月(\d{1,2})日?`)
var timePattern = regexp.MustCompile(`(?m)(上午|下午|晚上|早上)?\s*(\d{1,2})[:：](\d{2})`)
var locationPattern = regexp.MustCompile(`(?m)(地点|地址|会议地点)\s*[:：]\s*([^\n。；;，,]+)`)

func extractSchedule(title, content string) common.JSONMap {
	text := title + "\n" + content
	date := ""
	confidence := 0.35
	if match := datePattern.FindStringSubmatch(text); len(match) == 4 {
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		date = fmt.Sprintf("%s-%02d-%02d", match[1], month, day)
		confidence += 0.3
	} else if match := shortDatePattern.FindStringSubmatch(text); len(match) == 3 {
		year := time.Now().Year()
		month, _ := strconv.Atoi(match[1])
		day, _ := strconv.Atoi(match[2])
		date = fmt.Sprintf("%d-%02d-%02d", year, month, day)
		confidence += 0.2
	}
	clock := "09:00"
	if match := timePattern.FindStringSubmatch(text); len(match) == 4 {
		hour, _ := strconv.Atoi(match[2])
		minute := match[3]
		if (match[1] == "下午" || match[1] == "晚上") && hour < 12 {
			hour += 12
		}
		clock = fmt.Sprintf("%02d:%s", hour, minute)
		confidence += 0.2
	}
	location := ""
	if match := locationPattern.FindStringSubmatch(text); len(match) == 3 {
		location = strings.TrimSpace(match[2])
		for _, marker := range []string{"参加", "召开", "举行", "进行"} {
			if index := strings.Index(location, marker); index > 0 {
				location = strings.TrimSpace(location[:index])
				break
			}
		}
		confidence += 0.1
	}
	todos := make([]string, 0)
	for _, line := range strings.FieldsFunc(content, func(r rune) bool { return r == '\n' || r == '。' || r == '；' || r == ';' }) {
		line = strings.TrimSpace(line)
		if line != "" && strings.ContainsAny(line, "请需须截止完成提交报名参加") {
			todos = append(todos, line)
		}
	}
	if confidence > 1 {
		confidence = 1
	}
	note := strings.TrimSpace(content)
	if len([]rune(note)) > 480 {
		note = string([]rune(note)[:480])
	}
	return common.JSONMap{"title": strings.TrimSpace(title), "date": date, "time": clock, "location": location, "todos": todos, "note": note, "confidence": confidence}
}

func (s *smartService) AnnouncementDraft(userID, announcementID uint) (model.SmartDraft, error) {
	if userID == 0 || announcementID == 0 {
		return model.SmartDraft{}, errors.New("公告或用户参数不正确")
	}
	now := time.Now()
	dedupKey := fmt.Sprintf("schedule:%d:announcement:%d", userID, announcementID)
	var existing model.SmartDraft
	lookupErr := global.GVA_DB.Where("user_id = ? AND draft_type = ? AND source_id = ?", userID, model.DraftTypeSchedule, announcementID).Order("id DESC").First(&existing).Error
	if lookupErr == nil {
		if existing.Status == model.DraftStatusAccepted || existing.Status == model.DraftStatusProcessing || (existing.Status == model.DraftStatusDraft && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now))) {
			return existing, nil
		}
	}
	if lookupErr != nil && !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
		return model.SmartDraft{}, lookupErr
	}
	item, err := announcementService.Info.GetInfo(strconv.Itoa(int(announcementID)))
	if err != nil {
		return model.SmartDraft{}, err
	}
	if item.Status != "published" {
		return model.SmartDraft{}, errors.New("只有已发布公告可以提取日程")
	}
	payload := extractSchedule(item.Title, item.Content)
	expires := now.Add(24 * time.Hour)
	payload["announcementId"] = announcementID
	payload["clientKey"] = fmt.Sprintf("smart-announcement-%d", announcementID)
	confidence := payload["confidence"].(float64)
	if lookupErr == nil {
		if err := global.GVA_DB.Model(&existing).Updates(map[string]any{"dedup_key": dedupKey, "status": model.DraftStatusDraft, "payload": payload, "confidence": confidence, "expires_at": expires, "last_error": ""}).Error; err != nil {
			return model.SmartDraft{}, err
		}
		existing.DedupKey, existing.Status, existing.Payload, existing.Confidence, existing.ExpiresAt, existing.LastError = &dedupKey, model.DraftStatusDraft, payload, confidence, &expires, ""
		return existing, nil
	}
	draft := model.SmartDraft{UserID: userID, DedupKey: &dedupKey, DraftType: model.DraftTypeSchedule, SourceID: announcementID, Status: model.DraftStatusDraft, Payload: payload, Confidence: confidence, ExpiresAt: &expires}
	if err := global.GVA_DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&draft).Error; err != nil {
		return model.SmartDraft{}, err
	}
	if draft.ID == 0 {
		if err := global.GVA_DB.Where("dedup_key = ?", dedupKey).First(&draft).Error; err != nil {
			return model.SmartDraft{}, err
		}
	}
	return draft, nil
}

type OperationDraftInput struct {
	OperationType   string `json:"operationType"`
	Instruction     string `json:"instruction"`
	AssetIDs        []uint `json:"assetIds"`
	TargetLocation  string `json:"targetLocation"`
	TargetCustodian string `json:"targetCustodian"`
	Reason          string `json:"reason"`
	Remarks         string `json:"remarks"`
	BusinessDate    string `json:"businessDate"`
}

type OperationAssetCandidateInput struct {
	OperationType string
	Keyword       string
}

func (s *smartService) OperationAssetCandidates(authorityID uint, input OperationAssetCandidateInput) ([]assetModel.Asset, error) {
	if !permissionAllowed(authorityID, "/assetOperation/assetOptions", "GET") {
		return nil, errors.New("当前角色没有查询可流转资产的权限")
	}
	if _, err := assetOperationType(input.OperationType); err != nil {
		return nil, err
	}
	return assetService.Operation.AssetOptions(assetRequest.OperationAssetSearch{Type: input.OperationType, Keyword: strings.TrimSpace(input.Keyword)})
}

func inferOperationType(text string) string {
	for _, item := range []struct{ key, value string }{{"入库", "inbound"}, {"领用", "issue"}, {"调拨", "transfer"}, {"归还", "return"}, {"维修", "maintenance"}, {"报废", "scrap"}} {
		if strings.Contains(text, item.key) {
			return item.value
		}
	}
	return ""
}

func (s *smartService) OperationDraft(userID, authorityID uint, input OperationDraftInput) (model.SmartDraft, error) {
	if userID == 0 {
		return model.SmartDraft{}, errors.New("用户未登录")
	}
	if !permissionAllowed(authorityID, "/assetOperation/create", "POST") {
		return model.SmartDraft{}, errors.New("当前角色没有创建资产业务草稿的权限")
	}
	input.Instruction = strings.TrimSpace(input.Instruction)
	if input.OperationType == "" {
		input.OperationType = inferOperationType(input.Instruction)
	}
	if input.OperationType == "" {
		return model.SmartDraft{}, errors.New("无法识别业务单类型，请指定 operationType")
	}
	if _, err := assetOperationType(input.OperationType); err != nil {
		return model.SmartDraft{}, err
	}
	if len(input.AssetIDs) == 0 {
		return model.SmartDraft{}, errors.New("业务草稿至少需要一项资产")
	}
	if len(input.AssetIDs) > 100 {
		return model.SmartDraft{}, errors.New("单张业务草稿最多选择 100 项资产")
	}
	var assetCount int64
	if err := global.GVA_DB.Model(&assetModel.Asset{}).Where("id IN ?", input.AssetIDs).Count(&assetCount).Error; err != nil {
		return model.SmartDraft{}, err
	}
	if assetCount != int64(len(input.AssetIDs)) {
		return model.SmartDraft{}, errors.New("部分资产不存在或已删除")
	}
	date := input.BusinessDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if parsed, err := time.ParseInLocation("2006-01-02", date, time.Local); err != nil || parsed.Format("2006-01-02") != date {
		return model.SmartDraft{}, errors.New("业务日期格式不正确")
	}
	payload := common.JSONMap{"operationType": input.OperationType, "assetIds": input.AssetIDs, "targetLocation": strings.TrimSpace(input.TargetLocation), "targetCustodian": strings.TrimSpace(input.TargetCustodian), "reason": strings.TrimSpace(input.Reason), "remarks": strings.TrimSpace(input.Remarks), "businessDate": date, "instruction": input.Instruction}
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	draft := model.SmartDraft{UserID: userID, DraftType: model.DraftTypeOperation, Status: model.DraftStatusDraft, Payload: payload, Confidence: 0.8, ExpiresAt: &expires}
	return draft, global.GVA_DB.Create(&draft).Error
}

func assetOperationType(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "inbound", "issue", "transfer", "return", "maintenance", "scrap":
		return value, nil
	default:
		return "", errors.New("业务类型不正确")
	}
}

func asUintSlice(value any) []uint {
	result := make([]uint, 0)
	switch items := value.(type) {
	case []uint:
		return items
	case []any:
		for _, item := range items {
			switch number := item.(type) {
			case float64:
				result = append(result, uint(number))
			case int:
				result = append(result, uint(number))
			}
		}
	}
	return result
}

func scheduleDraftNote(payload common.JSONMap) string {
	parts := make([]string, 0, 3)
	if location := strings.TrimSpace(fmt.Sprint(payload["location"])); location != "" {
		parts = append(parts, "地点："+location)
	}
	todos := make([]string, 0)
	switch values := payload["todos"].(type) {
	case []string:
		todos = append(todos, values...)
	case []any:
		for _, value := range values {
			if text := strings.TrimSpace(fmt.Sprint(value)); text != "" {
				todos = append(todos, text)
			}
		}
	}
	if len(todos) > 0 {
		parts = append(parts, "待办："+strings.Join(todos, "；"))
	}
	if note := strings.TrimSpace(fmt.Sprint(payload["note"])); note != "" {
		parts = append(parts, "公告原文："+note)
	}
	result := strings.Join(parts, "\n")
	if len([]rune(result)) > 500 {
		result = string([]rune(result)[:500])
	}
	return result
}

func (s *smartService) AcceptDraft(ctx context.Context, userID, authorityID uint, userName string, draftID uint) (any, error) {
	var draft model.SmartDraft
	if err := global.GVA_DB.Where("id = ? AND user_id = ? AND status = ?", draftID, userID, model.DraftStatusDraft).First(&draft).Error; err != nil {
		return nil, err
	}
	if draft.ExpiresAt != nil && draft.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("草稿已过期，请重新生成")
	}
	switch draft.DraftType {
	case model.DraftTypeSchedule:
		if !permissionAllowed(authorityID, "/workSchedule/create", "POST") {
			return nil, errors.New("当前角色没有创建日程的权限")
		}
	case model.DraftTypeOperation:
		if !permissionAllowed(authorityID, "/assetOperation/create", "POST") {
			return nil, errors.New("当前角色没有创建资产业务草稿的权限")
		}
	default:
		return nil, errors.New("不支持的草稿类型")
	}
	claim := global.GVA_DB.Model(&model.SmartDraft{}).Where("id = ? AND user_id = ? AND status = ?", draftID, userID, model.DraftStatusDraft).Updates(map[string]any{"status": model.DraftStatusProcessing, "last_error": ""})
	if claim.Error != nil {
		return nil, claim.Error
	}
	if claim.RowsAffected == 0 {
		return nil, errors.New("草稿正在处理或已经确认")
	}
	revert := func(err error) (any, error) {
		_ = global.GVA_DB.Model(&model.SmartDraft{}).Where("id = ? AND user_id = ? AND status = ?", draftID, userID, model.DraftStatusProcessing).Updates(map[string]any{"status": model.DraftStatusDraft, "last_error": err.Error()})
		return nil, err
	}
	var result any
	switch draft.DraftType {
	case model.DraftTypeSchedule:
		payload := draft.Payload
		clientKey := fmt.Sprint(payload["clientKey"])
		created, err := scheduleService.WorkSchedule.Create(ctx, userID, scheduleRequest.WorkScheduleUpsert{ClientKey: clientKey, Title: fmt.Sprint(payload["title"]), Date: fmt.Sprint(payload["date"]), Time: fmt.Sprint(payload["time"]), Type: "announcement", Note: scheduleDraftNote(payload)})
		if err != nil {
			var existing scheduleModel.WorkSchedule
			if lookupErr := global.GVA_DB.WithContext(ctx).Where("user_id = ? AND client_key = ?", userID, clientKey).First(&existing).Error; lookupErr == nil {
				result = existing
				break
			}
			return revert(err)
		}
		result = created
	case model.DraftTypeOperation:
		payload := draft.Payload
		date, err := time.ParseInLocation("2006-01-02", fmt.Sprint(payload["businessDate"]), time.Local)
		if err != nil {
			return revert(errors.New("业务日期格式不正确"))
		}
		order, err := assetService.Operation.Create(assetRequest.SaveOperation{Type: fmt.Sprint(payload["operationType"]), BusinessDate: date, TargetLocation: fmt.Sprint(payload["targetLocation"]), TargetCustodian: fmt.Sprint(payload["targetCustodian"]), Reason: fmt.Sprint(payload["reason"]), Remarks: fmt.Sprint(payload["remarks"]), AssetIDs: asUintSlice(payload["assetIds"]), Submit: false}, userID, userName)
		if err != nil {
			return revert(err)
		}
		result = order
	default:
		return revert(errors.New("不支持的草稿类型"))
	}
	if err := global.GVA_DB.Model(&draft).Where("status = ?", model.DraftStatusProcessing).Updates(map[string]any{"status": model.DraftStatusAccepted, "last_error": ""}).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (s *smartService) Drafts(userID uint, draftType string) ([]model.SmartDraft, error) {
	var list []model.SmartDraft
	db := global.GVA_DB.Where("user_id = ?", userID)
	if draftType != "" {
		db = db.Where("draft_type = ?", draftType)
	}
	err := db.Order("created_at DESC").Limit(100).Find(&list).Error
	return list, err
}
