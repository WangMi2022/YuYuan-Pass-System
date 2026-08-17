package router

import (
	"github.com/WangMi2022/mit-assets-admin/server/router/example"
	"github.com/WangMi2022/mit-assets-admin/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System  system.RouterGroup
	Example example.RouterGroup
}
