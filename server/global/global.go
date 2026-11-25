package global

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"

	"github.com/WangMi2022/mit-assets-admin/server/utils/timer"
	"github.com/songzhibin97/gkit/cache/local_cache"

	"golang.org/x/sync/singleflight"

	"go.uber.org/zap"

	"github.com/WangMi2022/mit-assets-admin/server/config"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	MAA_DB          *gorm.DB
	MAA_DBList      map[string]*gorm.DB
	MAA_REDIS       redis.UniversalClient
	MAA_REDISList   map[string]redis.UniversalClient
	MAA_MONGO       *qmgo.QmgoClient
	MAA_CONFIG      config.Server
	MAA_CONFIG_LOCK sync.Mutex
	MAA_VP          *viper.Viper
	// MAA_LOG    *oplogging.Logger
	MAA_LOG                 *zap.Logger
	MAA_Timer               timer.Timer = timer.NewTimerTask()
	MAA_Concurrency_Control             = &singleflight.Group{}
	MAA_ROUTERS             gin.RoutesInfo
	MAA_ACTIVE_DBNAME       *string
	MAA_MCP_SERVER          *server.MCPServer
	BlackCache              local_cache.Cache
	lock                    sync.RWMutex
)

// GetGlobalDBByDBName 通过名称获取db list中的db
func GetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	return MAA_DBList[dbname]
}

// MustGetGlobalDBByDBName 通过名称获取db 如果不存在则panic
func MustGetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := MAA_DBList[dbname]
	if !ok || db == nil {
		panic("db no init")
	}
	return db
}

func GetRedis(name string) redis.UniversalClient {
	redis, ok := MAA_REDISList[name]
	if !ok || redis == nil {
		panic(fmt.Sprintf("redis `%s` no init", name))
	}
	return redis
}
