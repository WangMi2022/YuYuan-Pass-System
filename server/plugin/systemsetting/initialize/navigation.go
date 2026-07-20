package initialize

import (
	"context"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

const collaborationMenuName = "collaborationCenter"
const monitorMenuName = "monitorCenter"
const permissionMenuName = "permissionManagement"

type navigationItem struct {
	name  string
	title string
	icon  string
	sort  int
}

// syncBusinessNavigation 将已有菜单迁移为二开业务信息架构。
// 该过程幂等执行，既适用于新安装，也适用于已经运行的数据库。
func syncBusinessNavigation(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		collaboration := system.SysBaseMenu{
			ParentId: 0,
			Path:     "collaborationCenter", Name: collaborationMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 3,
			Meta: system.Meta{Title: "协同办公", Icon: "briefcase"},
		}
		if err := tx.Where("name = ?", collaboration.Name).FirstOrCreate(&collaboration).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", collaboration.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": collaboration.Path,
			"component": collaboration.Component, "hidden": false, "sort": 3,
			"title": "协同办公", "icon": "briefcase",
		}).Error; err != nil {
			return err
		}

		monitor := system.SysBaseMenu{
			ParentId: 0,
			Path:     "monitorCenter", Name: monitorMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 4,
			Meta: system.Meta{Title: "监控状态", Icon: "monitor"},
		}
		if err := tx.Where("name = ?", monitor.Name).FirstOrCreate(&monitor).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", monitor.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": monitor.Path,
			"component": monitor.Component, "hidden": false, "sort": monitor.Sort,
			"title": monitor.Meta.Title, "icon": monitor.Meta.Icon,
		}).Error; err != nil {
			return err
		}

		canonicalMenus := []navigationItem{
			{name: "dashboard", title: "首页驾驶舱", icon: "odometer", sort: 1},
			{name: "assetCenter", title: "资产管理", icon: "box", sort: 2},
			{name: "superAdmin", title: "系统管理", icon: "setting", sort: 5},
		}
		for _, item := range canonicalMenus {
			if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", item.name).Updates(map[string]any{
				"parent_id": 0, "menu_level": 0, "hidden": false,
				"title": item.title, "icon": item.icon, "sort": item.sort,
			}).Error; err != nil {
				return err
			}
		}

		assetMenus := []navigationItem{
			{name: "assetInventory", title: "资产档案", icon: "list", sort: 1},
			{name: "assetCategories", title: "分类管理", icon: "collection-tag", sort: 2},
		}
		var assetParent system.SysBaseMenu
		if err := tx.Where("name = ?", "assetCenter").First(&assetParent).Error; err == nil {
			for _, item := range assetMenus {
				if err := updateChildMenu(tx, assetParent.ID, item); err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", "assetDashboard").Update("hidden", true).Error; err != nil {
			return err
		}

		collaborationMenus := []navigationItem{
			{name: "documentViewer", title: "文档管理", icon: "document-copy", sort: 1},
			{name: "siteBookmarks", title: "站点收藏", icon: "link", sort: 2},
			{name: "anInfo", title: "公告管理", icon: "bell", sort: 3},
			{name: "upload", title: "媒体库", icon: "picture", sort: 4},
		}
		for _, item := range collaborationMenus {
			if err := updateChildMenu(tx, collaboration.ID, item); err != nil {
				return err
			}
		}

		monitorMenus := []navigationItem{
			{name: "state", title: "服务器负载", icon: "cpu", sort: 1},
		}
		for _, item := range monitorMenus {
			if err := updateChildMenu(tx, monitor.ID, item); err != nil {
				return err
			}
		}

		var systemParent system.SysBaseMenu
		if err := tx.Where("name = ?", "superAdmin").First(&systemParent).Error; err == nil {
			permissionParent := system.SysBaseMenu{
				ParentId: systemParent.ID,
				Path:     "permissionManagement", Name: permissionMenuName, Hidden: false,
				Component: "view/routerHolder.vue", Sort: 1, MenuLevel: 1,
				Meta: system.Meta{Title: "权限管理", Icon: "lock"},
			}
			if err := tx.Where("name = ?", permissionParent.Name).FirstOrCreate(&permissionParent).Error; err != nil {
				return err
			}
			if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", permissionParent.Name).Updates(map[string]any{
				"parent_id": systemParent.ID, "menu_level": 1, "path": permissionParent.Path,
				"component": permissionParent.Component, "hidden": false, "sort": permissionParent.Sort,
				"title": permissionParent.Meta.Title, "icon": permissionParent.Meta.Icon,
			}).Error; err != nil {
				return err
			}

			permissionMenus := []navigationItem{
				{name: "user", title: "用户管理", icon: "coordinate", sort: 1},
				{name: "authority", title: "角色管理", icon: "avatar", sort: 2},
				{name: "api", title: "API 管理", icon: "platform", sort: 3},
				{name: "menu", title: "菜单管理", icon: "tickets", sort: 4},
			}
			for _, item := range permissionMenus {
				if err := updateNestedMenu(tx, permissionParent.ID, item); err != nil {
					return err
				}
			}

			systemMenus := []navigationItem{
				{name: "dictionary", title: "字典管理", icon: "notebook", sort: 5},
				{name: "operation", title: "操作历史", icon: "pie-chart", sort: 6},
				{name: "sysParams", title: "参数管理", icon: "compass", sort: 7},
				{name: "system", title: "运行配置", icon: "operation", sort: 8},
				{name: "apiToken", title: "API Token", icon: "key", sort: 9},
				{name: "loginLog", title: "登录日志", icon: "monitor", sort: 10},
				{name: "sysVersion", title: "版本管理", icon: "server", sort: 11},
				{name: "sysError", title: "错误日志", icon: "warn", sort: 12},
				{name: "systemSettings", title: "系统设置", icon: "setting", sort: 13},
			}
			for _, item := range systemMenus {
				if err := updateChildMenu(tx, systemParent.ID, item); err != nil {
					return err
				}
			}
			if err := migrateAuthoritiesForParent(tx, permissionParent.ID, permissionMenus); err != nil {
				return err
			}
			if err := migrateAuthoritiesForParent(tx, systemParent.ID, []navigationItem{{name: permissionMenuName}}); err != nil {
				return err
			}
		}

		hiddenMenus := []string{
			"about", "example", "systemTools", "https://www.gin-vue-admin.com", "plugin", "AutoRoot",
			"documentManagement", "siteManagement", "assetDashboard", "breakpoint", "customer",
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name IN ?", hiddenMenus).Update("hidden", true).Error; err != nil {
			return err
		}

		if err := migrateAuthoritiesForParent(tx, collaboration.ID, collaborationMenus); err != nil {
			return err
		}
		return migrateAuthoritiesForParent(tx, monitor.ID, monitorMenus)
	})
}

