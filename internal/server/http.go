package server

import (
	v1 "conduit/api/v1"
	"conduit/internal/conf"
	"conduit/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type HttpServer http.Server
type AdminServer http.Server

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, conduit *service.ConduitServer) *HttpServer {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// Proxy
	srv.HandlePrefix("", conduit)
	return (*HttpServer)(srv)
}

// NewAdminHTTPServer new an HTTP server.
func NewAdminHTTPServer(
	c *conf.Server,
	logger log.Logger,
	conduit *service.ConduitServer,
	plugin *service.PluginServer,
) *AdminServer {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.AdminHttp.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.AdminHttp.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.AdminHttp.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// register
	v1.RegisterPluginHTTPServer(srv, plugin)
	return (*AdminServer)(srv)
}
