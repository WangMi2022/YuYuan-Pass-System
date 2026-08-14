package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
)

const (
	riskRulePendingTooLong        = "ASSET_PENDING_TOO_LONG"
	riskRuleInUseWithoutCustodian = "ASSET_IN_USE_WITHOUT_CUSTODIAN"
	riskRuleIdleWithCustodian     = "ASSET_IDLE_WITH_CUSTODIAN"
	riskRuleRetiredWithCustodian  = "ASSET_RETIRED_WITH_CUSTODIAN"
	riskRuleRetiredValueNonzero   = "ASSET_RETIRED_VALUE_NONZERO"
	riskRuleValueOverOriginal     = "ASSET_VALUE_OVER_ORIGINAL"
	riskRuleHighValueIdle         = "ASSET_HIGH_VALUE_IDLE"
	riskRuleWarrantyMissing       = "ASSET_WARRANTY_MISSING"
	riskRuleWarrantyExpiring      = "ASSET_WARRANTY_EXPIRING"
	riskRuleWarrantyExpired       = "ASSET_WARRANTY_EXPIRED"
	riskRuleMaintenanceOverdue    = "ASSET_MAINTENANCE_OVERDUE"
	riskRuleFrequentMaintenance   = "ASSET_FREQUENT_MAINTENANCE"
	riskRuleFrequentTransfer      = "ASSET_FREQUENT_TRANSFER"
	riskRuleLongTermInUse         = "ASSET_LONG_TERM_IN_USE"
	riskRuleDuplicateSerial       = "ASSET_DUPLICATE_SERIAL"
	riskRuleSimilarCode           = "ASSET_SIMILAR_CODE"
	riskRuleStateRecordMismatch   = "ASSET_STATE_RECORD_MISMATCH"
)

type riskMatch struct {
	asset          model.Asset
	rule           model.AssetRiskRule
	evidenceKey    string
	title          string
	description    string
	recommendation string
	evidence       common.JSONMap
}

type similarAsset struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
}

type riskEvaluationContext struct {
	now            time.Time
	recordsByAsset map[uint][]model.AssetOperationRecord
	serialGroups   map[string][]similarAsset
	similarByAsset map[uint][]similarAsset
}

