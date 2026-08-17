package service

import "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"

var Smart = new(smartService)

type smartService struct{}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ReadOnly    bool   `json:"readOnly"`
}

type Citation struct {
	Type   string `json:"type"`
	ID     uint   `json:"id"`
	Label  string `json:"label"`
	Path   string `json:"path"`
	Params string `json:"params,omitempty"`
}

type CopilotResult struct {
	SessionID   uint                   `json:"sessionId"`
	Question    string                 `json:"question"`
	Intent      string                 `json:"intent"`
	Tool        string                 `json:"tool"`
	Scope       string                 `json:"scope"`
	Answer      string                 `json:"answer"`
	Data        any                    `json:"data"`
	Citations   []Citation             `json:"citations"`
	GeneratedAt string                 `json:"generatedAt"`
	ReadOnly    bool                   `json:"readOnly"`
	ModelUsed   bool                   `json:"modelUsed"`
	Messages    []model.CopilotMessage `json:"messages,omitempty"`
}

type ReportListInput struct {
	Page     int
	PageSize int
}

type DraftAcceptInput struct {
	ID uint `json:"id" binding:"required"`
}

// Keep context in the service API so every query remains cancellable and
// future model/tool calls can carry the authenticated actor context.
