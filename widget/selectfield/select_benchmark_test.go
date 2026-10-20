package selectfield

import (
	"fmt"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BenchmarkSelectOpenSteadyState measures repeated open popup frames after
// flattened item data, style declarations, and list state are warmed up.
func BenchmarkSelectOpenSteadyState(b *testing.B) {
	items := make([]SelectItem, 100)
	for i := range items {
		items[i] = SelectItem{Key: fmt.Sprintf("key%d", i), Label: fmt.Sprintf("Option %d", i)}
	}
	selectWidget := New(items).Key("bench-open").SelectedKey("key0").
		DefaultOpen(true).
		DataVersion(1)
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	now := time.Unix(1, 0)
	layoutSelectTestFrame(ctx, router, selectWidget, now)
	layoutSelectTestFrame(ctx, router, selectWidget, now.Add(time.Millisecond))
	state := testComponentState[selectState](ctx, "bench-open", stateSlotSelect)
	if state == nil {
		b.Fatal("select state was not initialized")
	}
	baseline := state.styleCache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutSelectTestFrame(ctx, router, selectWidget, now.Add(time.Duration(i+2)*time.Millisecond))
	}
	b.StopTimer()
	reportSelectDeclarationCacheStats(b, state.styleCache.Stats(), baseline)
}

func reportSelectDeclarationCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "select-declaration-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "select-declaration-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "select-declaration-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "select-declaration-cache-key-invalidations/op")
}

// BenchmarkSelectProviderDeclarationCacheHit isolates the warmed Select
// Provider declaration path from layout, input routing, and Part resolution.
func BenchmarkSelectProviderDeclarationCacheHit(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	var cache computed.ProviderCache[SelectVariant, selectDeclarationBundle]
	selectStyleDeclarationsCached(&activeTheme, SelectPrimary, &cache)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = selectStyleDeclarationsCached(&activeTheme, SelectPrimary, &cache)
	}
	b.StopTimer()
	reportSelectDeclarationCacheStats(b, cache.Stats(), computed.ProviderCacheStats{})
}

func BenchmarkSelectProviderDeclarationCacheInvalidation(b *testing.B) {
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()
	var cache computed.ProviderCache[SelectVariant, selectDeclarationBundle]
	selectStyleDeclarationsCached(&light, SelectPrimary, &cache)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			selectStyleDeclarationsCached(&light, SelectPrimary, &cache)
		case 1:
			selectStyleDeclarationsCached(&dark, SelectPrimary, &cache)
		case 2:
			selectStyleDeclarationsCached(&dark, SelectSecondary, &cache)
		default:
			selectStyleDeclarationsCached(&dark, SelectPrimary, &cache)
		}
	}
	b.StopTimer()
	reportSelectDeclarationCacheStats(b, cache.Stats(), computed.ProviderCacheStats{})
}

// BenchmarkSelectPartResolveCacheHit measures the warmed static cascade path
// for the Select trigger's compound Parts. Transitions and Gio layout are
// intentionally excluded so the cache-hit cost remains independently visible.
func BenchmarkSelectPartResolveCacheHit(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declarations := selectStyleDeclarationsCached(&activeTheme, SelectPrimary)
	state := style.State{Selected: true, Open: true}
	parts := [...]style.Part{
		theme.SelectPartTrigger,
		theme.SelectPartPlaceholder,
		theme.SelectPartSelection,
		theme.SelectPartIndicator,
		theme.SelectPartIcon,
		theme.SelectPartDescription,
	}
	for _, part := range parts {
		computed.ResolvePartStatic(ctx, part, state, declarations.visual.Defaults, style.Style{}, style.Style{}, style.Style{})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, part := range parts {
			_ = computed.ResolvePartStatic(ctx, part, state, declarations.visual.Defaults, style.Style{}, style.Style{}, style.Style{})
		}
	}
}

// BenchmarkSelectKeyedPartResolveCacheHit measures the animation-aware
// compound Part path used by the real trigger layout after its keyed states
// have settled. Settled transition declarations should share their immutable
// cascade branches instead of cloning them on every frame.
func BenchmarkSelectKeyedPartResolveCacheHit(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	declarations := selectStyleDeclarationsCached(&activeTheme, SelectPrimary)
	state := style.State{Selected: true, Open: true}
	parts := [...]style.Part{
		theme.SelectPartTrigger,
		theme.SelectPartPlaceholder,
		theme.SelectPartSelection,
		theme.SelectPartIndicator,
		theme.SelectPartIcon,
		theme.SelectPartDescription,
	}
	keys := [len(parts)]string{}
	for index, part := range parts {
		keys[index] = "select-benchmark/" + string(part)
	}
	gtx := testLayoutContext()
	gtx.Now = time.Unix(1, 0)
	resolve := func() {
		for index, part := range parts {
			_ = computed.ResolvePartWithKey(ctx, gtx, keys[index], part, state, declarations.visual.Defaults, style.Style{}, style.Style{}, style.Style{})
		}
	}
	core.BeginFrame(ctx)
	resolve()
	core.EndFrame(ctx)
	core.BeginFrame(ctx)
	resolve()
	core.EndFrame(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		gtx.Now = gtx.Now.Add(time.Millisecond)
		core.BeginFrame(ctx)
		resolve()
		core.EndFrame(ctx)
	}
}
