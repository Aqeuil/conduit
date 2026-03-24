package pre

import (
	"conduit/internal/plugins"
	"net/http"
)

type Redirect struct {
}

func (r Redirect) Key() plugins.FuncKey {
	return "redirect"
}

func (r Redirect) Execute(req *http.Request, params map[string]any) error {
	req.URL.Host = params["host"].(string)
	return nil
}

func (r Redirect) Help() string {
	return "重定向访问域名"
}

func (r Redirect) ParamRules() []plugins.ParamRule {
	return []plugins.ParamRule{
		{
			Type: plugins.String,
			Name: "host",
		},
	}
}
