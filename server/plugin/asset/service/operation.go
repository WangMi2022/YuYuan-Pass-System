package service

import (
	"errors"
	"fmt"
	"sort"
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
	"inbound":     {prefix: "RK", label: "入库", statuses: []string{model.AssetStatusPendingInbound}},
	"issue":       {prefix: "LY", label: "领用", statuses: []string{model.AssetStatusIdle}},
	"transfer":    {prefix: "DB", label: "调拨", statuses: []string{model.AssetStatusIdle, model.AssetStatusInUse}},
	"return":      {prefix: "GH", label: "归还", statuses: []string{model.AssetStatusInUse, model.AssetStatusMaintenance}},
	"maintenance": {prefix: "WX", label: "维修", statuses: []string{model.AssetStatusIdle, model.AssetStatusInUse}},
	"scrap":       {prefix: "BF", label: "报废", statuses: []string{model.AssetStatusIdle, model.AssetStatusInUse, model.AssetStatusMaintenance}},
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
		return model.AssetStatusIdle, nil
	case "issue":
		return model.AssetStatusInUse, nil
	case "transfer":
		return currentStatus, nil
	case "maintenance":
		return model.AssetStatusMaintenance, nil
	case "scrap":
		return model.AssetStatusRetired, nil
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
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, req.ID).Error; err != nil {
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

type assetUpdate struct {
	setStatus       bool
	status          string
	setLocation     bool
	location        string
	setCustodian    bool
	custodian       string
	resetAssetValue bool
}

func newAssetUpdate(asset model.Asset, toStatus, toLocation, toCustodian string, resetAssetValue bool) assetUpdate {
	return assetUpdate{
		setStatus:       asset.Status != toStatus,
		status:          toStatus,
		setLocation:     asset.Location != toLocation,
		location:        toLocation,
		setCustodian:    asset.Custodian != toCustodian,
		custodian:       toCustodian,
		resetAssetValue: resetAssetValue && asset.CurrentValue != 0,
	}
}

func (u assetUpdate) values() map[string]any {
	updates := make(map[string]any, 4)
	if u.setStatus {
		updates["status"] = u.status
	}
	if u.setLocation {
		updates["location"] = u.location
	}
	if u.setCustodian {
		updates["custodian"] = u.custodian
	}
	if u.resetAssetValue {
		updates["current_value"] = 0
	}
	return updates
}

func (u assetUpdate) empty() bool {
	return !u.setStatus && !u.setLocation && !u.setCustodian && !u.resetAssetValue
}

type assetUpdateGroup struct {
	update   assetUpdate
	assetIDs []uint
}

func lockOperationAssets(tx *gorm.DB, items []model.AssetOperationItem) ([]model.Asset, error) {
	assetIDs := make([]uint, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.AssetID]; exists {
			return nil, errors.New("业务单包含重复资产")
		}
		seen[item.AssetID] = struct{}{}
		assetIDs = append(assetIDs, item.AssetID)
	}
	sort.Slice(assetIDs, func(i, j int) bool { return assetIDs[i] < assetIDs[j] })

	var assets []model.Asset
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", assetIDs).
		Order("id ASC").
		Find(&assets).Error; err != nil {
		return nil, err
	}
	if len(assets) != len(assetIDs) {
		found := make(map[uint]struct{}, len(assets))
		for _, asset := range assets {
			found[asset.ID] = struct{}{}
		}
		for _, item := range items {
			if _, exists := found[item.AssetID]; !exists {
				return nil, fmt.Errorf("资产 %s 不存在", item.AssetCode)
			}
		}
		return nil, errors.New("部分资产不存在或已删除")
	}
	return assets, nil
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
	if len(order.Items) == 0 {
		return errors.New("业务单没有资产明细")
	}
	assets, err := lockOperationAssets(tx, order.Items)
	if err != nil {
		return err
	}
	assetsByID := make(map[uint]model.Asset, len(assets))
	for _, asset := range assets {
		assetsByID[asset.ID] = asset
	}

	now := time.Now()
	records := make([]model.AssetOperationRecord, 0, len(order.Items))
	groups := make([]assetUpdateGroup, 0, 1)
	groupIndex := make(map[assetUpdate]int)
	for i := range order.Items {
		item := &order.Items[i]
		asset := assetsByID[item.AssetID]
		req := assetRequest.SaveOperation{
			Type: order.Type, TargetLocation: order.TargetLocation,
			TargetCustodian: order.TargetCustodian, Reason: order.Reason,
		}
		toStatus, toLocation, toCustodian, err := targetState(order.Type, asset, req)
		if err != nil {
			return fmt.Errorf("资产 %s：%w", asset.AssetCode, err)
		}
		update := newAssetUpdate(asset, toStatus, toLocation, toCustodian, order.Type == "scrap")
		if !update.empty() {
			index, exists := groupIndex[update]
			if !exists {
				index = len(groups)
				groupIndex[update] = index
				groups = append(groups, assetUpdateGroup{update: update})
			}
			groups[index].assetIDs = append(groups[index].assetIDs, asset.ID)
		}
		item.FromStatus = asset.Status
		item.ToStatus = toStatus
		item.FromLocation = asset.Location
		item.ToLocation = toLocation
		item.FromCustodian = asset.Custodian
		item.ToCustodian = toCustodian
		item.Quantity = asset.Quantity
		item.AssetCode = asset.AssetCode
		item.AssetName = asset.Name
		records = append(records, model.AssetOperationRecord{
			OrderID: order.ID, OrderNo: order.OrderNo, Type: order.Type,
			AssetID: asset.ID, AssetCode: asset.AssetCode, AssetName: asset.Name, Quantity: asset.Quantity,
			FromStatus: item.FromStatus, ToStatus: item.ToStatus,
			FromLocation: item.FromLocation, ToLocation: item.ToLocation,
			FromCustodian: item.FromCustodian, ToCustodian: item.ToCustodian,
			OperatorID: userID, OperatorName: userName, OperatedAt: now,
		})
	}

	for _, group := range groups {
		result := tx.Model(&model.Asset{}).Where("id IN ?", group.assetIDs).Updates(group.update.values())
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(group.assetIDs)) {
			return errors.New("部分资产更新失败")
		}
	}
	if err := tx.Omit("Asset").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"quantity", "asset_code", "asset_name", "from_status", "to_status",
			"from_location", "to_location", "from_custodian", "to_custodian", "updated_at",
		}),
	}).Create(&order.Items).Error; err != nil {
		return err
	}
	if len(records) > 0 {
		if err := tx.Create(&records).Error; err != nil {
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
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
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
