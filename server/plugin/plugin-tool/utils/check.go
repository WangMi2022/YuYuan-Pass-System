package utils

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ApiMap  = make(map[string][]system.SysApi)
	MenuMap = make(map[string][]system.SysBaseMenu)
	DictMap = make(map[string][]system.SysDictionary)
	rw      sync.RWMutex
)

func getPluginName() string {
	_, file, _, ok := runtime.Caller(2)
	pluginName := ""
	if ok {
		file = filepath.ToSlash(file)
		const key = "server/plugin/"
		if idx := strings.Index(file, key); idx != -1 {
			remain := file[idx+len(key):]
			parts := strings.Split(remain, "/")
			if len(parts) > 0 {
				pluginName = parts[0]
			}
		}
	}
	return pluginName
}

func RegisterApis(apis ...system.SysApi) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		ApiMap[name] = apis
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error { return registerApis(tx, apis) })
	if err != nil {
		zap.L().Error("注册API失败", zap.Error(err))
	}
}

type apiIdentity struct {
	path, method, group string
}

func registerApis(db *gorm.DB, apis []system.SysApi) error {
	if len(apis) == 0 {
		return nil
	}
	paths := make([]string, 0, len(apis))
	pathSet := make(map[string]struct{}, len(apis))
	for _, api := range apis {
		if _, exists := pathSet[api.Path]; !exists {
			pathSet[api.Path] = struct{}{}
			paths = append(paths, api.Path)
		}
	}

	var existing []system.SysApi
	if err := db.Where("path IN ?", paths).Find(&existing).Error; err != nil {
		return err
	}
	existingSet := make(map[apiIdentity]struct{}, len(existing))
	for _, api := range existing {
		existingSet[apiIdentity{path: api.Path, method: api.Method, group: api.ApiGroup}] = struct{}{}
	}
	missing := make([]system.SysApi, 0, len(apis))
	for _, api := range apis {
		identity := apiIdentity{path: api.Path, method: api.Method, group: api.ApiGroup}
		if _, exists := existingSet[identity]; exists {
			continue
		}
		existingSet[identity] = struct{}{}
		missing = append(missing, api)
	}
	if len(missing) == 0 {
		return nil
	}
	return db.Create(&missing).Error
}

func RegisterMenus(menus ...system.SysBaseMenu) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		MenuMap[name] = menus
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error { return registerMenus(tx, menus) })
	if err != nil {
		zap.L().Error("注册菜单失败", zap.Error(err))
	}
}

func registerMenus(db *gorm.DB, menus []system.SysBaseMenu) error {
	if len(menus) == 0 {
		return nil
	}
	parent := menus[0]
	if err := db.Where("name = ?", parent.Name).First(&parent).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "注册父菜单失败")
		}
		parent = menus[0]
		if err := db.Create(&parent).Error; err != nil {
			return errors.Wrap(err, "注册父菜单失败")
		}
	}
	if len(menus) == 1 {
		return nil
	}

	names := make([]string, 0, len(menus)-1)
	nameSet := make(map[string]struct{}, len(menus)-1)
	for i := 1; i < len(menus); i++ {
		menus[i].ParentId = parent.ID
		if _, exists := nameSet[menus[i].Name]; !exists {
			nameSet[menus[i].Name] = struct{}{}
			names = append(names, menus[i].Name)
		}
	}
	var existing []system.SysBaseMenu
	if err := db.Where("name IN ?", names).Find(&existing).Error; err != nil {
		return errors.Wrap(err, "查询插件菜单失败")
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, menu := range existing {
		existingSet[menu.Name] = struct{}{}
	}
	missing := make([]system.SysBaseMenu, 0, len(menus)-1)
	for i := 1; i < len(menus); i++ {
		if _, exists := existingSet[menus[i].Name]; exists {
			continue
		}
		existingSet[menus[i].Name] = struct{}{}
		missing = append(missing, menus[i])
	}
	if len(missing) == 0 {
		return nil
	}
	if err := db.Create(&missing).Error; err != nil {
		return errors.Wrap(err, "注册插件菜单失败")
	}
	return nil
}

