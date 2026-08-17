package main

import (
	"context"
	"os"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/core"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/initialize"
	systemService "github.com/WangMi2022/mit-assets-admin/server/service/system"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

// 这部分 @Tag 设置用于排序, 需要排序的接口请按照下面的格式添加
// swag init 对 @Tag 只会从入口文件解析, 默认 main.go
// 也可通过 --generalInfo flag 指定其他文件
// @Tag.Name        Base
// @Tag.Name        SysUser
// @Tag.Description 用户

// @title                       mit-assets-admin Swagger API 接口文档
// @version                     v2.9.2
// @description                 MIT 资产管理系统 API
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-token
// @BasePath                    /
func main() {
	// 初始化系统
	initializeSystem()
	// 运行服务器
	core.RunServer()
}

// initializeSystem 初始化系统所有组件
// 提取为单独函数以便于系统重载时调用
func initializeSystem() {
	global.GVA_VP = core.Viper() // 初始化Viper
	initialize.OtherInit()
	global.GVA_LOG = core.Zap() // 初始化zap日志库
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm() // gorm连接数据库
	if signingKey := strings.TrimSpace(os.Getenv("GVA_JWT_SIGNING_KEY")); signingKey != "" {
		global.GVA_CONFIG.JWT.SigningKey = signingKey
	}
	if global.GVA_DB != nil && utils.IsWeakJWTSigningKey(global.GVA_CONFIG.JWT.SigningKey) {
		panic("生产数据库已初始化，但 JWT signing-key 仍为默认弱密钥；请先通过安全配置注入随机密钥")
	}
	initialize.Timer()
	initialize.DBList()
	initialize.SetupHandlers() // 注册全局函数
	if global.GVA_DB != nil {
		initialize.RegisterTables() // 初始化表
		if err := systemService.CasbinServiceApp.EnsureCoreAdminPermissions(context.Background(), global.GVA_DB); err != nil {
			global.GVA_LOG.Error("repair core administrator permissions failed", zap.Error(err))
		}
	}
}
