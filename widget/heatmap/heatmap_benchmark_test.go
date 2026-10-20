package heatmap

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type heatmapBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	ops    op.Ops
	gtx    layout.Context
}

func newHeatmapBenchmarkFrame(count int) *heatmapBenchmarkFrame {
	active := *testtheme.Behavior()
	frame := &heatmapBenchmarkFrame{ctx: core.New(nil, &active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Unix(1, 0),
		Ops:         &frame.ops,
	}
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 90)
	values := make([]CalendarValue, 0, 91)
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		values = append(values, CalendarValue{Date: day, Value: float64(day.Day() % 11)})
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New("heatmap-"+strconv.Itoa(index), start, end, values)
	}
	return frame
}

func (frame *heatmapBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	dimensions := value.Layout(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	return dimensions
}

func (frame *heatmapBenchmarkFrame) layoutTree() layout.Dimensions {
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
	return dimensions
}

func (frame *heatmapBenchmarkFrame) cacheStats() (hits, misses, themeInvalidations, keyInvalidations uint64) {
	for index := range frame.values {
		stats := frame.values[index].visualCache.Stats()
		hits += stats.Hits
		misses += stats.Misses
		themeInvalidations += stats.ThemeInvalidations
		keyInvalidations += stats.KeyInvalidations
	}
	return hits, misses, themeInvalidations, keyInvalidations
}

func reportHeatmapCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "heatmap-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "heatmap-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "heatmap-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "heatmap-visual-cache-key-invalidations/op")
}

func BenchmarkHeatmapLayoutSteadyState(b *testing.B) {
	frame := newHeatmapBenchmarkFrame(1)
	frame.layout(frame.values[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(frame.values[0])
	}
}

func BenchmarkHeatmapComponentTreeSteadyState(b *testing.B) {
	frame := newHeatmapBenchmarkFrame(64)
	frame.layoutTree()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkHeatmapComponentTreeVisualCacheHit(b *testing.B) {
	frame := newHeatmapBenchmarkFrame(64)
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
	reportHeatmapCacheStats(b, after, before)
}

func BenchmarkHeatmapComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newHeatmapBenchmarkFrame(64)
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
	reportHeatmapCacheStats(b, after, before)
}
