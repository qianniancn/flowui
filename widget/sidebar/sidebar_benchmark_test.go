package sidebar

import (
	"fmt"
	"image"
	"runtime"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type sidebarBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newSidebarBenchmarkFrame(maximum image.Point) *sidebarBenchmarkFrame {
	frame := &sidebarBenchmarkFrame{ctx: core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(), Ops: &frame.ops, Now: time.Unix(1, 0),
	}
	return frame
}

func (f *sidebarBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func sidebarBenchmarkTree() Widget {
	sections := make([]Section, 4)
	for sectionIndex := range sections {
		sections[sectionIndex].Title = fmt.Sprintf("Section %d", sectionIndex)
		sections[sectionIndex].Items = make([]Item, 8)
		for itemIndex := range sections[sectionIndex].Items {
			sections[sectionIndex].Items[itemIndex] = Item{
				Key: fmt.Sprintf("item-%d-%d", sectionIndex, itemIndex), Label: "Navigation item",
			}
		}
	}
	return NewSections("benchmark-tree", "item-0-0", sections).DataVersion(1)
}

func reportSidebarVisualStats(b *testing.B, after, before computed.ProviderCacheStats) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "sidebar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "sidebar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "sidebar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "sidebar-visual-cache-key-invalidations/op")
}

func BenchmarkSidebarProviderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	cache := sidebarVisualCache{}
	cache.resolve(active, theme.SidebarRequest{}, func() theme.SidebarVisual {
		provider, err := theme.ComponentProvider(active)
		if err != nil {
			b.Fatal(err)
		}
		visual := provider.Sidebar(active, theme.SidebarRequest{})
		theme.ValidateSidebarVisual(visual)
		return visual
	})
	before := cache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.resolve(active, theme.SidebarRequest{}, func() theme.SidebarVisual {
			b.Helper()
			panic("sidebar visual cache miss")
		})
	}
	b.StopTimer()
	reportSidebarVisualStats(b, cache.stats(), before)
}

func BenchmarkSidebarComponentTreeSteadyState(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSidebarComponentTreeVisualCacheHit(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	state, ok := core.PeekState[sidebarState](frame.ctx, "benchmark-tree", stateSlotSidebar)
	if !ok {
		b.Fatal("sidebar state was not initialized")
	}
	before := state.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportSidebarVisualStats(b, state.visualCache.stats(), before)
}

func BenchmarkSidebarComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := sidebarBenchmarkTree()
	frame.layout(value)
	state, ok := core.PeekState[sidebarState](frame.ctx, "benchmark-tree", stateSlotSidebar)
	if !ok {
		b.Fatal("sidebar state was not initialized")
	}
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	before := state.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, first)
		} else {
			core.ReplaceTheme(frame.ctx, second)
		}
		frame.layout(value)
	}
	b.StopTimer()
	reportSidebarVisualStats(b, state.visualCache.stats(), before)
}

func BenchmarkSidebarStyledOverrideBypass(b *testing.B) {
	frame := newSidebarBenchmarkFrame(image.Pt(320, 800))
	value := (sidebarBenchmarkTree()).Style(style.Style{}.PaddingX(2))
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSidebarItemVisualCacheHit(b *testing.B) {
	active := theme.Theme{}
	state := new(sidebarState)
	visual := theme.SidebarVisual{Item: func(theme.SidebarItemRequest) theme.SidebarItemVisual {
		return theme.SidebarItemVisual{Opacity: 1}
	}}
	state.itemVisualCache.resolve(&active, visual, theme.SidebarItemRequest{Selected: true})
	b.ReportAllocs()
	for b.Loop() {
		_ = state.itemVisualCache.resolve(&active, visual, theme.SidebarItemRequest{Selected: true})
	}
}

func BenchmarkSidebarItemVisualCacheMiss(b *testing.B) {
	active := theme.Theme{}
	state := new(sidebarState)
	visual := theme.SidebarVisual{Item: func(request theme.SidebarItemRequest) theme.SidebarItemVisual {
		return theme.SidebarItemVisual{Opacity: 1}
	}}
	b.ReportAllocs()
	for b.Loop() {
		state.itemVisualCache.valid = [32]bool{}
		_ = state.itemVisualCache.resolve(&active, visual, theme.SidebarItemRequest{Selected: true})
	}
}

func BenchmarkSidebarCloneSections(b *testing.B) {
	sections := make([]Section, 8)
	for sectionIndex := range sections {
		sections[sectionIndex].Title = fmt.Sprintf("Section %d", sectionIndex)
		sections[sectionIndex].Items = make([]Item, 8)
		for itemIndex := range sections[sectionIndex].Items {
			sections[sectionIndex].Items[itemIndex] = Item{
				Key:   fmt.Sprintf("item-%d-%d", sectionIndex, itemIndex),
				Label: "Item",
			}
		}
	}

	b.ReportAllocs()
	for b.Loop() {
		cloned := cloneSections(sections)
		runtime.KeepAlive(cloned)
	}
}
