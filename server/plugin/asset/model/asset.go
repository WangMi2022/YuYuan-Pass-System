package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// Photo 保存资产图片在 RustFS/MinIO 中的对象信息。
type Photo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	URL  string `json:"url"`
}

// Category 资产分类。
type Category struct {
	global.GVA_MODEL
	Name        string `json:"name" form:"name" gorm:"size:100;not null;uniqueIndex;comment:分类名称"`
	Code        string `json:"code" form:"code" gorm:"size:50;not null;uniqueIndex;comment:分类编码"`
	Description string `json:"description" form:"description" gorm:"size:500;comment:分类说明"`
	Color       string `json:"color" form:"color" gorm:"size:20;default:#334155;comment:展示颜色"`
	Sort        int    `json:"sort" form:"sort" gorm:"default:0;comment:排序"`
	Enabled     bool   `json:"enabled" form:"enabled" gorm:"default:true;comment:是否启用"`
}

func (Category) TableName() string { return "asset_categories" }

// Asset 资产档案，支持按数量和单价计算原值，并独立记录当前估值。
type Asset struct {
	global.GVA_MODEL
	AssetCode       string     `json:"assetCode" form:"assetCode" gorm:"size:80;not null;uniqueIndex;comment:资产编号"`
	Name            string     `json:"name" form:"name" gorm:"size:150;not null;index;comment:资产名称"`
	CategoryID      uint       `json:"categoryId" form:"categoryId" gorm:"not null;index;comment:资产分类ID"`
	Category        Category   `json:"category" gorm:"foreignKey:CategoryID"`
	Brand           string     `json:"brand" form:"brand" gorm:"size:100;comment:品牌"`
	Model           string     `json:"model" form:"model" gorm:"size:120;comment:规格型号"`
	SerialNumber    string     `json:"serialNumber" form:"serialNumber" gorm:"size:120;index;comment:序列号"`
	Quantity        int        `json:"quantity" form:"quantity" gorm:"not null;default:1;comment:数量"`
	Unit            string     `json:"unit" form:"unit" gorm:"size:30;default:件;comment:计量单位"`
	UnitPrice       float64    `json:"unitPrice" form:"unitPrice" gorm:"type:numeric(16,2);not null;default:0;comment:采购单价"`
	OriginalValue   float64    `json:"originalValue" gorm:"type:numeric(18,2);not null;default:0;comment:资产原值"`
	CurrentValue    float64    `json:"currentValue" form:"currentValue" gorm:"type:numeric(18,2);not null;default:0;comment:当前估值"`
	Status          string     `json:"status" form:"status" gorm:"size:30;not null;default:in_use;index;comment:资产状态"`
	Location        string     `json:"location" form:"location" gorm:"size:150;index;comment:存放位置"`
	Custodian       string     `json:"custodian" form:"custodian" gorm:"size:100;index;comment:保管人"`
	Supplier        string     `json:"supplier" form:"supplier" gorm:"size:150;comment:供应商"`
	PurchaseDate    *time.Time `json:"purchaseDate" form:"purchaseDate" gorm:"type:date;comment:购置日期"`
	WarrantyEndDate *time.Time `json:"warrantyEndDate" form:"warrantyEndDate" gorm:"type:date;comment:质保到期日"`
	Photos          []Photo    `json:"photos" gorm:"serializer:json;type:jsonb;comment:资产图片"`
	Remarks         string     `json:"remarks" form:"remarks" gorm:"type:text;comment:备注"`
}

func (Asset) TableName() string { return "assets" }
