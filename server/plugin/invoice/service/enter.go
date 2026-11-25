package service

type ServiceGroup struct {
	InvoiceService
	CategoryService
	RuleService
	RecognitionService
	VerificationService
	QualityService
}

var Services = new(ServiceGroup)
