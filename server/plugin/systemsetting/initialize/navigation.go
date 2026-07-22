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
const auditMenuName = "auditPlatform"

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

		permissionParent := system.SysBaseMenu{
			ParentId: 0,
			Path:     "permissionManagement", Name: permissionMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 5,
			Meta: system.Meta{Title: "权限管理", Icon: "lock"},
		}
		if err := tx.Where("name = ?", permissionParent.Name).FirstOrCreate(&permissionParent).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", permissionParent.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": permissionParent.Path,
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
			if err := updateChildMenu(tx, permissionParent.ID, item); err != nil {
				return err
			}
		}

		auditParent := system.SysBaseMenu{
			ParentId: 0,
			Path:     "auditPlatform", Name: auditMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 6,
			Meta: system.Meta{Title: "审计平台", Icon: "document-checked"},
		}
		if err := tx.Where("name = ?", auditParent.Name).FirstOrCreate(&auditParent).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", auditParent.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": auditParent.Path,
			"component": auditParent.Component, "hidden": false, "sort": auditParent.Sort,
			"title": auditParent.Meta.Title, "icon": auditParent.Meta.Icon,
		}).Error; err != nil {
			return err
		}
		auditMenus := []navigationItem{
			{name: "operation", title: "操作历史", icon: "pie-chart", sort: 1},
			{name: "loginLog", title: "登录日志", icon: "monitor", sort: 2},
			{name: "sysError", title: "错误日志", icon: "warn", sort: 3},
		}
		for _, item := range auditMenus {
			if err := updateChildMenu(tx, auditParent.ID, item); err != nil {
				return err
			}
		}

		canonicalMenus := []navigationItem{
			{name: "dashboard", title: "首页驾驶舱", icon: "odometer", sort: 1},
			{name: "assetCenter", title: "资产管理", icon: "box", sort: 2},
			{name: "superAdmin", title: "系统管理", icon: "setting", sort: 7},
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
			systemMenus := []navigationItem{
				{name: "dictionary", title: "字典管理", icon: "notebook", sort: 1},
				{name: "sysParams", title: "参数管理", icon: "compass", sort: 2},
				{name: "system", title: "运行配置", icon: "operation", sort: 3},
				{name: "apiToken", title: "API Token", icon: "key", sort: 4},
				{name: "sysVersion", title: "版本管理", icon: "server", sort: 5},
				{name: "systemSettings", title: "系统设置", icon: "setting", sort: 6},
			}
			for _, item := range systemMenus {
				if err := updateChildMenu(tx, systemParent.ID, item); err != nil {
					return err
				}
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
		if err := migrateAuthoritiesForParent(tx, monitor.ID, monitorMenus); err != nil {
			return err
		}
		if err := migrateAuthoritiesForParent(tx, permissionParent.ID, permissionMenus); err != nil {
			return err
		}
		if err := migrateAuthoritiesForParent(tx, auditParent.ID, auditMenus); err != nil {
			return err
		}
		if systemParent.ID != 0 {
			return removeParentAuthoritiesWithoutChildren(tx, systemParent.ID)
		}
		return nil
	})
}

func updateChildMenu(tx *gorm.DB, parentID uint, item navigationItem) error {
	return tx.Model(&system.SysBaseMenu{}).Where("name = ?", item.name).Updates(map[string]any{
		"parent_id": parentID, "menu_level": 1, "hidden": false,
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

// removeParentAuthoritiesWithoutChildren 清理由菜单重组产生的空父菜单授权。
// 仅根据父菜单当前实际拥有的子菜单判断，以兼容后续新增的自定义系统菜单。
func removeParentAuthoritiesWithoutChildren(tx *gorm.DB, parentID uint) error {
	var childMenuIDs []uint
	if err := tx.Model(&system.SysBaseMenu{}).
		Where("parent_id = ?", parentID).
		Pluck("id", &childMenuIDs).Error; err != nil {
		return err
	}
	if len(childMenuIDs) == 0 {
		return nil
	}

	childIDs := make([]string, 0, len(childMenuIDs))
	for _, menuID := range childMenuIDs {
		childIDs = append(childIDs, strconv.Itoa(int(menuID)))
	}

	var authoritiesWithChildren []string
	if err := tx.Model(&system.SysAuthorityMenu{}).
		Distinct("sys_authority_authority_id").
		Where("sys_base_menu_id IN ?", childIDs).
		Pluck("sys_authority_authority_id", &authoritiesWithChildren).Error; err != nil {
		return err
	}

	query := tx.Where("sys_base_menu_id = ?", strconv.Itoa(int(parentID)))
	if len(authoritiesWithChildren) > 0 {
		query = query.Where("sys_authority_authority_id NOT IN ?", authoritiesWithChildren)
	}
	return query.Delete(&system.SysAuthorityMenu{}).Error
}
