package candlestick

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

type candlestickBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	gtx    layout.Context
	ops    op.Ops
}

func newCandlestickBenchmarkFrame(count int) *candlestickBenchmarkFrame {
	active := *testtheme.Behavior()
	frame := &candlestickBenchmarkFrame{ctx: core.New(nil, &active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 360)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Unix(1, 0),
		Ops:         &frame.ops,
	}
	values := make([]Candle, 24)
	for index := range values {
		base := float64((index*7)%19 + 10)
		values[index] = OHLC(base, base+float64(index%3-1), base-2, base+4)
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New("candle-"+strconv.Itoa(index), values).
			Categories([]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X"})
	}
	return frame
}

func (frame *candlestickBenchmarkFrame) layout(value core.Widget) {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	value.Layout(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
}

func (frame *candlestickBenchmarkFrame) layoutTree() {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	children := make([]layout.FlexChild, len(frame.values))
	for index := range frame.values {
		value := frame.values[index]
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(frame.ctx, gtx)
		})
	}
	layout.Flex{Axis: layout.Vertical}.Layout(frame.gtx, children...)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
}

func (frame *candlestickBenchmarkFrame) cacheStats() (hits, misses, themeInvalidations, keyInvalidations uint64) {
	for index := range frame.values {
		stats := frame.values[index].visualCache.Stats()
		hits += stats.Hits
		misses += stats.Misses
		themeInvalidations += stats.ThemeInvalidations
		keyInvalidations += stats.KeyInvalidations
	}
	return hits, misses, themeInvalidations, keyInvalidations
}

func reportCandlestickCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "candlestick-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "candlestick-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "candlestick-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "candlestick-visual-cache-key-invalidations/op")
}

func BenchmarkCandlestickLayoutSteadyState(b *testing.B) {
	frame := newCandlestickBenchmarkFrame(1)
	frame.layout(frame.values[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(frame.values[0])
	}
}

func BenchmarkCandlestickComponentTreeSteadyState(b *testing.B) {
	frame := newCandlestickBenchmarkFrame(64)
	frame.layoutTree()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkCandlestickComponentTreeVisualCacheHit(b *testing.B) {
	frame := newCandlestickBenchmarkFrame(64)
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
	reportCandlestickCacheStats(b, after, before)
}

func BenchmarkCandlestickComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newCandlestickBenchmarkFrame(64)
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
	reportCandlestickCacheStats(b, after, before)
}
