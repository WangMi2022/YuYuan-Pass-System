package response

import "github.com/WangMi2022/mit-assets-admin/server/config"

type SysConfigResponse struct {
	Config            config.Server   `json:"config"`
	ConfiguredSecrets map[string]bool `json:"configuredSecrets"`
}

type SysConfigSecretResponse struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}
