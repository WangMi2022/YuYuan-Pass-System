package service

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/site/model"
	siteRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/site/model/request"
	"gorm.io/gorm"
)

var Site = new(siteService)

type siteService struct{}

func normalizeURL(raw string) (string, error) {
	u := strings.TrimSpace(raw)
	if u == "" {
		return "", errors.New("请输入站点地址")
	}
	parsed, err := url.ParseRequestURI(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("站点地址格式不正确")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("仅支持 http 或 https 站点")
	}
	return u, nil
}

func normalizeSite(site *model.SiteBookmark) error {
	if site == nil {
		return errors.New("站点数据为空")
	}
	site.Name = strings.TrimSpace(site.Name)
	if site.Name == "" {
		return errors.New("请输入站点名称")
	}
	normalizedURL, err := normalizeURL(site.URL)
	if err != nil {
		return err
	}
	site.URL = normalizedURL
	site.Category = strings.TrimSpace(site.Category)
	if site.Category == "" {
		site.Category = "常用站点"
	}
	site.Description = strings.TrimSpace(site.Description)
	site.Color = strings.TrimSpace(site.Color)
	if site.Color == "" {
		site.Color = "#2563eb"
	}
	return nil
}

func (s *siteService) Create(site *model.SiteBookmark) error {
	if err := normalizeSite(site); err != nil {
		return err
	}
	return global.GVA_DB.Create(site).Error
}

func (s *siteService) Update(site *model.SiteBookmark) error {
	if site == nil || site.ID == 0 {
		return errors.New("缺少站点 ID")
	}
	if err := normalizeSite(site); err != nil {
		return err
	}
	updates := map[string]any{
		"name":        site.Name,
		"url":         site.URL,
		"category":    site.Category,
		"description": site.Description,
		"color":       site.Color,
		"sort":        site.Sort,
		"enabled":     site.Enabled,
	}
	return global.GVA_DB.Model(&model.SiteBookmark{}).Where("id = ?", site.ID).Updates(updates).Error
}

func (s *siteService) Delete(id uint) error {
	if id == 0 {
		return errors.New("缺少站点 ID")
	}
	return global.GVA_DB.Delete(&model.SiteBookmark{}, id).Error
}

func (s *siteService) Get(id uint) (model.SiteBookmark, error) {
	var site model.SiteBookmark
	err := global.GVA_DB.First(&site, id).Error
	return site, err
}

func (s *siteService) List(search siteRequest.SiteSearch) ([]model.SiteBookmark, int64, error) {
	var list []model.SiteBookmark
	var total int64
	db := global.GVA_DB.Model(&model.SiteBookmark{})
	if keyword := strings.TrimSpace(search.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name ILIKE ? OR url ILIKE ? OR category ILIKE ? OR description ILIKE ?", like, like, like, like)
	}
	if category := strings.TrimSpace(search.Category); category != "" {
		db = db.Where("category = ?", category)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("sort ASC, updated_at DESC, created_at DESC").Scopes(search.Paginate()).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *siteService) Categories() ([]string, error) {
	var categories []string
	err := global.GVA_DB.Model(&model.SiteBookmark{}).
		Where("category <> ''").
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories).Error
	return categories, err
}

func (s *siteService) Visit(id uint) (model.SiteBookmark, error) {
	if id == 0 {
		return model.SiteBookmark{}, errors.New("缺少站点 ID")
	}
	now := time.Now()
	if err := global.GVA_DB.Model(&model.SiteBookmark{}).Where("id = ?", id).Updates(map[string]any{
		"visit_count":     gorm.Expr("visit_count + 1"),
		"last_visited_at": &now,
	}).Error; err != nil {
		return model.SiteBookmark{}, err
	}
	return s.Get(id)
}
