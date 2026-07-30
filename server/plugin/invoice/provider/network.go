package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// providerHTTPClient centralizes outbound provider transport policy. Private
// endpoints are opt-in because these URLs are saved in runtime configuration
// and probes run from the server process.
func providerHTTPClient(timeout time.Duration, allowPrivateEndpoints bool) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:       nil,
			DialContext: providerDialContext(allowPrivateEndpoints),
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func providerDialContext(allowPrivateEndpoints bool) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if allowPrivateEndpoints {
			return dialer.DialContext(ctx, network, address)
		}
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("解析服务地址失败: %w", err)
		}
		var restricted []string
		for _, resolved := range ips {
			if restrictedProviderIP(resolved.IP) {
				restricted = append(restricted, resolved.IP.String())
				continue
			}
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.IP.String(), port))
			if dialErr == nil {
				return connection, nil
			}
		}
		if len(restricted) == len(ips) && len(restricted) > 0 {
			return nil, fmt.Errorf("服务地址解析到受限内网地址（%s），如确认是自建网关请开启允许内网端点", strings.Join(restricted, ", "))
		}
		return nil, fmt.Errorf("无法连接服务地址 %s", host)
	}
}

func restrictedProviderIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	// Carrier-grade NAT and the IPv4 metadata endpoint are not covered by
	// net.IP.IsPrivate on all supported Go versions.
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}
	return ipv4[0] == 0 ||
		(ipv4[0] == 100 && ipv4[1] >= 64 && ipv4[1] <= 127) ||
		(ipv4[0] == 169 && ipv4[1] == 254)
}
