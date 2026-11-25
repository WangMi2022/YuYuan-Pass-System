package api

import "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/service"

var (
	Api                 = new(apiGroup)
	serviceInvoice      = service.Services.InvoiceService
	serviceCategory     = service.Services.CategoryService
	serviceRule         = service.Services.RuleService
	serviceRecognition  = service.Services.RecognitionService
	serviceVerification = service.Services.VerificationService
	serviceQuality      = service.Services.QualityService
)

type apiGroup struct {
	Invoice  invoiceAPI
	Category categoryAPI
	Rule     ruleAPI
	Quality  qualityAPI
}
