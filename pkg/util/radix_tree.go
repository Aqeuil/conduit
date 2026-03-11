package util

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
	if key.Len() == 0 { // 完全匹配
		v.v = &value
		return 1
	}

	for _, child := range v.children {
		i := 0
		keyLen, childLen := key.Len(), child.k.Len()

		for ; i < childLen && i < keyLen; i++ {
			if child.k.At(i) != key.At(i) {
				break
			}
		}

		if i == childLen {
			slice := key.Slice(i, keyLen)

			return child.Save(slice.(K), value)
		}

		if i <= 0 {
			continue
		}

		// 分裂
		v.fission(child, key, value)
		return 1
	}

	// 新增子节点
	node := &RadixNode[K, V]{
		v: &value,
		k: key,
	}
	v.children = append(v.children, node)
	return 1
}

func (v *RadixNode[K, V]) fission(oldNode *RadixNode[K, V], key K, value V) bool {
	i := 0
	keyLen, childLen := key.Len(), oldNode.k.Len()

	for ; i < childLen && i < keyLen; i++ {
		if oldNode.k.At(i) != key.At(i) {
			break
		}
	}

	prefix := key.Slice(0, i)

	newNode := &RadixNode[K, V]{
		v: &value,
		k: key.Slice(i, keyLen).(K),
	}

	newChild := &RadixNode[K, V]{
		v:        oldNode.v,
		k:        oldNode.k.Slice(i, childLen).(K),
		children: oldNode.children,
	}
	oldNode.children = []*RadixNode[K, V]{newNode, newChild}
	oldNode.v = nil
	oldNode.k = prefix.(K)
	return true
}

func (v *RadixNode[K, V]) Delete(key K) bool {
	if key.Len() == 0 {
		v.v = nil
		return true
	}

	index := -1
	for idx, child := range v.children {
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
			if i == keyLen {
				// 就是这个节点
				index = idx
				break
			}

			slice := key.Slice(i, keyLen)
			return child.Delete(slice.(K))
		}
	}

	// remove
	if index != -1 {
		v.children[index].v = nil
		//v.children[index] = v.children[len(v.children)-1]
		//v.children = v.children[:len(v.children)-1]

		return true
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
