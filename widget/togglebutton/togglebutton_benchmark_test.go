package togglebutton

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
	"github.com/qianniancn/flowui/theme"
)

type toggleBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newToggleBenchmarkFrame(maximum image.Point) *toggleBenchmarkFrame {
	active := *testtheme.Behavior()
	frame := &toggleBenchmarkFrame{
		ctx: core.New(nil, &active, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *toggleBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkToggleButtonLayoutSteadyState(b *testing.B) {
	frame := newToggleBenchmarkFrame(image.Pt(320, 80))
	value := New("Notifications").Key("notifications").Selected(false).Size(Medium)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkToggleButtonVisualCacheHit(b *testing.B) {
	active := *testtheme.Behavior()
	var cache VisualCache
	request := theme.ToggleRequest{Variant: theme.ToggleButtonDefault, Size: theme.ToggleButtonMedium}
	cache.Resolve(&active, request)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.Resolve(&active, request)
	}
	b.StopTimer()
	reportToggleVisualCacheStats(b, cache.Stats(), VisualCacheStats{})
}

func BenchmarkToggleButtonVisualCacheMiss(b *testing.B) {
	active := *testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache VisualCache
		request := theme.ToggleRequest{Variant: theme.ToggleButtonDefault, Size: theme.ToggleButtonMedium}
		if index&1 != 0 {
			request.Variant = theme.ToggleButtonGhost
		}
		_ = cache.Resolve(&active, request)
	}
	b.StopTimer()
	// Each iteration owns a fresh cache, so every request is one miss without
	// introducing artificial key invalidations between unrelated instances.
	b.ReportMetric(0, "togglebutton-visual-cache-hits/op")
	b.ReportMetric(1, "togglebutton-visual-cache-misses/op")
	b.ReportMetric(0, "togglebutton-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "togglebutton-visual-cache-key-invalidations/op")
}

func BenchmarkToggleButtonVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	var cache VisualCache
	cache.cache.SetCapacity(1)
	request := theme.ToggleRequest{Variant: theme.ToggleButtonDefault, Size: theme.ToggleButtonMedium}
	cache.Resolve(first, request)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			request.Variant = theme.ToggleButtonDefault
			cache.Resolve(first, request)
		case 1:
			request.Variant = theme.ToggleButtonGhost
			cache.Resolve(first, request)
		case 2:
			request.Variant = theme.ToggleButtonDefault
			cache.Resolve(second, request)
		case 3:
			request.Variant = theme.ToggleButtonGhost
			cache.Resolve(second, request)
		}
	}
	b.StopTimer()
	reportToggleVisualCacheStats(b, cache.Stats(), VisualCacheStats{})
}

func BenchmarkToggleButtonLayoutThemeInvalidation(b *testing.B) {
	frame := newToggleBenchmarkFrame(image.Pt(320, 80))
	value := New("Notifications").Key("notifications").Selected(false).Size(Medium)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		active := *testtheme.Behavior()
		core.ReplaceTheme(frame.ctx, active)
		frame.layout(value)
	}
}

func BenchmarkToggleButtonComponentTreeSteadyState(b *testing.B) {
	frame, value, values := newToggleButtonBenchmarkTree(image.Pt(512, 96))
	frame.layout(value)
	baseline := toggleButtonCacheStats(frame.ctx, values)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		frame.gtx.Now = time.Unix(0, int64(index+1)*int64(time.Millisecond))
		frame.layout(value)
	}
	b.StopTimer()
	reportToggleVisualCacheStats(b, toggleButtonCacheStats(frame.ctx, values), baseline)
}

func BenchmarkToggleButtonComponentTreeThemeSwitch(b *testing.B) {
	frame, value, values := newToggleButtonBenchmarkTree(image.Pt(512, 96))
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	core.ReplaceTheme(frame.ctx, first)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, first)
		} else {
			core.ReplaceTheme(frame.ctx, second)
		}
		frame.gtx.Now = time.Unix(0, int64(index+1)*int64(time.Millisecond))
		frame.layout(value)
	}
	b.StopTimer()
	reportToggleVisualCacheStats(b, toggleButtonCacheStats(frame.ctx, values), VisualCacheStats{})
}

func newToggleButtonBenchmarkTree(maximum image.Point) (*toggleBenchmarkFrame, core.Widget, []Widget) {
	frame := newToggleBenchmarkFrame(maximum)
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		value := New("Toggle").Key("toggle-" + string(rune('a'+index))).Selected(index%3 == 0)
		switch index % 6 {
		case 1:
			value = value.Variant(Ghost)
		case 2:
			value = value.Size(Small)
		case 3:
			value = value.Size(Large)
		case 4:
			value = value.IconOnly()
		case 5:
			value = value.Variant(Ghost).IconOnly().Size(Small)
		}
		values[index] = value
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(frame.ctx, gtx)
		})
	}
	value := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	})
	return frame, value, values
}

func toggleButtonCacheStats(ctx *core.Context, values []Widget) (total VisualCacheStats) {
	for _, value := range values {
		state := toggleButtonStateFor(ctx, value.key)
		stats := state.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportToggleVisualCacheStats(b *testing.B, after, before VisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "togglebutton-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "togglebutton-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "togglebutton-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "togglebutton-visual-cache-key-invalidations/op")
}
