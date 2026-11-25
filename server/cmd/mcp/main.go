package main

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	mcpTool "github.com/WangMi2022/mit-assets-admin/server/mcp"
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

	authToken := strings.TrimSpace(global.MAA_CONFIG.MCP.AuthToken)
	if authToken == "" {
		authToken = strings.TrimSpace(os.Getenv("MAA_MCP_AUTH_TOKEN"))
	}
	if authToken == "" {
		panic("独立 MCP 必须配置 MAA_MCP_AUTH_TOKEN 或 mcp.auth_token")
	}
	host := strings.TrimSpace(os.Getenv("MAA_MCP_BIND"))
	if host == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(global.MAA_CONFIG.MCP.Addr))
	server := mcpTool.NewStreamableHTTPServer()

	global.MAA_LOG.Info("mcp独立服务启动",
		zap.String("config", configPath),
		zap.String("addr", addr),
		zap.String("path", global.MAA_CONFIG.MCP.Path),
		zap.String("upstream", global.MAA_CONFIG.MCP.UpstreamBaseURL),
	)

	if err := server.Start(addr); err != nil {
		global.MAA_LOG.Fatal("mcp独立服务启动失败", zap.Error(err))
	}
}
