package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/document/service"

var (
	Api             = new(apiGroup)
	serviceDocument = service.Service.Document
)

type apiGroup struct {
	Document documentAPI
}
