package server

import (
	v1 "conduit/api/v1/conduit"
	adminV1 "conduit/api/v1/conduit-admin"
	"conduit/internal/conf"
	"conduit/internal/service"
	"conduit/internal/service/plugin"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type HttpServer http.Server
type ConduitServer http.Server
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

// NewConduitHTTPServer new an HTTP server.
func NewConduitHTTPServer(
	c *conf.Server,
	logger log.Logger,
	plugin *plugin.PluginServer,
) *ConduitServer {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.InnerHttp.Network != "" {
		opts = append(opts, http.Network(c.InnerHttp.Network))
	}
	if c.InnerHttp.Addr != "" {
		opts = append(opts, http.Address(c.InnerHttp.Addr))
	}
	if c.InnerHttp.Timeout != nil {
		opts = append(opts, http.Timeout(c.InnerHttp.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// register
	v1.RegisterPluginHTTPServer(srv, plugin)
	return (*ConduitServer)(srv)
}

// NewAdminHTTPServer new an HTTP server.
func NewAdminHTTPServer(
	c *conf.Server,
	logger log.Logger,
	plugin *plugin.PluginAdminServer,
) *AdminServer {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.AdminHttp.Network != "" {
		opts = append(opts, http.Network(c.AdminHttp.Network))
	}
	if c.AdminHttp.Addr != "" {
		opts = append(opts, http.Address(c.AdminHttp.Addr))
	}
	if c.AdminHttp.Timeout != nil {
		opts = append(opts, http.Timeout(c.AdminHttp.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// register
	adminV1.RegisterPluginHTTPServer(srv, plugin)
	return (*AdminServer)(srv)
}
