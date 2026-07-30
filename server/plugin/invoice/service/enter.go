package service

type ServiceGroup struct {
	InvoiceService
	CategoryService
	RuleService
	RecognitionService
	VerificationService
}

var Services = new(ServiceGroup)
