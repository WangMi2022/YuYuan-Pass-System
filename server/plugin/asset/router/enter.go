package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/api"

var (
	Router       = new(routerGroup)
	apiAsset     = api.Api.Asset
	apiCategory  = api.Api.Category
	apiOperation = api.Api.Operation
)

type routerGroup struct {
	Asset     assetRouter
	Category  categoryRouter
	Operation operationRouter
}
