package mcpTool

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WangMi2022/mit-assets-admin/server/global"
)

func TestRequireMCPAuth(t *testing.T) {
	previous := global.GVA_CONFIG.MCP.AuthToken
	global.GVA_CONFIG.MCP.AuthToken = "test-mcp-token"
	t.Cleanup(func() {
		global.GVA_CONFIG.MCP.AuthToken = previous
	})

	handler := requireMCPAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	authorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	request.Header.Set("X-Token", "test-mcp-token")
	handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusNoContent {
		t.Fatalf("authorized status = %d, want %d", authorized.Code, http.StatusNoContent)
	}
}
