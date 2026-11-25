package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/aioperations/api"

var (
	Router        = new(routerGroup)
	apiOperations = api.Api.Operations
)

type routerGroup struct{ Operations operationsRouter }
