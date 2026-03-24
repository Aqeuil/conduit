package matcher

import (
	"conduit/internal/biz"
)

// RouterMatcher url匹配核心逻辑
type RouterMatcher interface {
	Match(path string) (*biz.ServiceUnit, error)

	Add(unit *biz.ServiceUnit, path ...string) error
}
