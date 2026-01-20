package util

import (
	"math/rand"
	"time"
)

// InSlice 元素是否存在数组中
func InSlice[E comparable](v E, slice []E) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}

	return false
}

// MergeSlice merge two slice
func MergeSlice[E any](slice []E, otherSlices ...[]E) []E {
	var m []E
	total := len(slice)
	for _, e := range otherSlices {
		total += len(e)
	}

	m = make([]E, 0, total)
	m = append(m, slice...)
	for _, slc := range otherSlices {
		m = append(m, slc...)
	}
	return m
}

// MergeMap merge two map
func MergeMap[E comparable, T any](slice map[E]T, otherSlices ...map[E]T) map[E]T {
	for _, slc := range otherSlices {
		for k, v := range slc {
			slice[k] = v
		}
	}
	return slice
}

func RemoveDuplicates[T comparable](s []T) []T {
	m := make(map[T]bool)
	for i := 0; i < len(s); i++ {
		if m[s[i]] {
			s = append(s[:i], s[i+1:]...)
			i--
		} else {
			m[s[i]] = true
		}
	}
	return s
}

// IsSubset 判断一个int类型的切片是不是另一个int类型切片的子集 s2是否包含s1
func IsSubset(s1, s2 []int) bool {
	set := make(map[int]bool)
	for _, v := range s2 {
		set[v] = true
	}
	for _, v := range s1 {
		if !set[v] {
			return false
		}
	}
	return true
}

// SliceIntersect 取交集
func SliceIntersect[E comparable](slice1, slice2 []E) []E {
	m := make(map[E]bool)
	for _, v := range slice1 {
		m[v] = true
	}
	var res []E
	for _, v := range slice2 {
		if m[v] {
			res = append(res, v)
		}
	}
	return res
}

// SliceDifference 取差集
func SliceDifference[E comparable](slice1, slice2 []E) []E {
	m := make(map[E]bool)
	for _, v := range slice2 {
		m[v] = true
	}
	var res []E
	for _, v := range slice1 {
		if !m[v] {
			res = append(res, v)
		}
	}
	return res
}

// RandomSpecialSelection /* 随机取独特的selection个元素
func RandomSpecialSelection[E any](items []E, selection int, isEqual func(E, E) bool) []E {
	// 如果长度小于等于selection，则返回全部
	if len(items) <= selection {
		return items
	}

	rand.Seed(time.Now().UnixNano())
	selected := make([]E, 0, selection)
	indices := rand.Perm(len(items))

	for _, idx := range indices {
		item := items[idx]
		unique := true
		for _, sel := range selected {
			if isEqual(item, sel) {
				unique = false
				break
			}
		}
		if unique {
			selected = append(selected, item)
		}

		// enough
		if len(selected) == selection {
			break
		}
	}

	return selected
}

// SliceUnique 切片去重
func SliceUnique[E comparable](slice []E) []E {
	m := make(map[E]bool)
	var res []E
	for _, v := range slice {
		if !m[v] {
			m[v] = true
			res = append(res, v)
		}
	}
	return res
}

// GroupBy 分组
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Filter 过滤
func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