func defaultRiskRules() []model.AssetRiskRule {
	return []model.AssetRiskRule{
		{Code: riskRulePendingTooLong, Name: "待入库超期", Category: "status", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{"days": 7}, Description: "资产创建后长时间停留在待入库状态。", Recommendation: "核对验收资料并尽快完成入库，或修正无效资产档案。", Enabled: true, Version: 1},
		{Code: riskRuleInUseWithoutCustodian, Name: "使用中无保管人", Category: "status", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{}, Description: "资产处于使用中但未记录保管人。", Recommendation: "补充实际保管人并通过资产流转单修正责任关系。", Enabled: true, Version: 1},
		{Code: riskRuleIdleWithCustodian, Name: "闲置资产仍有保管人", Category: "status", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{}, Description: "资产已闲置但仍保留保管人。", Recommendation: "核对资产是否已归还，必要时补录归还单。", Enabled: true, Version: 1},
		{Code: riskRuleRetiredWithCustodian, Name: "已处置资产仍有保管人", Category: "status", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{}, Description: "资产已处置但保管责任尚未清除。", Recommendation: "核对处置记录并清理保管关系。", Enabled: true, Version: 1},
		{Code: riskRuleRetiredValueNonzero, Name: "已处置资产估值未归零", Category: "value", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{}, Description: "已处置资产仍保留非零当前估值。", Recommendation: "复核报废流程和财务处理，修正当前估值。", Enabled: true, Version: 1},
		{Code: riskRuleValueOverOriginal, Name: "当前估值高于原值", Category: "value", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{}, Description: "资产当前估值高于登记原值。", Recommendation: "复核估值依据和资产原值，修正异常金额。", Enabled: true, Version: 1},
		{Code: riskRuleHighValueIdle, Name: "高价值资产长期闲置", Category: "value", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{"minValue": 10000, "days": 90}, Description: "高价值资产处于长期闲置状态。", Recommendation: "评估调配、共享或处置方案，提高资产利用率。", Enabled: true, Version: 1},
		{Code: riskRuleWarrantyMissing, Name: "重要资产缺少质保信息", Category: "warranty", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{"minValue": 10000, "categoryCodes": []string{"IT-COMPUTER", "IT-NETWORK", "PROD-EQUIP"}}, Description: "达到重要资产条件但未登记质保到期日。", Recommendation: "补录质保信息、合同或供应商服务期限。", Enabled: true, Version: 1},
		{Code: riskRuleWarrantyExpiring, Name: "质保即将到期", Category: "warranty", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{"days": 30}, Description: "资产质保将在阈值天数内到期。", Recommendation: "安排质保检查并评估续保、延保或备件计划。", Enabled: true, Version: 1},
		{Code: riskRuleWarrantyExpired, Name: "质保已过期", Category: "warranty", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{}, Description: "资产质保已经过期。", Recommendation: "确认维保方案并更新续保或替代计划。", Enabled: true, Version: 1},
		{Code: riskRuleMaintenanceOverdue, Name: "维修时间过长", Category: "maintenance", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{"days": 14}, Description: "资产进入维修状态后超过处理时限。", Recommendation: "联系维修责任人，更新维修进度或制定替代方案。", Enabled: true, Version: 1},
		{Code: riskRuleFrequentMaintenance, Name: "短期频繁维修", Category: "maintenance", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{"windowDays": 180, "count": 3}, Description: "资产在统计窗口内多次进入维修。", Recommendation: "开展故障根因分析，评估更换或报废。", Enabled: true, Version: 1},
		{Code: riskRuleFrequentTransfer, Name: "短期频繁调拨", Category: "return", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{"windowDays": 90, "count": 3}, Description: "资产在统计窗口内发生多次调拨。", Recommendation: "核对资产配置合理性和责任归属。", Enabled: true, Version: 1},
		{Code: riskRuleLongTermInUse, Name: "长期未归还", Category: "return", Severity: model.RiskSeverityHigh, Parameters: common.JSONMap{"days": 180}, Description: "资产领用后长时间未归还。", Recommendation: "联系保管人确认在用状态并补充归还计划。", Enabled: true, Version: 1},
		{Code: riskRuleDuplicateSerial, Name: "重复序列号", Category: "duplicate", Severity: model.RiskSeverityCritical, Parameters: common.JSONMap{}, Description: "多个资产档案使用相同非空序列号。", Recommendation: "核验铭牌和采购资料，合并或修正重复档案。", Enabled: true, Version: 1},
		{Code: riskRuleSimilarCode, Name: "疑似重复资产编号", Category: "duplicate", Severity: model.RiskSeverityMedium, Parameters: common.JSONMap{"maxDistance": 1, "minLength": 5}, Description: "资产编号高度相似，可能是录入错误或重复建档。", Recommendation: "对照实物标签和资产档案，确认编号是否准确。", Enabled: true, Version: 1},
		{Code: riskRuleStateRecordMismatch, Name: "状态与流转记录不一致", Category: "status", Severity: model.RiskSeverityCritical, Parameters: common.JSONMap{}, Description: "资产当前状态、位置或保管人与最近流转记录不一致。", Recommendation: "检查绕过流程的数据修改并恢复正确的资产状态。", Enabled: true, Version: 1},
	}
}

