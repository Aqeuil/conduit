package matcher

import (
	"conduit/internal/biz"
	"conduit/pkg/util"
	"errors"
)

// matcherSnapshot 不可变快照，构建一次后只读。
type matcherSnapshot struct {
	tree  *util.RadixTree[util.StringKey, string]
	units map[string]*biz.ServiceUnit
}

func buildSnapshot(units map[string]*biz.ServiceUnit) *matcherSnapshot {
	tree := util.NewRadixTree[util.StringKey, string]()
	for _, u := range units {
		for _, path := range u.Routers {
			tree.Save(util.StringKey(path), u.Id)
		}
	}
	return &matcherSnapshot{tree: tree, units: units}
}

func (s *matcherSnapshot) match(path string) (*biz.ServiceUnit, error) {
	v, ok := s.tree.Find(util.StringKey(path))
	if !ok {
		return nil, errors.New("not found")
	}
	return s.units[*v], nil
}
