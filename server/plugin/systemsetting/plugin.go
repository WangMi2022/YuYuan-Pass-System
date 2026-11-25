package systemsetting

import (
	"context"

	"github.com/WangMi2022/mit-assets-admin/server/plugin/systemsetting/initialize"
	interfaces "github.com/WangMi2022/mit-assets-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
)

var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Gorm(ctx)
	initialize.Permission(ctx)
	initialize.Router(engine)
}
