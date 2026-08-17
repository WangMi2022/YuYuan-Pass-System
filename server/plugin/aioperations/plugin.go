package aioperations

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/aioperations/initialize"
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
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Permission(ctx)
	initialize.Router(engine)
}
