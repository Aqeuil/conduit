package matcher

import (
	"conduit/internal/biz"
	"conduit/pkg/util"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type RadixMatcher struct {
	// snap 当前快照
	snap    atomic.Pointer[matcherSnapshot]
	writeMu sync.Mutex

	// ResourceVersion
	ResourceVersion int64
}

type matcherSnapshot struct {
	tree  *util.RadixTree[util.StringKey, string]
	units map[string]*biz.ServiceUnit
}

func NewRadixMatcher() *RadixMatcher {
	m := &RadixMatcher{}
	initialSnap := &matcherSnapshot{
		tree:  util.NewRadixTree[util.StringKey, string](),
		units: make(map[string]*biz.ServiceUnit),
	}
	m.snap.Store(initialSnap)
	return m
}

func (r *RadixMatcher) Match(path string) (*biz.ServiceUnit, error) {
	snap := r.snap.Load()
	v, ok := snap.tree.Find(util.StringKey(path))
	if !ok {
		return nil, errors.New("not found")
	}
	unit := snap.units[*v]
	return unit, nil
}

func (r *RadixMatcher) Add(unit *biz.ServiceUnit) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	oldSnap := r.snap.Load()

	newTree := util.Copy(oldSnap.tree).(*util.RadixTree[util.StringKey, string])
	newUnits := make(map[string]*biz.ServiceUnit, len(oldSnap.units)+1)
	for k, v := range oldSnap.units {
		newUnits[k] = v
	}

	for _, path := range unit.Routers {
		if _, ok := newTree.Find(util.StringKey(path)); ok {
			return errors.New(fmt.Sprintf("duplicate path: %s", path))
		}
	}

	for _, path := range unit.Routers {
		newTree.Save(util.StringKey(path), unit.Id)
	}
	newUnits[unit.Id] = unit

	newSnap := &matcherSnapshot{
		tree:  newTree,
		units: newUnits,
	}
	r.snap.Store(newSnap)
	return nil
}

func (r *RadixMatcher) Rebuild(unit []*biz.ServiceUnit) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	newTree := util.NewRadixTree[util.StringKey, string]()
	newUnits := make(map[string]*biz.ServiceUnit, len(unit))
	for _, v := range unit {
		newUnits[v.Id] = v

		for _, path := range v.Routers {
			newTree.Save(util.StringKey(path), v.Id)
		}
	}

	newSnap := &matcherSnapshot{
		tree:  newTree,
		units: newUnits,
	}
	r.snap.Store(newSnap)
	return nil
}

func (r *RadixMatcher) Update(unit *biz.ServiceUnit) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	oldSnap := r.snap.Load()

	newTree := util.Copy(oldSnap.tree).(*util.RadixTree[util.StringKey, string])
	newUnits := make(map[string]*biz.ServiceUnit, len(oldSnap.units))
	for k, v := range oldSnap.units {
		newUnits[k] = v
	}
	newUnits[unit.Id] = unit

	oldUnit := newUnits[unit.Id]

	waitDel := util.SliceDifference(oldUnit.Routers, unit.Routers)
	waitAdd := util.SliceDifference(unit.Routers, oldUnit.Routers)
	for _, path := range waitDel {
		_ = newTree.Delete(util.StringKey(path))
	}
	for _, path := range waitAdd {
		newTree.Save(util.StringKey(path), unit.Id)
	}

	newSnap := &matcherSnapshot{
		tree:  newTree,
		units: newUnits,
	}
	r.snap.Store(newSnap)
	return nil
}

func (r *RadixMatcher) Delete(unit *biz.ServiceUnit) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	oldSnap := r.snap.Load()

	newTree := util.Copy(oldSnap.tree).(*util.RadixTree[util.StringKey, string])
	newUnits := make(map[string]*biz.ServiceUnit, len(oldSnap.units)-1)
	for k, v := range oldSnap.units {
		if k == unit.Id {
			continue
		}
		newUnits[k] = v
	}

	for _, path := range unit.Routers {
		_ = newTree.Delete(util.StringKey(path))
	}

	newSnap := &matcherSnapshot{
		tree:  newTree,
		units: newUnits,
	}
	r.snap.Store(newSnap)
	return nil
}
