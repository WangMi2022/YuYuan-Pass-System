package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/api"

var (
	Router      = new(routerGroup)
	apiDocument = api.Api.Document
)

type routerGroup struct {
	Document documentRouter
}
