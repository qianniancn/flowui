package progress

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

type progressBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newProgressBenchmarkFrame(maximum image.Point) *progressBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &progressBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *progressBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkProgressBarLayoutSteadyState(b *testing.B) {
	frame := newProgressBenchmarkFrame(image.Pt(480, 80))
	value := Bar("upload", 42).Label("Upload").ShowValue()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkProgressCircleLayoutSteadyState(b *testing.B) {
	frame := newProgressBenchmarkFrame(image.Pt(120, 120))
	value := Circle("upload", 42).Label("Upload")
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkProgressComponentTreeSteadyState(b *testing.B) {
	frame := newProgressBenchmarkFrame(image.Pt(640, 1200))
	values := make([]core.Widget, 128)
	for index := range values {
		if index%3 == 0 {
			values[index] = Circle(fmt.Sprintf("circle-%d", index), float64(index%100))
		} else if index%3 == 1 {
			values[index] = Bar(fmt.Sprintf("bar-%d", index), float64(index%100)).ShowValue()
		} else {
			values[index] = Meter(fmt.Sprintf("meter-%d", index), float64(index%100))
		}
	}
	value := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, len(values))
		for _, item := range values {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return item.Layout(ctx, gtx)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkProgressBarVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	var cache progressBarVisualCache
	key := progressDeclarationKey{color: BarAccent, size: BarMedium}
	cache.resolve(active, key)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.resolve(active, key)
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "progress-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "progress-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "progress-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "progress-visual-cache-key-invalidations/op")
}

func BenchmarkProgressCircleVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	var cache progressCircleVisualCache
	key := progressDeclarationKey{color: CircleAccent, size: CircleMedium}
	cache.resolve(active, key)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.resolve(active, key)
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "progress-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "progress-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "progress-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "progress-visual-cache-key-invalidations/op")
}

func BenchmarkProgressBarVisualCacheMiss(b *testing.B) {
	active := testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache progressBarVisualCache
		key := progressDeclarationKey{color: BarAccent, size: BarMedium}
		if index&1 != 0 {
			key.color = BarSuccess
		}
		_ = cache.resolve(active, key)
	}
	b.StopTimer()
	b.ReportMetric(0, "progress-visual-cache-hits/op")
	b.ReportMetric(1, "progress-visual-cache-misses/op")
	b.ReportMetric(1, "progress-visual-cache-key-invalidations/op")
}

func BenchmarkProgressCircleVisualCacheMiss(b *testing.B) {
	active := testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache progressCircleVisualCache
		key := progressDeclarationKey{color: CircleAccent, size: CircleMedium}
		if index&1 != 0 {
			key.size = CircleLarge
		}
		_ = cache.resolve(active, key)
	}
	b.StopTimer()
	b.ReportMetric(0, "progress-visual-cache-hits/op")
	b.ReportMetric(1, "progress-visual-cache-misses/op")
	b.ReportMetric(1, "progress-visual-cache-key-invalidations/op")
}

func BenchmarkProgressBarVisualCacheThemeInvalidation(b *testing.B) {
	active := testtheme.Behavior()
	replacement := testtheme.Behavior()
	var cache progressBarVisualCache
	key := progressDeclarationKey{color: BarAccent, size: BarMedium}
	cache.resolve(active, key)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			_ = cache.resolve(active, key)
		} else {
			_ = cache.resolve(replacement, key)
		}
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "progress-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "progress-visual-cache-key-invalidations/op")
}

func BenchmarkProgressCircleVisualCacheThemeInvalidation(b *testing.B) {
	active := testtheme.Behavior()
	replacement := testtheme.Behavior()
	var cache progressCircleVisualCache
	key := progressDeclarationKey{color: CircleAccent, size: CircleMedium}
	cache.resolve(active, key)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			_ = cache.resolve(active, key)
		} else {
			_ = cache.resolve(replacement, key)
		}
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "progress-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "progress-visual-cache-key-invalidations/op")
}

var _ theme.ProgressBarColor = BarAccent
