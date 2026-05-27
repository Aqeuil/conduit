package post

import (
	"conduit/internal/plugins"
	"conduit/pkg/xlog"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"net/http"
	"strings"
	"time"
)

type AccessLog struct {
	logger *log.Helper
}

func NewAccessLog() *AccessLog {
	// Initialize with a default logger
	l := xlog.NewXLogger(xlog.WithLevel("info"), xlog.WithFileName("logs/access.log"))
	return &AccessLog{
		logger: log.NewHelper(l),
	}
}

func (a *AccessLog) Key() plugins.FuncKey {
	return "access_log"
}

func (a *AccessLog) Execute(req *http.Request, resp *http.Response, params map[string]any) error {
	start, ok := req.Context().Value("start_time").(time.Time)
	duration := ""
	if ok {
		duration = time.Since(start).String()
	}

	ip := a.getRealIP(req)
	
	a.logger.Infow(
		"kind", "access",
		"method", req.Method,
		"path", req.URL.Path,
		"query", req.URL.RawQuery,
		"ip", ip,
		"status", resp.StatusCode,
		"duration", duration,
		"user_agent", req.UserAgent(),
	)
	
	return nil
}

func (a *AccessLog) getRealIP(req *http.Request) string {
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

func (a *AccessLog) Help() string {
	return "记录详细的访问日志"
}

func (a *AccessLog) ParamRules() []plugins.ParamRule {
	return nil
}