func updateChildMenu(tx *gorm.DB, parentID uint, item navigationItem) error {
	return tx.Model(&system.SysBaseMenu{}).Where("name = ?", item.name).Updates(map[string]any{
		"parent_id": parentID, "menu_level": 1, "hidden": false,
		"title": item.title, "icon": item.icon, "sort": item.sort,
	}).Error
}

func updateNestedMenu(tx *gorm.DB, parentID uint, item navigationItem) error {
	return tx.Model(&system.SysBaseMenu{}).Where("name = ?", item.name).Updates(map[string]any{
		"parent_id": parentID, "menu_level": 2, "hidden": false,
		"title": item.title, "icon": item.icon, "sort": item.sort,
	}).Error
}

func migrateAuthoritiesForParent(tx *gorm.DB, parentID uint, items []navigationItem) error {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.name)
	}
	var menus []system.SysBaseMenu
	if err := tx.Where("name IN ?", names).Find(&menus).Error; err != nil {
		return err
	}
	menuIDs := make([]string, 0, len(menus))
	for _, menu := range menus {
		menuIDs = append(menuIDs, strconv.Itoa(int(menu.ID)))
	}
	if len(menuIDs) == 0 {
		return nil
	}
	var authorityIDs []string
	if err := tx.Model(&system.SysAuthorityMenu{}).
		Distinct("sys_authority_authority_id").
		Where("sys_base_menu_id IN ?", menuIDs).
		Pluck("sys_authority_authority_id", &authorityIDs).Error; err != nil {
		return err
	}
	for _, authorityID := range authorityIDs {
		relation := system.SysAuthorityMenu{
			MenuId: strconv.Itoa(int(parentID)), AuthorityId: authorityID,
		}
		if err := tx.Where(
			"sys_base_menu_id = ? AND sys_authority_authority_id = ?", relation.MenuId, relation.AuthorityId,
		).FirstOrCreate(&relation).Error; err != nil {
			return err
		}
	}
	return nil
}
