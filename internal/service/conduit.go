package service

import (
	"conduit/internal/biz/matcher"
	"conduit/internal/biz/response"
	"conduit/internal/data"
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

func NewConduitServer(
	logger log.Logger,
	watcher *data.UnitWatcher,
) *ConduitServer {
	return &ConduitServer{
		log:     log.NewHelper(logger),
		matcher: watcher.Matcher(),
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
			req.URL.Host = targetHost.Upstream
			req.Host = targetHost.Upstream
			req.URL.Scheme = "http"

			err2 := targetHost.PreProgress(req)
			if err2 != nil {
				panic(err2)
			}
		},
		ModifyResponse: func(res *http.Response) error {
			return targetHost.PostProgress(res.Request, res)
		},
	}
	proxy.ServeHTTP(w, r)
}
