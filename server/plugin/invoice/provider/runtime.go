package provider

import (
	"sync/atomic"

	"github.com/WangMi2022/mit-assets-admin/server/config"
)

var runtimeInvoiceRecognition atomic.Pointer[config.InvoiceRecognition]

func SetRuntimeInvoiceRecognition(configuration config.InvoiceRecognition) {
	snapshot := configuration
	snapshot.Normalize()
	runtimeInvoiceRecognition.Store(&snapshot)
}

func RuntimeInvoiceRecognition() config.InvoiceRecognition {
	snapshot := runtimeInvoiceRecognition.Load()
	if snapshot == nil {
		return config.InvoiceRecognition{}
	}
	return *snapshot
}
