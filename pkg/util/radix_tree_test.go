package util

import (
	"strconv"
	"sync"
	"testing"
)

func TestRadixTree(t *testing.T) {
	t.Run("Save and Find", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("foo")
		val := "bar"

		ok := tree.Save(key, val)
		if !ok {
			t.Errorf("Save returned false, expected true")
		}

		v, found := tree.Find(key)
		if !found {
			t.Errorf("Find failed to find key after Save")
		}
		if *v != val {
			t.Errorf("Find returned value %q, expected %q", *v, val)
		}
	})

	t.Run("Find non-existent key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("baz")

		v, found := tree.Find(key)
		if found {
			t.Errorf("Find returned true for non-existent key")
		}
		if v != nil {
			t.Errorf("Find returned non-nil value for non-existent key")
		}
	})

	t.Run("Update existing key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("foo")
		tree.Save(key, "old")

		ok := tree.Save(key, "new")
		if !ok {
			t.Errorf("Save returned false when updating existing key")
		}

		v, found := tree.Find(key)
		if !found {
			t.Errorf("Find failed after update")
		}
		if *v != "new" {
			t.Errorf("After update, value = %q, expected %q", *v, "new")
		}
	})

	t.Run("Delete existing key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("foo")
		tree.Save(key, "bar")

		ok := tree.Delete(key)
		if !ok {
			t.Errorf("Delete returned false for existing key")
		}

		v, found := tree.Find(key)
		if found {
			t.Errorf("Find returned true for deleted key")
		}
		if v != nil {
			t.Errorf("Find returned non-nil value for deleted key")
		}
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("nonexistent")

		ok := tree.Delete(key)
		if ok {
			t.Errorf("Delete returned true for non-existent key")
		}
	})

	t.Run("Multiple keys with common prefix", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		keys := []StringKey{"cat", "car", "cart", "dog"}
		values := []string{"meow", "vroom", "cartvalue", "woof"}

		for i, k := range keys {
			tree.Save(k, values[i])
		}

		// Verify all exist
		for i, k := range keys {
			v, found := tree.Find(k)
			if !found {
				t.Errorf("Key %q not found after insertion", k)
			} else if *v != values[i] {
				t.Errorf("Key %q has value %q, expected %q", k, *v, values[i])
			}
		}

		// Delete one with common prefix
		ok := tree.Delete(keys[1]) // "car"
		if !ok {
			t.Errorf("Delete of %q failed", keys[1])
		}

		// Ensure it's gone
		_, found := tree.Find(keys[1])
		if found {
			t.Errorf("Key %q still found after deletion", keys[1])
		}

		// Ensure others remain
		for i, k := range keys {
			if i == 1 {
				continue
			}
			v, found := tree.Find(k)
			if !found {
				t.Errorf("Key %q missing after deleting another key", k)
			} else if *v != values[i] {
				t.Errorf("Key %q has value %q, expected %q", k, *v, values[i])
			}
		}
	})

	t.Run("Empty key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("")
		val := "empty"

		ok := tree.Save(key, val)
		if !ok {
			t.Errorf("Save empty key failed")
		}

		v, found := tree.Find(key)
		if !found {
			t.Errorf("Find empty key failed")
		}
		if *v != val {
			t.Errorf("Empty key value mismatch: got %q, want %q", *v, val)
		}

		ok = tree.Delete(key)
		if !ok {
			t.Errorf("Delete empty key failed")
		}

		v, found = tree.Find(key)
		if found {
			t.Errorf("Empty key still found after delete")
		}
	})

	t.Run("Long key", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("thisisaverylongkeythatmightcauseproblemsiftheradixtreeisnotimplementedproperly")
		val := "longvalue"

		tree.Save(key, val)
		v, found := tree.Find(key)
		if !found {
			t.Errorf("Long key not found")
		}
		if *v != val {
			t.Errorf("Long key value mismatch: got %q, want %q", *v, val)
		}
	})

	t.Run("Overwrite and delete", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key := StringKey("overwrite")
		val1 := "first"
		val2 := "second"

		tree.Save(key, val1)
		tree.Save(key, val2) // overwrite

		v, found := tree.Find(key)
		if !found {
			t.Errorf("Key not found after overwrite")
		}
		if *v != val2 {
			t.Errorf("After overwrite, value = %q, expected %q", *v, val2)
		}

		tree.Delete(key)
		v, found = tree.Find(key)
		if found {
			t.Errorf("Key found after delete")
		}
	})

	t.Run("Delete all keys", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		keys := []StringKey{"a", "b", "c"}
		for _, k := range keys {
			tree.Save(k, "value")
		}

		for _, k := range keys {
			ok := tree.Delete(k)
			if !ok {
				t.Errorf("Delete of %q failed", k)
			}
		}

		for _, k := range keys {
			v, found := tree.Find(k)
			if found {
				t.Errorf("Key %q still present after deletion", k)
			}
			if v != nil {
				t.Errorf("Find returned non-nil for deleted key %q", k)
			}
		}
	})

	t.Run("Interleaved operations", func(t *testing.T) {
		tree := NewRadixTree[StringKey, string]()
		key1 := StringKey("key1")
		key2 := StringKey("key2")
		key3 := StringKey("key3")

		tree.Save(key1, "val1")
		tree.Save(key2, "val2")
		tree.Delete(key1)
		tree.Save(key3, "val3")
		tree.Save(key1, "newval1") // reinsert

		check := func(k StringKey, expected string) {
			v, found := tree.Find(k)
			if !found {
				t.Errorf("Key %q not found", k)
			} else if *v != expected {
				t.Errorf("Key %q = %q, want %q", k, *v, expected)
			}
		}

		check(key1, "newval1")
		check(key2, "val2")
		check(key3, "val3")
	})
}
func TestSave(t *testing.T) {
	tree := NewRadixTree[StringKey, string]()
	keys := []StringKey{"cat", "car", "cart", "dog"}
	values := []string{"meow", "vroom", "cartvalue", "woof"}

	for i, k := range keys {
		tree.Save(k, values[i])
	}

	// Verify all exist
	for i, k := range keys {
		v, found := tree.Find(k)
		if !found {
			t.Errorf("Key %q not found after insertion", k)
		} else if *v != values[i] {
			t.Errorf("Key %q has value %q, expected %q", k, *v, values[i])
		}
	}

	// Delete one with common prefix
	ok := tree.Delete(keys[1]) // "car"
	if !ok {
		t.Errorf("Delete of %q failed", keys[1])
	}

	// Ensure it's gone
	_, found := tree.Find(keys[1])
	if found {
		t.Errorf("Key %q still found after deletion", keys[1])
	}

	// Ensure others remain
	for i, k := range keys {
		if i == 1 {
			continue
		}
		v, found := tree.Find(k)
		if !found {
			t.Errorf("Key %q missing after deleting another key", k)
		} else if *v != values[i] {
			t.Errorf("Key %q has value %q, expected %q", k, *v, values[i])
		}
	}
}

