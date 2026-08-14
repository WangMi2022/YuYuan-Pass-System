package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/aioperations/service"

var (
	Api               = new(apiGroup)
	serviceOperations = service.Service.Operations
)

type apiGroup struct{ Operations operationsAPI }
