package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"
)

func TestPlannerEvaluationDataset(t *testing.T) {
	content, err := os.ReadFile("../eval/planner_cases.json")
	if err != nil {
		t.Fatalf("read planner evaluation cases: %v", err)
	}
	var cases []PlannerEvaluationCase
	if err := json.Unmarshal(content, &cases); err != nil {
		t.Fatalf("decode planner evaluation cases: %v", err)
	}
	location := mustShanghaiLocation()
	planner := NewRulePlanner(func() time.Time { return time.Date(2026, 8, 17, 10, 0, 0, 0, location) })
	report := EvaluatePlanner(context.Background(), planner, cases)
	if report.ExactMatches != len(cases) || report.ToolRecall != 1 {
		t.Fatalf("planner evaluation failed: %#v", report)
	}
}

type failingPlanner struct{}

func (failingPlanner) Plan(context.Context, PlanRequest) (AssistantPlan, error) {
	return AssistantPlan{}, errors.New("planner unavailable")
}

func TestPlannerEvaluationCountsExpectedToolsWhenPlannerFails(t *testing.T) {
	report := EvaluatePlanner(context.Background(), failingPlanner{}, []PlannerEvaluationCase{{
		Name:          "planner-error",
		Question:      "今天有哪些日程和未读公告？",
		ExpectedTools: []string{"schedule.today", "announcement.unread"},
	}})
	if report.ExpectedTools != 2 || report.MatchedTools != 0 || report.ToolRecall != 0 {
		t.Fatalf("planner error must contribute zero matches to the expected-tool denominator: %#v", report)
	}
	if len(report.Failures) != 1 || report.Failures[0] != "planner-error: planner unavailable" {
		t.Fatalf("unexpected planner failure report: %#v", report)
	}
}

type fakeLangGraphPlannerClient struct {
	response LangGraphPlanResponse
}

func (f fakeLangGraphPlannerClient) InvokePlan(context.Context, LangGraphPlanRequest) (LangGraphPlanResponse, error) {
	return f.response, nil
}

func TestLangGraphPlannerAdapterValidatesToolCalls(t *testing.T) {
	registry := NewToolRegistry(func(uint, string, string) bool { return true })
	adapter := LangGraphPlannerAdapter{
		Client: fakeLangGraphPlannerClient{response: LangGraphPlanResponse{Calls: []ToolCall{
			{Name: "asset.delete"},
			{Name: "schedule.today", Arguments: map[string]any{"date": "2026-08-17"}},
			{Name: "schedule.today"},
			{Name: "announcement.unread"},
		}}},
		Registry: registry,
	}
	plan, err := adapter.Plan(context.Background(), PlanRequest{Question: "今天有哪些日程和未读公告？"})
	if err != nil {
		t.Fatalf("langgraph adapter plan: %v", err)
	}
	if len(plan.Calls) != 2 || plan.Calls[0].Name != "schedule.today" || plan.Calls[1].Name != "announcement.unread" {
		t.Fatalf("adapter did not enforce registry: %#v", plan)
	}
	if plan.Planner != langGraphPlannerName {
		t.Fatalf("planner name = %q", plan.Planner)
	}
}
