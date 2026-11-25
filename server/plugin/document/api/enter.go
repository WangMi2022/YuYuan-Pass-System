package api

import "github.com/WangMi2022/mit-assets-admin/server/plugin/document/service"

var (
	Api             = new(apiGroup)
	serviceDocument = service.Service.Document
)

type apiGroup struct {
	Document documentAPI
}
