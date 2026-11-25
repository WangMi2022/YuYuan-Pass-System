package ast

import (
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"path/filepath"
)

func init() {
	global.MAA_CONFIG.AutoCode.Root, _ = filepath.Abs("../../../")
	global.MAA_CONFIG.AutoCode.Server = "server"
}
