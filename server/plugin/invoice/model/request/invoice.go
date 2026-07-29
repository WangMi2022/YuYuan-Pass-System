package request

import (
	"time"

	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/model"
)

type InvoiceSearch struct {
	commonRequest.PageInfo
	Status     string     `json:"status" form:"status"`
	CategoryID uint       `json:"categoryId" form:"categoryId"`
	Direction  string     `json:"direction" form:"direction"`
	StartDate  *time.Time `json:"startDate" form:"startDate" time_format:"2006-01-02"`
	EndDate    *time.Time `json:"endDate" form:"endDate" time_format:"2006-01-02"`
}

type InvoiceUpdate struct {
	ID            uint                `json:"ID"`
	Direction     string              `json:"direction"`
	InvoiceType   string              `json:"invoiceType"`
	InvoiceCode   string              `json:"invoiceCode"`
	InvoiceNumber string              `json:"invoiceNumber"`
	IssueDate     *time.Time          `json:"issueDate"`
	BuyerName     string              `json:"buyerName"`
	BuyerTaxNo    string              `json:"buyerTaxNo"`
	SellerName    string              `json:"sellerName"`
	SellerTaxNo   string              `json:"sellerTaxNo"`
	AmountCents   int64               `json:"amountCents"`
	TaxCents      int64               `json:"taxCents"`
	TotalCents    int64               `json:"totalCents"`
	CategoryID    *uint               `json:"categoryId"`
	ReviewNotes   string              `json:"reviewNotes"`
	Items         []model.InvoiceItem `json:"items"`
}

type CategorySearch struct {
	commonRequest.PageInfo
	Enabled *bool `json:"enabled" form:"enabled"`
}

type RuleSearch struct {
	commonRequest.PageInfo
	CategoryID uint  `json:"categoryId" form:"categoryId"`
	Enabled    *bool `json:"enabled" form:"enabled"`
}
