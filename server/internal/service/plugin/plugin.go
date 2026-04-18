package plugin

import (
	"conduit/api/v1/common"
	"conduit/api/v1/conduit"
	"conduit/internal/plugins"
	"context"
)

type PluginServer struct {
	conduit.UnimplementedPluginServer
}

func NewPluginServer() *PluginServer {
	return &PluginServer{}
}

func (p PluginServer) FindPlugins(context.Context, *conduit.FindPluginsReq) (*conduit.FindPluginsResp, error) {
	resp := &conduit.FindPluginsResp{}
	for _, k := range plugins.PrePlugins {
		paramRule, _ := p.paramRule(k.ParamRules())

		resp.PrePlugins = append(resp.PrePlugins, &common.PluginInfo{
			Key:  string(k.Key()),
			Desc: k.Help(),
			Rule: paramRule,
		})
	}

	for _, k := range plugins.PostPlugins {
		paramRule, _ := p.paramRule(k.ParamRules())

		resp.PostPlugins = append(resp.PostPlugins, &common.PluginInfo{
			Key:  string(k.Key()),
			Desc: k.Help(),
			Rule: paramRule,
		})
	}

	return resp, nil
}

func (p PluginServer) paramRule(rule []plugins.ParamRule) (rules []*common.PluginParamRule, err error) {
	for _, r := range rule {
		paramRule, _ := p.paramRule(r.Children)

		rules = append(rules, &common.PluginParamRule{
			Name:     r.Name,
			Type:     string(r.Type),
			Children: paramRule,
		})
	}
	return
}
