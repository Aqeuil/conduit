package matcher

import (
	"conduit/internal/biz"
	"conduit/pkg/util"
	"errors"
)

type RadixMatcher struct {
	matcher *util.SafeRadixTree[util.StringKey, biz.ServiceUnit]
}

func NewRadixMatcher() *RadixMatcher {
	return &RadixMatcher{
		matcher: util.NewSafeRadixTree[util.StringKey, biz.ServiceUnit](),
	}
}

func (r RadixMatcher) Match(path string) (*biz.ServiceUnit, error) {
	v, ok := r.matcher.Find(util.StringKey(path))
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (r RadixMatcher) Add(unit *biz.ServiceUnit, paths ...string) error {
	for _, path := range paths {
		_, ok := r.matcher.Find(util.StringKey(path))
		if ok {
			return errors.New("path exists")
		}
	}

	for _, path := range paths {
		r.matcher.Save(util.StringKey(path), *unit)
	}
	return nil
}
