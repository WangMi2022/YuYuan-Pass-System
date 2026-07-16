package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/api"

var (
	Router       = new(routerGroup)
	apiAsset     = api.Api.Asset
	apiCategory  = api.Api.Category
	apiLocation  = api.Api.Location
	apiOperation = api.Api.Operation
)

type routerGroup struct {
	Asset     assetRouter
	Category  categoryRouter
	Location  locationRouter
	Operation operationRouter
}
