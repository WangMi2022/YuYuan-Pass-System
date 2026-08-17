package api

import "github.com/WangMi2022/mit-assets-admin/server/plugin/site/service"

var (
	Api         = new(apiGroup)
	serviceSite = service.Service.Site
)

type apiGroup struct {
	Site siteAPI
}
