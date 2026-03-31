package data

import (
	"conduit/internal/conf"
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdClient(c *conf.Etcd, ctx context.Context) (*clientv3.Client, func(), error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.Endpoints,
		DialTimeout: time.Second * 3,
		Context:     ctx,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = cli.Status(ctx, c.Endpoints[0])
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = cli.Close()
	}
	return cli, cleanup, nil
}
