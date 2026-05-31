package router

import (
	"conduit/internal/biz/unit"
	"conduit/internal/model"
	"context"
	"encoding/json"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// syncAppToEtcd writes the full ServiceApplication for an app to etcd.
// The key is {prefix}/{appID}; only routers with Type=="router" are included.
func syncAppToEtcd(ctx context.Context, cli *clientv3.Client, prefix string, app *model.Application, routers []*model.Router) error {
	paths := make([]string, 0, len(routers))
	for _, r := range routers {
		if r.Type == "router" && r.Path != "" {
			paths = append(paths, r.Path)
		}
	}

	sa := unit.ServiceApplication{
		Id:       app.ID,
		Upstream: app.Upstream,
		Routers:  paths,
	}

	val, err := json.Marshal(sa)
	if err != nil {
		return err
	}

	_, err = cli.Put(ctx, prefix+"/"+app.ID, string(val))
	return err
}
