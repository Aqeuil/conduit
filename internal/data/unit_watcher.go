package data

import (
	"conduit/internal/biz/matcher"
	"conduit/internal/biz/unit"
	"conduit/internal/conf"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ matcher.Watcher = (*UnitWatcher)(nil)

// UnitWatcher 实现 matcher.Watcher，负责 list+watch etcd，将变更推送到 ch。
type UnitWatcher struct {
	etcdConf *conf.Etcd
	prefix   string
	log      *log.Helper

	// 运行时字段，由 Watch 内部初始化
	client          *clientv3.Client
	cleanUp         func()
	resourceVersion int64
}

func NewUnitWatcher(c *conf.Etcd, logger log.Logger) *UnitWatcher {
	return &UnitWatcher{
		etcdConf: c,
		prefix:   c.WatchPrefix,
		log:      log.NewHelper(logger),
	}
}

// Watch 实现 matcher.Watcher。先 list 全量，再 watch 增量；watch 断连后等 2s 重新 list+watch。
// ctx 取消时安静退出（返回 nil），client 生命周期由内部 defer 管理。
func (w *UnitWatcher) Watch(ctx context.Context, ch chan<- matcher.Event) error {
	client, cleanup, err := NewEtcdClient(w.etcdConf)
	if err != nil {
		return err
	}
	defer cleanup()
	w.client = client

	if err = w.list(ctx, ch); err != nil {
		return err
	}

	for {
		watchErr := w.watch(ctx, ch)
		if watchErr == nil {
			// ctx 结束，正常退出
			return nil
		}
		w.log.Errorf("etcd watch error, retrying: %v", watchErr)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
			if err = w.list(ctx, ch); err != nil {
				return err
			}
		}
	}
}

func (w *UnitWatcher) list(ctx context.Context, ch chan<- matcher.Event) error {
	resp, err := w.client.Get(ctx, w.prefix, clientv3.WithPrefix())
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("etcd list %s: %w", w.prefix, err)
	}

	units := make([]*unit.ServiceApplication, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		unit, err2 := decodeUnit(kv.Value, kv.ModRevision)
		if err2 != nil {
			return fmt.Errorf("decode key=%s: %w", kv.Key, err2)
		}
		units = append(units, unit)
	}

	w.resourceVersion = resp.Header.Revision
	w.log.Infof("etcd list %s done, revision=%d count=%d", w.prefix, w.resourceVersion, len(units))

	ch <- matcher.Event{Type: matcher.EventRebuild, Units: units, ResourceVersion: w.resourceVersion}
	return nil
}

func (w *UnitWatcher) watch(ctx context.Context, ch chan<- matcher.Event) error {
	wch := w.client.Watch(
		ctx,
		w.prefix,
		clientv3.WithPrefix(),
		clientv3.WithRev(w.resourceVersion+1),
		clientv3.WithPrevKV(),
	)
	for {
		select {
		case <-ctx.Done():
			return nil
		case wresp, ok := <-wch:
			if !ok {
				// ctx 取消时 etcd client 会关闭 wch
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("watch channel closed")
			}
			if err := wresp.Err(); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
			for _, ev := range wresp.Events {
				event, err := w.toEvent(ev)
				if err != nil {
					w.log.Errorf("toEvent key=%s: %v", ev.Kv.Key, err)
					continue
				}
				ch <- event
			}
			w.resourceVersion = wresp.Header.Revision
		default:
		}
	}
}

func (w *UnitWatcher) toEvent(ev *clientv3.Event) (matcher.Event, error) {
	switch ev.Type {
	case clientv3.EventTypePut:
		u, err := decodeUnit(ev.Kv.Value, ev.Kv.ModRevision)
		if err != nil {
			return matcher.Event{}, err
		}
		if ev.IsCreate() {
			return matcher.Event{Type: matcher.EventAdd, Units: []*unit.ServiceApplication{u}, ResourceVersion: ev.Kv.ModRevision}, nil
		}
		return matcher.Event{Type: matcher.EventUpdate, Units: []*unit.ServiceApplication{u}, ResourceVersion: ev.Kv.ModRevision}, nil

	case clientv3.EventTypeDelete:
		u, err := decodeUnit(ev.PrevKv.Value, ev.Kv.ModRevision)
		if err != nil {
			return matcher.Event{}, err
		}
		return matcher.Event{Type: matcher.EventDelete, Units: []*unit.ServiceApplication{u}, ResourceVersion: ev.Kv.ModRevision}, nil

	default:
		return matcher.Event{}, fmt.Errorf("unknown event type %v", ev.Type)
	}
}

func decodeUnit(val []byte, rv int64) (*unit.ServiceApplication, error) {
	var unit unit.ServiceApplication
	if err := json.Unmarshal(val, &unit); err != nil {
		return nil, err
	}
	unit.ResourceVersion = rv
	return &unit, nil
}
