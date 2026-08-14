package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/service"

var (
	Api                = new(apiGroup)
	serviceAsset       = service.Service.Asset
	serviceCategory    = service.Service.Category
	serviceLocation    = service.Service.Location
	serviceOperation   = service.Service.Operation
	serviceRisk        = service.Service.Risk
	serviceRecognition = service.Service.Recognition
)

type apiGroup struct {
	Asset       assetAPI
	Category    categoryAPI
	Location    locationAPI
	Operation   operationAPI
	Risk        riskAPI
	Recognition recognitionAPI
}
