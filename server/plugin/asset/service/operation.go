package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Operation = new(operationService)

type operationService struct{}

type operationRule struct {
	prefix   string
	label    string
	statuses []string
}

var operationRules = map[string]operationRule{
	"inbound":     {prefix: "RK", label: "入库", statuses: []string{"idle"}},
	"issue":       {prefix: "LY", label: "领用", statuses: []string{"idle"}},
	"transfer":    {prefix: "DB", label: "调拨", statuses: []string{"idle", "in_use"}},
	"return":      {prefix: "GH", label: "归还", statuses: []string{"in_use", "maintenance"}},
	"maintenance": {prefix: "WX", label: "维修", statuses: []string{"idle", "in_use"}},
	"scrap":       {prefix: "BF", label: "报废", statuses: []string{"idle", "in_use", "maintenance"}},
}

func operationRuleFor(operationType string) (operationRule, error) {
	rule, ok := operationRules[strings.TrimSpace(operationType)]
	if !ok {
		return operationRule{}, errors.New("业务类型不正确")
	}
	return rule, nil
}

func transitionStatus(operationType, currentStatus string) (string, error) {
	rule, err := operationRuleFor(operationType)
	if err != nil {
		return "", err
	}
	allowed := false
	for _, status := range rule.statuses {
		if status == currentStatus {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("当前状态不允许执行%s", rule.label)
	}
	switch operationType {
	case "inbound", "return":
		return "idle", nil
	case "issue":
		return "in_use", nil
	case "transfer":
		return currentStatus, nil
	case "maintenance":
		return "maintenance", nil
	case "scrap":
		return "retired", nil
	default:
		return "", errors.New("业务类型不正确")
	}
}

func normalizeOperation(req *assetRequest.SaveOperation) error {
	if _, err := operationRuleFor(req.Type); err != nil {
		return err
	}
	req.TargetLocation = strings.TrimSpace(req.TargetLocation)
	req.TargetCustodian = strings.TrimSpace(req.TargetCustodian)
	req.Reason = strings.TrimSpace(req.Reason)
	req.Remarks = strings.TrimSpace(req.Remarks)
	if len(req.AssetIDs) == 0 {
		return errors.New("请至少选择一项资产")
	}
	if len(req.AssetIDs) > 100 {
		return errors.New("单张业务单最多选择 100 项资产")
	}
	switch req.Type {
	case "inbound", "transfer", "return":
		if req.TargetLocation == "" {
			return errors.New("请填写目标位置")
		}
	case "issue":
		if req.TargetCustodian == "" {
			return errors.New("请填写领用人或责任人")
		}
	case "maintenance", "scrap":
		if req.Reason == "" {
			return errors.New("请填写业务原因")
		}
	}
	seen := make(map[uint]struct{}, len(req.AssetIDs))
	for _, id := range req.AssetIDs {
		if id == 0 {
			return errors.New("资产参数不正确")
		}
		if _, exists := seen[id]; exists {
			return errors.New("不能重复选择同一项资产")
		}
		seen[id] = struct{}{}
	}
	if req.BusinessDate.IsZero() {
		now := time.Now()
		req.BusinessDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	return nil
}

func targetState(operationType string, asset model.Asset, req assetRequest.SaveOperation) (string, string, string, error) {
	status, err := transitionStatus(operationType, asset.Status)
	if err != nil {
		return "", "", "", err
	}
	location := asset.Location
	custodian := asset.Custodian
	switch operationType {
	case "inbound":
		location = req.TargetLocation
		custodian = ""
	case "issue":
		if req.TargetLocation != "" {
			location = req.TargetLocation
		}
		custodian = req.TargetCustodian
	case "transfer":
		location = req.TargetLocation
		if req.TargetCustodian != "" {
			custodian = req.TargetCustodian
		}
	case "return":
		location = req.TargetLocation
		custodian = ""
	case "maintenance":
		if req.TargetLocation != "" {
			location = req.TargetLocation
		}
		if req.TargetCustodian != "" {
			custodian = req.TargetCustodian
		}
	case "scrap":
		if req.TargetLocation != "" {
			location = req.TargetLocation
		}
		custodian = ""
	}
	return status, location, custodian, nil
}

func buildOperationItems(tx *gorm.DB, req assetRequest.SaveOperation) ([]model.AssetOperationItem, error) {
	var assets []model.Asset
	if err := tx.Where("id IN ?", req.AssetIDs).Find(&assets).Error; err != nil {
		return nil, err
	}
	if len(assets) != len(req.AssetIDs) {
		return nil, errors.New("部分资产不存在或已删除")
	}
	assetMap := make(map[uint]model.Asset, len(assets))
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}
	items := make([]model.AssetOperationItem, 0, len(req.AssetIDs))
	for _, id := range req.AssetIDs {
		asset := assetMap[id]
		toStatus, toLocation, toCustodian, err := targetState(req.Type, asset, req)
		if err != nil {
			return nil, fmt.Errorf("资产 %s：%w", asset.AssetCode, err)
		}
		items = append(items, model.AssetOperationItem{
			AssetID: asset.ID, Quantity: asset.Quantity, AssetCode: asset.AssetCode, AssetName: asset.Name,
			FromStatus: asset.Status, ToStatus: toStatus, FromLocation: asset.Location, ToLocation: toLocation,
			FromCustodian: asset.Custodian, ToCustodian: toCustodian,
		})
	}
	return items, nil
}

func operationNumber(rule operationRule) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%s", rule.prefix, now.Format("20060102"), strings.ToUpper(uuid.NewString()[:8]))
}

