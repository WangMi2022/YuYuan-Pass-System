package mcpTool

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	mcpServer "github.com/mark3labs/mcp-go/server"
)

func NewMCPServer() *mcpServer.MCPServer {
	config := global.GVA_CONFIG.MCP

	s := mcpServer.NewMCPServer(
		config.Name,
		config.Version,
	)

	global.GVA_MCP_SERVER = s
	RegisterAllTools(s)

	return s
}

func NewStreamableHTTPServer() *mcpServer.StreamableHTTPServer {
	config := global.GVA_CONFIG.MCP
	path := config.Path
	if path == "" {
		path = "/mcp"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	mux := http.NewServeMux()
	httpSrv := &http.Server{
		Handler: mux,
	}

	handler := mcpServer.NewStreamableHTTPServer(
		NewMCPServer(),
		mcpServer.WithHTTPContextFunc(WithHTTPRequestContext),
		mcpServer.WithStreamableHTTPServer(httpSrv),
	)
	mux.Handle(path, http.MaxBytesHandler(requireMCPAuth(handler), 64*1024))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return handler
}

func requireMCPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := strings.TrimSpace(global.GVA_CONFIG.MCP.AuthToken)
		if expected == "" {
			expected = strings.TrimSpace(os.Getenv("GVA_MCP_AUTH_TOKEN"))
		}
		provided := extractIncomingAuthToken(r.Header)
		if expected == "" || len(expected) != len(provided) || subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
