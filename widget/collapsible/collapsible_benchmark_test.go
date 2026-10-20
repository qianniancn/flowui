package collapsible

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/widget/text"
)

type collapsibleBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newCollapsibleBenchmarkFrame(count int) *collapsibleBenchmarkFrame {
	active := *testtheme.Behavior()
	frame := &collapsibleBenchmarkFrame{ctx: core.New(nil, &active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New("Section "+strconv.Itoa(index), text.New("Additional details are available.")).Key("collapsible-" + strconv.Itoa(index)).DefaultExpanded(index%5 == 0)
	}
	return frame
}

func (frame *collapsibleBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	dimensions := value.Layout(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	return dimensions
}

func (frame *collapsibleBenchmarkFrame) layoutTree() layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	children := make([]layout.FlexChild, len(frame.values))
	for index := range frame.values {
		value := frame.values[index]
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(frame.ctx, gtx)
		})
	}
	dimensions := layout.Flex{Axis: layout.Vertical}.Layout(frame.gtx, children...)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	return dimensions
}

func (frame *collapsibleBenchmarkFrame) cacheStats() (hits, misses, themeInvalidations, keyInvalidations uint64) {
	for index := range frame.values {
		state, ok := core.PeekState[collapsibleState](frame.ctx, frame.values[index].key, stateSlotCollapsible)
		if !ok {
			continue
		}
		stats := state.visualCache.Stats()
		hits += stats.Hits
		misses += stats.Misses
		themeInvalidations += stats.ThemeInvalidations
		keyInvalidations += stats.KeyInvalidations
	}
	return hits, misses, themeInvalidations, keyInvalidations
}

func reportCollapsibleCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "collapsible-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "collapsible-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "collapsible-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "collapsible-visual-cache-key-invalidations/op")
}

func BenchmarkCollapsibleLayoutSteadyState(b *testing.B) {
	frame := newCollapsibleBenchmarkFrame(1)
	frame.layout(frame.values[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(frame.values[0])
	}
}

func BenchmarkCollapsibleComponentTreeSteadyState(b *testing.B) {
	frame := newCollapsibleBenchmarkFrame(64)
	frame.layoutTree()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkCollapsibleComponentTreeVisualCacheHit(b *testing.B) {
	frame := newCollapsibleBenchmarkFrame(64)
	frame.layoutTree()
	before := [4]uint64{}
	before[0], before[1], before[2], before[3] = frame.cacheStats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
	b.StopTimer()
	after := [4]uint64{}
	after[0], after[1], after[2], after[3] = frame.cacheStats()
	reportCollapsibleCacheStats(b, after, before)
}

func BenchmarkCollapsibleComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newCollapsibleBenchmarkFrame(64)
	frame.layoutTree()
	before := [4]uint64{}
	before[0], before[1], before[2], before[3] = frame.cacheStats()
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.layoutTree()
	}
	b.StopTimer()
	after := [4]uint64{}
	after[0], after[1], after[2], after[3] = frame.cacheStats()
	reportCollapsibleCacheStats(b, after, before)
}
