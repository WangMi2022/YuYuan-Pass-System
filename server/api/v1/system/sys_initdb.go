package system

import (
	"crypto/subtle"
	"net"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type DBApi struct{}

var (
	initDBMu             sync.Mutex
	installTokenConsumed bool
)

func allowDatabaseInitialization(c *gin.Context) bool {
	if global.GVA_DB != nil {
		response.FailWithMessage("已存在数据库配置", c)
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("GVA_INSTALL_MODE")), "true") {
		response.FailWithMessage("初始化未授权，请在本机显式启用安装模式", c)
		return false
	}
	if installTokenConsumed {
		response.FailWithMessage("安装令牌已使用", c)
		return false
	}
	token := strings.TrimSpace(os.Getenv("GVA_INSTALL_TOKEN"))
	supplied := strings.TrimSpace(c.GetHeader("X-Install-Token"))
	remoteHost, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		remoteHost = strings.TrimSpace(c.Request.RemoteAddr)
	}
	remoteIP := net.ParseIP(remoteHost)
	if token == "" || remoteIP == nil || !remoteIP.IsLoopback() || len(token) != len(supplied) || subtle.ConstantTimeCompare([]byte(token), []byte(supplied)) != 1 {
		response.FailWithMessage("初始化未授权，请使用本机一次性安装令牌", c)
		return false
	}
	return true
}

// InitDB
// @Tags     InitDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Param    data  body      request.InitDB                  true  "初始化数据库参数"
// @Success  200   {object}  response.Response{data=string}  "初始化用户数据库"
// @Router   /init/initdb [post]
func (i *DBApi) InitDB(c *gin.Context) {
	initDBMu.Lock()
	defer initDBMu.Unlock()
	if !allowDatabaseInitialization(c) {
		return
	}
	if utils.IsWeakJWTSigningKey(global.GVA_CONFIG.JWT.SigningKey) {
		response.FailWithMessage("初始化未授权，请先配置至少 32 字节的 JWT signing-key", c)
		return
	}
	if global.GVA_DB != nil {
		global.GVA_LOG.Error("已存在数据库配置!")
		response.FailWithMessage("已存在数据库配置", c)
		return
	}
	var dbInfo request.InitDB
	if err := c.ShouldBindJSON(&dbInfo); err != nil {
		global.GVA_LOG.Error("参数校验不通过!", zap.Error(err))
		response.FailWithMessage("参数校验不通过", c)
		return
	}
	if utf8.RuneCountInString(strings.TrimSpace(dbInfo.AdminPassword)) < 12 {
		response.FailWithMessage("管理员密码至少需要 12 个字符", c)
		return
	}
	if err := initDBService.InitDB(dbInfo); err != nil {
		global.GVA_LOG.Error("自动创建数据库失败!", zap.Error(err))
		response.FailWithMessage("自动创建数据库失败，请查看后台日志，检查后在进行初始化", c)
		return
	}
	installTokenConsumed = true
	response.OkWithMessage("自动创建数据库成功", c)
}

// CheckDB
// @Tags     CheckDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Success  200  {object}  response.Response{data=map[string]interface{},msg=string}  "初始化用户数据库"
// @Router   /init/checkdb [post]
func (i *DBApi) CheckDB(c *gin.Context) {
	var (
		message  = "前往初始化数据库"
		needInit = true
	)

	if global.GVA_DB != nil {
		message = "数据库无需初始化"
		needInit = false
	}
	global.GVA_LOG.Info(message)
	response.OkWithDetailed(gin.H{"needInit": needInit}, message, c)
}
