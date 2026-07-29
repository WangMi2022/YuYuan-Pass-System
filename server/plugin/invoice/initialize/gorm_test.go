package initialize

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBackfillInvoiceStorageLocations(t *testing.T) {
	previousDB := global.GVA_DB
	previousConfig := global.GVA_CONFIG
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err = db.AutoMigrate(&model.Invoice{}); err != nil {
		t.Fatalf("migrate invoice table: %v", err)
	}
	global.GVA_DB = db
	global.GVA_CONFIG.Local.StorePath = filepath.Join(t.TempDir(), "invoices")
	global.GVA_CONFIG.Minio.Endpoint = "minio.internal:9000"
	global.GVA_CONFIG.Minio.BucketName = "private-invoices"
	global.GVA_CONFIG.Minio.UseSSL = true
	t.Cleanup(func() {
		global.GVA_DB = previousDB
		global.GVA_CONFIG = previousConfig
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	invoices := []model.Invoice{
		{FileName: "local.png", FileKey: "local.png", FileHash: "local", MimeType: "image/png", FileSize: 1, StorageType: "local", CreatedBy: 1, AuthorityID: 1},
		{FileName: "minio.png", FileKey: "minio.png", FileHash: "minio", MimeType: "image/png", FileSize: 1, StorageType: "minio", CreatedBy: 1, AuthorityID: 1},
	}
	if err = db.Create(&invoices).Error; err != nil {
		t.Fatalf("create legacy invoices: %v", err)
	}
	if err = backfillInvoiceStorageLocations(context.Background()); err != nil {
		t.Fatalf("backfill storage locations: %v", err)
	}

	var localInvoice, minioInvoice model.Invoice
	if err = db.First(&localInvoice, invoices[0].ID).Error; err != nil {
		t.Fatalf("reload local invoice: %v", err)
	}
	if err = db.First(&minioInvoice, invoices[1].ID).Error; err != nil {
		t.Fatalf("reload minio invoice: %v", err)
	}
	expectedRoot, _ := filepath.Abs(global.GVA_CONFIG.Local.StorePath)
	if localInvoice.StorageRoot != expectedRoot {
		t.Fatalf("local root was not backfilled: %q", localInvoice.StorageRoot)
	}
	if minioInvoice.StorageEndpoint != "minio.internal:9000" || minioInvoice.StorageBucket != "private-invoices" || !minioInvoice.StorageUseSSL {
		t.Fatalf("minio location was not backfilled: %#v", minioInvoice)
	}
}
