package service

import (
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
)

func TestClassifyInvoiceReturnsExplainableWinner(t *testing.T) {
	rules := []model.ClassificationRule{
		{Name: "软件关键词", Field: "item", Pattern: "软件", MatchType: "contains", Weight: 80, Enabled: true, CategoryID: 2},
		{Name: "云服务关键词", Field: "raw", Pattern: "云服务", MatchType: "contains", Weight: 30, Enabled: true, CategoryID: 2},
		{Name: "办公关键词", Field: "all", Pattern: "办公", MatchType: "contains", Weight: 70, Enabled: true, CategoryID: 1},
	}
	invoice := model.Invoice{RawText: "年度云服务", Items: []model.InvoiceItem{{Name: "协作软件订阅"}}}
	got := classifyInvoice(invoice, rules)
	if got.CategoryID != 2 || got.Confidence != 0.99 || got.Reason == "" {
		t.Fatalf("unexpected classification: %#v", got)
	}
}

func TestClassifyInvoiceKeepsLowScoreAsSuggestionOnly(t *testing.T) {
	rules := []model.ClassificationRule{{Name: "弱规则", Field: "seller", Pattern: "公司", MatchType: "contains", Weight: 30, Enabled: true, CategoryID: 3}}
	got := classifyInvoice(model.Invoice{SellerName: "示例公司"}, rules)
	if got.CategoryID != 0 || got.CandidateID != 3 || got.Confidence != 0.3 {
		t.Fatalf("unexpected low-score classification: %#v", got)
	}
}
