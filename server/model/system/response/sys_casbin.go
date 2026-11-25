package response

import (
	"github.com/WangMi2022/mit-assets-admin/server/model/system/request"
)

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}
