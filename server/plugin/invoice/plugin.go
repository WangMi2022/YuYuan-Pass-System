package invoice

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/provider"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/invoice/service"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (*plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	provider.SetRuntimeInvoiceRecognition(global.GVA_CONFIG.InvoiceRecognition)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Permission(ctx)
	initialize.Router(engine)
	service.Services.RecognitionService.StartWorker()
}
