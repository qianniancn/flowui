package spinner

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type spinnerBenchmarkFrame struct {
	ctx    *core.Context
	gtx    layout.Context
	ops    op.Ops
	router input.Router
}

func newSpinnerBenchmarkFrame(maximum image.Point) *spinnerBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &spinnerBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(0, 0),
	}
	return frame
}

func (f *spinnerBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkSpinnerLayoutSteadyState(b *testing.B) {
	frame := newSpinnerBenchmarkFrame(image.Pt(320, 160))
	value := New().Color(Success).Size(Medium).Label("Loading")
	frame.layout(value)
	baseline := value.visualCache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		frame.gtx.Now = time.Unix(0, int64(index+1)*int64(time.Millisecond))
		frame.layout(value)
	}
	b.StopTimer()
	reportSpinnerVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkSpinnerVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New().Color(Success).Size(Medium)
	value.visualWithCache(ctx, nil)
	baseline := value.visualCache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportSpinnerVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkSpinnerVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New().Color(Accent).Size(Medium)
	value.visualWithCache(firstContext, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 7 {
		case 0, 1:
			value.visualWithCache(firstContext, nil)
		case 2, 3:
			value.visualWithCache(secondContext, nil)
		case 4, 5:
			value.Color(Success).visualWithCache(secondContext, nil)
		case 6, 7:
			value.Size(Large).visualWithCache(secondContext, nil)
		}
	}
	b.StopTimer()
	reportSpinnerVisualCacheStats(b, value.visualCache.Stats(), SpinnerVisualCacheStats{})
}

func BenchmarkSpinnerComponentTreeSteadyState(b *testing.B) {
	frame := newSpinnerBenchmarkFrame(image.Pt(512, 64))
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New().Size(Size(index % 4)).Color(Color(index % 5))
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(frame.ctx, gtx)
		})
	}
	value := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	})
	frame.layout(value)
	baseline := spinnerCacheStats(values)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		frame.gtx.Now = time.Unix(0, int64(index+1)*int64(time.Millisecond))
		frame.layout(value)
	}
	b.StopTimer()
	reportSpinnerVisualCacheStats(b, spinnerCacheStats(values), baseline)
}

func BenchmarkSpinnerComponentTreeThemeSwitch(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	active := firstContext
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New().Size(Size(index % 4)).Color(Color(index % 5))
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(active, gtx)
		})
	}
	value := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	})
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(512, 64)},
		Ops:         new(op.Ops),
		Now:         time.Unix(0, 0),
	}
	layoutSpinnerTree(firstContext, &gtx, value)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			active = firstContext
		} else {
			active = secondContext
		}
		gtx.Now = time.Unix(0, int64(index+1)*int64(time.Millisecond))
		layoutSpinnerTree(active, &gtx, value)
	}
	b.StopTimer()
	reportSpinnerVisualCacheStats(b, spinnerCacheStats(values), SpinnerVisualCacheStats{})
}

func layoutSpinnerTree(ctx *core.Context, gtx *layout.Context, value core.Widget) {
	gtx.Ops.Reset()
	core.BeginFrame(ctx)
	value.Layout(ctx, *gtx)
	core.ApplyFrameCommands(ctx, *gtx)
	core.EndFrame(ctx)
}

func spinnerCacheStats(values []Widget) (total SpinnerVisualCacheStats) {
	for _, value := range values {
		stats := value.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportSpinnerVisualCacheStats(b *testing.B, after, before SpinnerVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "spinner-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "spinner-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "spinner-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "spinner-visual-cache-key-invalidations/op")
}
