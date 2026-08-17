package initialize

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/smart/service"
)

func Worker(_ context.Context) {
	service.StartReportWorker()
}
