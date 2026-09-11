package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	assetService "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/service"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
)

func assetQueryFromCall(call ToolCall) (assetService.AssetQuery, error) {
	var query assetService.AssetQuery
	if raw, ok := call.Arguments["query"]; ok {
		encoded, err := json.Marshal(raw)
		if err != nil || raw == nil || decodeQueryJSON(string(encoded), &query) != nil {
			return query, errors.New(assetQueryClarification)
		}
		return query, assetService.ValidateAssetQuery(query)
	}
	if call.Name == "asset.warranty.expiring" {
		days := intArgument(call.Arguments, "days", 30)
		if days < 1 || days > 3660 {
			return query, errors.New("请将质保到期范围设为 1 至 3660 天")
		}
		today := time.Now().In(mustShanghaiLocation())
		query.Where = &assetService.AssetFilter{All: []assetService.AssetFilter{
			{Field: "warrantyEndDate", Op: "gte", Value: today.Format("2006-01-02")},
			{Field: "warrantyEndDate", Op: "lte", Value: today.AddDate(0, 0, days).Format("2006-01-02")},
		}}
		query.OrderBy = []assetService.AssetSort{{Field: "warrantyEndDate", Direction: "asc"}}
		return query, nil
	}
	filters := []assetService.AssetFilter{}
	if holder := stringArgument(call.Arguments, "custodian"); holder != "" {
		filters = append(filters, assetService.AssetFilter{Field: "custodian", Op: "eq", Value: holder})
	}
	if keyword := stringArgument(call.Arguments, "keyword"); keyword != "" {
		alternatives := []assetService.AssetFilter{}
		for _, field := range []string{"assetCode", "name", "brand", "model", "serialNumber", "custodian"} {
			alternatives = append(alternatives, assetService.AssetFilter{Field: field, Op: "contains", Value: keyword})
		}
		filters = append(filters, assetService.AssetFilter{Any: alternatives})
	}
	if len(filters) > 0 {
		query.Where = &assetService.AssetFilter{All: filters}
	} else if call.Question != "" {
		var message string
		query, message, _ = ruleAssetQuery(call.Question, time.Now().In(mustShanghaiLocation()))
		if message != "" {
			return query, errors.New(message)
		}
	}
	return query, assetService.ValidateAssetQuery(query)
}

func (s *smartService) executeAssetQuery(ctx context.Context, call ToolCall) (toolResult, error) {
	query, err := assetQueryFromCall(call)
	if err != nil {
		return toolResult{}, err
	}
	queried, err := assetService.Asset.Query(ctx, query)
	if err != nil {
		return toolResult{}, errors.New("资产查询未完成，请稍后重试；这不代表匹配数量为零")
	}
	data := map[string]any{"list": queried.List, "total": queried.Total, "quantity": queried.Quantity, "originalValue": queried.OriginalValue, "currentValue": queried.CurrentValue, "criteria": queried.Criteria, "query": queried.Query, "page": queried.Page, "pageSize": queried.PageSize, "keyword": stringArgument(call.Arguments, "keyword")}
	if call.Name == "asset.warranty.expiring" && query.Where != nil {
		for _, filter := range query.Where.All {
			if filter.Field == "warrantyEndDate" && filter.Op == "lte" {
				data["until"] = filter.Value
			}
		}
	}
	if holder := stringArgument(call.Arguments, "custodian"); holder != "" {
		data["custodian"] = holder
	}
	answer := fmt.Sprintf("查询条件：%s。\n\n匹配 %d 条资产记录，数量合计 %d，当前估值合计 %.2f 元。", queried.Criteria, queried.Total, queried.Quantity, queried.CurrentValue)
	result := toolResult{Data: data, Answer: answer, Authoritative: true}
	if query.GroupBy != "" {
		data["groups"], data["groupTotal"] = queried.Groups, queried.GroupTotal
		result.Answer += fmt.Sprintf("\n\n共 %d 个分组，本页显示 %d 个：", queried.GroupTotal, len(queried.Groups))
		for _, group := range queried.Groups {
			label := group.Key
			if query.GroupBy == "status" {
				if translated := map[string]string{"pending_inbound": "待入库", "idle": "闲置", "in_use": "在用", "maintenance": "维修中", "retired": "已报废"}[label]; translated != "" {
					label = translated
				}
			}
			if label == "" {
				label = "未填写"
			}
			result.Answer += fmt.Sprintf("\n- %s：%d 条记录，数量 %d，当前估值 %.2f 元。", label, group.Total, group.Quantity, group.CurrentValue)
		}
	} else {
		if queried.Total > int64(len(queried.List)) {
			result.Answer += fmt.Sprintf("\n\n当前为第 %d 页，显示 %d 条；上述统计包含全部匹配记录。", queried.Page, len(queried.List))
		}
		for _, item := range queried.List {
			result.Citations = append(result.Citations, Citation{Type: "asset", ID: item.ID, Label: item.AssetCode + " " + item.Name, Path: "/assetInventory", Params: "id=" + strconv.Itoa(int(item.ID))})
		}
	}
	return result, nil
}

