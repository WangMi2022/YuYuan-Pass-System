package service

type ServiceGroup struct {
	InvoiceService
	CategoryService
	RuleService
	RecognitionService
}

var Services = new(ServiceGroup)
