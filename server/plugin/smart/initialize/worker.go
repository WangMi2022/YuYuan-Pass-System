package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/smart/service"
)

func Worker(_ context.Context) {
	service.StartReportWorker()
}
