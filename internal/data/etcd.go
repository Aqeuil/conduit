package data

import (
	"conduit/internal/conf"
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewEtcdClient(c *conf.Etcd) (*clientv3.Client, func(), error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.Endpoints,
		DialTimeout: time.Second * 3,
		// 不绑定 lifecycle ctx，client 生命周期由 cleanup 控制
	})
	if err != nil {
		return nil, nil, err
	}

	statusCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = cli.Status(statusCtx, c.Endpoints[0])
	if err != nil {
		_ = cli.Close()
		return nil, nil, err
	}

	return cli, func() { _ = cli.Close() }, nil
}
