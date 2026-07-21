package service

import (
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/model/request"
	announcementResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/announcement/model/response"
	"gorm.io/gorm"
)

var Info = new(info)

type info struct{}

func normalizeInfo(item *model.Info) error {
	if item == nil {
		return errors.New("公告数据为空")
	}
	item.Title = strings.TrimSpace(item.Title)
	item.Content = strings.TrimSpace(item.Content)
	if item.Title == "" {
		return errors.New("请输入公告标题")
	}
	if item.Content == "" {
		return errors.New("请输入公告内容")
	}
	if item.Status == "" {
		item.Status = "published"
	}
	if item.Status != "draft" && item.Status != "published" {
		return errors.New("公告状态不正确")
	}
	return nil
}

// CreateInfo 创建公告；返回值 published 表示本次操作需要触发实时提醒。
func (s *info) CreateInfo(item *model.Info) (published bool, err error) {
	if err = normalizeInfo(item); err != nil {
		return false, err
	}
	if item.Status == "published" {
		now := time.Now()
		item.PublishedAt = &now
		published = true
	}
	err = global.GVA_DB.Create(item).Error
	return published && err == nil, err
}

func (s *info) DeleteInfo(ID string) (err error) {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("announcement_id = ?", ID).Delete(&model.Read{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Info{}, "id = ?", ID).Error
	})
}

func (s *info) DeleteInfoByIds(IDs []string) (err error) {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("announcement_id IN ?", IDs).Delete(&model.Read{}).Error; err != nil {
			return err
		}
		return tx.Delete(&[]model.Info{}, "id in ?", IDs).Error
	})
}

// UpdateInfo 更新公告；仅首次从草稿切换为发布状态时触发提醒。
func (s *info) UpdateInfo(item model.Info) (published bool, err error) {
	if item.ID == 0 {
		return false, errors.New("缺少公告 ID")
	}
	if err = normalizeInfo(&item); err != nil {
		return false, err
	}
	var old model.Info
	if err = global.GVA_DB.First(&old, item.ID).Error; err != nil {
		return false, err
	}
	updates := map[string]any{
		"title":       item.Title,
		"content":     item.Content,
		"user_id":     item.UserID,
		"attachments": item.Attachments,
		"status":      item.Status,
	}
	if item.Status == "published" && old.Status != "published" {
		now := time.Now()
		updates["published_at"] = &now
		published = true
	}
	if item.Status == "draft" {
		updates["published_at"] = nil
	}
	err = global.GVA_DB.Model(&model.Info{}).Where("id = ?", item.ID).Updates(updates).Error
	return published && err == nil, err
}

func (s *info) GetInfo(ID string) (item model.Info, err error) {
	err = global.GVA_DB.Where("id = ?", ID).First(&item).Error
	return
}

func (s *info) GetInfoInfoList(search request.InfoSearch) (list []model.Info, total int64, err error) {
	db := global.GVA_DB.Model(&model.Info{})
	if search.StartCreatedAt != nil && search.EndCreatedAt != nil {
		db = db.Where("created_at BETWEEN ? AND ?", search.StartCreatedAt, search.EndCreatedAt)
	}
	if status := strings.TrimSpace(search.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	err = db.Order("COALESCE(published_at, created_at) DESC").Scopes(search.Paginate()).Find(&list).Error
	return list, total, err
}

func (s *info) GetInfoDataSource() (res map[string][]map[string]any, err error) {
	res = make(map[string][]map[string]any)
	userID := make([]map[string]any, 0)
	global.GVA_DB.Table("sys_users").Select("nick_name as label,id as value").Scan(&userID)
	res["userID"] = userID
	return
}

func (s *info) Notifications(userID uint, limit int) (announcementResponse.NotificationResult, error) {
	result := announcementResponse.NotificationResult{List: make([]announcementResponse.NotificationItem, 0)}
	if userID == 0 {
		return result, errors.New("用户信息无效")
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	base := global.GVA_DB.Table("gva_announcements_info AS a").
		Where("a.deleted_at IS NULL AND a.status = ?", "published")
	if err := base.Select("COUNT(*)").
		Joins("LEFT JOIN gva_announcement_reads AS r ON r.announcement_id = a.id AND r.user_id = ? AND r.deleted_at IS NULL", userID).
		Where("r.id IS NULL").Scan(&result.UnreadCount).Error; err != nil {
		return result, err
	}
	err := global.GVA_DB.Table("gva_announcements_info AS a").
		Select(`a.id, a.created_at, a.updated_at, a.title, a.content, a.attachments, a.published_at,
			COALESCE(u.nick_name, u.username, '系统管理员') AS publisher,
			CASE WHEN r.id IS NULL THEN false ELSE true END AS is_read`).
		Joins("LEFT JOIN sys_users AS u ON u.id = a.user_id").
		Joins("LEFT JOIN gva_announcement_reads AS r ON r.announcement_id = a.id AND r.user_id = ? AND r.deleted_at IS NULL", userID).
		Where("a.deleted_at IS NULL AND a.status = ?", "published").
		Order("COALESCE(a.published_at, a.created_at) DESC").
		Limit(limit).
		Scan(&result.List).Error
	return result, err
}

func (s *info) MarkRead(userID, announcementID uint) error {
	if userID == 0 || announcementID == 0 {
		return errors.New("公告或用户信息无效")
	}
	var count int64
	if err := global.GVA_DB.Model(&model.Info{}).
		Where("id = ? AND status = ?", announcementID, "published").Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	now := time.Now()
	read := model.Read{UserID: userID, AnnouncementID: announcementID, ReadAt: now}
	return global.GVA_DB.Where("user_id = ? AND announcement_id = ?", userID, announcementID).
		FirstOrCreate(&read).Error
}

func (s *info) MarkAllRead(userID uint) error {
	if userID == 0 {
		return errors.New("用户信息无效")
	}
	return global.GVA_DB.Exec(`
		INSERT INTO gva_announcement_reads (created_at, updated_at, user_id, announcement_id, read_at)
		SELECT NOW(), NOW(), ?, a.id, NOW()
		FROM gva_announcements_info AS a
		WHERE a.deleted_at IS NULL AND a.status = 'published'
		ON CONFLICT (user_id, announcement_id) DO NOTHING`, userID).Error
}
