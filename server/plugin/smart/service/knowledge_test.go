package service

import (
	"context"
	"strings"
	"testing"
)

func TestKnowledgeStoreIndexesAndScopesPrivateKnowledge(t *testing.T) {
	setupSmartTestDB(t)
	ctx := context.Background()
	owner := AssistantActor{UserID: 1, AuthorityID: 888, TenantID: 0, DepartmentID: 0}
	input := KnowledgeSourceInput{
		SourceType: "document", SourceID: "42", Title: "费用报销制度",
		Content: "员工差旅报销需要在行程结束后十个工作日内提交发票和审批单。",
	}
	if err := Knowledge.ReplaceSource(ctx, owner, input); err != nil {
		t.Fatalf("index knowledge source: %v", err)
	}

	results, err := Knowledge.Search(ctx, owner, "知识库里的报销制度怎么规定？", 5)
	if err != nil || len(results) != 1 || results[0].SourceID != "42" || !strings.Contains(results[0].Snippet, "十个工作日") {
		t.Fatalf("unexpected owner results: %#v, err=%v", results, err)
	}
	otherUserResults, err := Knowledge.Search(ctx, AssistantActor{UserID: 2, AuthorityID: 888}, "报销制度", 5)
	if err != nil || len(otherUserResults) != 0 {
		t.Fatalf("private knowledge leaked across users: %#v, err=%v", otherUserResults, err)
	}
	otherAuthorityResults, err := Knowledge.Search(ctx, AssistantActor{UserID: 1, AuthorityID: 889}, "报销制度", 5)
	if err != nil || len(otherAuthorityResults) != 0 {
		t.Fatalf("private knowledge leaked across authorities: %#v, err=%v", otherAuthorityResults, err)
	}
}

func TestKnowledgeStoreReplaceSourceRemovesStaleChunks(t *testing.T) {
	setupSmartTestDB(t)
	ctx := context.Background()
	actor := AssistantActor{UserID: 1, AuthorityID: 888}
	if err := Knowledge.ReplaceSource(ctx, actor, KnowledgeSourceInput{SourceType: "manual", SourceID: "policy", Title: "采购制度", Content: "旧规则要求纸质审批。"}); err != nil {
		t.Fatal(err)
	}
	if err := Knowledge.ReplaceSource(ctx, actor, KnowledgeSourceInput{SourceType: "manual", SourceID: "policy", Title: "采购制度", Content: "新规则要求线上审批。"}); err != nil {
		t.Fatal(err)
	}
	oldResults, err := Knowledge.Search(ctx, actor, "纸质审批", 5)
	if err != nil || len(oldResults) != 0 {
		t.Fatalf("stale chunks remain: %#v, err=%v", oldResults, err)
	}
	newResults, err := Knowledge.Search(ctx, actor, "线上审批", 5)
	if err != nil || len(newResults) != 1 {
		t.Fatalf("replacement content missing: %#v, err=%v", newResults, err)
	}
}

func TestRulePlannerRoutesExplicitKnowledgeQuestion(t *testing.T) {
	plan, err := NewRulePlanner(nil).Plan(context.Background(), PlanRequest{Question: "知识库里的报销制度怎么规定？"})
	if err != nil || len(plan.Calls) != 1 || plan.Calls[0].Name != "knowledge.search" {
		t.Fatalf("unexpected knowledge plan: %#v, err=%v", plan, err)
	}
}
