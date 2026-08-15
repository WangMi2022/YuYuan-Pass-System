package smart

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/smart/initialize"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)
var Plugin = new(plugin)

func init() { interfaces.Register(Plugin) }

type plugin struct{}

func (*plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Permission(ctx)
	initialize.Router(engine)
	initialize.Worker(ctx)
}
