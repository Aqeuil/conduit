package matcher

import (
	"conduit/internal/biz"
	"context"
)

// RouterMatcher 只负责路由匹配，不感知数据来源。
type RouterMatcher interface {
	Match(path string) (*biz.ServiceUnit, error)
}

// EventType 变更类型
type EventType uint8

const (
	// EventRebuild 全量重建（初始化或 watch 断连重连后）
	EventRebuild EventType = iota
	EventAdd
	EventUpdate
	EventDelete
)

// Event watcher 推送给 manager 的变更事件
type Event struct {
	Type            EventType
	Units           []*biz.ServiceUnit // Rebuild
	ResourceVersion int64
}

// Watcher 数据来源抽象，负责推送变更事件到 ch，直到 ctx 结束后返回 nil。
type Watcher interface {
	Watch(ctx context.Context, ch chan<- Event) error
}
