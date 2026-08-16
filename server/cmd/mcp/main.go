package main

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	mcpTool "github.com/flipped-aurora/gin-vue-admin/server/mcp"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	configPath, err := loadStandaloneConfig()
	if err != nil {
		panic(err)
	}

	if err := initializeStandaloneLogger(); err != nil {
		panic(err)
	}

	authToken := strings.TrimSpace(global.GVA_CONFIG.MCP.AuthToken)
	if authToken == "" {
		authToken = strings.TrimSpace(os.Getenv("GVA_MCP_AUTH_TOKEN"))
	}
	if authToken == "" {
		panic("独立 MCP 必须配置 GVA_MCP_AUTH_TOKEN 或 mcp.auth_token")
	}
	host := strings.TrimSpace(os.Getenv("GVA_MCP_BIND"))
	if host == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(global.GVA_CONFIG.MCP.Addr))
	server := mcpTool.NewStreamableHTTPServer()

	global.GVA_LOG.Info("mcp独立服务启动",
		zap.String("config", configPath),
		zap.String("addr", addr),
		zap.String("path", global.GVA_CONFIG.MCP.Path),
		zap.String("upstream", global.GVA_CONFIG.MCP.UpstreamBaseURL),
	)

	if err := server.Start(addr); err != nil {
		global.GVA_LOG.Fatal("mcp独立服务启动失败", zap.Error(err))
	}
}
