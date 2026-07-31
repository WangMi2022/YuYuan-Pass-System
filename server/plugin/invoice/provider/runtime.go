package provider

import (
	"sync/atomic"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
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
