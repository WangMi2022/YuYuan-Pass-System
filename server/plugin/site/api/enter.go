package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/site/service"

var (
	Api         = new(apiGroup)
	serviceSite = service.Service.Site
)

type apiGroup struct {
	Site siteAPI
}
