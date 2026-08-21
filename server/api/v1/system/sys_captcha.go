package system

import (
	"context"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	systemRes "github.com/WangMi2022/mit-assets-admin/server/model/system/response"
	redisCaptcha "github.com/WangMi2022/mit-assets-admin/server/utils/captcha"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

const captchaChallengeTTL = redisCaptcha.DefaultExpiration

// The memory store keeps local development and single-process installations
// functional. Production instances with Redis enabled use the shared store.
var memoryCaptchaStore base64Captcha.Store = redisCaptcha.NewMemoryStore(captchaChallengeTTL)

type BaseApi struct{}

// Captcha
// @Tags      Base
// @Summary   生成验证码
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{data=systemRes.SysCaptchaResponse,msg=string}  "生成验证码,返回包括随机数id,base64,验证码长度,是否开启验证码"
// @Router    /base/captcha [post]
func (b *BaseApi) Captcha(c *gin.Context) {
	captchaConfig := currentCaptchaConfig()
	failures := captchaFailureCount(c.ClientIP(), captchaConfig.OpenCaptchaTimeOut)
	required := captchaConfig.Required(failures)

	driver := base64Captcha.NewDriverDigit(
		captchaConfig.ImgHeight,
		captchaConfig.ImgWidth,
		captchaConfig.KeyLong,
		0.7,
		80,
	)
	cp := base64Captcha.NewCaptcha(driver, captchaStoreWithContext(c.Request.Context()))
	id, b64s, _, err := cp.Generate()
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("验证码获取失败!", zap.Error(err))
		}
		response.FailWithMessage("验证码获取失败", c)
		return
	}
	response.OkWithDetailed(systemRes.SysCaptchaResponse{
		CaptchaId:     id,
		PicPath:       b64s,
		CaptchaLength: captchaConfig.KeyLong,
		OpenCaptcha:   required,
	}, "验证码获取成功", c)
}

func currentCaptchaConfig() config.Captcha {
	global.GVA_CONFIG_LOCK.Lock()
	captchaConfig := global.GVA_CONFIG.Captcha
	global.GVA_CONFIG_LOCK.Unlock()
	captchaConfig.Normalize()
	return captchaConfig
}

func captchaFailureCount(ip string, timeoutSeconds int) int {
	value, ok := global.BlackCache.Get(ip)
	if !ok {
		global.BlackCache.Set(ip, 1, time.Second*time.Duration(timeoutSeconds))
		return 1
	}
	return interfaceToInt(value)
}

func captchaStoreWithContext(ctx context.Context) base64Captcha.Store {
	global.GVA_CONFIG_LOCK.Lock()
	useRedis := global.GVA_CONFIG.System.UseRedis
	global.GVA_CONFIG_LOCK.Unlock()
	if useRedis {
		return redisCaptcha.NewDefaultRedisStore().UseWithCtx(ctx)
	}
	return memoryCaptchaStore
}

func captchaStore() base64Captcha.Store {
	return captchaStoreWithContext(context.Background())
}

func verifyCaptcha(id, answer string) bool {
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if id == "" || answer == "" {
		return false
	}
	return captchaStore().Verify(id, answer, true)
}

// 类型转换
func interfaceToInt(v interface{}) (i int) {
	switch v := v.(type) {
	case int:
		i = v
	default:
		i = 0
	}
	return
}
