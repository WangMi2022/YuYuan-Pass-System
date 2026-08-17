package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/asset/api"

var (
	Router         = new(routerGroup)
	apiAsset       = api.Api.Asset
	apiCategory    = api.Api.Category
	apiLocation    = api.Api.Location
	apiOperation   = api.Api.Operation
	apiRisk        = api.Api.Risk
	apiRecognition = api.Api.Recognition
)

type routerGroup struct {
	Asset       assetRouter
	Category    categoryRouter
	Location    locationRouter
	Operation   operationRouter
	Risk        riskRouter
	Recognition recognitionRouter
}
