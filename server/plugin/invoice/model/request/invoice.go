package request

import (
	"time"

	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
)

type InvoiceSearch struct {
	commonRequest.PageInfo
	Status        string     `json:"status" form:"status"`
	ExcludeStatus string     `json:"excludeStatus" form:"excludeStatus"`
	CategoryID    uint       `json:"categoryId" form:"categoryId"`
	Direction     string     `json:"direction" form:"direction"`
	StartDate     *time.Time `json:"startDate" form:"startDate" time_format:"2006-01-02"`
	EndDate       *time.Time `json:"endDate" form:"endDate" time_format:"2006-01-02"`
}

// Normalize removes zero-value dates produced by Gin when optional query
// parameters are present with an empty value (for example, startDate=).
func (search *InvoiceSearch) Normalize() {
	if search.StartDate != nil && search.StartDate.IsZero() {
		search.StartDate = nil
	}
	if search.EndDate != nil && search.EndDate.IsZero() {
		search.EndDate = nil
	}
}

type InvoiceUpdate struct {
	ID                     uint                `json:"ID"`
	Direction              string              `json:"direction"`
	InvoiceType            string              `json:"invoiceType"`
	VerificationType       string              `json:"verificationType"`
	VerificationAmountMode string              `json:"verificationAmountMode"`
	InvoiceCode            string              `json:"invoiceCode"`
	InvoiceNumber          string              `json:"invoiceNumber"`
	CheckCode              string              `json:"checkCode"`
	IssueDate              *time.Time          `json:"issueDate"`
	BuyerName              string              `json:"buyerName"`
	BuyerTaxNo             string              `json:"buyerTaxNo"`
	SellerName             string              `json:"sellerName"`
	SellerTaxNo            string              `json:"sellerTaxNo"`
	AmountCents            int64               `json:"amountCents"`
	TaxCents               int64               `json:"taxCents"`
	TotalCents             int64               `json:"totalCents"`
	CategoryID             *uint               `json:"categoryId"`
	ReviewNotes            string              `json:"reviewNotes"`
	Items                  []model.InvoiceItem `json:"items"`
}

type InvoiceConfirm struct {
	ID                       uint   `json:"id"`
	VerificationBypass       bool   `json:"verificationBypass"`
	VerificationBypassReason string `json:"verificationBypassReason"`
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