func (s *operationService) Create(req assetRequest.SaveOperation, userID uint, userName string) (model.AssetOperationOrder, error) {
	var order model.AssetOperationOrder
	if err := normalizeOperation(&req); err != nil {
		return order, err
	}
	rule, _ := operationRuleFor(req.Type)
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		items, err := buildOperationItems(tx, req)
		if err != nil {
			return err
		}
		order = model.AssetOperationOrder{
			OrderNo: operationNumber(rule), Type: req.Type, Status: model.OperationStatusDraft,
			BusinessDate: req.BusinessDate, TargetLocation: req.TargetLocation, TargetCustodian: req.TargetCustodian,
			Reason: req.Reason, Remarks: req.Remarks, CreatedBy: userID, CreatedByName: userName,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		order.Items = items
		if req.Submit {
			return s.complete(tx, &order, userID, userName)
		}
		return nil
	})
	return order, err
}

func (s *operationService) Update(req assetRequest.SaveOperation, userID uint, userName string) (model.AssetOperationOrder, error) {
	var order model.AssetOperationOrder
	if req.ID == 0 {
		return order, errors.New("缺少业务单 ID")
	}
	if err := normalizeOperation(&req); err != nil {
		return order, err
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&order, req.ID).Error; err != nil {
			return errors.New("业务单不存在")
		}
		if order.Status != model.OperationStatusDraft {
			return errors.New("只有草稿单据可以编辑")
		}
		if order.Type != req.Type {
			return errors.New("不能变更业务单类型")
		}
		items, err := buildOperationItems(tx, req)
		if err != nil {
			return err
		}
		if err := tx.Model(&order).Updates(map[string]any{
			"business_date": req.BusinessDate, "target_location": req.TargetLocation,
			"target_custodian": req.TargetCustodian, "reason": req.Reason, "remarks": req.Remarks,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("order_id = ?", order.ID).Delete(&model.AssetOperationItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		order.BusinessDate = req.BusinessDate
		order.TargetLocation = req.TargetLocation
		order.TargetCustodian = req.TargetCustodian
		order.Reason = req.Reason
		order.Remarks = req.Remarks
		order.Items = items
		if req.Submit {
			return s.complete(tx, &order, userID, userName)
		}
		return nil
	})
	return order, err
}

func (s *operationService) complete(tx *gorm.DB, order *model.AssetOperationOrder, userID uint, userName string) error {
	if order.Status != model.OperationStatusDraft {
		return errors.New("该业务单已经提交")
	}
	if len(order.Items) == 0 {
		if err := tx.Where("order_id = ?", order.ID).Order("id ASC").Find(&order.Items).Error; err != nil {
			return err
		}
	}
	now := time.Now()
	for i := range order.Items {
		item := &order.Items[i]
		var asset model.Asset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&asset, item.AssetID).Error; err != nil {
			return fmt.Errorf("资产 %s 不存在", item.AssetCode)
		}
		req := assetRequest.SaveOperation{
			Type: order.Type, TargetLocation: order.TargetLocation,
			TargetCustodian: order.TargetCustodian, Reason: order.Reason,
		}
		toStatus, toLocation, toCustodian, err := targetState(order.Type, asset, req)
		if err != nil {
			return fmt.Errorf("资产 %s：%w", asset.AssetCode, err)
		}
		updates := map[string]any{"status": toStatus, "location": toLocation, "custodian": toCustodian}
		if order.Type == "scrap" {
			updates["current_value"] = 0
		}
		if err := tx.Model(&asset).Updates(updates).Error; err != nil {
			return err
		}
		item.FromStatus = asset.Status
		item.ToStatus = toStatus
		item.FromLocation = asset.Location
		item.ToLocation = toLocation
		item.FromCustodian = asset.Custodian
		item.ToCustodian = toCustodian
		item.Quantity = asset.Quantity
		if err := tx.Model(item).Updates(map[string]any{
			"quantity": item.Quantity, "asset_code": asset.AssetCode, "asset_name": asset.Name,
			"from_status": item.FromStatus, "to_status": item.ToStatus,
			"from_location": item.FromLocation, "to_location": item.ToLocation,
			"from_custodian": item.FromCustodian, "to_custodian": item.ToCustodian,
		}).Error; err != nil {
			return err
		}
		record := model.AssetOperationRecord{
			OrderID: order.ID, OrderNo: order.OrderNo, Type: order.Type,
			AssetID: asset.ID, AssetCode: asset.AssetCode, AssetName: asset.Name, Quantity: asset.Quantity,
			FromStatus: item.FromStatus, ToStatus: item.ToStatus,
			FromLocation: item.FromLocation, ToLocation: item.ToLocation,
			FromCustodian: item.FromCustodian, ToCustodian: item.ToCustodian,
			OperatorID: userID, OperatorName: userName, OperatedAt: now,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
	}
	order.Status = model.OperationStatusCompleted
	order.CompletedBy = userID
	order.CompletedByName = userName
	order.CompletedAt = &now
	return tx.Model(order).Updates(map[string]any{
		"status": model.OperationStatusCompleted, "completed_by": userID,
		"completed_by_name": userName, "completed_at": now,
	}).Error
}

func (s *operationService) Submit(id, userID uint, userName string) error {
	if id == 0 {
		return errors.New("缺少业务单 ID")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var order model.AssetOperationOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			return errors.New("业务单不存在")
		}
		return s.complete(tx, &order, userID, userName)
	})
}

