package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/document/api"

var (
	Router      = new(routerGroup)
	apiDocument = api.Api.Document
)

type routerGroup struct {
	Document documentRouter
}
