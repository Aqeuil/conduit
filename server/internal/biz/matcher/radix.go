package matcher

import (
	"conduit/internal/biz/unit"
	"conduit/pkg/util"
	"errors"
)

// MatcherSnapshot 不可变快照，构建一次后只读。
type MatcherSnapshot struct {
	tree  *util.RadixTree[util.StringKey, string]
	units map[string]*unit.ServiceApplication
}

func BuildSnapshot(units map[string]*unit.ServiceApplication) *MatcherSnapshot {
	tree := util.NewRadixTree[util.StringKey, string]()
	for _, u := range units {
		for _, path := range u.Routers {
			tree.Save(util.StringKey(path), u.Id)
		}
	}
	return &MatcherSnapshot{tree: tree, units: units}
}

func (s *MatcherSnapshot) Match(path string) (*unit.ServiceApplication, error) {
	v, ok := s.tree.Find(util.StringKey(path))
	if !ok {
		return nil, errors.New("not found")
	}
	return s.units[*v], nil
}
