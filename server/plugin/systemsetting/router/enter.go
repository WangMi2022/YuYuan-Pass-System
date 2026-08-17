package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/api"

var (
	Router             = new(routerGroup)
	apiLoginBackground = api.Api.LoginBackground
	apiLoginLogo       = api.Api.LoginLogo
)

type routerGroup struct {
	LoginBackground loginBackgroundRouter
	LoginLogo       loginLogoRouter
}
