package state

import (
	"testing"

	"gioui.org/io/key"
)

func BenchmarkFrameMap(b *testing.B) {
	b.Run("UseFrameMap", func(b *testing.B) {
		var values map[string]*int
		var seen map[string]struct{}
		BeginFrameMap(&seen)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = UseFrameMap(&values, &seen, "test-key")
		}
	})

	b.Run("EnsureFrameMap", func(b *testing.B) {
		var values map[string]*int
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = EnsureFrameMap(&values, "test-key")
		}
	})

	b.Run("SweepFrameMap", func(b *testing.B) {
		values := make(map[string]*int)
		seen := make(map[string]struct{})
		for i := range 100 {
			key := string(rune('a' + i%26))
			values[key] = new(int)
			if i%2 == 0 {
				seen[key] = struct{}{}
			}
		}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			SweepFrameMap(values, seen)
		}
	})
}

func BenchmarkStore(b *testing.B) {
	b.Run("Use", func(b *testing.B) {
		store := &Store{}
		id := Identity{Key: "component", Slot: "state"}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Use(store, id, func() *int {
				val := 42
				return &val
			})
		}
	})

	b.Run("Peek", func(b *testing.B) {
		store := &Store{}
		id := Identity{Key: "component", Slot: "state"}
		Use(store, id, func() *int {
			val := 42
			return &val
		})
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Peek[int](store, id)
		}
	})
}

func BenchmarkMemo(b *testing.B) {
	b.Run("CacheHit", func(b *testing.B) {
		store := &Store{}
		computeCount := 0
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			memo := UseMemo[int](store, "test", "memo", 1, 100)
			_ = memo.Compute(func() int {
				computeCount++
				return 42
			})
		}
		if computeCount != 1 {
			b.Errorf("expected 1 computation, got %d", computeCount)
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		store := &Store{}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			memo := UseMemo[int](store, "test", "memo", 1, uint64(i))
			_ = memo.Compute(func() int {
				return 42
			})
		}
	})

	b.Run("ThemeInvalidation", func(b *testing.B) {
		store := &Store{}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			memo := UseMemo[int](store, "test", "memo", uint64(i), 100)
			_ = memo.Compute(func() int {
				return 42
			})
		}
	})
}

func BenchmarkKeyFilterCache(b *testing.B) {
	type testTag struct{}
	tag := &testTag{}
	cache := &KeyFilterCache{}
	names := []key.Name{key.NameReturn, key.NameSpace, key.NameEscape}

	b.Run("CacheHit", func(b *testing.B) {
		cache.Resolve(tag, names...)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = cache.Resolve(tag, names...)
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		tags := make([]testTag, 100)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tag := &tags[i%len(tags)]
			_ = cache.Resolve(tag, names...)
		}
	})
}

func BenchmarkKeys(b *testing.B) {
	b.Run("Claim", func(b *testing.B) {
		keys := Keys{}
		keys.BeginFrame()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = keys.Claim(KindClickable, "button")
			keys.BeginFrame()
		}
	})

	b.Run("ClaimDerived", func(b *testing.B) {
		keys := Keys{}
		keys.BeginFrame()
		parent := keys.Claim(KindClickable, "parent")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = keys.ClaimDerivedResolved(KindCustom, parent, "child")
			keys.BeginFrame()
			parent = keys.Claim(KindClickable, "parent")
		}
	})

	b.Run("FullKey", func(b *testing.B) {
		keys := Keys{}
		defer keys.Push("parent")()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = keys.FullKey("child")
		}
	})
}

func BenchmarkFocus(b *testing.B) {
	type testTag struct{}
	tag1 := &testTag{}
	tag2 := &testTag{}

	b.Run("Request", func(b *testing.B) {
		focus := Focus{}
		focus.BeginFrame()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			focus.Request(tag1, FocusOriginKeyboard)
		}
	})

	b.Run("Observe", func(b *testing.B) {
		focus := Focus{}
		focus.BeginFrame()
		focus.Request(tag1, FocusOriginKeyboard)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = focus.Observe(tag1, true)
		}
	})

	b.Run("CommitObservations", func(b *testing.B) {
		focus := Focus{}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			focus.BeginFrame()
			focus.Request(tag1, FocusOriginKeyboard)
			focus.Observe(tag1, true)
			focus.Observe(tag2, false)
			focus.CommitObservations()
		}
	})
}
