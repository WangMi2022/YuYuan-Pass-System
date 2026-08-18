package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/ai"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	announcementModel "github.com/WangMi2022/mit-assets-admin/server/plugin/announcement/model"
	announcementResponse "github.com/WangMi2022/mit-assets-admin/server/plugin/announcement/model/response"
	scheduleModel "github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/model"
	scheduleResponse "github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/model/response"
	smartModel "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
)

func TestRulePlannerPlansCompositeScheduleAndAnnouncementQuestion(t *testing.T) {
	location := mustShanghaiLocation()
	now := time.Date(2026, 8, 17, 9, 30, 0, 0, location)
	planner := NewRulePlanner(func() time.Time { return now })

	plan, err := planner.Plan(context.Background(), PlanRequest{Question: "今天有哪些日程和未读公告？"})
	if err != nil {
		t.Fatalf("plan composite question: %v", err)
	}
	if len(plan.Calls) != 2 {
		t.Fatalf("planned calls = %#v, want 2 calls", plan.Calls)
	}
	if plan.Calls[0].Name != "schedule.today" || plan.Calls[1].Name != "announcement.unread" {
		t.Fatalf("unexpected planned calls: %#v", plan.Calls)
	}
	if got := plan.Calls[0].Arguments["date"]; got != "2026-08-17" {
		t.Fatalf("schedule date = %#v, want 2026-08-17", got)
	}
}

func TestRulePlannerParsesTomorrowWithoutAddingScheduleToOtherDomains(t *testing.T) {
	location := mustShanghaiLocation()
	now := time.Date(2026, 8, 17, 23, 30, 0, 0, location)
	planner := NewRulePlanner(func() time.Time { return now })

	plan, err := planner.Plan(context.Background(), PlanRequest{Question: "明天有哪些日程？"})
	if err != nil || len(plan.Calls) != 1 {
		t.Fatalf("unexpected tomorrow plan: %#v, err=%v", plan, err)
	}
	if got := plan.Calls[0].Arguments["date"]; got != "2026-08-18" {
		t.Fatalf("tomorrow date = %#v, want 2026-08-18", got)
	}

	invoicePlan, err := planner.Plan(context.Background(), PlanRequest{Question: "今天有多少待复核发票？"})
	if err != nil || len(invoicePlan.Calls) != 1 || invoicePlan.Calls[0].Name != "invoice.pending_reviews" {
		t.Fatalf("today invoice question must not add schedule tool: %#v, err=%v", invoicePlan, err)
	}
}

func TestScheduleToolReturnsOnlyRequestedDayAndCurrentUser(t *testing.T) {
	database := setupSmartTestDB(t)
	items := []scheduleModel.WorkSchedule{
		{UserID: 1, ClientKey: "daily-user-1", Title: "每日巡检", ScheduleDate: "2026-08-03", ScheduleTime: "08:30", Type: "task", RecurrenceEnabled: true, RecurrenceMode: scheduleModel.RecurrenceDaily},
		{UserID: 1, ClientKey: "today-user-1", Title: "今日盘点", ScheduleDate: "2026-08-17", ScheduleTime: "09:00", Type: "work"},
		{UserID: 1, ClientKey: "tomorrow-user-1", Title: "明日会议", ScheduleDate: "2026-08-18", ScheduleTime: "10:00", Type: "meeting"},
		{UserID: 2, ClientKey: "today-user-2", Title: "他人日程", ScheduleDate: "2026-08-17", ScheduleTime: "11:00", Type: "work"},
	}
	if err := database.Create(&items).Error; err != nil {
		t.Fatal(err)
	}

	result, err := Smart.executeRegisteredTool(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, ToolCall{
		Name: "schedule.today", Arguments: map[string]any{"date": "2026-08-17", "label": "今天"},
	})
	if err != nil {
		t.Fatalf("execute schedule tool: %v", err)
	}
	data := result.Data.(map[string]any)
	list := data["list"].([]scheduleResponse.WorkSchedule)
	if len(list) != 2 || list[0].Date != "2026-08-17" || list[1].Date != "2026-08-17" {
		t.Fatalf("unexpected schedule list: %#v", list)
	}
}

