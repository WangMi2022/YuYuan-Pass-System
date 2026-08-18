package model

import "github.com/WangMi2022/mit-assets-admin/server/global"

const KnowledgeVisibilityPrivate = "private"

// KnowledgeChunk is a searchable, explicitly-owned slice of non-structured
// knowledge. Dynamic schedules, announcements, assets and invoices never enter
// this table; they stay behind real-time business Tools.
type KnowledgeChunk struct {
	global.GVA_MODEL
	TenantID      uint   `json:"tenantId" gorm:"not null;default:0;index:idx_knowledge_scope"`
	DepartmentID  uint   `json:"departmentId" gorm:"not null;default:0;index:idx_knowledge_scope"`
	OwnerUserID   uint   `json:"ownerUserId" gorm:"not null;index:idx_knowledge_scope;uniqueIndex:idx_knowledge_source_chunk"`
	AuthorityID   uint   `json:"authorityId" gorm:"not null;index:idx_knowledge_scope;uniqueIndex:idx_knowledge_source_chunk"`
	Visibility    string `json:"visibility" gorm:"size:20;not null;default:private;index:idx_knowledge_scope"`
	SourceType    string `json:"sourceType" gorm:"size:40;not null;index;uniqueIndex:idx_knowledge_source_chunk"`
	SourceID      string `json:"sourceId" gorm:"size:120;not null;index;uniqueIndex:idx_knowledge_source_chunk"`
	SourceVersion string `json:"sourceVersion" gorm:"size:64;not null;default:1"`
	ChunkIndex    int    `json:"chunkIndex" gorm:"not null;uniqueIndex:idx_knowledge_source_chunk"`
	Title         string `json:"title" gorm:"size:300;not null"`
	Content       string `json:"content" gorm:"type:text;not null"`
	ContentHash   string `json:"contentHash" gorm:"size:64;not null;index"`
}

func (KnowledgeChunk) TableName() string { return "smart_knowledge_chunks" }
