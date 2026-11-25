package api

import "github.com/WangMi2022/mit-assets-admin/server/plugin/aioperations/service"

var (
	Api               = new(apiGroup)
	serviceOperations = service.Service.Operations
)

type apiGroup struct{ Operations operationsAPI }
