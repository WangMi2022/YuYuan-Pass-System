package service

var Service = new(serviceGroup)

type serviceGroup struct {
	Asset     assetService
	Category  categoryService
	Operation operationService
}
