package data

import (
	"conduit/internal/biz"
	"conduit/internal/biz/matcher"
	"conduit/internal/conf"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ transport.Server = (*UnitWatcher)(nil)

// UnitWatcher lists and watches a ServiceUnit key prefix in etcd,
// keeping the RouterMatcher in sync.
type UnitWatcher struct {
	etcdConf *conf.Etcd
	client   *clientv3.Client
	cleanUp  func()

	prefix  string
	matcher matcher.RouterMatcher
	log     *log.Helper

	resourceVersion int64
}

func NewUnitWatcher(c *conf.Etcd, logger log.Logger) *UnitWatcher {
	return &UnitWatcher{
		//client:  client,
		etcdConf: c,
		prefix:   c.WatchPrefix,
		matcher:  matcher.NewRadixMatcher(),
		log:      log.NewHelper(logger),
	}
}

func (w *UnitWatcher) Matcher() matcher.RouterMatcher {
	return w.matcher
}

func (w *UnitWatcher) Watch(ctx context.Context) error {
	for {
		// --- Watch ---
		wch := w.client.Watch(ctx, w.prefix,
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
					w.log.Info("etcd watch done.")
					return nil
				}
				if wresp.Err() != nil {
					w.log.Errorf("etcd watch error: %v", wresp.Err())
					return wresp.Err()
				}

				for _, ev := range wresp.Events {
					switch ev.Type {
					case clientv3.EventTypePut:
						unit, err := w.decodeUnit(ev.Kv.Value, ev.Kv.ModRevision)
						if err != nil {
							w.log.Errorf("decode unit key=%s: %v", ev.Kv.Key, err)
							continue
						}
						if ev.IsCreate() {
							if err := w.matcher.Add(unit); err != nil {
								w.log.Errorf("matcher.Add %s: %v", unit.Id, err)
							}
						} else {
							if err := w.matcher.Update(unit); err != nil {
								w.log.Errorf("matcher.Update %s: %v", unit.Id, err)
							}
						}
					case clientv3.EventTypeDelete:
						unit, err := w.decodeUnit(ev.PrevKv.Value, ev.Kv.ModRevision)
						if err != nil {
							w.log.Errorf("decode prev unit key=%s: %v", ev.Kv.Key, err)
							continue
						}
						if err := w.matcher.Delete(unit); err != nil {
							w.log.Errorf("matcher.Delete %s: %v", unit.Id, err)
						}
					}
				}
			}
		}
	}
}

func (w *UnitWatcher) List(ctx context.Context) error {
	resp, err := w.client.Get(ctx, w.prefix, clientv3.WithPrefix())
	if err != nil {
		w.log.Errorf("etcd list %s failed: %v", w.prefix, err)
		return err
	}

	firstUnits := make([]*biz.ServiceUnit, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		unit, err2 := w.decodeUnit(kv.Value, kv.ModRevision)
		if err2 != nil {
			panic(fmt.Sprintf("decode unit key=%s: %v", kv.Key, err2))
		}

		firstUnits = append(firstUnits, unit)
	}
	if err = w.matcher.Rebuild(firstUnits); err != nil {
		panic(fmt.Sprintf("matcher.Batch : %v", err))
	}
	w.resourceVersion = resp.Header.Revision

	w.log.Infof("etcd watch %s start revision %d", w.prefix, w.resourceVersion)
	return nil
}

func (w *UnitWatcher) decodeUnit(val []byte, rv int64) (*biz.ServiceUnit, error) {
	var unit biz.ServiceUnit
	if err := json.Unmarshal(val, &unit); err != nil {
		return nil, err
	}

	unit.ResourceVersion = rv
	w.log.Infof("Received unit: %+v", unit)
	return &unit, nil
}

func (w *UnitWatcher) Start(ctx context.Context) error {
	var err error
	w.client, w.cleanUp, err = NewEtcdClient(w.etcdConf, ctx)
	if err != nil {
		return err
	}

	err = w.List(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			w.log.Info("watcher exited gracefully")
			return nil
		default:
			if err2 := w.Watch(ctx); err2 == nil {
				return nil
			}

			break
		}

		// 防止惊群
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 2):
			// 重新 List 补齐 Watch 期间可能的缝隙
			if err2 := w.List(ctx); err2 != nil {
				return err2
			}
		}
	}
}

func (w *UnitWatcher) Stop(ctx context.Context) error {
	if w.cleanUp != nil {
		w.cleanUp()
	}

	return nil
}
