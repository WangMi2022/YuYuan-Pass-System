package response

import "github.com/WangMi2022/mit-assets-admin/server/model/example"

type ExaCustomerResponse struct {
	Customer example.ExaCustomer `json:"customer"`
}