func evaluateRiskRule(rule model.AssetRiskRule, asset model.Asset, evaluation riskEvaluationContext) []riskMatch {
	records := evaluation.recordsByAsset[asset.ID]
	base := func(key, title, description string, evidence common.JSONMap) []riskMatch {
		return []riskMatch{{asset: asset, rule: rule, evidenceKey: key, title: title, description: description, recommendation: rule.Recommendation, evidence: evidence}}
	}
	ageDays := func(since time.Time) int { return elapsedDays(evaluation.now, since) }

	switch rule.Code {
	case riskRulePendingTooLong:
		days := paramInt(rule.Parameters, "days", 7)
		age := ageDays(asset.CreatedAt)
		if asset.Status == model.AssetStatusPendingInbound && age >= days {
			return base("pending", "资产待入库超期", fmt.Sprintf("%s 已待入库 %d 天，超过 %d 天阈值。", asset.Name, age, days), common.JSONMap{"status": asset.Status, "pendingDays": age, "thresholdDays": days, "createdAt": asset.CreatedAt})
		}
	case riskRuleInUseWithoutCustodian:
		if asset.Status == model.AssetStatusInUse && strings.TrimSpace(asset.Custodian) == "" {
			return base("missing-custodian", "使用中资产缺少保管人", fmt.Sprintf("%s 当前处于使用中，但未登记保管人。", asset.Name), common.JSONMap{"status": asset.Status, "custodian": asset.Custodian})
		}
	case riskRuleIdleWithCustodian:
		if asset.Status == model.AssetStatusIdle && strings.TrimSpace(asset.Custodian) != "" {
			return base("idle-custodian", "闲置资产仍保留保管人", fmt.Sprintf("%s 已闲置，但保管人仍为 %s。", asset.Name, asset.Custodian), common.JSONMap{"status": asset.Status, "custodian": asset.Custodian})
		}
	case riskRuleRetiredWithCustodian:
		if asset.Status == model.AssetStatusRetired && strings.TrimSpace(asset.Custodian) != "" {
			return base("retired-custodian", "已处置资产仍保留保管人", fmt.Sprintf("%s 已处置，但保管人仍为 %s。", asset.Name, asset.Custodian), common.JSONMap{"status": asset.Status, "custodian": asset.Custodian})
		}
	case riskRuleRetiredValueNonzero:
		if asset.Status == model.AssetStatusRetired && math.Abs(asset.CurrentValue) >= 0.01 {
			return base("retired-value", "已处置资产估值未归零", fmt.Sprintf("%s 已处置，当前估值仍为 %.2f。", asset.Name, asset.CurrentValue), common.JSONMap{"status": asset.Status, "currentValue": asset.CurrentValue})
		}
	case riskRuleValueOverOriginal:
		if asset.CurrentValue-asset.OriginalValue >= 0.01 {
			return base("value-over-original", "当前估值高于原值", fmt.Sprintf("%s 当前估值 %.2f 高于原值 %.2f。", asset.Name, asset.CurrentValue, asset.OriginalValue), common.JSONMap{"currentValue": asset.CurrentValue, "originalValue": asset.OriginalValue, "difference": asset.CurrentValue - asset.OriginalValue})
		}
	case riskRuleHighValueIdle:
		minValue := paramFloat(rule.Parameters, "minValue", 10000)
		days := paramInt(rule.Parameters, "days", 90)
		if asset.Status == model.AssetStatusIdle && asset.CurrentValue >= minValue {
			since, source := stateEnteredAt(asset, records, model.AssetStatusIdle)
			age := ageDays(since)
			if age >= days {
				return base("high-value-idle", "高价值资产长期闲置", fmt.Sprintf("%s 当前估值 %.2f，已闲置 %d 天。", asset.Name, asset.CurrentValue, age), common.JSONMap{"currentValue": asset.CurrentValue, "minValue": minValue, "idleDays": age, "thresholdDays": days, "idleSince": since, "timeSource": source})
			}
		}
	case riskRuleWarrantyMissing:
		if asset.WarrantyEndDate == nil && isImportantAsset(asset, rule.Parameters) {
			return base("warranty-missing", "重要资产缺少质保信息", fmt.Sprintf("%s 达到重要资产条件，但未登记质保到期日。", asset.Name), common.JSONMap{"categoryCode": asset.Category.Code, "currentValue": asset.CurrentValue, "originalValue": asset.OriginalValue})
		}
	case riskRuleWarrantyExpiring:
		if asset.WarrantyEndDate != nil {
			today := startOfRiskDay(evaluation.now)
			end := startOfRiskDay(*asset.WarrantyEndDate)
			days := paramInt(rule.Parameters, "days", 30)
			remaining := elapsedDays(end, today)
			if !end.Before(today) && remaining <= days {
				return base("warranty-expiring", "资产质保即将到期", fmt.Sprintf("%s 的质保将在 %d 天后到期。", asset.Name, remaining), common.JSONMap{"warrantyEndDate": end.Format("2006-01-02"), "remainingDays": remaining, "thresholdDays": days})
			}
		}
	case riskRuleWarrantyExpired:
		if asset.WarrantyEndDate != nil {
			today := startOfRiskDay(evaluation.now)
			end := startOfRiskDay(*asset.WarrantyEndDate)
			if end.Before(today) {
				overdue := elapsedDays(today, end)
				return base("warranty-expired", "资产质保已过期", fmt.Sprintf("%s 的质保已过期 %d 天。", asset.Name, overdue), common.JSONMap{"warrantyEndDate": end.Format("2006-01-02"), "overdueDays": overdue})
			}
		}
	case riskRuleMaintenanceOverdue:
		days := paramInt(rule.Parameters, "days", 14)
		if asset.Status == model.AssetStatusMaintenance {
			since, source := stateEnteredAt(asset, records, model.AssetStatusMaintenance)
			age := ageDays(since)
			if age >= days {
				return base("maintenance-overdue", "资产维修时间过长", fmt.Sprintf("%s 已处于维修状态 %d 天。", asset.Name, age), common.JSONMap{"maintenanceDays": age, "thresholdDays": days, "maintenanceSince": since, "timeSource": source})
			}
		}
	case riskRuleFrequentMaintenance:
		windowDays := paramInt(rule.Parameters, "windowDays", 180)
		threshold := paramInt(rule.Parameters, "count", 3)
		count, timestamps := countOperations(records, "maintenance", evaluation.now.AddDate(0, 0, -windowDays))
		if count >= threshold {
			return base("frequent-maintenance", "资产短期频繁维修", fmt.Sprintf("%s 在最近 %d 天内维修 %d 次。", asset.Name, windowDays, count), common.JSONMap{"count": count, "thresholdCount": threshold, "windowDays": windowDays, "operatedAt": timestamps})
		}
	case riskRuleFrequentTransfer:
		windowDays := paramInt(rule.Parameters, "windowDays", 90)
		threshold := paramInt(rule.Parameters, "count", 3)
		count, timestamps := countOperations(records, "transfer", evaluation.now.AddDate(0, 0, -windowDays))
		if count >= threshold {
			return base("frequent-transfer", "资产短期频繁调拨", fmt.Sprintf("%s 在最近 %d 天内调拨 %d 次。", asset.Name, windowDays, count), common.JSONMap{"count": count, "thresholdCount": threshold, "windowDays": windowDays, "operatedAt": timestamps})
		}
	case riskRuleLongTermInUse:
		days := paramInt(rule.Parameters, "days", 180)
		if asset.Status == model.AssetStatusInUse {
			since, source := stateEnteredAt(asset, records, model.AssetStatusInUse)
			age := ageDays(since)
			if age >= days {
				return base("long-term-in-use", "资产长期未归还", fmt.Sprintf("%s 已连续使用 %d 天，当前保管人为 %s。", asset.Name, age, displayValue(asset.Custodian)), common.JSONMap{"inUseDays": age, "thresholdDays": days, "inUseSince": since, "timeSource": source, "custodian": asset.Custodian})
			}
		}
	case riskRuleDuplicateSerial:
		normalized := normalizeIdentity(asset.SerialNumber)
		duplicates := evaluation.serialGroups[normalized]
		if normalized != "" && len(duplicates) > 1 {
			return base("serial:"+normalized, "资产序列号重复", fmt.Sprintf("%s 的序列号 %s 与另外 %d 项资产重复。", asset.Name, asset.SerialNumber, len(duplicates)-1), common.JSONMap{"serialNumber": asset.SerialNumber, "normalizedSerial": normalized, "assets": duplicates})
		}
	case riskRuleSimilarCode:
		matches := evaluation.similarByAsset[asset.ID]
		result := make([]riskMatch, 0, len(matches))
		for _, other := range matches {
			pair := []string{normalizeIdentity(asset.AssetCode), normalizeIdentity(other.Code)}
			sort.Strings(pair)
			result = append(result, riskMatch{asset: asset, rule: rule, evidenceKey: "code:" + strings.Join(pair, ":"), title: "资产编号疑似重复", description: fmt.Sprintf("%s 的资产编号 %s 与 %s 高度相似。", asset.Name, asset.AssetCode, other.Code), recommendation: rule.Recommendation, evidence: common.JSONMap{"assetCode": asset.AssetCode, "similarAssetId": other.ID, "similarAssetCode": other.Code}})
		}
		return result
	case riskRuleStateRecordMismatch:
		if len(records) == 0 {
			if asset.Status != model.AssetStatusPendingInbound {
				return base("missing-record", "资产状态缺少流转依据", fmt.Sprintf("%s 当前状态为 %s，但没有任何已完成流转记录。", asset.Name, asset.Status), common.JSONMap{"currentStatus": asset.Status, "currentLocation": asset.Location, "currentCustodian": asset.Custodian})
			}
			break
		}
		latest := records[0]
		if strings.TrimSpace(asset.Status) != strings.TrimSpace(latest.ToStatus) || strings.TrimSpace(asset.Location) != strings.TrimSpace(latest.ToLocation) || strings.TrimSpace(asset.Custodian) != strings.TrimSpace(latest.ToCustodian) {
			return base("record:"+strconv.FormatUint(uint64(latest.ID), 10), "资产状态与最近流转不一致", fmt.Sprintf("%s 的当前状态或责任信息与最近流转记录不一致。", asset.Name), common.JSONMap{"current": common.JSONMap{"status": asset.Status, "location": asset.Location, "custodian": asset.Custodian}, "latestRecord": common.JSONMap{"id": latest.ID, "orderNo": latest.OrderNo, "type": latest.Type, "status": latest.ToStatus, "location": latest.ToLocation, "custodian": latest.ToCustodian, "operatedAt": latest.OperatedAt}})
		}
	}
	return nil
}

