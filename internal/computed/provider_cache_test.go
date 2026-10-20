package computed

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type testProviderCache interface {
	Resolve(*theme.Theme, int, func() int) int
	Stats() ProviderCacheStats
}

func providerCacheCases() []struct {
	name string
	new  func() testProviderCache
} {
	return []struct {
		name string
		new  func() testProviderCache
	}{
		{name: "single", new: func() testProviderCache { return &ProviderCache[int, int]{} }},
		{name: "map", new: func() testProviderCache { return &ProviderCacheMap[int, int]{} }},
		{name: "bounded map", new: func() testProviderCache {
			cache := &ProviderCacheMap[int, int]{}
			cache.SetCapacity(2)
			return cache
		}},
	}
}

type cacheTestProvider struct {
	theme.ComponentVisualProvider
	identity any
}

type cacheSliceProvider struct {
	theme.ComponentVisualProvider
	identity []int
}

type cacheButtonOnlyProvider struct{}

func (cacheButtonOnlyProvider) Button(*theme.Theme, theme.ButtonRequest) theme.ButtonVisual {
	return theme.ButtonVisual{}
}

func TestProviderCachesInvalidateOnProviderReplacement(t *testing.T) {
	for _, test := range providerCacheCases() {
		t.Run(test.name, func(t *testing.T) {
			cache := test.new()
			active := *testtheme.Behavior()
			loads := 0
			load := func() int { loads++; return loads }
			cache.Resolve(&active, 1, load)
			cache.Resolve(&active, 1, load)
			cache.Resolve(&active, 2, load)

			active.VisualProvider = cacheTestProvider{testtheme.BehaviorProvider(), 1}
			if got := cache.Resolve(&active, 1, load); got != 3 {
				t.Fatalf("provider replacement returned stale value %d, want 3", got)
			}
			if got := cache.Resolve(&active, 1, load); got != 3 {
				t.Fatalf("stable provider returned %d, want cached 3", got)
			}
			if got := cache.Resolve(&active, 2, load); got != 4 {
				t.Fatalf("old provider's second key survived replacement: %d", got)
			}
			stats := cache.Stats()
			if stats.Hits != 2 || stats.Misses != 4 || stats.ThemeInvalidations != 1 {
				t.Fatalf("replacement stats = %#v", stats)
			}
		})
	}
}

func TestProviderCachesDoNotReuseNonComparableProviders(t *testing.T) {
	for _, test := range providerCacheCases() {
		t.Run(test.name, func(t *testing.T) {
			for _, provider := range []struct {
				name  string
				value theme.VisualProvider
			}{
				{name: "slice field", value: cacheSliceProvider{testtheme.BehaviorProvider(), []int{1}}},
				{name: "interface containing slice", value: cacheTestProvider{testtheme.BehaviorProvider(), []int{1}}},
			} {
				t.Run(provider.name, func(t *testing.T) {
					cache := test.new()
					active := *testtheme.Behavior()
					active.VisualProvider = cacheTestProvider{testtheme.BehaviorProvider(), 1}
					loads := 0
					load := func() int { loads++; return loads }
					cache.Resolve(&active, 1, load)
					active.VisualProvider = provider.value
					cache.Resolve(&active, 1, load)
					if got := cache.Resolve(&active, 1, load); got != 3 {
						t.Fatalf("non-comparable provider reused cached value %d", got)
					}
					active.VisualProvider = cacheTestProvider{testtheme.BehaviorProvider(), 2}
					cache.Resolve(&active, 1, load)
					if got := cache.Resolve(&active, 1, load); got != 4 {
						t.Fatalf("comparable replacement returned %d, want cached 4", got)
					}
				})
			}
		})
	}
}

func TestProviderCachesRejectInvalidReplacementWithoutDiscardingPreviousValue(t *testing.T) {
	for _, test := range providerCacheCases() {
		t.Run(test.name, func(t *testing.T) {
			for _, replacement := range []struct {
				name     string
				provider theme.VisualProvider
				want     error
			}{
				{name: "removed", want: theme.ErrMissingVisualProvider},
				{name: "incomplete", provider: cacheButtonOnlyProvider{}, want: theme.ErrIncompleteVisualProvider},
			} {
				t.Run(replacement.name, func(t *testing.T) {
					cache := test.new()
					active := *testtheme.Behavior()
					original := active.VisualProvider
					loads := 0
					load := func() int {
						if _, err := theme.ComponentProvider(&active); err != nil {
							panic(err)
						}
						loads++
						return loads
					}
					cache.Resolve(&active, 1, load)
					active.VisualProvider = replacement.provider
					func() {
						defer func() {
							err, ok := recover().(error)
							if !ok || !errors.Is(err, replacement.want) {
								t.Fatalf("replacement panic = %v, want %v", err, replacement.want)
							}
						}()
						cache.Resolve(&active, 1, load)
					}()
					active.VisualProvider = original
					if got := cache.Resolve(&active, 1, load); got != 1 || loads != 1 {
						t.Fatalf("failed replacement discarded previous entry: value=%d, loads=%d", got, loads)
					}
				})
			}
		})
	}
}

