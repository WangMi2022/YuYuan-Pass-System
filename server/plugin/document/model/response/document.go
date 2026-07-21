package response

import "time"

// DocumentListItem 只包含文档列表渲染所需的元数据，避免批量读取正文和存储地址。
type DocumentListItem struct {
	ID           uint      `json:"ID"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
	Title        string    `json:"title"`
	OriginalName string    `json:"originalName"`
	FileExt      string    `json:"fileExt"`
	FileSize     int64     `json:"fileSize"`
	MimeType     string    `json:"mimeType"`
	StorageType  string    `json:"storageType"`
	Editable     bool      `json:"editable"`
}
