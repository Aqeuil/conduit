package util

import (
	"sync"
	"sync/atomic"
)

type RadixKey interface {
	// Len 返回 key 的字节长度（radix tree 通常按字节处理）
	Len() int

	// At 返回第 i 个字节（0 <= i < Len()），用于逐字节比较前缀
	At(i int) byte

	// Slice 返回 key 中 [lo:hi) 的子视图（不复制底层数据）
	// 这是 radix tree 压缩边（edge label）时必须的，用于 split 节点时创建 prefix/suffix
	Slice(lo, hi int) RadixKey
}

type StringKey string

func (k StringKey) Len() int                  { return len(k) }
func (k StringKey) At(i int) byte             { return k[i] }
func (k StringKey) Slice(lo, hi int) RadixKey { return StringKey(k[lo:hi]) }

type BytesKey []byte

func (k BytesKey) Len() int                  { return len(k) }
func (k BytesKey) At(i int) byte             { return k[i] }
func (k BytesKey) Slice(lo, hi int) RadixKey { return k[lo:hi] } // 共享底层数组，插入后不要修改！

// RadixNode 树节点
type RadixNode[K RadixKey, V any] struct {
	v        *V
	k        K
	children []*RadixNode[K, V]
}

func (v *RadixNode[K, V]) Find(key K) (*V, bool) {
	if key.Len() == 0 {
		return v.v, v.v != nil
	}

	for _, child := range v.children {
		i := 0
		keyLen, childLen := key.Len(), child.k.Len()
		if keyLen < childLen {
			continue
		}

		for ; i < childLen && i < keyLen; i++ {
			if child.k.At(i) != key.At(i) {
				break
			}
		}

		if i == childLen {
			slice := key.Slice(i, keyLen)

			return child.Find(slice.(K))
		}
	}

	return nil, false
}
func (v *RadixNode[K, V]) Save(key K, value V) int {
	keyLen, nodeLen := key.Len(), v.k.Len()

	matchLen := 0
	for matchLen < nodeLen && matchLen < keyLen && v.k.At(matchLen) == key.At(matchLen) {
		matchLen++
	}

	if matchLen == 0 && nodeLen > 0 {
		return 0
	}

	if matchLen < nodeLen {
		newChild := &RadixNode[K, V]{
			k:        v.k.Slice(matchLen, nodeLen).(K),
			v:        v.v,
			children: v.children,
		}
		v.k = v.k.Slice(0, matchLen).(K)
		v.v = nil
		v.children = []*RadixNode[K, V]{newChild}
	}

	if matchLen == keyLen {
		v.v = &value
		return 1
	}

	suffixKey := key.Slice(matchLen, keyLen).(K)

	for _, child := range v.children {
		if child.Save(suffixKey, value) == 1 {
			return 1
		}
	}

	v.children = append(v.children, &RadixNode[K, V]{
		k: suffixKey,
		v: &value,
	})

	return 1
}

func (v *RadixNode[K, V]) Delete(key K) bool {
	keyLen, nodeLen := key.Len(), v.k.Len()

	if keyLen < nodeLen {
		return false
	}

	matchLen := 0
	for matchLen < nodeLen && matchLen < keyLen && v.k.At(matchLen) == key.At(matchLen) {
		matchLen++
	}

	if matchLen < nodeLen {
		return false
	}

	if matchLen == keyLen {
		v.v = nil
		return true
	}

	suffixKey := key.Slice(matchLen, keyLen).(K)
	for _, child := range v.children {
		if child.Delete(suffixKey) == true {
			return true
		}
	}

	return false
}

type RadixTree[K RadixKey, V any] struct {
	root *RadixNode[K, V]
	size int
}

func NewRadixTree[K RadixKey, V any]() *RadixTree[K, V] {
	return &RadixTree[K, V]{
		root: &RadixNode[K, V]{},
	}
}

func (r *RadixTree[K, V]) Find(key K) (v *V, ok bool) {
	return r.root.Find(key)
}

func (r *RadixTree[K, V]) Save(key K, value V) (ok bool) {
	rowEffected := r.root.Save(key, value)
	if rowEffected >= 1 {
		r.size++
	}
	return true
}

