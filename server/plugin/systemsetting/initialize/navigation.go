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
const workCalendarMenuName = "workCalendar"
const systemDataMenuName = "systemData"
const systemConfigurationMenuName = "systemConfiguration"
const systemIntegrationMenuName = "systemIntegration"
const systemIntelligenceMenuName = "systemIntelligence"

type navigationItem struct {
	name  string
	title string
	icon  string
	sort  int
}

type systemNavigationGroup struct {
	name  string
	path  string
	title string
	icon  string
	sort  int
	menus []navigationItem
}

// syncBusinessNavigation 将已有菜单迁移为二开业务信息架构。
// 该过程幂等执行，既适用于新安装，也适用于已经运行的数据库。
func syncBusinessNavigation(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		collaboration := system.SysBaseMenu{
			ParentId: 0,
			Path:     "collaborationCenter", Name: collaborationMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 5,
			Meta: system.Meta{Title: "协同办公", Icon: "briefcase"},
		}
		if err := tx.Where("name = ?", collaboration.Name).FirstOrCreate(&collaboration).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", collaboration.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": collaboration.Path,
			"component": collaboration.Component, "hidden": false, "sort": collaboration.Sort,
			"title": "协同办公", "icon": "briefcase",
		}).Error; err != nil {
			return err
		}

		monitor := system.SysBaseMenu{
			ParentId: 0,
			Path:     "monitorCenter", Name: monitorMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 6,
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
			Component: "view/routerHolder.vue", Sort: 7,
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
			Component: "view/routerHolder.vue", Sort: 8,
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

		workCalendar := system.SysBaseMenu{
			ParentId: 0,
			Path:     workCalendarMenuName, Name: workCalendarMenuName, Hidden: false,
			Component: "view/routerHolder.vue", Sort: 4,
			Meta: system.Meta{Title: "工作日历", Icon: "calendar"},
		}
		if err := tx.Where("name = ?", workCalendar.Name).FirstOrCreate(&workCalendar).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", workCalendar.Name).Updates(map[string]any{
			"parent_id": 0, "menu_level": 0, "path": workCalendar.Path,
			"component": workCalendar.Component, "hidden": false, "sort": workCalendar.Sort,
			"title": workCalendar.Meta.Title, "icon": workCalendar.Meta.Icon,
		}).Error; err != nil {
			return err
		}
		workSchedule := system.SysBaseMenu{
			ParentId: workCalendar.ID, MenuLevel: 1,
			Path: "schedule", Name: "workSchedule", Hidden: false,
			Component: "view/workCalendar/index.vue", Sort: 1,
			Meta: system.Meta{Title: "日程总览", Icon: "calendar", KeepAlive: true},
		}
		if err := tx.Where("name = ?", workSchedule.Name).FirstOrCreate(&workSchedule).Error; err != nil {
			return err
		}
		if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", workSchedule.Name).Updates(map[string]any{
			"parent_id": workCalendar.ID, "menu_level": 1, "path": workSchedule.Path,
			"component": workSchedule.Component, "hidden": false, "sort": workSchedule.Sort,
			"title": workSchedule.Meta.Title, "icon": workSchedule.Meta.Icon, "keep_alive": true,
		}).Error; err != nil {
			return err
		}
		workCalendarMenus := []navigationItem{
			{name: "workSchedule", title: "日程总览", icon: "calendar", sort: 1},
		}

		canonicalMenus := []navigationItem{
			{name: "dashboard", title: "首页驾驶舱", icon: "odometer", sort: 1},
			{name: "assetCenter", title: "资产管理", icon: "box", sort: 2},
			{name: "invoiceCenter", title: "流水管理", icon: "wallet", sort: 3},
			{name: "superAdmin", title: "系统管理", icon: "setting", sort: 9},
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
			systemGroups := []systemNavigationGroup{
				{
					name: systemDataMenuName, path: systemDataMenuName, title: "基础数据", icon: "collection-tag", sort: 1,
					menus: []navigationItem{
						{name: "dictionary", title: "数据字典", icon: "notebook", sort: 1},
						{name: "sysParams", title: "系统参数", icon: "compass", sort: 2},
					},
				},
				{
					name: systemConfigurationMenuName, path: systemConfigurationMenuName, title: "平台设置", icon: "setting", sort: 2,
					menus: []navigationItem{
						{name: "systemSettings", title: "登录外观", icon: "picture", sort: 1},
						{name: "system", title: "运行配置", icon: "operation", sort: 2},
					},
				},
				{
					name: systemIntegrationMenuName, path: systemIntegrationMenuName, title: "开放与运维", icon: "connection", sort: 3,
					menus: []navigationItem{
						{name: "apiToken", title: "接口令牌", icon: "key", sort: 1},
						{name: "sysVersion", title: "配置版本", icon: "server", sort: 2},
					},
				},
				{
					name: systemIntelligenceMenuName, path: systemIntelligenceMenuName, title: "智能服务", icon: "cpu", sort: 4,
					menus: []navigationItem{
						{name: "aiOperations", title: "智能能力配置", icon: "cpu", sort: 1},
					},
				},
			}
			parentItems := make([]navigationItem, 0, len(systemGroups))
			for _, group := range systemGroups {
				groupMenu, err := upsertSystemNavigationGroup(tx, systemParent.ID, group)
				if err != nil {
					return err
				}
				for _, item := range group.menus {
					if err := updateSystemNavigationMenu(tx, groupMenu.ID, item); err != nil {
						return err
					}
				}
				if err := migrateAuthoritiesForParent(tx, groupMenu.ID, group.menus); err != nil {
					return err
				}
				parentItems = append(parentItems, navigationItem{name: group.name})
			}
			if err := migrateAuthoritiesForParent(tx, systemParent.ID, parentItems); err != nil {
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
		if err := migrateAuthoritiesForParent(tx, monitor.ID, monitorMenus); err != nil {
			return err
		}
		if err := migrateAuthoritiesForParent(tx, permissionParent.ID, permissionMenus); err != nil {
			return err
		}
		if err := migrateAuthoritiesForParent(tx, auditParent.ID, auditMenus); err != nil {
			return err
		}
		if err := migrateAuthoritiesForParent(tx, workCalendar.ID, workCalendarMenus); err != nil {
			return err
		}
		if systemParent.ID != 0 {
			legacySystemMenus := append(permissionMenus, auditMenus...)
			return removeLegacySystemParentAuthorities(tx, systemParent.ID, legacySystemMenus)
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

func upsertSystemNavigationGroup(tx *gorm.DB, systemParentID uint, group systemNavigationGroup) (system.SysBaseMenu, error) {
	menu := system.SysBaseMenu{
		ParentId: systemParentID, MenuLevel: 1, Path: group.path, Name: group.name, Hidden: false,
		Component: "view/routerHolder.vue", Sort: group.sort,
		Meta: system.Meta{Title: group.title, Icon: group.icon},
	}
	if err := tx.Where("name = ?", group.name).FirstOrCreate(&menu).Error; err != nil {
		return system.SysBaseMenu{}, err
	}
	if err := tx.Model(&system.SysBaseMenu{}).Where("name = ?", group.name).Updates(map[string]any{
		"parent_id": systemParentID, "menu_level": 1, "path": group.path,
		"component": "view/routerHolder.vue", "hidden": false, "sort": group.sort,
		"title": group.title, "icon": group.icon,
	}).Error; err != nil {
		return system.SysBaseMenu{}, err
	}
	return menu, nil
}

func updateSystemNavigationMenu(tx *gorm.DB, parentID uint, item navigationItem) error {
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

// removeLegacySystemParentAuthorities 仅移除已迁出系统管理的旧菜单造成的冗余父级授权。
// 未关联任何旧迁出菜单的显式父级授权必须保留，避免升级时静默修改角色配置。
func removeLegacySystemParentAuthorities(tx *gorm.DB, parentID uint, legacyItems []navigationItem) error {
	legacyNames := make([]string, 0, len(legacyItems))
	for _, item := range legacyItems {
		legacyNames = append(legacyNames, item.name)
	}
	if len(legacyNames) == 0 {
		return nil
	}

	var legacyMenuIDs []uint
	if err := tx.Model(&system.SysBaseMenu{}).
		Where("name IN ?", legacyNames).
		Pluck("id", &legacyMenuIDs).Error; err != nil {
		return err
	}
	if len(legacyMenuIDs) == 0 {
		return nil
	}

	legacyIDs := make([]string, 0, len(legacyMenuIDs))
	for _, menuID := range legacyMenuIDs {
		legacyIDs = append(legacyIDs, strconv.Itoa(int(menuID)))
	}
	var authoritiesWithLegacyMenus []string
	if err := tx.Model(&system.SysAuthorityMenu{}).
		Distinct("sys_authority_authority_id").
		Where("sys_base_menu_id IN ?", legacyIDs).
		Pluck("sys_authority_authority_id", &authoritiesWithLegacyMenus).Error; err != nil {
		return err
	}
	if len(authoritiesWithLegacyMenus) == 0 {
		return nil
	}

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

	query := tx.Where(
		"sys_base_menu_id = ? AND sys_authority_authority_id IN ?",
		strconv.Itoa(int(parentID)), authoritiesWithLegacyMenus,
	)
	if len(authoritiesWithChildren) > 0 {
		query = query.Where("sys_authority_authority_id NOT IN ?", authoritiesWithChildren)
	}
	return query.Delete(&system.SysAuthorityMenu{}).Error
}
