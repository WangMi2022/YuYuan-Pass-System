package initialize

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGormMigratesCopilotRunAndKnowledgeChunk(t *testing.T) {
	previousDB := global.GVA_DB
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	global.GVA_DB = database
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		if sqlDB, dbErr := database.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	Gorm(context.Background())
	if !database.Migrator().HasTable(&model.CopilotRun{}) || !database.Migrator().HasTable(&model.KnowledgeChunk{}) {
		t.Fatal("smart plugin migration did not create observability and knowledge tables")
	}
}
