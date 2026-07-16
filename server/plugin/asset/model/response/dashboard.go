package response

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/asset/model"

type CategorySummary struct {
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Color        string  `json:"color"`
	AssetKinds   int64   `json:"assetKinds"`
	Quantity     int64   `json:"quantity"`
	Original     float64 `json:"originalValue"`
	Current      float64 `json:"currentValue"`
}

type StatusSummary struct {
	Status     string `json:"status"`
	AssetKinds int64  `json:"assetKinds"`
	Quantity   int64  `json:"quantity"`
}

type LocationSummary struct {
	Location string  `json:"location"`
	Quantity int64   `json:"quantity"`
	Value    float64 `json:"value"`
}

type CategoryItem struct {
	model.Category
	AssetKinds int64 `json:"assetKinds"`
	Quantity   int64 `json:"quantity"`
}

type Dashboard struct {
	AssetKinds      int64             `json:"assetKinds"`
	TotalQuantity   int64             `json:"totalQuantity"`
	CategoryCount   int64             `json:"categoryCount"`
	OriginalValue   float64           `json:"originalValue"`
	CurrentValue    float64           `json:"currentValue"`
	Depreciation    float64           `json:"depreciation"`
	CategorySummary []CategorySummary `json:"categorySummary"`
	StatusSummary   []StatusSummary   `json:"statusSummary"`
	LocationSummary []LocationSummary `json:"locationSummary"`
	RecentAssets    []model.Asset     `json:"recentAssets"`
}
