package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	defaultMaxPlannedTools = 3
	rulePlannerName        = "rule"
	llmPlannerName         = "llm"
	langGraphPlannerName   = "langgraph"
)

type AssistantActor struct {
	UserID       uint
	AuthorityID  uint
	TenantID     uint
	DepartmentID uint
}

type ToolCall struct {
	Name      string         `json:"name"`
	Intent    string         `json:"intent"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Question  string         `json:"-"`
}

type PlanRequest struct {
	Actor    AssistantActor
	Question string
}

type AssistantPlan struct {
	Intent  string     `json:"intent"`
	Calls   []ToolCall `json:"calls"`
	Planner string     `json:"planner"`
}

type Planner interface {
	Plan(context.Context, PlanRequest) (AssistantPlan, error)
}

type RulePlanner struct {
	now      func() time.Time
	location *time.Location
	maxTools int
}

func NewRulePlanner(now func() time.Time) *RulePlanner {
	if now == nil {
		now = time.Now
	}
	return &RulePlanner{now: now, location: mustShanghaiLocation(), maxTools: defaultMaxPlannedTools}
}

func mustShanghaiLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return location
	}
	return time.FixedZone("Asia/Shanghai", 8*60*60)
}

func (p *RulePlanner) Plan(_ context.Context, request PlanRequest) (AssistantPlan, error) {
	question := strings.TrimSpace(request.Question)
	if question == "" {
		return AssistantPlan{}, errors.New("规划问题不能为空")
	}
	q := strings.ToLower(question)
	calls := make([]ToolCall, 0, p.maxTools)
	seen := make(map[string]struct{}, p.maxTools)
	appendCall := func(name, intent string, arguments map[string]any) {
		if len(calls) >= p.maxTools {
			return
		}
		if _, exists := seen[name]; exists {
			return
		}
		seen[name] = struct{}{}
		calls = append(calls, ToolCall{Name: name, Intent: intent, Arguments: arguments})
	}

	knowledgeDomain := containsAny(q, "制度", "手册", "流程文档", "合同", "会议纪要", "知识库", "文档里", "文档中")
	assetDomain := containsAny(q, "资产", "设备", "电脑", "打印机")
	invoiceDomain := containsAny(q, "发票", "金额") || (strings.Contains(q, "报销") && !knowledgeDomain)
	scheduleDomain := containsAny(q, "日程", "安排", "行程", "会议安排")
	announcementDomain := containsAny(q, "公告", "通知", "未读")

	switch {
	case containsAny(q, "质保", "保修"):
		appendCall("asset.warranty.expiring", "warranty", nil)
	case containsAny(q, "风险", "异常"):
		appendCall("asset.risk.list", "risk", nil)
	case containsAny(q, "保管人", "负责人"):
		appendCall("asset.custodian.summary", "custodian", nil)
	case containsAny(q, "流转", "维修", "领用", "报废"):
		appendCall("asset.operation.summary", "operation", nil)
	case assetDomain:
		if containsAny(q, "详情", "明细") {
			appendCall("asset.detail", "asset_detail", nil)
		} else {
			appendCall("asset.search", "asset", map[string]any{"keyword": extractKeyword(question)})
		}
	}

	if invoiceDomain {
		switch {
		case containsAny(q, "质量", "provider", "供应商"):
			appendCall("invoice.provider_quality", "invoice_quality", nil)
		case containsAny(q, "待复核", "待审核"):
			appendCall("invoice.pending_reviews", "invoice_review", nil)
		case containsAny(q, "失败", "识别"):
			appendCall("invoice.failed_recognitions", "invoice_failed", nil)
		default:
			appendCall("invoice.summary", "invoice", nil)
		}
	}

	if scheduleDomain {
		appendCall("schedule.today", "schedule", p.scheduleArguments(q))
	}
	if announcementDomain {
		appendCall("announcement.unread", "announcement", nil)
	}
	if knowledgeDomain {
		appendCall("knowledge.search", "knowledge", map[string]any{"query": question, "limit": 5})
	}

	if len(calls) == 0 {
		appendCall("asset.search", "asset", map[string]any{"keyword": extractKeyword(question)})
	}
	intents := make([]string, 0, len(calls))
	for _, call := range calls {
		intents = append(intents, call.Intent)
	}
	return AssistantPlan{Intent: strings.Join(intents, "+"), Calls: calls, Planner: rulePlannerName}, nil
}

func (p *RulePlanner) scheduleArguments(question string) map[string]any {
	now := p.now().In(p.location)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, p.location)
	end := start
	label := "今天"
	switch {
	case strings.Contains(question, "后天"):
		start = start.AddDate(0, 0, 2)
		end = start
		label = "后天"
	case strings.Contains(question, "明天"):
		start = start.AddDate(0, 0, 1)
		end = start
		label = "明天"
	case strings.Contains(question, "本周") || strings.Contains(question, "这周"):
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = start.AddDate(0, 0, 1-weekday)
		end = start.AddDate(0, 0, 6)
		label = "本周"
	}
	arguments := map[string]any{
		"from":  start.Format("2006-01-02"),
		"to":    end.Format("2006-01-02"),
		"label": label,
	}
	if start.Equal(end) {
		arguments["date"] = start.Format("2006-01-02")
	}
	return arguments
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

// PlannerModel is a safe seam for an optional model planner. The adapter always
// validates model-produced calls against the Tool Registry before execution.
type PlannerModel interface {
	Plan(context.Context, PlanRequest, []ToolDefinition) ([]ToolCall, error)
}

type LLMPlannerAdapter struct {
	Model    PlannerModel
	Registry *ToolRegistry
	Fallback Planner
	MaxTools int
}

func (p LLMPlannerAdapter) Plan(ctx context.Context, request PlanRequest) (AssistantPlan, error) {
	if p.Fallback == nil {
		p.Fallback = NewRulePlanner(nil)
	}
	if p.Model == nil || p.Registry == nil {
		return p.Fallback.Plan(ctx, request)
	}
	calls, err := p.Model.Plan(ctx, request, p.Registry.Definitions())
	if err != nil || len(calls) == 0 {
		return p.Fallback.Plan(ctx, request)
	}
	return validatedExternalPlan(calls, llmPlannerName, p.Registry, p.MaxTools)
}

type LangGraphPlanRequest struct {
	Question string           `json:"question"`
	Actor    AssistantActor   `json:"actor"`
	Tools    []ToolDefinition `json:"tools"`
}

type LangGraphPlanResponse struct {
	Calls []ToolCall `json:"calls"`
}

type LangGraphPlannerClient interface {
	InvokePlan(context.Context, LangGraphPlanRequest) (LangGraphPlanResponse, error)
}

// LangGraphPlannerAdapter is intentionally not wired into the default path.
// It provides a narrow seam for a future Python/JS graph runtime without
// importing that runtime into the Go business process.
type LangGraphPlannerAdapter struct {
	Client   LangGraphPlannerClient
	Registry *ToolRegistry
	Fallback Planner
	MaxTools int
}

func (p LangGraphPlannerAdapter) Plan(ctx context.Context, request PlanRequest) (AssistantPlan, error) {
	if p.Fallback == nil {
		p.Fallback = NewRulePlanner(nil)
	}
	if p.Client == nil || p.Registry == nil {
		return p.Fallback.Plan(ctx, request)
	}
	response, err := p.Client.InvokePlan(ctx, LangGraphPlanRequest{Question: request.Question, Actor: request.Actor, Tools: p.Registry.Definitions()})
	if err != nil || len(response.Calls) == 0 {
		return p.Fallback.Plan(ctx, request)
	}
	return validatedExternalPlan(response.Calls, langGraphPlannerName, p.Registry, p.MaxTools)
}

func validatedExternalPlan(calls []ToolCall, plannerName string, registry *ToolRegistry, maxTools int) (AssistantPlan, error) {
	if maxTools <= 0 || maxTools > defaultMaxPlannedTools {
		maxTools = defaultMaxPlannedTools
	}
	validated := make([]ToolCall, 0, maxTools)
	seen := make(map[string]struct{}, maxTools)
	for _, call := range calls {
		if len(validated) >= maxTools {
			break
		}
		spec, ok := registry.Spec(call.Name)
		if !ok || !spec.Definition.ReadOnly {
			continue
		}
		if _, exists := seen[call.Name]; exists {
			continue
		}
		seen[call.Name] = struct{}{}
		call.Intent = spec.Intent
		validated = append(validated, call)
	}
	if len(validated) == 0 {
		return AssistantPlan{}, errors.New("外部规划器未返回可执行的只读 Tool")
	}
	intents := make([]string, 0, len(validated))
	for _, call := range validated {
		intents = append(intents, call.Intent)
	}
	return AssistantPlan{Intent: strings.Join(intents, "+"), Calls: validated, Planner: plannerName}, nil
}