func TestAnnouncementToolReturnsOnlyUnreadItems(t *testing.T) {
	database := setupSmartTestDB(t)
	publishedAt := time.Date(2026, 8, 17, 8, 0, 0, 0, time.Local)
	items := []announcementModel.Info{
		{Title: "未读公告", Content: "内容一", Status: "published", PublishedAt: &publishedAt},
		{Title: "已读公告", Content: "内容二", Status: "published", PublishedAt: &publishedAt},
		{Title: "草稿公告", Content: "内容三", Status: "draft"},
	}
	if err := database.Create(&items).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&announcementModel.Read{UserID: 1, AnnouncementID: items[1].ID, ReadAt: publishedAt}).Error; err != nil {
		t.Fatal(err)
	}

	result, err := Smart.executeRegisteredTool(context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, ToolCall{Name: "announcement.unread"})
	if err != nil {
		t.Fatalf("execute announcement tool: %v", err)
	}
	data := result.Data.(announcementResponse.NotificationResult)
	if data.UnreadCount != 1 || len(data.List) != 1 || data.List[0].Title != "未读公告" || data.List[0].IsRead {
		t.Fatalf("unexpected unread announcements: %#v", data)
	}
}

func TestModelSummaryCoverageRequiresEveryExecutedTool(t *testing.T) {
	executions := []ToolExecution{
		{Call: ToolCall{Name: "schedule.today"}, Result: toolResult{Answer: "今天暂无个人日程。", Coverage: []string{"日程", "0"}}},
		{Call: ToolCall{Name: "announcement.unread"}, Result: toolResult{Answer: "你目前没有未读公告。", Coverage: []string{"公告", "0"}}},
	}
	if modelSummaryCovers("今天没有日程。", executions) {
		t.Fatal("summary missing announcement facts must be rejected")
	}
	if !modelSummaryCovers("今天有 0 项日程，并且有 0 条未读公告。", executions) {
		t.Fatal("summary covering both tools should be accepted")
	}
}

func TestCopilotQueryExecutesCompositeToolsAndPersistsRun(t *testing.T) {
	database := setupSmartTestDB(t)
	previousRegistry := defaultToolRegistry
	defaultToolRegistry = NewToolRegistry(func(uint, string, string) bool { return true })
	t.Cleanup(func() { defaultToolRegistry = previousRegistry })

	result, err := Smart.Query(context.Background(), 1, 888, "今天有哪些日程和未读公告？", 0)
	if err != nil {
		t.Fatalf("query composite question: %v", err)
	}
	if len(result.Tools) != 2 || result.Tools[0] != "schedule.today" || result.Tools[1] != "announcement.unread" {
		t.Fatalf("unexpected response tools: %#v", result.Tools)
	}
	if result.Answer != "今天暂无个人日程；你目前没有未读公告。" || result.ModelUsed || result.Partial {
		t.Fatalf("unexpected deterministic response: %#v", result)
	}
	data, ok := result.Data.(map[string]any)
	if !ok || data["schedule.today"] == nil || data["announcement.unread"] == nil {
		t.Fatalf("unexpected composite data: %#v", result.Data)
	}
	var runCount int64
	if err := database.Model(&smartModel.CopilotRun{}).Where("session_id = ? AND planner = ?", result.SessionID, rulePlannerName).Count(&runCount).Error; err != nil || runCount != 1 {
		t.Fatalf("copilot run not persisted: count=%d err=%v", runCount, err)
	}
}

