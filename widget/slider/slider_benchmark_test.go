package slider

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

type sliderBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newSliderBenchmarkFrame(maximum image.Point) *sliderBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &sliderBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *sliderBenchmarkFrame) layout(value sliderLayoutWidget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkSliderLayoutSteadyState(b *testing.B) {
	frame := newSliderBenchmarkFrame(image.Pt(320, 80))
	value := New("volume", 50).ShowValue()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSliderComponentTreeSteadyState(b *testing.B) {
	frame := newSliderBenchmarkFrame(image.Pt(640, 900))
	values := make([]Widget, 128)
	for index := range values {
		values[index] = New(fmt.Sprintf("slider-%d", index), float64(index%100))
		if index%2 == 0 {
			values[index] = values[index].ShowValue()
		}
	}
	value := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, func() []layout.FlexChild {
			children := make([]layout.FlexChild, 0, len(values))
			for _, slider := range values {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return slider.Layout(ctx, gtx)
				}))
			}
			return children
		}()...)
	})
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkSliderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	var cache VisualCache
	cache.Resolve(active, Horizontal)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.Resolve(active, Horizontal)
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "slider-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "slider-visual-cache-misses/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "slider-visual-cache-key-invalidations/op")
	b.ReportMetric(0, "slider-visual-cache-theme-invalidations/op")
}

func BenchmarkSliderVisualCacheMiss(b *testing.B) {
	active := testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache VisualCache
		orientation := Horizontal
		if index&1 != 0 {
			orientation = Vertical
		}
		_ = cache.Resolve(active, orientation)
	}
	b.StopTimer()
	b.ReportMetric(0, "slider-visual-cache-hits/op")
	b.ReportMetric(1, "slider-visual-cache-misses/op")
	b.ReportMetric(0, "slider-visual-cache-key-invalidations/op")
	b.ReportMetric(0, "slider-visual-cache-theme-invalidations/op")
}

func BenchmarkSliderVisualCacheThemeInvalidation(b *testing.B) {
	active := testtheme.Behavior()
	replacement := testtheme.Behavior()
	var cache VisualCache
	cache.Resolve(active, Horizontal)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			_ = cache.Resolve(active, Horizontal)
		} else {
			_ = cache.Resolve(replacement, Horizontal)
		}
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "slider-visual-cache-theme-invalidations/op")
}

var _ theme.SliderOrientation = Horizontal
