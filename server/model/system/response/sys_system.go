package response

import "github.com/flipped-aurora/gin-vue-admin/server/config"

type SysConfigResponse struct {
	Config            config.Server   `json:"config"`
	ConfiguredSecrets map[string]bool `json:"configuredSecrets"`
}

type SysConfigSecretResponse struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}