func TestOrchestratorChecksEachToolPermissionIndependently(t *testing.T) {
	setupSmartTestDB(t)
	registry := NewToolRegistry(func(_ uint, path, _ string) bool {
		return path == "/workSchedule/list"
	})
	location := mustShanghaiLocation()
	planner := NewRulePlanner(func() time.Time { return time.Date(2026, 8, 17, 10, 0, 0, 0, location) })
	result, err := NewAssistantOrchestrator(Smart, planner, registry).Ask(
		context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, "今天有哪些日程和未读公告？",
	)
	if err != nil {
		t.Fatalf("partial composite query: %v", err)
	}
	if !result.Partial || len(result.Executions) != 2 || result.Executions[0].Err != nil || result.Executions[1].Err == nil {
		t.Fatalf("permissions were not checked independently: %#v", result.Executions)
	}
	if !strings.Contains(result.Answer, "announcement.unread") {
		t.Fatalf("partial answer must disclose the unavailable tool: %s", result.Answer)
	}
}

type fakeCopilotGateway struct {
	content string
	err     error
}

func (f fakeCopilotGateway) Complete(context.Context, ai.CompletionRequest) (ai.CompletionResult, error) {
	if f.err != nil {
		return ai.CompletionResult{}, f.err
	}
	return ai.CompletionResult{Content: f.content}, nil
}

func (fakeCopilotGateway) Vision(context.Context, ai.VisionRequest) (ai.VisionResult, error) {
	return ai.VisionResult{}, nil
}

func (fakeCopilotGateway) Stream(context.Context, ai.CompletionRequest) (ai.StreamResult, error) {
	return ai.StreamResult{}, nil
}

func TestOrchestratorDoesNotUseModelWhenAToolPartiallyFails(t *testing.T) {
	setupSmartTestDB(t)
	previousGateway := ai.Default
	ai.Default = fakeCopilotGateway{content: "今天没有日程。"}
	global.GVA_CONFIG.AI.Enabled = true
	t.Cleanup(func() { ai.Default = previousGateway })

	registry := NewToolRegistry(func(_ uint, path, _ string) bool {
		return path == "/workSchedule/list"
	})
	location := mustShanghaiLocation()
	planner := NewRulePlanner(func() time.Time { return time.Date(2026, 8, 17, 10, 0, 0, 0, location) })
	result, err := NewAssistantOrchestrator(Smart, planner, registry).Ask(
		context.Background(), AssistantActor{UserID: 1, AuthorityID: 888}, "今天有哪些日程和未读公告？",
	)
	if err != nil {
		t.Fatalf("partial query: %v", err)
	}
	if !result.Partial || result.ModelUsed || result.FallbackReason != "partial-tool-failure" {
		t.Fatalf("partial query must retain deterministic disclosure: %#v", result)
	}
	if !strings.Contains(result.Answer, "announcement.unread") || !strings.Contains(result.Answer, "查询未完成") {
		t.Fatalf("partial result must name the failed tool: %s", result.Answer)
	}
}

func TestCopilotFallsBackWhenModelFailsOrMissesToolFacts(t *testing.T) {
	for _, item := range []struct {
		name    string
		gateway fakeCopilotGateway
	}{
		{name: "provider-failure", gateway: fakeCopilotGateway{err: errors.New("provider unavailable")}},
		{name: "missing-announcement-coverage", gateway: fakeCopilotGateway{content: "今天没有日程。"}},
	} {
		t.Run(item.name, func(t *testing.T) {
			setupSmartTestDB(t)
			previousRegistry := defaultToolRegistry
			previousGateway := ai.Default
			defaultToolRegistry = NewToolRegistry(func(uint, string, string) bool { return true })
			ai.Default = item.gateway
			global.GVA_CONFIG.AI.Enabled = true
			t.Cleanup(func() {
				defaultToolRegistry = previousRegistry
				ai.Default = previousGateway
			})
			result, err := Smart.Query(context.Background(), 1, 888, "今天有哪些日程和未读公告？", 0)
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if result.ModelUsed || result.Answer != "今天暂无个人日程；你目前没有未读公告。" {
				t.Fatalf("must retain deterministic answer: %#v", result)
			}
		})
	}
}
