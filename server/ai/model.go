package ai

import (
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

const (
	InvocationStatusSuccess = "success"
	InvocationStatusFailed  = "failed"
	InvocationStatusBlocked = "blocked"

	QuotaScopeUser      = "user"
	QuotaScopeAuthority = "authority"
	QuotaScopeModule    = "module"
	QuotaScopeGlobal    = "global"

	PromptStatusDraft   = "draft"
	PromptStatusActive  = "active"
	PromptStatusRetired = "retired"
)

// ModelInvocation contains only an auditable summary. Prompt, model output,
// images, and credentials are deliberately never persisted here.
type ModelInvocation struct {
	global.GVA_MODEL
	RequestID           string `json:"requestId" gorm:"size:64;not null;uniqueIndex;comment:调用追踪ID"`
	UserID              uint   `json:"userId" gorm:"not null;index;comment:调用用户"`
	AuthorityID         uint   `json:"authorityId" gorm:"not null;index;comment:调用角色"`
	Module              string `json:"module" gorm:"size:80;not null;index;comment:业务模块"`
	Operation           string `json:"operation" gorm:"size:100;not null;index;comment:业务动作"`
	Provider            string `json:"provider" gorm:"size:80;not null;index;comment:模型供应商"`
	Model               string `json:"model" gorm:"size:200;comment:模型名称"`
	PromptKey           string `json:"promptKey" gorm:"size:120;index;comment:Prompt标识"`
	PromptVersion       int    `json:"promptVersion" gorm:"default:0;comment:Prompt版本"`
	InputTokens         int64  `json:"inputTokens" gorm:"not null;default:0;comment:输入Token"`
	OutputTokens        int64  `json:"outputTokens" gorm:"not null;default:0;comment:输出Token"`
	EstimatedCostMicros int64  `json:"estimatedCostMicros" gorm:"not null;default:0;comment:估算费用微单位"`
	DurationMS          int64  `json:"durationMs" gorm:"not null;default:0;comment:耗时毫秒"`
	Status              string `json:"status" gorm:"size:20;not null;index;comment:调用状态"`
	ErrorType           string `json:"errorType" gorm:"size:40;index;comment:错误类型"`
	ObjectType          string `json:"objectType" gorm:"size:80;index;comment:关联对象类型"`
	ObjectID            string `json:"objectId" gorm:"size:120;index;comment:关联对象ID"`
	RedactionCount      int    `json:"redactionCount" gorm:"not null;default:0;comment:脱敏数量"`
	InputHash           string `json:"inputHash" gorm:"size:64;comment:输入哈希"`
	OutputHash          string `json:"outputHash" gorm:"size:64;comment:输出哈希"`
}

func (ModelInvocation) TableName() string { return "ai_model_invocations" }

type UsageQuota struct {
	global.GVA_MODEL
	ScopeType         string `json:"scopeType" gorm:"size:20;not null;uniqueIndex:idx_ai_quota_scope;comment:配额范围"`
	ScopeID           string `json:"scopeId" gorm:"size:120;not null;uniqueIndex:idx_ai_quota_scope;comment:范围标识"`
	DailyRequests     int64  `json:"dailyRequests" gorm:"not null;default:0;comment:每日请求数"`
	DailyTokens       int64  `json:"dailyTokens" gorm:"not null;default:0;comment:每日Token数"`
	MonthlyCostMicros int64  `json:"monthlyCostMicros" gorm:"not null;default:0;comment:每月预算微单位"`
	MaxConcurrency    int    `json:"maxConcurrency" gorm:"not null;default:0;comment:最大并发"`
	Enabled           bool   `json:"enabled" gorm:"not null;default:true;index;comment:是否启用"`
}

func (UsageQuota) TableName() string { return "ai_usage_quotas" }

type PromptTemplate struct {
	global.GVA_MODEL
	PromptKey    string     `json:"promptKey" gorm:"size:120;not null;uniqueIndex:idx_ai_prompt_version;comment:Prompt标识"`
	Version      int        `json:"version" gorm:"not null;uniqueIndex:idx_ai_prompt_version;comment:版本"`
	Content      string     `json:"content" gorm:"type:text;not null;comment:Prompt内容"`
	OutputSchema string     `json:"outputSchema" gorm:"type:text;comment:输出Schema"`
	Status       string     `json:"status" gorm:"size:20;not null;default:draft;index;comment:状态"`
	CreatedBy    uint       `json:"createdBy" gorm:"not null;index;comment:创建人"`
	ActivatedAt  *time.Time `json:"activatedAt" gorm:"index;comment:启用时间"`
}

func (PromptTemplate) TableName() string { return "ai_prompt_templates" }
