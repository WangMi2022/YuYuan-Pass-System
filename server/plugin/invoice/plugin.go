package invoice

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/initialize"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/provider"
	"github.com/WangMi2022/mit-assets-admin/server/plugin/invoice/service"
	interfaces "github.com/WangMi2022/mit-assets-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (*plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	provider.SetRuntimeInvoiceRecognition(global.GVA_CONFIG.AI.Invoice)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Permission(ctx)
	initialize.Router(engine)
	service.Services.RecognitionService.StartWorker()
}
