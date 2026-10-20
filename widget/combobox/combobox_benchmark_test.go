package combobox

import (
	"fmt"
	"testing"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func TestComboBoxDerivedItemKeyCache(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	state := &comboBoxState{}
	first := state.derivedItemKey(ctx, "combo", "one")
	if got := state.derivedItemKey(ctx, "combo", "one"); got != first {
		t.Fatalf("cached item key = %q, want %q", got, first)
	}
	if got := state.derivedItemKey(ctx, "other", "one"); got == first {
		t.Fatalf("base change reused stale item key %q", got)
	}
}

// BenchmarkComboBoxOpenSteadyState measures a repeated open popup frame after
// item validation, style declarations, and item states have been warmed up.
func BenchmarkComboBoxOpenSteadyState(b *testing.B) {
	items := make([]Item, 100)
	for i := range items {
		items[i] = Item{Key: fmt.Sprintf("key%d", i), Label: fmt.Sprintf("Option %d", i)}
	}
	combo := New(items).Key("bench-open").SelectedKey("key0").DataVersion(1)
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	layoutComboBoxFrame(ctx, router, combo)
	state := testComponentState[comboBoxState](ctx, "bench-open", stateSlotComboBox)
	if state == nil {
		b.Fatal("combobox state was not initialized")
	}
	state.open = true
	layoutComboBoxFrame(ctx, router, combo)
	baseline := state.styleCache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutComboBoxFrame(ctx, router, combo)
	}
	b.StopTimer()
	reportComboDeclarationCacheStats(b, state.styleCache.Stats(), baseline)
}

func reportComboDeclarationCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "combo-declaration-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "combo-declaration-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "combo-declaration-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "combo-declaration-cache-key-invalidations/op")
}

// BenchmarkComboBoxProviderDeclarationCacheHit measures the warmed provider
// declaration path independently from Gio layout and item interaction.
func BenchmarkComboBoxProviderDeclarationCacheHit(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	var cache computed.ProviderCache[field.Variant, comboDeclarationBundle]
	comboStyleDeclarationsCached(&activeTheme, field.Primary, &cache)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		comboStyleDeclarationsCached(&activeTheme, field.Primary, &cache)
	}
	b.StopTimer()
	reportComboDeclarationCacheStats(b, cache.Stats(), computed.ProviderCacheStats{})
}

func BenchmarkComboBoxProviderDeclarationCacheInvalidation(b *testing.B) {
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()
	var cache computed.ProviderCache[field.Variant, comboDeclarationBundle]
	comboStyleDeclarationsCached(&light, field.Primary, &cache)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			comboStyleDeclarationsCached(&light, field.Primary, &cache)
		case 1:
			comboStyleDeclarationsCached(&dark, field.Primary, &cache)
		case 2:
			comboStyleDeclarationsCached(&dark, field.Secondary, &cache)
		default:
			comboStyleDeclarationsCached(&dark, field.Primary, &cache)
		}
	}
	b.StopTimer()
	reportComboDeclarationCacheStats(b, cache.Stats(), computed.ProviderCacheStats{})
}

// Benchmark for #15 - visible items caching optimization

// BenchmarkVisibleItems_CacheHit benchmarks the cache hit path
func BenchmarkVisibleItems_CacheHit(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]Item, 1000)
	for i := range items {
		items[i] = Item{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := New(items).Key("bench").SelectedKey("key0")

	// Populate cache
	_ = s.visibleItems(widget, "", "Option 0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Same query and selection should hit cache
		_ = s.visibleItems(widget, "", "Option 0")
	}
}

// BenchmarkVisibleItems_Filtering benchmarks the filtering logic
func BenchmarkVisibleItems_Filtering(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]Item, 1000)
	for i := range items {
		items[i] = Item{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := New(items).Key("bench").SelectedKey("key0").InputValue("Option 1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Query that matches a subset (forces full scan and filtering)
		_ = s.visibleItems(widget, "Option 1", "Option 0")
	}
}

// BenchmarkVisibleItems_NoFilter benchmarks when all items are visible
func BenchmarkVisibleItems_NoFilter(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]Item, 1000)
	for i := range items {
		items[i] = Item{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := New(items).Key("bench").SelectedKey("key0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Empty query shows all items
		_ = s.visibleItems(widget, "", "Option 0")
	}
}

// BenchmarkVisibleItems_CaseInsensitiveMatch benchmarks case-insensitive filtering
func BenchmarkVisibleItems_CaseInsensitiveMatch(b *testing.B) {
	s := &comboBoxState{}

	// Create mixed-case dataset
	items := make([]Item, 1000)
	for i := range items {
		items[i] = Item{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := New(items).Key("bench").SelectedKey("key0").InputValue("option")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Case-insensitive query (forces strings.ToLower on each comparison)
		_ = s.visibleItems(widget, "option", "Option 0")
	}
}
