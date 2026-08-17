package example

import "github.com/WangMi2022/mit-assets-admin/server/service"

type ApiGroup struct {
	CustomerApi

	AttachmentCategoryApi
	FileUploadAndDownloadApi
}

var (
	customerService = service.ServiceGroupApp.ExampleServiceGroup.CustomerService

	attachmentCategoryService    = service.ServiceGroupApp.ExampleServiceGroup.AttachmentCategoryService
	fileUploadAndDownloadService = service.ServiceGroupApp.ExampleServiceGroup.FileUploadAndDownloadService
)
