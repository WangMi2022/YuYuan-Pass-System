package request

import (
	common "github.com/WangMi2022/mit-assets-admin/server/model/common"
	commonReq "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	system "github.com/WangMi2022/mit-assets-admin/server/model/system"
)

type SysAIWorkflowSessionUpsert struct {
	ID             uint                       `json:"id"`
	Tab            string                     `json:"tab"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	ConversationID string                     `json:"conversationId"`
	MessageID      string                     `json:"messageId"`
	CurrentNodeID  string                     `json:"currentNodeId"`
	Settings       common.JSONMap             `json:"settings"`
	FormData       common.JSONMap             `json:"formData"`
	ResultData     common.JSONMap             `json:"resultData"`
	Messages       []system.AIWorkflowMessage `json:"messages"`
}

type SysAIWorkflowSessionSearch struct {
	commonReq.PageInfo
	Tab string `json:"tab" form:"tab"`
}

type SysAIWorkflowMarkdownDump struct {
	ID             uint                       `json:"id"`
	Tab            string                     `json:"tab"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	ConversationID string                     `json:"conversationId"`
	MessageID      string                     `json:"messageId"`
	CurrentNodeID  string                     `json:"currentNodeId"`
	Settings       common.JSONMap             `json:"settings"`
	FormData       common.JSONMap             `json:"formData"`
	ResultData     common.JSONMap             `json:"resultData"`
	Messages       []system.AIWorkflowMessage `json:"messages"`
}
