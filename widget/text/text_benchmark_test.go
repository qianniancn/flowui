package text

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	giotext "gioui.org/text"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type textBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newTextBenchmarkFrame(maximum image.Point) *textBenchmarkFrame {
	frame := &textBenchmarkFrame{
		ctx: core.New(nil, testtheme.Behavior(), locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *textBenchmarkFrame) measure(value Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Measure(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	return dimensions
}

func BenchmarkTextMeasureCacheMiss(b *testing.B) {
	frame := newTextBenchmarkFrame(image.Pt(240, 80))
	value := New("A measured label with stable text")

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		// The cache holds 256 entries. Cycling through 257 widths forces the
		// next access to evict the oldest key, keeping this a miss benchmark.
		frame.gtx.Constraints.Max.X = 200 + index%257
		frame.measure(value)
	}
	b.StopTimer()
	reportTextVisualCacheStats(b, value)
}

func BenchmarkTextMeasureCacheHit(b *testing.B) {
	frame := newTextBenchmarkFrame(image.Pt(240, 80))
	value := New("A measured label with stable text")
	frame.measure(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.measure(value)
	}
	b.StopTimer()
	reportTextVisualCacheStats(b, value)
}

func BenchmarkTextMeasureLongMultiline(b *testing.B) {
	frame := newTextBenchmarkFrame(image.Pt(240, 180))
	value := New("Line one of a long paragraph that wraps across the available width.\nLine two keeps the text layout path representative of a text area.").Wrap(giotext.WrapWords).MaxLines(0)
	frame.measure(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.measure(value)
	}
	b.StopTimer()
	reportTextVisualCacheStats(b, value)
}

func reportTextVisualCacheStats(b *testing.B, value Widget) {
	b.Helper()
	stats := value.visualCache.stats()
	for _, metric := range []struct {
		name  string
		value uint64
	}{
		{name: "text-visual-cache-hits/op", value: stats.Hits},
		{name: "text-visual-cache-misses/op", value: stats.Misses},
		{name: "text-visual-cache-theme-invalidations/op", value: stats.ThemeInvalidations},
		{name: "text-visual-cache-key-invalidations/op", value: stats.KeyInvalidations},
	} {
		b.ReportMetric(float64(metric.value)/float64(b.N), metric.name)
	}
}
