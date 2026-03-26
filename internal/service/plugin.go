package service

import (
	v1 "conduit/api/v1"
	"conduit/internal/plugins"
	"context"
)

type PluginServer struct {
	v1.UnimplementedPluginServer
}

func NewPluginServer() *PluginServer {
	return &PluginServer{}
}

func (p PluginServer) FindPlugins(context.Context, *v1.FindPluginsReq) (*v1.FindPluginsResp, error) {
	resp := &v1.FindPluginsResp{}
	for _, k := range plugins.PrePlugins {
		paramRule, _ := p.paramRule(k.ParamRules())

		resp.PrePlugins = append(resp.PrePlugins, &v1.PluginInfo{
			Key:  string(k.Key()),
			Desc: k.Help(),
			Rule: paramRule,
		})
	}

	for _, k := range plugins.PostPlugins {
		paramRule, _ := p.paramRule(k.ParamRules())

		resp.PostPlugins = append(resp.PostPlugins, &v1.PluginInfo{
			Key:  string(k.Key()),
			Desc: k.Help(),
			Rule: paramRule,
		})
	}

	return resp, nil
}

func (p PluginServer) paramRule(rule []plugins.ParamRule) (rules []*v1.PluginParamRule, err error) {
	for _, r := range rule {
		paramRule, _ := p.paramRule(r.Children)

		rules = append(rules, &v1.PluginParamRule{
			Name:     r.Name,
			Type:     string(r.Type),
			Children: paramRule,
		})
	}
	return
}
