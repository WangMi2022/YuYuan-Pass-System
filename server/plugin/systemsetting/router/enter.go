package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/systemsetting/api"

var (
	Router             = new(routerGroup)
	apiLoginBackground = api.Api.LoginBackground
	apiLoginLogo       = api.Api.LoginLogo
)

type routerGroup struct {
	LoginBackground loginBackgroundRouter
	LoginLogo       loginLogoRouter
}
