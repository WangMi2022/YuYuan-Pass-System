package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/config"
	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/common/response"
	systemRes "github.com/WangMi2022/mit-assets-admin/server/model/system/response"
	redisCaptcha "github.com/WangMi2022/mit-assets-admin/server/utils/captcha"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/songzhibin97/gkit/cache/local_cache"
)

func TestCaptchaEndpointRequiresChallengeForEveryLoginMode(t *testing.T) {
	restore := replaceCaptchaTestRuntime(config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 3600,
	})
	defer restore()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/base/captcha", nil)
	new(BaseApi).Captcha(context)

	var body struct {
		Code int                          `json:"code"`
		Data systemRes.SysCaptchaResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != response.SUCCESS || !body.Data.OpenCaptcha {
		t.Fatalf("captcha response = %s", recorder.Body.String())
	}
	if body.Data.CaptchaId == "" || body.Data.PicPath == "" || body.Data.CaptchaLength != 4 {
		t.Fatalf("backend challenge is incomplete: %+v", body.Data)
	}
}

func TestCaptchaChallengeIsSingleUse(t *testing.T) {
	restore := replaceCaptchaTestRuntime(config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 3600,
	})
	defer restore()

	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	id, _, answer := driver.GenerateIdQuestionAnswer()
	if err := captchaStore().Set(id, answer); err != nil {
		t.Fatal(err)
	}
	if !verifyCaptcha(id, answer) {
		t.Fatal("valid backend challenge was rejected")
	}
	if verifyCaptcha(id, answer) {
		t.Fatal("captcha challenge was accepted more than once")
	}
}

func TestVerifyCaptchaRejectsEmptyChallenge(t *testing.T) {
	restore := replaceCaptchaTestRuntime(config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 3600,
	})
	defer restore()

	for _, challenge := range []struct {
		id     string
		answer string
	}{
		{},
		{id: "challenge-id"},
		{answer: "1234"},
		{id: "   ", answer: "   "},
	} {
		if verifyCaptcha(challenge.id, challenge.answer) {
			t.Fatalf("empty captcha challenge was accepted: %#v", challenge)
		}
	}
}

func TestCaptchaStoreUsesRedisWhenEnabled(t *testing.T) {
	restore := replaceCaptchaTestRuntime(config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 3600,
	})
	defer restore()

	currentRedis := global.GVA_REDIS
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer client.Close()
	defer func() { global.GVA_REDIS = currentRedis }()
	global.GVA_CONFIG.System.UseRedis = true
	global.GVA_REDIS = client

	if _, ok := captchaStore().(*redisCaptcha.RedisStore); !ok {
		t.Fatalf("captcha store = %T, want RedisStore", captchaStore())
	}
}

func TestCaptchaEndpointFailsClosedWhenRedisIsUnavailable(t *testing.T) {
	restore := replaceCaptchaTestRuntime(config.Captcha{
		KeyLong: 4, ImgWidth: 240, ImgHeight: 80,
		OpenCaptcha: 0, OpenCaptchaTimeOut: 3600,
	})
	defer restore()

	currentRedis := global.GVA_REDIS
	defer func() { global.GVA_REDIS = currentRedis }()
	global.GVA_CONFIG.System.UseRedis = true
	global.GVA_REDIS = nil

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/base/captcha", nil)
	new(BaseApi).Captcha(context)

	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code == response.SUCCESS {
		t.Fatalf("captcha response unexpectedly succeeded without Redis: %s", recorder.Body.String())
	}
}

func replaceCaptchaTestRuntime(captcha config.Captcha) func() {
	currentConfig := global.GVA_CONFIG
	currentCache := global.BlackCache
	currentMemoryStore := memoryCaptchaStore
	global.GVA_CONFIG.Captcha = captcha
	global.GVA_CONFIG.System.UseRedis = false
	global.BlackCache = local_cache.NewCache()
	memoryCaptchaStore = redisCaptcha.NewMemoryStore(captchaChallengeTTL)
	return func() {
		global.GVA_CONFIG = currentConfig
		global.BlackCache = currentCache
		memoryCaptchaStore = currentMemoryStore
	}
}
