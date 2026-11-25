package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/api"

var (
	Router      = new(routerGroup)
	apiInvoice  = api.Api.Invoice
	apiCategory = api.Api.Category
	apiRule     = api.Api.Rule
	apiQuality  = api.Api.Quality
)

type routerGroup struct {
	Invoice invoiceRouter
}
