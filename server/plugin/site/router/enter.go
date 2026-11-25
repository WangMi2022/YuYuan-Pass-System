package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/site/api"

var (
	Router  = new(routerGroup)
	apiSite = api.Api.Site
)

type routerGroup struct {
	Site siteRouter
}
