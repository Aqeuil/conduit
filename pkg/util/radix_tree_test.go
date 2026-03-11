package util

import "testing"

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