func ValidateRiskRuleParameters(code string, parameters common.JSONMap) error {
	allowed := map[string]map[string]struct{}{
		riskRulePendingTooLong:      {"days": {}},
		riskRuleHighValueIdle:       {"minValue": {}, "days": {}},
		riskRuleWarrantyMissing:     {"minValue": {}, "categoryCodes": {}},
		riskRuleWarrantyExpiring:    {"days": {}},
		riskRuleMaintenanceOverdue:  {"days": {}},
		riskRuleFrequentMaintenance: {"windowDays": {}, "count": {}},
		riskRuleFrequentTransfer:    {"windowDays": {}, "count": {}},
		riskRuleLongTermInUse:       {"days": {}},
		riskRuleSimilarCode:         {"maxDistance": {}, "minLength": {}},
	}
	keys, exists := allowed[code]
	if !exists {
		keys = map[string]struct{}{}
	}
	for key := range parameters {
		if _, ok := keys[key]; !ok {
			return fmt.Errorf("规则 %s 不支持参数 %s", code, key)
		}
	}
	for _, key := range []string{"days", "windowDays"} {
		if value, ok := parameters[key]; ok {
			number, valid := numberValue(value)
			if !valid || number < 1 || number > 3650 {
				return fmt.Errorf("%s 必须是 1 到 3650 的整数", key)
			}
		}
	}
	if value, ok := parameters["count"]; ok {
		number, valid := numberValue(value)
		if !valid || number < 2 || number > 1000 {
			return errors.New("count 必须是 2 到 1000 的整数")
		}
	}
	if value, ok := parameters["minValue"]; ok {
		number, valid := floatValue(value)
		if !valid || number < 0 || number > 1e15 {
			return errors.New("minValue 超出允许范围")
		}
	}
	if value, ok := parameters["maxDistance"]; ok {
		number, valid := numberValue(value)
		if !valid || number < 1 || number > 2 {
			return errors.New("maxDistance 必须是 1 或 2")
		}
	}
	if value, ok := parameters["minLength"]; ok {
		number, valid := numberValue(value)
		if !valid || number < 3 || number > 50 {
			return errors.New("minLength 必须是 3 到 50 的整数")
		}
	}
	if value, ok := parameters["categoryCodes"]; ok {
		codes := stringSlice(value)
		if len(codes) == 0 || len(codes) > 100 {
			return errors.New("categoryCodes 必须包含 1 到 100 个分类编码")
		}
	}
	return nil
}

