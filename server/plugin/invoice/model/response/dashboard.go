package response

import "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"

type Dashboard struct {
	ConfirmedCount int64             `json:"confirmedCount"`
	PendingCount   int64             `json:"pendingCount"`
	FailedCount    int64             `json:"failedCount"`
	TotalCents     int64             `json:"totalCents"`
	AmountCents    int64             `json:"amountCents"`
	TaxCents       int64             `json:"taxCents"`
	MonthlyTrend   []MonthlySummary  `json:"monthlyTrend"`
	Categories     []CategorySummary `json:"categories"`
	Suppliers      []SupplierSummary `json:"suppliers"`
	Recent         []model.Invoice   `json:"recent"`
}

type MonthlySummary struct {
	Month      string `json:"month"`
	Count      int64  `json:"count"`
	TotalCents int64  `json:"totalCents"`
}

type CategorySummary struct {
	CategoryID uint   `json:"categoryId"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Count      int64  `json:"count"`
	TotalCents int64  `json:"totalCents"`
}

type SupplierSummary struct {
	Name       string `json:"name"`
	Count      int64  `json:"count"`
	TotalCents int64  `json:"totalCents"`
}
