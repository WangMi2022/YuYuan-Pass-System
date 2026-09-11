package service

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt-in dialect/production reconciliation. No migrations, inserts or updates;
// PostgreSQL enforces READ ONLY on the transaction used by every query below.
func TestAssetQueryPostgresReadOnly(t *testing.T) {
	path := os.Getenv("SMART_QUERY_POSTGRES_CONFIG")
	if path == "" {
		t.Skip("set SMART_QUERY_POSTGRES_CONFIG to opt into read-only PostgreSQL checks")
	}
	reader := viper.New()
	reader.SetConfigFile(path)
	if reader.ReadInConfig() != nil {
		t.Fatal("read database configuration")
	}
	var settings struct {
		Pgsql config.Pgsql `mapstructure:"pgsql"`
	}
	if reader.Unmarshal(&settings) != nil {
		t.Fatal("decode database configuration")
	}
	db, err := gorm.Open(postgres.Open(settings.Pgsql.Dsn()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("connect configured PostgreSQL")
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal("open database connection")
	}
	defer connection.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tx := db.WithContext(ctx).Begin(&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if tx.Error != nil {
		t.Fatal("begin read-only database transaction")
	}
	defer tx.Rollback()
	previous := global.GVA_DB
	global.GVA_DB = tx
	defer func() { global.GVA_DB = previous }()
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	today := time.Now().In(location).Format("2006-01-02")
	result, err := Asset.Query(ctx, AssetQuery{Where: &AssetFilter{Field: "warrantyEndDate", Op: "lt", Value: today}, PageSize: 1})
	if err != nil {
		t.Fatal("execute expired warranty query on PostgreSQL")
	}
	var expected int64
	if tx.Raw("SELECT COUNT(*) FROM assets WHERE deleted_at IS NULL AND warranty_end_date < ?", today).Scan(&expected).Error != nil {
		t.Fatal("read independent warranty total")
	}
	if result.Total != expected {
		t.Fatalf("warranty count differs: got=%d expected=%d", result.Total, expected)
	}
	for _, query := range []AssetQuery{
		{GroupBy: "category", PageSize: 1},
		{GroupBy: "status", PageSize: 1},
		{Where: &AssetFilter{Field: "category", Op: "contains", Value: "办公"}, OrderBy: []AssetSort{{Field: "currentValue", Direction: "desc"}}},
		{Where: &AssetFilter{Field: "maintenanceCount", Op: "gte", Value: 0}, MaintenanceFrom: "2026-01-01", MaintenanceTo: today},
		{Where: &AssetFilter{All: []AssetFilter{{Field: "department", Op: "eq", Value: "行政部"}, {Any: []AssetFilter{{Field: "status", Op: "eq", Value: "in_use"}, {Field: "status", Op: "eq", Value: "idle"}}}}}},
	} {
		if _, err := Asset.Query(ctx, query); err != nil {
			t.Fatalf("PostgreSQL query failed (group=%s)", query.GroupBy)
		}
	}
	t.Logf("read-only PostgreSQL reconciliation passed; expired warranty records=%d", expected)
}
