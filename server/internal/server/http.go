package server

import (
	v1 "conduit/api/v1/conduit"
	adminV1 "conduit/api/v1/conduit-admin"
	"conduit/internal/conf"
	"conduit/internal/service"
	"conduit/internal/service/application"
	"conduit/internal/service/plugin"
	"conduit/internal/service/router"
	"conduit/internal/service/workflow"
	"time"

	httpNet "net/http"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	http "github.com/go-kratos/kratos/v2/transport/http"
)

type HttpServer http.Server
type ConduitServer http.Server
type AdminServer http.Server

// NewHTTPServer 代理服务（conduit）
func NewHTTPServer(c *conf.Server, logger log.Logger, conduit *service.ConduitServer) *HttpServer {
	opts := serverOpts(c.Http)
	srv := http.NewServer(opts...)
	srv.HandlePrefix("", conduit)
	return (*HttpServer)(srv)
}

// NewConduitHTTPServer conduit 管理端内部 HTTP 服务
func NewConduitHTTPServer(c *conf.Server, logger log.Logger, pluginSvc *plugin.PluginServer) *ConduitServer {
	opts := serverOpts(c.InnerHttp)
	srv := http.NewServer(opts...)
	v1.RegisterPluginHTTPServer(srv, pluginSvc)
	return (*ConduitServer)(srv)
}

// NewAdminHTTPServer admin HTTP 服务，注册所有管理接口
func NewAdminHTTPServer(
	c *conf.Server,
	logger log.Logger,
	pluginSvc *plugin.PluginAdminServer,
	appSvc *application.ApplicationAdminServer,
	routerSvc *router.RouterAdminServer,
	wfSvc *workflow.WorkflowAdminServer,
) *AdminServer {
	opts := serverOpts(c.AdminHttp)
	srv := http.NewServer(opts...)

	adminV1.RegisterPluginHTTPServer(srv, pluginSvc)
	adminV1.RegisterApplicationAdminHTTPServer(srv, appSvc)
	adminV1.RegisterRouterAdminHTTPServer(srv, routerSvc)
	adminV1.RegisterWorkflowAdminHTTPServer(srv, wfSvc)

	return (*AdminServer)(srv)
}

func serverOpts(c *conf.HTTP) []http.ServerOption {
	opts := []http.ServerOption{
		http.Middleware(recovery.Recovery()),
	}
	if c == nil {
		return opts
	}
	if c.Network != "" {
		opts = append(opts, http.Network(c.Network))
	}
	if c.Addr != "" {
		opts = append(opts, http.Address(c.Addr))
	}
	if c.Timeout != nil {
		opts = append(opts, http.Timeout(c.Timeout.AsDuration()))
	}
	// 包装返回
	opts = append(opts, http.ErrorEncoder(ErrorEncoder))
	opts = append(opts, http.ResponseEncoder(ResponseEncoder))
	return opts
}

type Response struct {
	Code    int         `json:"code" form:"code" protobuf:"varint,1,opt,name=code"`
	Message string      `json:"message" form:"message" protobuf:"bytes,2,opt,name=message"`
	Ts      string      `json:"time" form:"time" protobuf:"bytes,3,opt,name=time"`
	Reason  string      `json:"reason" form:"reason" protobuf:"bytes,4,opt,name=reason"`
	Data    interface{} `json:"data" form:"data" protobuf:"bytes,5,opt,name=data"`
}

func ErrorEncoder(w httpNet.ResponseWriter, r *httpNet.Request, err error) {
	se := errors.FromError(err)
	reply := &Response{
		Code:    int(se.Code),
		Message: se.Message,
		Reason:  se.Reason,
		Data:    nil,
		Ts:      time.Now().Format("2006-01-02 15:04:05"),
	}

	codec, _ := http.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(reply)
	if err != nil {
		w.WriteHeader(httpNet.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpNet.StatusOK)
	w.Write(body)
}

func ResponseEncoder(w httpNet.ResponseWriter, r *httpNet.Request, v interface{}) error {
	reply := &Response{
		Code:    200,
		Message: "success",
		Reason:  "success",
		Data:    v,
		Ts:      time.Now().Format("2006-01-02 15:04:05"),
	}

	codec, _ := http.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(reply)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpNet.StatusOK)
	w.Write(data)
	return nil
}
