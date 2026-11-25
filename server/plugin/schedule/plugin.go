package schedule

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/initialize"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/schedule/service"
	interfaces "github.com/WangMi2022/mit-assets-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (*plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Permission(ctx)
	initialize.Gorm(ctx)
	initialize.Router(engine)
	service.WorkSchedule.StartReminderWorker()
}
