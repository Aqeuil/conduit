package manager

import (
	"conduit/internal/biz/matcher"
	"conduit/internal/biz/unit"
	"context"
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Manager)(nil)
var _ matcher.RouterMatcher = (*Manager)(nil)

// Manager 持有 units 的真实数据，监听 Watcher 推送的事件，驱动 matcher 快照重建。
type Manager struct {
	snap    atomic.Pointer[matcher.MatcherSnapshot]
	units   map[string]*unit.ServiceApplication
	watcher matcher.Watcher
	log     *log.Helper

	cancel context.CancelFunc
}

func NewManager(w matcher.Watcher, logger log.Logger) *Manager {
	m := &Manager{
		units:   make(map[string]*unit.ServiceApplication),
		watcher: w,
		log:     log.NewHelper(logger),
	}
	m.snap.Store(matcher.BuildSnapshot(m.units))
	return m
}

// Match 实现 RouterMatcher，无锁读。
func (m *Manager) Match(path string) (*unit.ServiceApplication, error) {
	return m.snap.Load().Match(path)
}

// Start 实现 transport.Server，启动 watcher 并在单 goroutine 内处理所有事件。
func (m *Manager) Start(ctx context.Context) error {
	watchCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	ch := make(chan matcher.Event, 64)
	go func() {
		if err := m.watcher.Watch(watchCtx, ch); err != nil {
			m.log.Errorf("watcher exited with error: %v", err)
		}
	}()

	for {
		select {
		case <-watchCtx.Done():
			return nil
		case ev := <-ch:
			m.handle(ev)
		}
	}
}

// Stop 实现 transport.Server。先取消 watcher ctx，Watch goroutine 感知后干净退出，
// etcd client 随 defer cleanup() 关闭，不会产生飞行中请求被强制中断的错误。
func (m *Manager) Stop(_ context.Context) error {
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}

func (m *Manager) handle(ev matcher.Event) {
	switch ev.Type {
	case matcher.EventRebuild:
		m.units = make(map[string]*unit.ServiceApplication, len(ev.Units))
		for _, u := range ev.Units {
			m.units[u.Id] = u
		}
	case matcher.EventAdd, matcher.EventUpdate:
		m.units[ev.Units[0].Id] = ev.Units[0]
	case matcher.EventDelete:
		delete(m.units, ev.Units[0].Id)
	}
	m.snap.Store(matcher.BuildSnapshot(m.units))
}
