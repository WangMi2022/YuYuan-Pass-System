package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/site/api"

var (
	Router  = new(routerGroup)
	apiSite = api.Api.Site
)

type routerGroup struct {
	Site siteRouter
}
