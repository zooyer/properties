package properties

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestNewHashtable(t *testing.T) {
	h := NewHashtable()
	h.Put("key1", "val1")
	h.Put("key2", 1000)
	h.Put("key3", 'a')
	diff := cmp.Diff(h.Size(), 3)
	if diff != "" {
		t.Fatal(diff)
	}
	t.Log("size:", h.Size())
	diff = cmp.Diff(len(h.Keys()), 3)
	if diff != "" {
		t.Fatal(diff)
	}
	t.Log(h.Keys())
	diff = cmp.Diff(h.Get("key1"), "val1")
	if diff != "" {
		t.Fatal(diff)
	}
	diff = cmp.Diff(h.Get("key2"), 1000)
	if diff != "" {
		t.Fatal(diff)
	}
	diff = cmp.Diff(h.Get("key3"), 'a')
	if diff != "" {
		t.Fatal(diff)
	}
	h.Remove("key2")
	diff = cmp.Diff(h.Size(), 2)
	if diff != "" {
		t.Fatal(diff)
	}
	t.Log("remove key2 size:", h.Size())
	t.Log(h.Keys())
}