func TestProviderCacheResolvesAndInvalidates(t *testing.T) {
	first := theme.NeutralTheme()
	second := theme.NeutralTheme()
	cache := ProviderCache[int, string]{}
	loads := 0
	load := func() string {
		loads++
		return "value"
	}

	if got := cache.Resolve(&first, 1, load); got != "value" {
		t.Fatalf("first resolve = %q, want value", got)
	}
	if got := cache.Resolve(&first, 1, load); got != "value" {
		t.Fatalf("cached resolve = %q, want value", got)
	}
	cache.Resolve(&first, 2, load)
	cache.Resolve(&second, 2, load)
	stats := cache.Stats()
	if loads != 3 {
		t.Fatalf("loader calls = %d, want 3", loads)
	}
	if stats.Hits != 1 || stats.Misses != 3 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %#v", stats)
	}
}

func TestProviderCachePanicDoesNotCommit(t *testing.T) {
	active := theme.NeutralTheme()
	cache := ProviderCache[int, string]{}
	cache.Resolve(&active, 1, func() string { return "stable" })

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected loader panic")
			}
		}()
		cache.Resolve(&active, 2, func() string { panic("provider failure") })
	}()

	if got := cache.Resolve(&active, 1, func() string { return "unexpected" }); got != "stable" {
		t.Fatalf("value after loader panic = %q, want stable", got)
	}
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.KeyInvalidations != 1 {
		t.Fatalf("panic cache stats = %#v", stats)
	}
}

func TestProviderCacheNilReceiverLoadsWithoutCaching(t *testing.T) {
	active := theme.NeutralTheme()
	var cache *ProviderCache[int, string]
	loads := 0
	load := func() string {
		loads++
		return "value"
	}
	cache.Resolve(&active, 1, load)
	cache.Resolve(&active, 1, load)
	if loads != 2 {
		t.Fatalf("nil cache loader calls = %d, want 2", loads)
	}
}

func TestProviderCacheMapResolvesAndInvalidates(t *testing.T) {
	first := theme.NeutralTheme()
	second := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	loads := 0
	load := func() string {
		loads++
		return "value"
	}

	cache.Resolve(&first, 1, load)
	cache.Resolve(&first, 1, load)
	cache.Resolve(&first, 2, load)
	cache.Resolve(&second, 1, load)
	stats := cache.Stats()
	if loads != 3 {
		t.Fatalf("loader calls = %d, want 3", loads)
	}
	if stats.Hits != 1 || stats.Misses != 3 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 1 {
		t.Fatalf("cache stats = %#v", stats)
	}
}

func TestProviderCacheMapResolveWithStatusReportsHits(t *testing.T) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	loads := 0
	load := func() string {
		loads++
		return "value"
	}

	if value, hit := cache.ResolveWithStatus(&active, 1, load); hit || value != "value" {
		t.Fatalf("first resolve = %q, hit=%v; want miss with value", value, hit)
	}
	if value, hit := cache.ResolveWithStatus(&active, 1, load); !hit || value != "value" {
		t.Fatalf("second resolve = %q, hit=%v; want hit with value", value, hit)
	}
	if loads != 1 {
		t.Fatalf("loader calls = %d, want 1", loads)
	}
}

func TestProviderCacheMapPanicDoesNotCommit(t *testing.T) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	cache.Resolve(&active, 1, func() string { return "stable" })

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected loader panic")
			}
		}()
		cache.Resolve(&active, 2, func() string { panic("provider failure") })
	}()

	if got := cache.Resolve(&active, 1, func() string { return "unexpected" }); got != "stable" {
		t.Fatalf("value after loader panic = %q, want stable", got)
	}
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.KeyInvalidations != 1 {
		t.Fatalf("panic cache stats = %#v", stats)
	}
}

func TestProviderCacheMapNilReceiverLoadsWithoutCaching(t *testing.T) {
	active := theme.NeutralTheme()
	var cache *ProviderCacheMap[int, string]
	loads := 0
	load := func() string {
		loads++
		return "value"
	}
	cache.Resolve(&active, 1, load)
	cache.Resolve(&active, 1, load)
	if loads != 2 {
		t.Fatalf("nil cache loader calls = %d, want 2", loads)
	}
}

func TestProviderCacheMapBoundedLRU(t *testing.T) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	cache.SetCapacity(2)
	loads := 0
	load := func(value string) func() string {
		return func() string {
			loads++
			return value
		}
	}

	cache.Resolve(&active, 1, load("one"))
	cache.Resolve(&active, 2, load("two"))
	cache.Resolve(&active, 1, load("unexpected"))
	cache.Resolve(&active, 3, load("three"))
	cache.Resolve(&active, 1, load("unexpected"))
	if loads != 3 {
		t.Fatalf("bounded cache loader calls = %d, want 3", loads)
	}
	stats := cache.Stats()
	if stats.Hits != 2 || stats.Misses != 3 || stats.Evictions != 1 || stats.Size != 2 {
		t.Fatalf("bounded cache stats = %#v", stats)
	}
}

func TestProviderCacheMapBoundedPanicDoesNotEvict(t *testing.T) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	cache.SetCapacity(1)
	cache.Resolve(&active, 1, func() string { return "stable" })

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected loader panic")
			}
		}()
		cache.Resolve(&active, 2, func() string { panic("provider failure") })
	}()

	if got := cache.Resolve(&active, 1, func() string { return "unexpected" }); got != "stable" {
		t.Fatalf("bounded value after loader panic = %q, want stable", got)
	}
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.Evictions != 0 || stats.Size != 1 {
		t.Fatalf("bounded panic cache stats = %#v", stats)
	}
}
