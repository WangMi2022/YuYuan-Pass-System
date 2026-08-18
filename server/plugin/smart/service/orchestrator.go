package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ToolExecution struct {
	Call       ToolCall
	Result     toolResult
	DurationMS int64
	Err        error
}

type AssistantAnswer struct {
	Plan           AssistantPlan
	Executions     []ToolExecution
	Tools          []string
	Data           any
	Answer         string
	Citations      []Citation
	Partial        bool
	ModelUsed      bool
	FallbackReason string
	DurationMS     int64
}

type AssistantOrchestrator struct {
	service     *smartService
	planner     Planner
	registry    *ToolRegistry
	maxParallel int
}

func NewAssistantOrchestrator(service *smartService, planner Planner, registry *ToolRegistry) *AssistantOrchestrator {
	if planner == nil {
		planner = NewRulePlanner(nil)
	}
	if registry == nil {
		registry = defaultToolRegistry
	}
	return &AssistantOrchestrator{service: service, planner: planner, registry: registry, maxParallel: defaultMaxPlannedTools}
}

func (o *AssistantOrchestrator) Ask(ctx context.Context, actor AssistantActor, question string) (AssistantAnswer, error) {
	startedAt := time.Now()
	plan, err := o.planner.Plan(ctx, PlanRequest{Actor: actor, Question: question})
	if err != nil {
		return AssistantAnswer{}, err
	}
	if len(plan.Calls) == 0 {
		return AssistantAnswer{}, errors.New("未规划出可执行的只读查询")
	}
	for index := range plan.Calls {
		plan.Calls[index].Question = question
	}
	executions := o.execute(ctx, actor, plan.Calls)
	successCount := 0
	partial := false
	tools := make([]string, 0, len(executions))
	dataByTool := make(map[string]any, len(executions))
	citations := make([]Citation, 0)
	for _, execution := range executions {
		tools = append(tools, execution.Call.Name)
		if execution.Err != nil {
			partial = true
			dataByTool[execution.Call.Name] = map[string]any{"error": execution.Err.Error()}
			continue
		}
		successCount++
		dataByTool[execution.Call.Name] = execution.Result.Data
		citations = append(citations, execution.Result.Citations...)
	}
	if successCount == 0 {
		for _, execution := range executions {
			if execution.Err != nil {
				return AssistantAnswer{}, execution.Err
			}
		}
		return AssistantAnswer{}, errors.New("业务查询未返回结果")
	}

	deterministic := combineDeterministic(executions)
	responseData := any(dataByTool)
	if len(executions) == 1 && executions[0].Err == nil {
		responseData = executions[0].Result.Data
	}
	fallbackReason := ""
	answer := deterministic
	modelUsed := false
	if partial {
		fallbackReason = "partial-tool-failure"
	} else {
		modelAnswer, available := o.service.tryModelSummary(ctx, actor.UserID, actor.AuthorityID, question, strings.Join(tools, ","), deterministic, dataByTool)
		if available && modelSummaryCovers(modelAnswer, executions) {
			answer = modelAnswer
			modelUsed = true
		} else if available {
			fallbackReason = "model-summary-missing-tool-coverage"
		} else {
			fallbackReason = "model-unavailable"
		}
	}
	return AssistantAnswer{
		Plan: plan, Executions: executions, Tools: tools, Data: responseData, Answer: answer,
		Citations: citations, Partial: partial, ModelUsed: modelUsed,
		FallbackReason: fallbackReason, DurationMS: time.Since(startedAt).Milliseconds(),
	}, nil
}

func (o *AssistantOrchestrator) execute(ctx context.Context, actor AssistantActor, calls []ToolCall) []ToolExecution {
	maxParallel := o.maxParallel
	if maxParallel <= 0 || maxParallel > defaultMaxPlannedTools {
		maxParallel = defaultMaxPlannedTools
	}
	results := make([]ToolExecution, len(calls))
	semaphore := make(chan struct{}, maxParallel)
	var wait sync.WaitGroup
	for index, call := range calls {
		wait.Add(1)
		go func(index int, call ToolCall) {
			defer wait.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				results[index] = ToolExecution{Call: call, Err: ctx.Err()}
				return
			}
			startedAt := time.Now()
			result, err := o.registry.Execute(ctx, o.service, actor, call)
			results[index] = ToolExecution{Call: call, Result: result, DurationMS: time.Since(startedAt).Milliseconds(), Err: err}
		}(index, call)
	}
	wait.Wait()
	return results
}

func combineDeterministic(executions []ToolExecution) string {
	parts := make([]string, 0, len(executions))
	for _, execution := range executions {
		if execution.Err != nil {
			parts = append(parts, fmt.Sprintf("%s 查询未完成：%s", toolDisplayName(execution.Call.Name), execution.Err.Error()))
			continue
		}
		part := strings.TrimSpace(execution.Result.Answer)
		part = strings.TrimRight(part, "。；; ")
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "未查询到可用的业务结果。"
	}
	return strings.Join(parts, "；") + "。"
}

func modelSummaryCovers(answer string, executions []ToolExecution) bool {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return false
	}
	for _, execution := range executions {
		if execution.Err != nil {
			continue
		}
		terms := execution.Result.Coverage
		if len(terms) == 0 {
			terms = defaultCoverageTerms(execution.Call.Name, execution.Result.Answer)
		}
		for _, term := range terms {
			if term != "" && !strings.Contains(answer, term) {
				return false
			}
		}
	}
	return true
}

var answerNumberPattern = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?`)

func defaultCoverageTerms(tool, deterministic string) []string {
	terms := []string{toolDisplayName(tool)}
	if number := answerNumberPattern.FindString(deterministic); number != "" {
		terms = append(terms, number)
	}
	return terms
}

func toolDisplayName(tool string) string {
	switch tool {
	case "asset.search", "asset.detail", "asset.risk.list", "asset.warranty.expiring", "asset.custodian.summary", "asset.operation.summary":
		return "资产"
	case "invoice.summary", "invoice.pending_reviews", "invoice.failed_recognitions", "invoice.provider_quality":
		return "发票"
	case "schedule.today":
		return "日程"
	case "announcement.unread":
		return "公告"
	case "knowledge.search":
		return "知识库"
	default:
		return tool
	}
}

func stringArgument(arguments map[string]any, key string) string {
	if arguments == nil {
		return ""
	}
	value, _ := arguments[key].(string)
	return strings.TrimSpace(value)
}

func intArgument(arguments map[string]any, key string, fallback int) int {
	if arguments == nil {
		return fallback
	}
	switch value := arguments[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	}
	return fallback
}
