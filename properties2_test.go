package properties

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestNewHashtable2(t *testing.T) {
	table := NewHashtable2()
	table.Put("k1", "v1")
	table.Put("k2", 2)
	table.Put("k3", 'a')
	table.Put("k4", []byte("---"))
	table.Put("k5", struct{}{})

	diff := cmp.Diff(table.Keys(), []interface{}{"k1", "k2", "k3", "k4", "k5"})
	if diff != "" {
		t.Fatal(diff)
	}

	table.Remove("k1")
	table.Remove("k2")
	table.Remove("k3")
	table.Remove("k4")
	table.Remove("k5")
	diff = cmp.Diff(table.Keys(), []interface{}{})
	if diff != "" {
		t.Fatal(diff)
	}

	diff = cmp.Diff(table.Size(), 0)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestNewProperties2(t *testing.T) {
	prop := NewProperties2()
	prop.SetProperty("k1", "")
	prop.SetProperty("k2", "")
	prop.SetProperty("k3", "")
	prop.SetProperty("k4", "")
	prop.SetProperty("k5", "")
	diff := cmp.Diff(prop.Keys(), []interface{}{"k1", "k2", "k3", "k4", "k5"})
	if diff != "" {
		t.Fatal(diff)
	}
}
