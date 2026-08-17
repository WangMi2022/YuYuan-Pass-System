package initialize

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate()
	if err != nil {
		return err
	}
	return nil
}
