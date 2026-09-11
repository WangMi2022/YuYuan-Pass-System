package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/asset/model"
	"gorm.io/gorm"
)

// AssetQuery is a bounded, read-only query language. Identifiers and operators
// are resolved here; callers never supply SQL or ownership/permission scope.
type AssetQuery struct {
	Where           *AssetFilter `json:"where,omitempty"`
	OrderBy         []AssetSort  `json:"orderBy,omitempty"`
	GroupBy         string       `json:"groupBy,omitempty"`
	Page            int          `json:"page,omitempty"`
	PageSize        int          `json:"pageSize,omitempty"`
	MaintenanceFrom string       `json:"maintenanceFrom,omitempty"`
	MaintenanceTo   string       `json:"maintenanceTo,omitempty"`
}

type AssetFilter struct {
	All   []AssetFilter `json:"all,omitempty"`
	Any   []AssetFilter `json:"any,omitempty"`
	Field string        `json:"field,omitempty"`
	Op    string        `json:"op,omitempty"`
	Value any           `json:"value,omitempty"`
}

type AssetSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type AssetQueryTotals struct {
	Total         int64   `json:"total"`
	Quantity      int64   `json:"quantity"`
	OriginalValue float64 `json:"originalValue"`
	CurrentValue  float64 `json:"currentValue"`
}

type AssetQueryGroup struct {
	Key string `json:"key"`
	AssetQueryTotals
}

