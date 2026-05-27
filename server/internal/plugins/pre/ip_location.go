package pre

import (
	"conduit/internal/plugins"
	"net"
	"net/http"
	"strings"
)

type IpLocation struct {
}

func (i IpLocation) Key() plugins.FuncKey {
	return "ip_location"
}

func (i IpLocation) Execute(req *http.Request, params map[string]any) error {
	ip := i.getRealIP(req)
	
	headerName := "X-Geo-Location"
	if customHeader, ok := params["header_name"].(string); ok && customHeader != "" {
		headerName = customHeader
	}
	
	req.Header.Set(headerName, ip)
	return nil
}

func (i IpLocation) getRealIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}


func (i IpLocation) Help() string {
	return "解析客户端真实IP并在Header中注入地理位置信息"
}

func (i IpLocation) ParamRules() []plugins.ParamRule {
	return []plugins.ParamRule{
		{
			Type: plugins.String,
			Name: "header_name",
		},
	}
}
