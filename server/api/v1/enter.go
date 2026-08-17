package v1

import (
	"github.com/WangMi2022/mit-assets-admin/server/api/v1/example"
	"github.com/WangMi2022/mit-assets-admin/server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	ExampleApiGroup example.ApiGroup
}
