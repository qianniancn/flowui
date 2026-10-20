package dropdown

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type dropdownBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
	values []Widget
}

func newDropdownBenchmarkFrame(count int) *dropdownBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &dropdownBenchmarkFrame{
		ctx: core.New(nil, active, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New(
			"benchmark-dropdown-"+strconv.Itoa(index),
			fixedWidget{size: image.Pt(160, 36)},
			[]Item{{Key: "open", Label: "Open"}},
		)
	}
	return frame
}

func (frame *dropdownBenchmarkFrame) layout() layout.Dimensions {
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

func (frame *dropdownBenchmarkFrame) cacheStats() DropdownVisualCacheStats {
	var total DropdownVisualCacheStats
	for index := range frame.values {
		state, ok := core.PeekState[dropdownState](frame.ctx, frame.values[index].key, stateSlotDropdown)
		if !ok || state == nil {
			continue
		}
		stats := state.visualCache.stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportDropdownCacheStats(b *testing.B, stats DropdownVisualCacheStats) {
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "dropdown-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "dropdown-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "dropdown-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "dropdown-visual-cache-key-invalidations/op")
}

func dropdownCacheStatsDelta(after, before DropdownVisualCacheStats) DropdownVisualCacheStats {
	return DropdownVisualCacheStats{
		Hits:               after.Hits - before.Hits,
		Misses:             after.Misses - before.Misses,
		ThemeInvalidations: after.ThemeInvalidations - before.ThemeInvalidations,
		KeyInvalidations:   after.KeyInvalidations - before.KeyInvalidations,
	}
}

func BenchmarkDropdownLayoutSteadyState(b *testing.B) {
	frame := newDropdownBenchmarkFrame(1)
	frame.layout()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout()
	}
}

func BenchmarkDropdownComponentTreeSteadyState(b *testing.B) {
	frame := newDropdownBenchmarkFrame(128)
	frame.layout()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout()
	}
}

func BenchmarkDropdownVisualCacheHit(b *testing.B) {
	frame := newDropdownBenchmarkFrame(128)
	frame.layout()
	baseline := frame.cacheStats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout()
	}
	b.StopTimer()
	reportDropdownCacheStats(b, dropdownCacheStatsDelta(frame.cacheStats(), baseline))
}

func BenchmarkDropdownVisualCacheMiss(b *testing.B) {
	frame := newDropdownBenchmarkFrame(128)
	frame.layout()
	baseline := frame.cacheStats()
	light := *testtheme.Behavior()
	dark := light
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.layout()
	}
	b.StopTimer()
	reportDropdownCacheStats(b, dropdownCacheStatsDelta(frame.cacheStats(), baseline))
}
