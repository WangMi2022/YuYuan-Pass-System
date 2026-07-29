package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
)

type classificationResult struct {
	CategoryID  uint
	CandidateID uint
	Confidence  float64
	Reason      string
}

func classifyInvoice(invoice model.Invoice, rules []model.ClassificationRule) classificationResult {
	type score struct {
		categoryID uint
		points     int
		reasons    []string
	}
	scores := make(map[uint]*score)
	for _, rule := range rules {
		if !rule.Enabled || rule.CategoryID == 0 || !ruleMatches(rule, invoice) {
			continue
		}
		item := scores[rule.CategoryID]
		if item == nil {
			item = &score{categoryID: rule.CategoryID}
			scores[rule.CategoryID] = item
		}
		item.points += rule.Weight
		item.reasons = append(item.reasons, fmt.Sprintf("%s（+%d）", rule.Name, rule.Weight))
	}
	if len(scores) == 0 {
		return classificationResult{}
	}
	ordered := make([]*score, 0, len(scores))
	for _, item := range scores {
		ordered = append(ordered, item)
	}
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].points > ordered[j].points })
	best := ordered[0]
	confidence := float64(best.points) / 100
	if confidence > 0.99 {
		confidence = 0.99
	}
	if best.points < 60 {
		return classificationResult{CandidateID: best.categoryID, Confidence: confidence, Reason: strings.Join(best.reasons, "；")}
	}
	return classificationResult{CategoryID: best.categoryID, CandidateID: best.categoryID, Confidence: confidence, Reason: strings.Join(best.reasons, "；")}
}

func ruleMatches(rule model.ClassificationRule, invoice model.Invoice) bool {
	pattern := strings.ToLower(strings.TrimSpace(rule.Pattern))
	if pattern == "" {
		return false
	}
	values := classificationValues(rule.Field, invoice)
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		switch rule.MatchType {
		case "exact":
			if value == pattern {
				return true
			}
		default:
			if strings.Contains(value, pattern) {
				return true
			}
		}
	}
	return false
}

func classificationValues(field string, invoice model.Invoice) []string {
	items := make([]string, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		items = append(items, item.Name)
	}
	switch field {
	case "seller":
		return []string{invoice.SellerName}
	case "item":
		return items
	case "type":
		return []string{invoice.InvoiceType}
	case "raw":
		return []string{invoice.RawText}
	default:
		return append([]string{invoice.SellerName, invoice.InvoiceType, invoice.RawText}, items...)
	}
}
