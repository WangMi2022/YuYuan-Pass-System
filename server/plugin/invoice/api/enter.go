package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/service"

var (
	Api                 = new(apiGroup)
	serviceInvoice      = service.Services.InvoiceService
	serviceCategory     = service.Services.CategoryService
	serviceRule         = service.Services.RuleService
	serviceRecognition  = service.Services.RecognitionService
	serviceVerification = service.Services.VerificationService
)

type apiGroup struct {
	Invoice  invoiceAPI
	Category categoryAPI
	Rule     ruleAPI
}
