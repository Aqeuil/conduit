package plugin

import (
	"conduit/api/v1/common"
	admin "conduit/api/v1/conduit-admin"
	"conduit/internal/data"
	"conduit/internal/plugins"
	"context"
)

type PluginAdminServer struct {
	admin.UnimplementedPluginServer

	data *data.Data
}

func NewPluginAdminServer(data *data.Data) *PluginAdminServer {
	return &PluginAdminServer{
		data: data,
	}
}

func (p PluginAdminServer) FindBasePlugins(context.Context, *admin.FindPluginsReq) (*admin.FindPluginsResp, error) {
	resp := &admin.FindPluginsResp{}
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

func (p PluginAdminServer) paramRule(rule []plugins.ParamRule) (rules []*common.PluginParamRule, err error) {
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