// TestSafeRadixTree_ConcurrentInserts 测试多协程并发插入 + 最终一致性验证
// 每个协程插入独立范围的 key，避免冲突，但仍测试树内部锁的线程安全。
func TestSafeRadixTree_ConcurrentInserts(t *testing.T) {
	tree := NewSafeRadixTree[StringKey, string]()

	const numGoroutines = 20
	const keysPerGoroutine = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				keyStr := strconv.Itoa(goroutineID*keysPerGoroutine + j)
				key := StringKey(keyStr)
				tree.Save(key, "value_"+keyStr)
			}
		}(i)
	}

	wg.Wait()

	// 验证所有 key 均能正确查找到
	total := numGoroutines * keysPerGoroutine
	t.Attr(strconv.FormatInt(tree.size, 10), strconv.Itoa(total))
	for i := 0; i < total; i++ {
		keyStr := strconv.Itoa(i)
		key := StringKey(keyStr)
		v, ok := tree.Find(key)
		if !ok {
			t.Errorf("key %s 未找到", keyStr)
			continue
		}
		if *v != "value_"+keyStr {
			t.Errorf("key %s 值不匹配，期望 value_%s，实际 %s", keyStr, keyStr, *v)
		}
	}
	t.Attr(strconv.FormatInt(tree.size, 10), strconv.Itoa(total))

}

