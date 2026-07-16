package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const (
	OperationStatusDraft     = "draft"
	OperationStatusCompleted = "completed"
)

// AssetOperationOrder 统一承载入库、领用、调拨、归还、维修和报废单据。
type AssetOperationOrder struct {
	global.GVA_MODEL
	OrderNo         string               `json:"orderNo" gorm:"size:80;not null;uniqueIndex;comment:业务单号"`
	Type            string               `json:"type" gorm:"size:30;not null;index;comment:业务类型"`
	Status          string               `json:"status" gorm:"size:20;not null;default:draft;index;comment:单据状态"`
	BusinessDate    time.Time            `json:"businessDate" gorm:"type:date;not null;index;comment:业务日期"`
	TargetLocation  string               `json:"targetLocation" gorm:"size:150;comment:目标位置"`
	TargetCustodian string               `json:"targetCustodian" gorm:"size:100;comment:目标保管人"`
	Reason          string               `json:"reason" gorm:"size:500;comment:业务原因"`
	Remarks         string               `json:"remarks" gorm:"type:text;comment:备注"`
	CreatedBy       uint                 `json:"createdBy" gorm:"index;comment:创建用户"`
	CreatedByName   string               `json:"createdByName" gorm:"size:100;comment:创建人"`
	CompletedBy     uint                 `json:"completedBy" gorm:"index;comment:提交用户"`
	CompletedByName string               `json:"completedByName" gorm:"size:100;comment:提交人"`
	CompletedAt     *time.Time           `json:"completedAt" gorm:"comment:完成时间"`
	Items           []AssetOperationItem `json:"items" gorm:"foreignKey:OrderID"`
}

func (AssetOperationOrder) TableName() string { return "asset_operation_orders" }

// AssetOperationItem 保存单据资产明细和提交前后的快照。
type AssetOperationItem struct {
	global.GVA_MODEL
	OrderID       uint   `json:"orderId" gorm:"not null;index;comment:单据ID"`
	AssetID       uint   `json:"assetId" gorm:"not null;index;comment:资产ID"`
	Asset         Asset  `json:"asset" gorm:"foreignKey:AssetID"`
	Quantity      int    `json:"quantity" gorm:"not null;comment:流转数量"`
	AssetCode     string `json:"assetCode" gorm:"size:80;not null;index;comment:资产编号快照"`
	AssetName     string `json:"assetName" gorm:"size:150;not null;comment:资产名称快照"`
	FromStatus    string `json:"fromStatus" gorm:"size:30;comment:原状态"`
	ToStatus      string `json:"toStatus" gorm:"size:30;comment:目标状态"`
	FromLocation  string `json:"fromLocation" gorm:"size:150;comment:原位置"`
	ToLocation    string `json:"toLocation" gorm:"size:150;comment:目标位置"`
	FromCustodian string `json:"fromCustodian" gorm:"size:100;comment:原保管人"`
	ToCustodian   string `json:"toCustodian" gorm:"size:100;comment:目标保管人"`
}

func (AssetOperationItem) TableName() string { return "asset_operation_items" }

// AssetOperationRecord 是提交后生成的不可变资产流转审计记录。
type AssetOperationRecord struct {
	global.GVA_MODEL
	OrderID       uint      `json:"orderId" gorm:"not null;index;comment:单据ID"`
	OrderNo       string    `json:"orderNo" gorm:"size:80;not null;index;comment:业务单号"`
	Type          string    `json:"type" gorm:"size:30;not null;index;comment:业务类型"`
	AssetID       uint      `json:"assetId" gorm:"not null;index;comment:资产ID"`
	AssetCode     string    `json:"assetCode" gorm:"size:80;not null;index;comment:资产编号快照"`
	AssetName     string    `json:"assetName" gorm:"size:150;not null;comment:资产名称快照"`
	Quantity      int       `json:"quantity" gorm:"not null;comment:流转数量"`
	FromStatus    string    `json:"fromStatus" gorm:"size:30;comment:原状态"`
	ToStatus      string    `json:"toStatus" gorm:"size:30;comment:目标状态"`
	FromLocation  string    `json:"fromLocation" gorm:"size:150;comment:原位置"`
	ToLocation    string    `json:"toLocation" gorm:"size:150;comment:目标位置"`
	FromCustodian string    `json:"fromCustodian" gorm:"size:100;comment:原保管人"`
	ToCustodian   string    `json:"toCustodian" gorm:"size:100;comment:目标保管人"`
	OperatorID    uint      `json:"operatorId" gorm:"index;comment:操作用户"`
	OperatorName  string    `json:"operatorName" gorm:"size:100;comment:操作人"`
	OperatedAt    time.Time `json:"operatedAt" gorm:"not null;index;comment:操作时间"`
}

func (AssetOperationRecord) TableName() string { return "asset_operation_records" }
