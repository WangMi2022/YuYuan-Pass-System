package service

import "context"

type PlannerEvaluationCase struct {
	Name          string   `json:"name"`
	Question      string   `json:"question"`
	ExpectedTools []string `json:"expectedTools"`
}

type PlannerEvaluationReport struct {
	TotalCases    int      `json:"totalCases"`
	ExactMatches  int      `json:"exactMatches"`
	ExpectedTools int      `json:"expectedTools"`
	MatchedTools  int      `json:"matchedTools"`
	ExactRate     float64  `json:"exactRate"`
	ToolRecall    float64  `json:"toolRecall"`
	Failures      []string `json:"failures,omitempty"`
}

func EvaluatePlanner(ctx context.Context, planner Planner, cases []PlannerEvaluationCase) PlannerEvaluationReport {
	report := PlannerEvaluationReport{TotalCases: len(cases), Failures: make([]string, 0)}
	for _, item := range cases {
		report.ExpectedTools += len(item.ExpectedTools)
		plan, err := planner.Plan(ctx, PlanRequest{Question: item.Question})
		if err != nil {
			report.Failures = append(report.Failures, item.Name+": "+err.Error())
			continue
		}
		actual := make([]string, 0, len(plan.Calls))
		for _, call := range plan.Calls {
			actual = append(actual, call.Name)
		}
		report.MatchedTools += matchedToolCount(actual, item.ExpectedTools)
		if sameStrings(actual, item.ExpectedTools) {
			report.ExactMatches++
		} else {
			report.Failures = append(report.Failures, item.Name)
		}
	}
	if report.TotalCases > 0 {
		report.ExactRate = float64(report.ExactMatches) / float64(report.TotalCases)
	}
	if report.ExpectedTools > 0 {
		report.ToolRecall = float64(report.MatchedTools) / float64(report.ExpectedTools)
	}
	return report
}

func matchedToolCount(actual, expected []string) int {
	remaining := make(map[string]int, len(actual))
	for _, tool := range actual {
		remaining[tool]++
	}
	matched := 0
	for _, tool := range expected {
		if remaining[tool] > 0 {
			matched++
			remaining[tool]--
		}
	}
	return matched
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
