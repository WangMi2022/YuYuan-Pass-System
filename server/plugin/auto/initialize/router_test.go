package initialize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestInitializeRouterRegistersAutoCompatibilityPrefixes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	InitializeRouter(engine)

	foundAutoCode := false
	foundSkills := false
	for _, route := range engine.Routes() {
		if route.Path == "/autoCode/getDB" {
			foundAutoCode = true
		}
		if route.Path == "/skills/getTools" {
			foundSkills = true
		}
	}

	if !foundAutoCode {
		t.Fatalf("expected /autoCode/getDB to be registered by plugin router")
	}
	if !foundSkills {
		t.Fatalf("expected /skills/getTools to be registered by plugin router")
	}
}

func TestInitializeRouterProtectsAIAndInitializationEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousLogger := global.GVA_LOG
	global.GVA_LOG = zap.NewNop()
	t.Cleanup(func() { global.GVA_LOG = previousLogger })
	engine := gin.New()

	InitializeRouter(engine)

	for _, path := range []string{
		"/autoCode/llmAuto",
		"/autoCode/llmAutoSSE",
		"/autoCode/initMenu",
		"/autoCode/initAPI",
		"/autoCode/initDictionary",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, path, nil)
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("anonymous POST %s status = %d, want %d", path, response.Code, http.StatusUnauthorized)
			}
		})
	}
}