func RegisterDictionaries(dictionaries ...system.SysDictionary) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		DictMap[name] = dictionaries
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error { return registerDictionaries(tx, dictionaries) })
	if err != nil {
		zap.L().Error("注册字典失败", zap.Error(err))
	}
}

type dictionaryDetailIdentity struct {
	dictionaryID int
	value        string
}

func registerDictionaries(db *gorm.DB, dictionaries []system.SysDictionary) error {
	if len(dictionaries) == 0 {
		return nil
	}
	types := make([]string, 0, len(dictionaries))
	typeSet := make(map[string]struct{}, len(dictionaries))
	detailsByType := make(map[string][]system.SysDictionaryDetail, len(dictionaries))
	definitions := make(map[string]system.SysDictionary, len(dictionaries))
	for _, dictionary := range dictionaries {
		detailsByType[dictionary.Type] = append(detailsByType[dictionary.Type], dictionary.SysDictionaryDetails...)
		if _, exists := typeSet[dictionary.Type]; exists {
			continue
		}
		typeSet[dictionary.Type] = struct{}{}
		types = append(types, dictionary.Type)
		dictionary.SysDictionaryDetails = nil
		definitions[dictionary.Type] = dictionary
	}

	var existing []system.SysDictionary
	if err := db.Where("type IN ?", types).Find(&existing).Error; err != nil {
		return errors.Wrap(err, "查询插件字典失败")
	}
	dictionariesByType := make(map[string]system.SysDictionary, len(types))
	for _, dictionary := range existing {
		dictionariesByType[dictionary.Type] = dictionary
	}
	missingDictionaries := make([]system.SysDictionary, 0, len(types))
	for _, dictionaryType := range types {
		if _, exists := dictionariesByType[dictionaryType]; !exists {
			missingDictionaries = append(missingDictionaries, definitions[dictionaryType])
		}
	}
	if len(missingDictionaries) > 0 {
		if err := db.Omit("Children", "SysDictionaryDetails").Create(&missingDictionaries).Error; err != nil {
			return errors.Wrap(err, "注册插件字典失败")
		}
		for _, dictionary := range missingDictionaries {
			dictionariesByType[dictionary.Type] = dictionary
		}
	}

	dictionaryIDs := make([]int, 0, len(dictionariesByType))
	for _, dictionary := range dictionariesByType {
		dictionaryIDs = append(dictionaryIDs, int(dictionary.ID))
	}
	var existingDetails []system.SysDictionaryDetail
	if err := db.Where("sys_dictionary_id IN ?", dictionaryIDs).Find(&existingDetails).Error; err != nil {
		return errors.Wrap(err, "查询插件字典详情失败")
	}
	existingDetailSet := make(map[dictionaryDetailIdentity]struct{}, len(existingDetails))
	for _, detail := range existingDetails {
		existingDetailSet[dictionaryDetailIdentity{dictionaryID: detail.SysDictionaryID, value: detail.Value}] = struct{}{}
	}
	missingDetails := make([]system.SysDictionaryDetail, 0)
	for _, dictionaryType := range types {
		dictionaryID := int(dictionariesByType[dictionaryType].ID)
		for _, detail := range detailsByType[dictionaryType] {
			identity := dictionaryDetailIdentity{dictionaryID: dictionaryID, value: detail.Value}
			if _, exists := existingDetailSet[identity]; exists {
				continue
			}
			existingDetailSet[identity] = struct{}{}
			detail.SysDictionaryID = dictionaryID
			missingDetails = append(missingDetails, detail)
		}
	}
	if len(missingDetails) == 0 {
		return nil
	}
	if err := db.Create(&missingDetails).Error; err != nil {
		return errors.Wrap(err, "注册插件字典详情失败")
	}
	return nil
}

func Pointer[T any](in T) *T {
	return &in
}

func GetPluginData(pluginName string) ([]system.SysApi, []system.SysBaseMenu, []system.SysDictionary) {
	rw.RLock()
	defer rw.RUnlock()
	return ApiMap[pluginName], MenuMap[pluginName], DictMap[pluginName]
}