type AssetQueryResult struct {
	AssetQueryTotals
	List       []model.Asset     `json:"list"`
	Groups     []AssetQueryGroup `json:"groups,omitempty"`
	GroupTotal int64             `json:"groupTotal,omitempty"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	Criteria   string            `json:"criteria"`
	Query      AssetQuery        `json:"query"`
}

type assetQueryField struct{ column, label, kind string }

var assetQueryFields = map[string]assetQueryField{
	"id":               {"assets.id", "资产 ID", "number"},
	"assetCode":        {"assets.asset_code", "资产编号", "text"},
	"name":             {"assets.name", "名称", "text"},
	"brand":            {"assets.brand", "品牌", "text"},
	"model":            {"assets.model", "型号", "text"},
	"serialNumber":     {"assets.serial_number", "序列号", "text"},
	"category":         {"", "分类", "category"},
	"custodian":        {"assets.custodian", "保管人", "custodian"},
	"department":       {"assets.custodian", "保管部门", "department"},
	"location":         {"assets.location", "位置", "text"},
	"supplier":         {"assets.supplier", "供应商", "text"},
	"status":           {"assets.status", "状态", "status"},
	"quantity":         {"assets.quantity", "数量", "number"},
	"unitPrice":        {"assets.unit_price", "采购单价（元）", "number"},
	"originalValue":    {"assets.original_value", "资产原值（元）", "number"},
	"currentValue":     {"assets.current_value", "当前估值（元）", "number"},
	"purchaseDate":     {"assets.purchase_date", "购置日期", "date"},
	"productionDate":   {"assets.production_date", "生产日期", "date"},
	"warrantyEndDate":  {"assets.warranty_end_date", "质保到期日", "date"},
	"maintenanceCount": {"", "已完成维修次数", "number"},
}

var assetStatusLabels = map[string]string{
	"pending_inbound": "待入库", "idle": "闲置", "in_use": "在用",
	"maintenance": "维修中", "retired": "已报废",
}

func queryValidationError() error {
	return errors.New("资产筛选条件不完整或不受支持，请明确字段、比较方式和取值")
}

// ValidateAssetQuery also runs at execution time, including for non-model callers.
func ValidateAssetQuery(query AssetQuery) error {
	_, _, _, err := compileAssetQuery(query, "postgres")
	return err
}

func compileAssetQuery(query AssetQuery, dialect string) (string, []any, string, error) {
	if query.Page < 0 || query.Page > 10000 || query.PageSize < 0 || query.PageSize > 100 || len(query.OrderBy) > 3 {
		return "", nil, "", queryValidationError()
	}
	for _, date := range []string{query.MaintenanceFrom, query.MaintenanceTo} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return "", nil, "", queryValidationError()
			}
		}
	}
	if query.MaintenanceFrom != "" && query.MaintenanceTo != "" && query.MaintenanceFrom > query.MaintenanceTo {
		return "", nil, "", queryValidationError()
	}
	if (query.MaintenanceFrom != "" || query.MaintenanceTo != "") && !AssetQueryUsesMaintenance(query) {
		return "", nil, "", queryValidationError()
	}
	if query.GroupBy != "" && query.GroupBy != "status" && query.GroupBy != "custodian" && query.GroupBy != "category" && query.GroupBy != "location" {
		return "", nil, "", queryValidationError()
	}
	if query.GroupBy != "" && len(query.OrderBy) > 0 {
		return "", nil, "", queryValidationError()
	}
	for _, order := range query.OrderBy {
		field, ok := assetQueryFields[order.Field]
		if !ok || field.column == "" || order.Field == "department" || (order.Direction != "asc" && order.Direction != "desc") {
			return "", nil, "", queryValidationError()
		}
	}
	if query.Where == nil {
		return "", nil, "全部资产", nil
	}
	count := 0
	return compileAssetFilter(*query.Where, query, dialect, 0, &count)
}

func compileAssetFilter(filter AssetFilter, query AssetQuery, dialect string, depth int, count *int) (string, []any, string, error) {
	*count++
	if depth > 4 || *count > 32 {
		return "", nil, "", queryValidationError()
	}
	if len(filter.All) > 0 || len(filter.Any) > 0 {
		if (len(filter.All) > 0 && len(filter.Any) > 0) || filter.Field != "" || filter.Op != "" || filter.Value != nil {
			return "", nil, "", queryValidationError()
		}
		children, joiner, label := filter.All, " AND ", " 且 "
		if len(filter.Any) > 0 {
			children, joiner, label = filter.Any, " OR ", " 或 "
		}
		parts, labels, args := []string{}, []string{}, []any{}
		for _, child := range children {
			sql, values, description, err := compileAssetFilter(child, query, dialect, depth+1, count)
			if err != nil {
				return "", nil, "", err
			}
			parts, labels, args = append(parts, sql), append(labels, description), append(args, values...)
		}
		return "(" + strings.Join(parts, joiner) + ")", args, "（" + strings.Join(labels, label) + "）", nil
	}
	field, ok := assetQueryFields[filter.Field]
	if !ok {
		return "", nil, "", queryValidationError()
	}
	column, args := field.column, []any{}
	if filter.Field == "maintenanceCount" {
		column, args = maintenanceCountSQL(query, dialect)
	}
	if filter.Op == "isNull" || filter.Op == "notNull" {
		if field.kind != "date" || filter.Value != nil {
			return "", nil, "", queryValidationError()
		}
		if filter.Op == "isNull" {
			return column + " IS NULL", args, field.label + "未填写", nil
		}
		return column + " IS NOT NULL", args, field.label + "已填写", nil
	}
	ops := map[string]string{"eq": "=", "ne": "<>", "gt": ">", "gte": ">=", "lt": "<", "lte": "<="}
	labels := map[string]string{"eq": "为", "ne": "不为", "gt": "大于", "gte": "不小于", "lt": "小于", "lte": "不大于", "contains": "包含"}
	op, validOp := ops[filter.Op]
	if !validOp && filter.Op != "contains" {
		return "", nil, "", queryValidationError()
	}
	if field.kind == "number" {
		encoded, err := json.Marshal(filter.Value)
		if err != nil || !validOp {
			return "", nil, "", queryValidationError()
		}
		var number float64
		if json.Unmarshal(encoded, &number) != nil || string(encoded) == "null" || number < 0 || math.IsNaN(number) || math.IsInf(number, 0) || number > 1e15 {
			return "", nil, "", queryValidationError()
		}
		if (filter.Field == "id" || filter.Field == "quantity" || filter.Field == "maintenanceCount") && number != math.Trunc(number) {
			return "", nil, "", queryValidationError()
		}
		return column + " " + op + " ?", append(args, number), field.label + labels[filter.Op] + strconv.FormatFloat(number, 'f', -1, 64), nil
	}
	value, ok := filter.Value.(string)
	if !ok || strings.TrimSpace(value) == "" || len([]rune(value)) > 150 || strings.ContainsAny(value, "\x00\r\n") {
		return "", nil, "", queryValidationError()
	}
	value = strings.TrimSpace(value)
	description := field.label + labels[filter.Op] + "“" + value + "”"
	if field.kind == "date" {
		if _, err := time.Parse("2006-01-02", value); err != nil || !validOp {
			return "", nil, "", queryValidationError()
		}
		return assetQueryDateColumn(column, dialect) + " " + op + " ?", []any{value}, description, nil
	}
	if filter.Op != "eq" && filter.Op != "ne" && filter.Op != "contains" {
		return "", nil, "", queryValidationError()
	}
	if field.kind == "status" {
		label, valid := assetStatusLabels[value]
		if !valid || filter.Op == "contains" {
			return "", nil, "", queryValidationError()
		}
		return column + " " + op + " ?", []any{value}, field.label + labels[filter.Op] + label, nil
	}
	escaped := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value)
	if field.kind == "department" {
		if filter.Op != "eq" || strings.Contains(value, "-") {
			return "", nil, "", queryValidationError()
		}
		return column + " LIKE ? ESCAPE '!'", []any{escaped + "-%"}, description, nil
	}
	if field.kind == "custodian" && (filter.Op == "eq" || filter.Op == "ne") && !strings.Contains(value, "-") {
		predicate := "(" + column + " = ? OR " + column + " LIKE ? ESCAPE '!')"
		if filter.Op == "ne" {
			predicate = "NOT " + predicate
		}
		return predicate, []any{value, "%-" + escaped}, description, nil
	}
	if field.kind == "category" {
		if filter.Op == "ne" {
			return "", nil, "", queryValidationError()
		}
		predicate, arg := "c.name = ?", value
		if filter.Op == "contains" {
			predicate, arg = "LOWER(c.name) LIKE LOWER(?) ESCAPE '!'", "%"+escaped+"%"
		}
		return "EXISTS (SELECT 1 FROM asset_categories c WHERE c.id = assets.category_id AND c.deleted_at IS NULL AND " + predicate + ")", []any{arg}, description, nil
	}
	if filter.Op == "contains" {
		return "LOWER(" + column + ") LIKE LOWER(?) ESCAPE '!'", []any{"%" + escaped + "%"}, description, nil
	}
	return column + " " + op + " ?", []any{value}, description, nil
}

func assetQueryDateColumn(column, dialect string) string {
	// SQLite stores Go date fixtures as timestamp text; DATE() would shift a
	// midnight +08:00 value to the preceding UTC day. Keep its calendar portion.
	if dialect == "sqlite" {
		return "SUBSTR(" + column + ",1,10)"
	}
	return column
}

func maintenanceCountSQL(query AssetQuery, dialect string) (string, []any) {
	sql := `(SELECT COUNT(DISTINCT r.order_id) FROM asset_operation_records r JOIN asset_operation_orders o ON o.id = r.order_id WHERE r.asset_id = assets.id AND r.type = 'maintenance' AND o.type = 'maintenance' AND o.status = 'completed' AND r.deleted_at IS NULL AND o.deleted_at IS NULL`
	args := []any{}
	if query.MaintenanceFrom != "" {
		sql += " AND " + assetQueryDateColumn("o.business_date", dialect) + " >= ?"
		args = append(args, query.MaintenanceFrom)
	}
	if query.MaintenanceTo != "" {
		sql += " AND " + assetQueryDateColumn("o.business_date", dialect) + " <= ?"
		args = append(args, query.MaintenanceTo)
	}
	return sql + ")", args
}

func AssetQueryUsesMaintenance(query AssetQuery) bool {
	var visit func(*AssetFilter) bool
	visit = func(filter *AssetFilter) bool {
		if filter == nil {
			return false
		}
		if filter.Field == "maintenanceCount" {
			return true
		}
		for _, child := range append(append([]AssetFilter{}, filter.All...), filter.Any...) {
			if visit(&child) {
				return true
			}
		}
		return false
	}
	return visit(query.Where)
}

func (s *assetService) Query(ctx context.Context, query AssetQuery) (AssetQueryResult, error) {
	result := AssetQueryResult{List: []model.Asset{}}
	predicate, args, criteria, err := compileAssetQuery(query, global.GVA_DB.Dialector.Name())
	if err != nil {
		return result, err
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	result.Page, result.PageSize, result.Query = query.Page, query.PageSize, query
	if query.MaintenanceFrom != "" || query.MaintenanceTo != "" {
		criteria += fmt.Sprintf("；维修业务日期：%s 至 %s", query.MaintenanceFrom, query.MaintenanceTo)
	}
	for _, order := range query.OrderBy {
		direction := "升序"
		if order.Direction == "desc" {
			direction = "降序"
		}
		criteria += "；按" + assetQueryFields[order.Field].label + direction
	}
	result.Criteria = criteria
	db := global.GVA_DB.WithContext(ctx).Model(&model.Asset{})
	if predicate != "" {
		db = db.Where(predicate, args...)
	}
	const totalsSQL = "COUNT(*) AS total, COALESCE(SUM(assets.quantity),0) AS quantity, COALESCE(SUM(assets.original_value),0) AS original_value, COALESCE(SUM(assets.current_value),0) AS current_value"
	if err = db.Session(&gorm.Session{}).Select(totalsSQL).Scan(&result.AssetQueryTotals).Error; err != nil {
		return result, err
	}
	if query.GroupBy != "" {
		column := assetQueryFields[query.GroupBy].column
		if query.GroupBy == "category" {
			column = "COALESCE((SELECT c.name FROM asset_categories c WHERE c.id = assets.category_id AND c.deleted_at IS NULL), '未分类')"
		}
		groups := db.Session(&gorm.Session{}).Select(column + " AS key, " + totalsSQL).Group(column)
		if err = global.GVA_DB.WithContext(ctx).Table("(?) AS query_groups", groups).Count(&result.GroupTotal).Error; err != nil {
			return result, err
		}
		err = groups.Order("total DESC, key ASC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Scan(&result.Groups).Error
		result.Criteria += "；按" + assetQueryFields[query.GroupBy].label + "汇总"
		return result, err
	}
	listDB := db.Session(&gorm.Session{}).Preload("Category")
	for _, order := range query.OrderBy {
		listDB = listDB.Order(assetQueryFields[order.Field].column + " " + order.Direction)
	}
	if len(query.OrderBy) == 0 {
		listDB = listDB.Order("assets.created_at DESC")
	}
	err = listDB.Order("assets.id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.List).Error
	for index := range result.List {
		model.NormalizeAssetPhotos(&result.List[index])
	}
	return result, err
}

// AssetQuerySchema describes the public query language without exposing SQL.
func AssetQuerySchema() map[string]any {
	fields := make([]string, 0, len(assetQueryFields))
	for name := range assetQueryFields {
		fields = append(fields, name)
	}
	sort.Strings(fields)
	leaf := map[string]any{
		"type": "object", "additionalProperties": false, "required": []string{"field", "op"},
		"properties": map[string]any{
			"field": map[string]any{"type": "string", "enum": fields},
			"op":    map[string]any{"type": "string", "enum": []string{"eq", "ne", "gt", "gte", "lt", "lte", "contains", "isNull", "notNull"}},
			"value": map[string]any{"type": []string{"string", "number", "null"}},
		},
	}
	filterRef := map[string]any{"$ref": "#/$defs/filter"}
	group := func(key string) map[string]any {
		return map[string]any{"type": "object", "additionalProperties": false, "required": []string{key}, "properties": map[string]any{key: map[string]any{"type": "array", "minItems": 1, "maxItems": 32, "items": filterRef}}}
	}
	filter := map[string]any{"anyOf": []any{leaf, group("all"), group("any")}}
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"where":           filterRef,
		"orderBy":         map[string]any{"type": "array", "maxItems": 3, "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"field", "direction"}, "properties": map[string]any{"field": map[string]any{"type": "string", "enum": fields}, "direction": map[string]any{"type": "string", "enum": []string{"asc", "desc"}}}}},
		"groupBy":         map[string]any{"type": "string", "enum": []string{"status", "custodian", "category", "location"}},
		"page":            map[string]any{"type": "integer", "minimum": 1, "maximum": 10000},
		"pageSize":        map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
		"maintenanceFrom": map[string]any{"type": "string", "format": "date"},
		"maintenanceTo":   map[string]any{"type": "string", "format": "date"},
	}, "$defs": map[string]any{"filter": filter}}
}
