package initialize

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/example"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		&example.ExaFile{},
	)
	if err != nil {
		return err
	}
	return nil
}
