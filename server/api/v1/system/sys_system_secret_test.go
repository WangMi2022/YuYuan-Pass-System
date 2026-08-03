package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
)

func TestGetSystemConfigRedactsSecretsByDefault(t *testing.T) {
	restore := replaceSystemSecretTestConfig()
	defer restore()

	response := performSystemConfigRequest(t, 888, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, "top-secret") {
		t.Fatalf("default config response leaked a secret: %s", body)
	}
	if !strings.Contains(body, `"configuredSecrets":{"qiniu.secret-key":true`) {
		t.Fatalf("configured state is missing: %s", body)
	}
}

func TestGetSystemConfigSecretRequiresSuperAdmin(t *testing.T) {
	restore := replaceSystemSecretTestConfig()
	defer restore()

	response := performSystemConfigRequest(t, 8881, `{"secretPath":"qiniu.secret-key"}`)
	body := response.Body.String()
	if strings.Contains(body, "top-secret") || !strings.Contains(body, "仅超级管理员") {
		t.Fatalf("non-super-admin reveal response = %s", body)
	}
}

func TestGetSystemConfigSecretRevealsWhitelistedValue(t *testing.T) {
	restore := replaceSystemSecretTestConfig()
	defer restore()

	response := performSystemConfigRequest(t, 888, `{"secretPath":"qiniu.secret-key"}`)
	body := response.Body.String()
	if !strings.Contains(body, `"value":"top-secret"`) {
		t.Fatalf("super-admin reveal response = %s", body)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("reveal cache control = %q", response.Header().Get("Cache-Control"))
	}
}

func performSystemConfigRequest(t *testing.T, authorityID uint, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/system/getSystemConfig", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("claims", &systemReq.CustomClaims{BaseClaims: systemReq.BaseClaims{AuthorityId: authorityID}})
	new(SystemApi).GetSystemConfig(context)
	return recorder
}

func replaceSystemSecretTestConfig() func() {
	current := global.GVA_CONFIG
	global.GVA_CONFIG = config.Server{Qiniu: config.Qiniu{SecretKey: "top-secret"}}
	return func() { global.GVA_CONFIG = current }
}
