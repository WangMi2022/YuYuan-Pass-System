package initialize

import (
	"os"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/example"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.MAA_CONFIG.System.DbType {
	case "mysql":
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Mysql.Dbname
		return GormMysql()
	case "pgsql":
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Pgsql.Dbname
		return GormPgSql()
	case "oracle":
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Oracle.Dbname
		return GormOracle()
	case "mssql":
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Mssql.Dbname
		return GormMssql()
	case "sqlite":
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Sqlite.Dbname
		return GormSqlite()
	default:
		global.MAA_ACTIVE_DBNAME = &global.MAA_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

func RegisterTables() {
	if global.MAA_CONFIG.System.DisableAutoMigrate {
		global.MAA_LOG.Info("auto-migrate is disabled, skipping table registration")
		return
	}

	db := global.MAA_DB
	err := db.AutoMigrate(

		system.SysApi{},
		system.SysIgnoreApi{},
		system.SysUser{},
		system.SysBaseMenu{},
		system.JwtBlacklist{},
		system.SysAuthority{},
		system.SysDictionary{},
		system.SysOperationRecord{},
		system.SysAutoCodeHistory{},
		system.SysDictionaryDetail{},
		system.SysBaseMenuParameter{},
		system.SysBaseMenuBtn{},
		system.SysAuthorityBtn{},
		system.SysAutoCodePackage{},
		system.SysExportTemplate{},
		system.Condition{},
		system.JoinTemplate{},
		system.SysParams{},
		system.SysVersion{},
		system.SysError{},
		system.SysApiToken{},
		system.SysLoginLog{},

		example.ExaFile{},
		example.ExaCustomer{},
		example.ExaFileChunk{},
		example.ExaFileUploadAndDownload{},
		example.ExaAttachmentCategory{},
	)
	if err != nil {
		global.MAA_LOG.Error("register table failed", zap.Error(err))
		os.Exit(1)
	}

	err = bizModel()

	if err != nil {
		global.MAA_LOG.Error("register biz_table failed", zap.Error(err))
		os.Exit(1)
	}
	global.MAA_LOG.Info("register table success")
}