func stateEnteredAt(asset model.Asset, records []model.AssetOperationRecord, status string) (time.Time, string) {
	operationType := map[string]string{
		model.AssetStatusIdle: "return", model.AssetStatusInUse: "issue", model.AssetStatusMaintenance: "maintenance",
	}[status]
	for _, record := range records {
		if record.ToStatus != status {
			continue
		}
		if status == model.AssetStatusIdle && record.Type != "return" && record.Type != "inbound" {
			continue
		}
		if status != model.AssetStatusIdle && record.Type != operationType {
			continue
		}
		return record.OperatedAt, "operation_record"
	}
	if !asset.UpdatedAt.IsZero() {
		return asset.UpdatedAt, "asset_updated_at"
	}
	return asset.CreatedAt, "asset_created_at"
}

func countOperations(records []model.AssetOperationRecord, operationType string, since time.Time) (int, []time.Time) {
	timestamps := make([]time.Time, 0)
	for _, record := range records {
		if record.Type == operationType && !record.OperatedAt.Before(since) {
			timestamps = append(timestamps, record.OperatedAt)
		}
	}
	return len(timestamps), timestamps
}

func isImportantAsset(asset model.Asset, parameters common.JSONMap) bool {
	minValue := paramFloat(parameters, "minValue", 10000)
	if math.Max(asset.CurrentValue, asset.OriginalValue) >= minValue {
		return true
	}
	categoryCode := strings.ToUpper(strings.TrimSpace(asset.Category.Code))
	for _, code := range stringSlice(parameters["categoryCodes"]) {
		if strings.ToUpper(strings.TrimSpace(code)) == categoryCode {
			return true
		}
	}
	return false
}

func startOfRiskDay(value time.Time) time.Time {
	value = value.In(time.Local)
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func elapsedDays(later, earlier time.Time) int {
	return int(startOfRiskDay(later).Sub(startOfRiskDay(earlier)).Hours() / 24)
}

func displayValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未登记"
	}
	return value
}

func normalizeIdentity(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToUpper(strings.TrimSpace(value)) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func paramInt(parameters common.JSONMap, key string, fallback int) int {
	if value, ok := parameters[key]; ok {
		if number, valid := numberValue(value); valid {
			return number
		}
	}
	return fallback
}

func paramFloat(parameters common.JSONMap, key string, fallback float64) float64 {
	if value, ok := parameters[key]; ok {
		if number, valid := floatValue(value); valid {
			return number
		}
	}
	return fallback
}

func numberValue(value any) (int, bool) {
	number, ok := floatValue(value)
	if !ok || number != math.Trunc(number) {
		return 0, false
	}
	return int(number), true
}

func floatValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if ok && strings.TrimSpace(text) != "" {
				result = append(result, strings.TrimSpace(text))
			}
		}
		return result
	default:
		return nil
	}
}
