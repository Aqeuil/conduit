package service

import (
	"conduit/internal/biz"
	"conduit/internal/biz/matcher"
	"conduit/internal/biz/response"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

type ConduitServer struct {
	log *log.Helper

	matcher matcher.RouterMatcher
}

func NewConduitServer(logger log.Logger) *ConduitServer {
	return &ConduitServer{
		log:     log.NewHelper(logger),
		matcher: matcher.NewRadixMatcher(),
	}
}

func (c ConduitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "?")[0]
	targetHost, err := c.matcher.Match(path)
	if err != nil {
		// 返回404
		resp := response.FailWithCode(404, fmt.Sprintf("path %s not found", path))
		marshal, _ := json.Marshal(resp)
		w.WriteHeader(200)
		_, _ = w.Write(marshal)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			err2 := targetHost.PreProgress(req)
			if err2 != nil {
				panic(err2)
			}
		},
		ModifyResponse: func(res *http.Response) error {
			return targetHost.PostProgress(res)
		},
	}
	proxy.ServeHTTP(w, r)
}

func (c ConduitServer) Online(unit *biz.ServiceUnit, path ...string) error {
	for _, p := range path {
		err := c.matcher.Add(unit, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c ConduitServer) Offline(unit *biz.ServiceUnit, path ...string) error {
	for _, p := range path {
		err := c.matcher.Add(nil, p)
		if err != nil {
			return err
		}
	}
	return nil
}
