package system

import (
	"github.com/WangMi2022/mit-assets-admin/server/config"
)

// 配置文件结构体
type System struct {
	Config config.Server `json:"config"`
}

type GetSystemConfigRequest struct {
	SecretPath string `json:"secretPath"`
}