func (r *RadixTree[K, V]) Delete(key K) (ok bool) {
	ok = r.root.Delete(key)
	if ok {
		r.size--
	}
	return ok
}

// SafeRadixNode 树节点
type SafeRadixNode[K RadixKey, V any] struct {
	sync.RWMutex

	v        *V
	k        K
	children []*SafeRadixNode[K, V]
}

// Find 保持不变，纯读操作，使用 RLock 非常合适
func (v *SafeRadixNode[K, V]) Find(key K) (*V, bool) {
	v.RLock()
	defer v.RUnlock()

	if key.Len() == 0 {
		return v.v, v.v != nil
	}

	for _, child := range v.children {
		i := 0
		keyLen, childLen := key.Len(), child.k.Len()
		if keyLen < childLen {
			continue
		}

		for ; i < childLen && i < keyLen; i++ {
			if child.k.At(i) != key.At(i) {
				break
			}
		}

		if i == childLen {
			slice := key.Slice(i, keyLen)
			find, b := child.Find(slice.(K))
			return find, b
		}
	}

	return nil, false
}

// Save 作为写入操作，直接使用写锁，消除锁升级带来的竞争窗口
func (v *SafeRadixNode[K, V]) Save(key K, value V) int {
	v.Lock() // 直接使用写锁
	defer v.Unlock()

	keyLen, nodeLen := key.Len(), v.k.Len()

	matchLen := 0
	for matchLen < nodeLen && matchLen < keyLen && v.k.At(matchLen) == key.At(matchLen) {
		matchLen++
	}

	if matchLen == 0 && nodeLen > 0 {
		return 0
	}

	if matchLen < nodeLen {
		newChild := &SafeRadixNode[K, V]{
			k:        v.k.Slice(matchLen, nodeLen).(K),
			v:        v.v,
			children: v.children,
		}
		v.k = v.k.Slice(0, matchLen).(K)
		v.v = nil
		v.children = []*SafeRadixNode[K, V]{newChild}
	}

	if matchLen == keyLen {
		v.v = &value
		return 1
	}

	suffixKey := key.Slice(matchLen, keyLen).(K)

	for _, child := range v.children {
		if child.Save(suffixKey, value) == 1 {
			return 1
		}
	}

	v.children = append(v.children, &SafeRadixNode[K, V]{
		k: suffixKey,
		v: &value,
	})

	return 1
}

// Delete 作为写入操作，直接使用写锁
func (v *SafeRadixNode[K, V]) Delete(key K) bool {
	v.Lock()
	defer v.Unlock()

	keyLen, nodeLen := key.Len(), v.k.Len()

	if keyLen < nodeLen {
		return false
	}

	matchLen := 0
	for matchLen < nodeLen && matchLen < keyLen && v.k.At(matchLen) == key.At(matchLen) {
		matchLen++
	}

	if matchLen < nodeLen {
		return false
	}

	if matchLen == keyLen {
		v.v = nil
		return true
	}

	suffixKey := key.Slice(matchLen, keyLen).(K)
	for _, child := range v.children {
		if child.Delete(suffixKey) == true {
			return true
		}
	}

	return false
}

type SafeRadixTree[K RadixKey, V any] struct {
	root *SafeRadixNode[K, V]
	size int64
}

func NewSafeRadixTree[K RadixKey, V any]() *SafeRadixTree[K, V] {
	return &SafeRadixTree[K, V]{
		root: &SafeRadixNode[K, V]{},
	}
}

func (r *SafeRadixTree[K, V]) Find(key K) (v *V, ok bool) {
	return r.root.Find(key)
}

func (r *SafeRadixTree[K, V]) Save(key K, value V) (ok bool) {
	rowEffected := r.root.Save(key, value)
	if rowEffected >= 1 {
		atomic.AddInt64(&r.size, int64(rowEffected))
	}
	return true
}

func (r *SafeRadixTree[K, V]) Delete(key K) (ok bool) {
	ok = r.root.Delete(key)
	if ok {
		atomic.AddInt64(&r.size, -1)
	}
	return ok
}
