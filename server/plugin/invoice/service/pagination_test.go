package service

import (
	"fmt"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	commonRequest "github.com/WangMi2022/mit-assets-admin/server/model/common/request"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model"
	invoiceRequest "github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/model/request"
)

func TestInvoiceListCanPageTheUnconfirmedQueue(t *testing.T) {
	setupInvoiceServiceTestDB(t)
	for index := 0; index < 25; index++ {
		status := model.InvoiceStatusPendingReview
		if index < 5 {
			status = model.InvoiceStatusConfirmed
		}
		invoice := model.Invoice{
			Direction: "expense", Currency: "CNY", Status: status,
			FileName: fmt.Sprintf("invoice-%02d.pdf", index), FileKey: fmt.Sprintf("key-%02d", index),
			FileHash: fmt.Sprintf("hash-%02d", index), MimeType: "application/pdf", FileSize: 1,
			StorageType: "local", CreatedBy: 1, AuthorityID: 888,
		}
		if err := global.GVA_DB.Create(&invoice).Error; err != nil {
			t.Fatalf("create invoice %d: %v", index, err)
		}
	}

	search := invoiceRequest.InvoiceSearch{
		PageInfo:      commonRequest.PageInfo{Page: 1, PageSize: 10},
		ExcludeStatus: model.InvoiceStatusConfirmed,
	}
	list, total, err := (InvoiceService{}).List(search, AccessScope{All: true})
	if err != nil || total != 20 || len(list) != 10 {
		t.Fatalf("queue page = %d, total=%d, err=%v", len(list), total, err)
	}
	for _, item := range list {
		if item.Status == model.InvoiceStatusConfirmed {
			t.Fatalf("confirmed invoice %d leaked into recognition queue", item.ID)
		}
	}
}