func preloadOperation(db *gorm.DB) *gorm.DB {
	return db.Preload("Items", func(itemDB *gorm.DB) *gorm.DB { return itemDB.Order("id ASC") }).
		Preload("Items.Asset.Category")
}

func (s *operationService) Get(id uint) (model.AssetOperationOrder, error) {
	var order model.AssetOperationOrder
	err := preloadOperation(global.GVA_DB).First(&order, id).Error
	return order, err
}

func (s *operationService) List(search assetRequest.OperationSearch) ([]model.AssetOperationOrder, int64, error) {
	var list []model.AssetOperationOrder
	var total int64
	db := global.GVA_DB.Model(&model.AssetOperationOrder{})
	if search.Type != "" {
		db = db.Where("type = ?", search.Type)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(`order_no ILIKE ? OR EXISTS (
			SELECT 1 FROM asset_operation_items oi
			WHERE oi.order_id = asset_operation_orders.id AND oi.deleted_at IS NULL
			AND (oi.asset_code ILIKE ? OR oi.asset_name ILIKE ?)
		)`, like, like, like)
	}
	if search.StartDate != "" {
		db = db.Where("business_date >= ?", search.StartDate)
	}
	if search.EndDate != "" {
		db = db.Where("business_date <= ?", search.EndDate)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := preloadOperation(db).Order("business_date DESC, id DESC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *operationService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少业务单 ID")
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var order model.AssetOperationOrder
		if err := tx.First(&order, id).Error; err != nil {
			return errors.New("业务单不存在")
		}
		if order.Status != model.OperationStatusDraft {
			return errors.New("只有草稿单据可以删除")
		}
		if err := tx.Where("order_id = ?", id).Delete(&model.AssetOperationItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&order).Error
	})
}

func (s *operationService) AssetOptions(search assetRequest.OperationAssetSearch) ([]model.Asset, error) {
	rule, err := operationRuleFor(search.Type)
	if err != nil {
		return nil, err
	}
	var list []model.Asset
	db := global.GVA_DB.Preload("Category").Where("status IN ?", rule.statuses)
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("asset_code ILIKE ? OR name ILIKE ? OR serial_number ILIKE ?", like, like, like)
	}
	err = db.Order("asset_code ASC").Limit(200).Find(&list).Error
	return list, err
}
