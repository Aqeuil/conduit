package unit

import (
	"conduit/internal/plugins"
	"fmt"
	"net/http"
)

// ServiceApplication 服务单元, 最小的Deployment单位
type ServiceApplication struct {
	// Id DeploymentId
	Id string `json:"id,omitempty"`

	// Upstream 转发对应地址
	Upstream string `json:"upstream,omitempty"`
	// Routers 匹配的路由
	Routers []string `json:"routers,omitempty"`

	// 拦截器
	PreWork  []WorkFlow `json:"pre_work,omitempty"`
	PostWork []WorkFlow `json:"post_work,omitempty"`

	ResourceVersion int64 `json:"resource_version,omitempty"`
}

func (s ServiceApplication) PreProgress(req *http.Request) error {
	for _, w := range s.PreWork {
		key := w.FuncKey
		preFunc, ok := plugins.PrePlugins[plugins.FuncKey(key)]
		if !ok {
			return fmt.Errorf("pre plugin %s not found", key)
		}

		err := preFunc.Execute(req, w.Params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ServiceApplication) PostProgress(req *http.Request, resp *http.Response) error {
	for _, w := range s.PostWork {
		key := w.FuncKey
		postFunc, ok := plugins.PostPlugins[plugins.FuncKey(key)]
		if !ok {
			return fmt.Errorf("post plugin %s not found", key)
		}

		err := postFunc.Execute(req, resp, w.Params)
		if err != nil {
			return err
		}
	}
	return nil
}
