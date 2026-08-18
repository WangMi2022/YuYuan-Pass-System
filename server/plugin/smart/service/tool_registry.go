package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

type ToolPermissionChecker func(authorityID uint, path, method string) bool

type ToolSpec struct {
	Definition     ToolDefinition
	Intent         string
	PermissionPath string
}

type ToolRegistry struct {
	specs             map[string]ToolSpec
	orderedNames      []string
	permissionChecker ToolPermissionChecker
}

func NewToolRegistry(permissionChecker ToolPermissionChecker) *ToolRegistry {
	registry := &ToolRegistry{
		specs:             make(map[string]ToolSpec),
		permissionChecker: permissionChecker,
	}
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.search", Description: "按编号、名称、品牌、型号或序列号查询资产", ReadOnly: true, InputSchema: objectSchema("keyword", "string")}, Intent: "asset", PermissionPath: "/asset/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.detail", Description: "查询单项资产详情", ReadOnly: true, InputSchema: objectSchema("id", "integer")}, Intent: "asset_detail", PermissionPath: "/asset/detail"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.risk.list", Description: "查询开放资产风险和异常", ReadOnly: true}, Intent: "risk", PermissionPath: "/assetRisk/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.warranty.expiring", Description: "查询即将到期的资产质保", ReadOnly: true, InputSchema: objectSchema("days", "integer")}, Intent: "warranty", PermissionPath: "/asset/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.custodian.summary", Description: "按保管人汇总资产数量和价值", ReadOnly: true}, Intent: "custodian", PermissionPath: "/asset/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "asset.operation.summary", Description: "汇总资产流转单据状态", ReadOnly: true}, Intent: "operation", PermissionPath: "/assetOperation/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "invoice.summary", Description: "查询当前权限范围内发票汇总", ReadOnly: true}, Intent: "invoice", PermissionPath: "/invoice/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "invoice.pending_reviews", Description: "查询待复核发票", ReadOnly: true}, Intent: "invoice_review", PermissionPath: "/invoice/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "invoice.failed_recognitions", Description: "查询识别失败发票", ReadOnly: true}, Intent: "invoice_failed", PermissionPath: "/invoice/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "invoice.provider_quality", Description: "查询发票识别 Provider 质量", ReadOnly: true}, Intent: "invoice_quality", PermissionPath: "/invoiceQuality/providerMetrics"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "schedule.today", Description: "按问题中的日期范围查询本人日程", ReadOnly: true, InputSchema: dateRangeSchema()}, Intent: "schedule", PermissionPath: "/workSchedule/list"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "announcement.unread", Description: "查询本人未读公告", ReadOnly: true, InputSchema: objectSchema("limit", "integer")}, Intent: "announcement", PermissionPath: "/info/notifications"})
	registry.register(ToolSpec{Definition: ToolDefinition{Name: "knowledge.search", Description: "在当前用户明确归属的制度、手册和文档知识中检索", ReadOnly: true, InputSchema: objectSchema("query", "string")}, Intent: "knowledge", PermissionPath: "/document/list"})
	return registry
}

func objectSchema(property, propertyType string) map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{property: map[string]any{"type": propertyType}},
	}
}

func dateRangeSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"date": map[string]any{"type": "string", "format": "date"},
			"from": map[string]any{"type": "string", "format": "date"},
			"to":   map[string]any{"type": "string", "format": "date"},
		},
	}
}

func (r *ToolRegistry) register(spec ToolSpec) {
	if spec.Definition.Name == "" {
		return
	}
	if _, exists := r.specs[spec.Definition.Name]; !exists {
		r.orderedNames = append(r.orderedNames, spec.Definition.Name)
	}
	r.specs[spec.Definition.Name] = spec
}

func (r *ToolRegistry) Spec(name string) (ToolSpec, bool) {
	if r == nil {
		return ToolSpec{}, false
	}
	spec, ok := r.specs[name]
	return spec, ok
}

func (r *ToolRegistry) Definitions() []ToolDefinition {
	if r == nil {
		return nil
	}
	definitions := make([]ToolDefinition, 0, len(r.orderedNames))
	for _, name := range r.orderedNames {
		definitions = append(definitions, r.specs[name].Definition)
	}
	return definitions
}

func (r *ToolRegistry) AvailableDefinitions(authorityID uint) []ToolDefinition {
	definitions := make([]ToolDefinition, 0, len(r.orderedNames))
	for _, name := range r.orderedNames {
		spec := r.specs[name]
		if r.allowed(authorityID, spec) {
			definitions = append(definitions, spec.Definition)
		}
	}
	return definitions
}

func (r *ToolRegistry) allowed(authorityID uint, spec ToolSpec) bool {
	return spec.Definition.ReadOnly && r.permissionChecker != nil && r.permissionChecker(authorityID, spec.PermissionPath, "GET")
}

type ToolPermissionError struct {
	Tool string
}

func (e ToolPermissionError) Error() string {
	return fmt.Sprintf("当前角色没有 %s 的读取权限", e.Tool)
}

func (r *ToolRegistry) Execute(ctx context.Context, service *smartService, actor AssistantActor, call ToolCall) (toolResult, error) {
	spec, ok := r.Spec(call.Name)
	if !ok || !spec.Definition.ReadOnly {
		return toolResult{}, errors.New("未注册的只读业务助手 Tool")
	}
	if !r.allowed(actor.AuthorityID, spec) {
		return toolResult{}, ToolPermissionError{Tool: call.Name}
	}
	return service.executeRegisteredTool(ctx, actor, call)
}

func (r *ToolRegistry) Names() []string {
	names := append([]string(nil), r.orderedNames...)
	sort.Strings(names)
	return names
}

var defaultToolRegistry = NewToolRegistry(permissionAllowed)
var toolDefinitions = defaultToolRegistry.Definitions()

func toolPermissionAllowed(authorityID uint, tool string) bool {
	spec, ok := defaultToolRegistry.Spec(tool)
	return ok && defaultToolRegistry.allowed(authorityID, spec)
}

func toolPermissionPath(tool string) (string, bool) {
	spec, ok := defaultToolRegistry.Spec(tool)
	if !ok {
		return "", false
	}
	return spec.PermissionPath, true
}
