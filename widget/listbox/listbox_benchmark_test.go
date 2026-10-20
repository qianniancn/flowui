package listbox

import (
	"fmt"
	"image"
	"runtime"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkListBoxLargeData(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New(items).Key("large")
	for _, benchmark := range []struct {
		name   string
		widget Widget
	}{
		{name: "unversioned", widget: widget},
		{name: "versioned", widget: widget.DataVersion(1)},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			state := new(listBoxState)
			state.resolveEntries(benchmark.widget)
			b.ReportAllocs()
			for b.Loop() {
				entries, resolved := state.resolveEntries(benchmark.widget)
				runtime.KeepAlive(entries)
				runtime.KeepAlive(resolved)
			}
		})
	}
}

func BenchmarkListBoxProviderDeclaration(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	provider, err := theme.ComponentProvider(&activeTheme)
	if err != nil {
		b.Fatal(err)
	}
	request := theme.ListBoxRequest{}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = provider.ListBox(&activeTheme, request)
	}
}

func BenchmarkListBoxVisualCacheHit(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	state := new(listBoxState)
	state.listBoxVisualFor(ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = state.listBoxVisualFor(ctx)
	}
}

func BenchmarkListBoxLargeSelection(b *testing.B) {
	keys := make([]string, 10_000)
	for index := range keys {
		keys[index] = fmt.Sprintf("item-%d", index)
	}
	widget := New(nil).Key("large").SelectedKeys(keys)
	target := keys[len(keys)-1]
	state := new(listBoxState)
	b.ReportAllocs()
	for b.Loop() {
		widget.selectedKeySet = state.selectedKeys.Resolve(keys)
		runtime.KeepAlive(widget.isSelected(target))
	}
}

func BenchmarkListBoxLargeLayout(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New(items).Key("large").DataVersion(1)
	ctx := newContext(nil)
	b.ReportAllocs()
	for b.Loop() {
		gtx := testLayoutContext()
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		widget.Layout(ctx, gtx)
		core.EndFrame(ctx)
	}
}

func BenchmarkListBoxLargeLayoutSteadyState(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New(items).Key("large-steady").DataVersion(1)
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	var ops op.Ops
	viewport := image.Pt(300, 200)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
	}
	b.ReportAllocs()
	// Warm the real component tree before timing so the benchmark measures
	// steady-state Provider and item visual cache behavior.
	ops.Reset()
	core.BeginFrameWithViewport(ctx, viewport)
	widget.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	state := testComponentState[listBoxState](ctx, "large-steady", stateSlotListBox)
	if state == nil {
		b.Fatal("listbox state was not initialized")
	}
	baseline := state.itemVisualCache.Stats()
	resolvedBaseline := state.resolvedPartCache.stats()
	visualBaseline := state.visualCache.stats()
	b.ResetTimer()
	for b.Loop() {
		ops.Reset()
		core.BeginFrameWithViewport(ctx, viewport)
		widget.Layout(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(&ops)
	}
	stats := state.itemVisualCache.Stats()
	resolvedStats := state.resolvedPartCache.stats()
	if b.N > 0 {
		b.ReportMetric(float64(stats.Hits-baseline.Hits)/float64(b.N), "listbox-item-cache-hits/op")
		b.ReportMetric(float64(stats.Misses-baseline.Misses)/float64(b.N), "listbox-item-cache-misses/op")
		b.ReportMetric(float64(stats.Evictions-baseline.Evictions)/float64(b.N), "listbox-item-cache-evictions/op")
		b.ReportMetric(float64(stats.ThemeInvalidations-baseline.ThemeInvalidations)/float64(b.N), "listbox-item-cache-theme-invalidations/op")
		b.ReportMetric(float64(resolvedStats.Hits-resolvedBaseline.Hits)/float64(b.N), "listbox-resolved-cache-hits/op")
		b.ReportMetric(float64(resolvedStats.Misses-resolvedBaseline.Misses)/float64(b.N), "listbox-resolved-cache-misses/op")
		b.ReportMetric(float64(resolvedStats.ThemeInvalidations-resolvedBaseline.ThemeInvalidations)/float64(b.N), "listbox-resolved-cache-theme-invalidations/op")
		b.ReportMetric(float64(resolvedStats.ScopeInvalidations-resolvedBaseline.ScopeInvalidations)/float64(b.N), "listbox-resolved-cache-scope-invalidations/op")
		b.ReportMetric(float64(resolvedStats.OverrideInvalidations-resolvedBaseline.OverrideInvalidations)/float64(b.N), "listbox-resolved-cache-override-invalidations/op")
		visualStats := state.visualCache.stats()
		b.ReportMetric(float64(visualStats.Hits-visualBaseline.Hits)/float64(b.N), "listbox-visual-cache-hits/op")
		b.ReportMetric(float64(visualStats.Misses-visualBaseline.Misses)/float64(b.N), "listbox-visual-cache-misses/op")
		b.ReportMetric(float64(visualStats.ThemeInvalidations-visualBaseline.ThemeInvalidations)/float64(b.N), "listbox-visual-cache-theme-invalidations/op")
	}
}

func BenchmarkListBoxLargeDisabledLayout(b *testing.B) {
	items := make([]Item, 10_000)
	disabled := make([]string, len(items))
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
		disabled[index] = items[index].Key
	}
	widget := New(items).Key("large-disabled").DisabledKeys(disabled).DataVersion(1)
	ctx := newContext(nil)
	b.ReportAllocs()
	for b.Loop() {
		gtx := testLayoutContext()
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		widget.Layout(ctx, gtx)
		core.EndFrame(ctx)
	}
}
