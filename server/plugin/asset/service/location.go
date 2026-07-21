package service

import (
	"errors"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"
	assetRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model/request"
	"gorm.io/gorm"
)

var Location = new(locationService)

type locationService struct{}

var validLocationTypes = map[string]struct{}{
	model.LocationTypeInbound:     {},
	model.LocationTypeUsage:       {},
	model.LocationTypeTransfer:    {},
	model.LocationTypeReturn:      {},
	model.LocationTypeMaintenance: {},
	model.LocationTypeDisposal:    {},
}

func normalizeLocation(location *model.Location) error {
	location.Name = strings.TrimSpace(location.Name)
	location.Type = strings.ToLower(strings.TrimSpace(location.Type))
	location.Code = strings.ToUpper(strings.TrimSpace(location.Code))
	location.Description = strings.TrimSpace(location.Description)
	if location.Name == "" {
		return errors.New("位置名称不能为空")
	}
	if _, ok := validLocationTypes[location.Type]; !ok {
		return errors.New("位置类型不正确")
	}
	return nil
}

func validateLocationType(locationType string) (string, error) {
	locationType = strings.ToLower(strings.TrimSpace(locationType))
	if _, ok := validLocationTypes[locationType]; !ok {
		return "", errors.New("位置类型不正确")
	}
	return locationType, nil
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint") || strings.Contains(message, "unique failed")
}

func (s *locationService) Create(location *model.Location) error {
	if err := normalizeLocation(location); err != nil {
		return err
	}
	if err := global.GVA_DB.Create(location).Error; err != nil {
		if isUniqueConstraintError(err) {
			return errors.New("该位置类型下已存在同名位置")
		}
		return err
	}
	return nil
}

func (s *locationService) Update(location *model.Location) error {
	if location.ID == 0 {
		return errors.New("缺少位置 ID")
	}
	if err := normalizeLocation(location); err != nil {
		return err
	}
	result := global.GVA_DB.Model(&model.Location{}).Where("id = ?", location.ID).
		Select("Name", "Type", "Code", "Description", "Sort", "Enabled").Updates(location)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return errors.New("该位置类型下已存在同名位置")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *locationService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少位置 ID")
	}
	result := global.GVA_DB.Unscoped().Delete(&model.Location{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *locationService) List(search assetRequest.LocationSearch) ([]model.Location, int64, error) {
	locationType, err := validateLocationType(search.Type)
	if err != nil {
		return nil, 0, err
	}
	var list []model.Location
	var total int64
	db := global.GVA_DB.Model(&model.Location{}).Where("type = ?", locationType)
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query := db.Order("sort ASC, id ASC").Scopes(search.Paginate())
	if err := query.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *locationService) Options(locationType string) ([]model.Location, error) {
	locationType, err := validateLocationType(locationType)
	if err != nil {
		return nil, err
	}
	var list []model.Location
	err = global.GVA_DB.Where("type = ? AND enabled = ?", locationType, true).
		Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}
