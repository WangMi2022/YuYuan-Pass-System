package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/aioperations/api"

var (
	Router        = new(routerGroup)
	apiOperations = api.Api.Operations
)

type routerGroup struct{ Operations operationsRouter }
