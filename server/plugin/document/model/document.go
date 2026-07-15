package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// Document 保存上传文档的对象存储信息和在线编辑内容。
type Document struct {
	global.GVA_MODEL
	Title        string `json:"title" form:"title" gorm:"size:180;not null;index;comment:文档标题"`
	OriginalName string `json:"originalName" form:"originalName" gorm:"size:255;not null;comment:原始文件名"`
	FileExt      string `json:"fileExt" form:"fileExt" gorm:"size:30;index;comment:文件扩展名"`
	FileSize     int64  `json:"fileSize" form:"fileSize" gorm:"comment:文件大小"`
	MimeType     string `json:"mimeType" form:"mimeType" gorm:"size:120;comment:MIME类型"`
	StorageType  string `json:"storageType" form:"storageType" gorm:"size:40;comment:存储类型"`
	FileKey      string `json:"fileKey" form:"fileKey" gorm:"size:500;not null;comment:对象存储Key"`
	FileURL      string `json:"fileUrl" form:"fileUrl" gorm:"size:800;comment:对象访问地址"`
	Content      string `json:"content" form:"content" gorm:"type:text;comment:在线编辑内容"`
	Editable     bool   `json:"editable" form:"editable" gorm:"default:true;comment:是否支持在线编辑"`
	Remarks      string `json:"remarks" form:"remarks" gorm:"type:text;comment:备注"`
}

func (Document) TableName() string { return "document_files" }
