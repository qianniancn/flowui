package core

import (
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
	optional := cache.ResolveOptional(tag, key.ModShift, key.NameHome, key.NameEnd)
	optionalAgain := cache.ResolveOptional(tag, key.ModShift, key.NameHome, key.NameEnd)
	if len(optionalAgain) != 2 || &optional[0] != &optionalAgain[0] {
		t.Fatal("optional key filter cache did not reuse its filters")
	}
	optionalFilter, ok := optionalAgain[0].(key.Filter)
	if !ok || optionalFilter.Optional != key.ModShift {
		t.Fatalf("optional filter = %#v", optionalAgain[0])
	}
	changed := cache.Resolve(tag, key.NameEscape)
	if len(changed) != 1 || changed[0].(key.Filter).Name != key.NameEscape {
		t.Fatalf("changed filters = %#v", changed)
	}
}