// TestSafeRadixTree_ConcurrentReadWrite 测试插入协程 + 读取协程并发执行（读写混合压力）
// 读取协程会在插入进行中不断调用 Find（可能部分 key 还未插入），仅在结束后做最终一致性检查。
func TestSafeRadixTree_ConcurrentReadWrite(t *testing.T) {
	tree := NewSafeRadixTree[StringKey, string]()

	const numWriters = 10
	const numReaders = 10
	const opsPerWriter = 100
	var wg sync.WaitGroup

	// Writers：并发插入
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWriter; j++ {
				keyStr := strconv.Itoa(writerID*opsPerWriter + j)
				key := StringKey(keyStr)
				tree.Save(key, "write_value_"+keyStr)
			}
		}(i)
	}

	// Readers：并发读取（压力测试读路径）
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			maxKey := numWriters * opsPerWriter
			for j := 0; j < opsPerWriter*numWriters*2; j++ { // 多读几次
				keyStr := strconv.Itoa(j % maxKey)
				key := StringKey(keyStr)
				_, _ = tree.Find(key) // 只调用，不在过程中断言（允许中间状态）
			}
		}()
	}

	wg.Wait()

	// 最终一致性验证
	maxKey := numWriters * opsPerWriter
	for i := 0; i < maxKey; i++ {
		keyStr := strconv.Itoa(i)
		key := StringKey(keyStr)
		v, ok := tree.Find(key)
		if !ok {
			t.Errorf("key %s 最终未找到", keyStr)
			continue
		}
		if *v != "write_value_"+keyStr {
			t.Errorf("key %s 值不匹配", keyStr)
		}
	}
}

// TestSafeRadixTree_ConcurrentDeleteAndRead 测试先顺序插入全部 key，再并发删除 + 读取
// 删除仅针对前半部分 key，后半部分保留，用于最终状态断言（验证 Delete 的线程安全与正确性）。
func TestSafeRadixTree_ConcurrentDeleteAndRead(t *testing.T) {
	tree := NewSafeRadixTree[StringKey, string]()

	const totalKeys = 500
	// Step 1: 顺序插入所有 key（建立基准状态）
	for i := 0; i < totalKeys; i++ {
		keyStr := strconv.Itoa(i)
		key := StringKey(keyStr)
		tree.Save(key, "val_"+keyStr)
	}

	const numDeleters = 10
	const numReaders = 10
	const opsPerGoroutine = 100
	var wg sync.WaitGroup

	// Deleters：并发删除前 250 个 key（每 key 会被删除多次，确保最终被删除）
	for i := 0; i < numDeleters; i++ {
		wg.Add(1)
		go func(deleterID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				keyIdx := (deleterID*opsPerGoroutine + j) % 250
				keyStr := strconv.Itoa(keyIdx)
				key := StringKey(keyStr)
				tree.Delete(key)
			}
		}(i)
	}

	// Readers：并发读取全部 key（压力测试读路径）
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine*totalKeys; j++ {
				keyIdx := j % totalKeys
				keyStr := strconv.Itoa(keyIdx)
				key := StringKey(keyStr)
				_, _ = tree.Find(key)
			}
		}()
	}

	wg.Wait()

	// 最终状态验证：
	// 0~249 应该全部被删除
	for i := 0; i < 250; i++ {
		keyStr := strconv.Itoa(i)
		key := StringKey(keyStr)
		_, ok := tree.Find(key)
		if ok {
			t.Errorf("key %d 应该已被删除，但仍存在", i)
		}
	}
	// 250~499 应该仍然存在
	for i := 250; i < totalKeys; i++ {
		keyStr := strconv.Itoa(i)
		key := StringKey(keyStr)
		v, ok := tree.Find(key)
		if !ok {
			t.Errorf("key %d 应该仍然存在，但未找到", i)
			continue
		}
		if *v != "val_"+keyStr {
			t.Errorf("key %d 值不匹配", i)
		}
	}
}