// Validate the session before planning (and before sending any text to a model).
// A short answer to an expiry clarification keeps the original asset conditions.
func (s *smartService) assetClarificationQuestion(ctx context.Context, userID, authorityID, sessionID uint, question string) (string, error) {
	if sessionID == 0 {
		return question, nil
	}
	var session model.CopilotSession
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ? AND authority_id = ?", sessionID, userID, authorityID).First(&session).Error; err != nil {
		return "", errors.New("业务助手会话不存在或无权访问")
	}
	var messages []model.CopilotMessage
	if err := global.GVA_DB.WithContext(ctx).Where("session_id = ? AND user_id = ? AND authority_id = ?", sessionID, userID, authorityID).Order("id DESC").Limit(2).Find(&messages).Error; err != nil {
		return "", errors.New("读取会话上下文失败，请重试")
	}
	if len(messages) != 2 || messages[0].Role != model.MessageRoleAssistant || messages[0].Intent != "clarification" || messages[1].Role != model.MessageRoleUser {
		return question, nil
	}
	reply := strings.Trim(question, " 。！!？?")
	if containsAny(messages[1].Content, "过期", "到期") {
		switch reply {
		case "质保", "保修", "质保过期", "质保已过期", "质保到期", "过保", "是质保", "是保修":
			return strings.NewReplacer("过期", "质保已过期", "到期", "质保已到期").Replace(messages[1].Content), nil
		case "使用年限", "借用逾期":
			return "查询超过" + reply + "的资产", nil
		}
	}
	return question, nil
}

type assetPagingPlanner struct{ plan AssistantPlan }

func (p assetPagingPlanner) Plan(context.Context, PlanRequest) (AssistantPlan, error) {
	return p.plan, nil
}

var assetPageRequest = regexp.MustCompile(`^(?:请)?(?:查看|显示)?第\s*(\d+)\s*页$`)

// Paging reuses the audited conditions, with authorization checked again at
// execution. It does not ask a model to reconstruct or broaden the query.
func (s *smartService) assetPagingPlan(ctx context.Context, userID, authorityID, sessionID uint, question string) (*AssistantPlan, error) {
	reply := strings.Trim(question, " 。！!？?")
	match := assetPageRequest.FindStringSubmatch(reply)
	if reply != "下一页" && reply != "上一页" && len(match) != 2 {
		return nil, nil
	}
	clarify := &AssistantPlan{Intent: "clarification", Planner: rulePlannerName, Clarification: "请先在本会话完成一次资产查询，再查看指定页。"}
	if sessionID == 0 {
		return clarify, nil
	}
	var runs []model.CopilotRun
	if err := global.GVA_DB.WithContext(ctx).Where("session_id = ? AND user_id = ? AND authority_id = ? AND status = ?", sessionID, userID, authorityID, "success").Order("id DESC").Limit(1).Find(&runs).Error; err != nil {
		return nil, errors.New("读取分页条件失败，请重试")
	}
	if len(runs) != 1 || runs[0].Status != "success" || len(runs[0].PlannedTools) != 1 {
		return clarify, nil
	}
	encoded, err := json.Marshal(runs[0].PlannedTools["0"])
	if err != nil {
		return clarify, nil
	}
	var call ToolCall
	if decodeQueryJSON(string(encoded), &call) != nil || (call.Name != "asset.search" && call.Name != "asset.warranty.expiring") {
		return clarify, nil
	}
	query, err := assetQueryFromCall(call)
	if err != nil {
		return clarify, nil
	}
	if query.Page == 0 {
		query.Page = 1
	}
	switch reply {
	case "下一页":
		query.Page++
	case "上一页":
		query.Page--
	default:
		query.Page, _ = strconv.Atoi(match[1])
	}
	if query.Page < 1 || query.Page > 10000 {
		clarify.Clarification = "页码应在 1 至 10000 之间。"
		return clarify, nil
	}
	call.Arguments = map[string]any{"query": query}
	return &AssistantPlan{Intent: call.Intent, Planner: rulePlannerName, Calls: []ToolCall{call}}, nil
}
