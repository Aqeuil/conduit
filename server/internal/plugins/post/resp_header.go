package post

import (
	"conduit/internal/plugins"
	"net/http"
)

type RespHeader struct {
}

func (r RespHeader) Key() plugins.FuncKey {
	return "resp_header"
}

func (r RespHeader) Execute(req *http.Request, resp *http.Response, params map[string]any) error {
	headers, ok := params["headers"].(map[string]any)
	if !ok {
		return nil
	}
	
	for k, v := range headers {
		if val, ok := v.(string); ok {
			resp.Header.Set(k, val)
		}
	}
	return nil
}

func (r RespHeader) Help() string {
	return "重写或增加响应头"
}

func (r RespHeader) ParamRules() []plugins.ParamRule {
	return []plugins.ParamRule{
		{
			Type: plugins.Object,
			Name: "headers",
		},
	}
}
