package state

import (
	"runtime"
	"testing"

	"gioui.org/io/key"
)

func TestKeyFilterCacheReusesConcreteFilters(t *testing.T) {
	tag := new(int)
	var cache KeyFilterCache
	first := cache.Resolve(tag, key.NameHome, key.NameEnd)
	second := cache.Resolve(tag, key.NameHome, key.NameEnd)
	if len(second) != 2 || &first[0] != &second[0] {
		t.Fatal("key filter cache did not reuse its filters")
	}
	filter, ok := second[0].(key.Filter)
	if !ok || filter.Focus != tag || filter.Name != key.NameHome {
		t.Fatalf("first filter = %#v", second[0])
	}
	changed := cache.Resolve(tag, key.NameEscape)
	if len(changed) != 1 || changed[0].(key.Filter).Name != key.NameEscape {
		t.Fatalf("changed filters = %#v", changed)
	}
}

func BenchmarkKeyFilterCache(b *testing.B) {
	tag := new(int)
	var cache KeyFilterCache
	cache.Resolve(tag, key.NameHome, key.NameEnd)
	b.ReportAllocs()
	for b.Loop() {
		runtime.KeepAlive(cache.Resolve(tag, key.NameHome, key.NameEnd))
	}
}
