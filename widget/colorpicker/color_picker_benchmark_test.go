package colorpicker

import (
	"image"
	"image/color"
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

type colorPickerBenchmarkFrame struct {
	ctx      *core.Context
	router   input.Router
	ops      op.Ops
	gtx      layout.Context
	now      time.Time
	viewport image.Point
}

func newColorPickerBenchmarkFrame() *colorPickerBenchmarkFrame {
	active := *testtheme.Behavior()
	frame := &colorPickerBenchmarkFrame{
		ctx:      core.New(nil, &active, locale.LanguageEnglish),
		now:      time.Unix(1, 0),
		viewport: image.Pt(400, 500),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: frame.viewport},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
	}
	return frame
}

func (frame *colorPickerBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	frame.ops.Reset()
	frame.gtx.Now = frame.now
	core.BeginFrameWithViewport(frame.ctx, frame.viewport)
	dimensions := value.Layout(frame.ctx, frame.gtx)
	core.LayoutOverlays(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	frame.now = frame.now.Add(time.Millisecond)
	return dimensions
}

func benchmarkColorPickerValue() Widget {
	return New("benchmark", color.NRGBA{R: 4, G: 133, B: 247, A: 180}).
		Alpha(true).
		Presets([]color.NRGBA{
			{R: 220, G: 60, B: 60, A: 255},
			{R: 50, G: 120, B: 220, A: 255},
			{R: 60, G: 180, B: 110, A: 255},
		}).
		OnChange(func(color.NRGBA) {})
}

func (frame *colorPickerBenchmarkFrame) openPicker(value Widget) {
	frame.layout(value)
	state := colorPickerTestState(frame.ctx, "benchmark")
	state.history = []color.NRGBA{{R: 240, G: 180, B: 40, A: 255}}
	state.open = true
	frame.now = frame.now.Add(colorPickerEnterDuration)
	frame.layout(value)
}

type colorPickerBenchmarkStats struct {
	hits               uint64
	misses             uint64
	themeInvalidations uint64
	keyInvalidations   uint64
	cacheCount         uint64
}

func (stats *colorPickerBenchmarkStats) add(value computed.ProviderCacheStats) {
	stats.hits += value.Hits
	stats.misses += value.Misses
	stats.themeInvalidations += value.ThemeInvalidations
	stats.keyInvalidations += value.KeyInvalidations
	stats.cacheCount++
}

func colorPickerBenchmarkVisualStats(ctx *core.Context) colorPickerBenchmarkStats {
	var stats colorPickerBenchmarkStats
	add := func(value *colorPickerVisualCache) {
		if value != nil {
			stats.add(value.stats())
		}
	}
	if state, ok := core.PeekState[colorPickerState](ctx, "benchmark", stateSlotColorPicker); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorAreaState](ctx, "benchmark/area", stateSlotColorArea); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorSliderState](ctx, "benchmark/hue", stateSlotColorSlider); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorSliderState](ctx, "benchmark/alpha", stateSlotColorSlider); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorFieldState](ctx, "benchmark/field", stateSlotColorField); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorSwatchPickerState](ctx, "benchmark/presets", stateSlotColorSwatchPicker); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorSwatchPickerState](ctx, "benchmark/history", stateSlotColorSwatchPicker); ok {
		add(&state.visualCache)
	}
	if state, ok := core.PeekState[colorPickerVisualCache](ctx, standaloneColorPickerVisualKey, standaloneColorPickerVisualSlot); ok {
		stats.add(state.stats())
	}
	return stats
}

func reportColorPickerBenchmarkStats(b *testing.B, after, before colorPickerBenchmarkStats) {
	divisor := float64(b.N)
	b.ReportMetric(float64(after.hits-before.hits)/divisor, "colorpicker-visual-cache-hits/op")
	b.ReportMetric(float64(after.misses-before.misses)/divisor, "colorpicker-visual-cache-misses/op")
	b.ReportMetric(float64(after.themeInvalidations-before.themeInvalidations)/divisor, "colorpicker-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.keyInvalidations-before.keyInvalidations)/divisor, "colorpicker-visual-cache-key-invalidations/op")
}

func BenchmarkColorPickerComponentTreeSteadyState(b *testing.B) {
	frame := newColorPickerBenchmarkFrame()
	value := benchmarkColorPickerValue()
	frame.openPicker(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkColorPickerComponentTreeVisualCacheHit(b *testing.B) {
	frame := newColorPickerBenchmarkFrame()
	value := benchmarkColorPickerValue()
	frame.openPicker(value)
	before := colorPickerBenchmarkVisualStats(frame.ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportColorPickerBenchmarkStats(b, colorPickerBenchmarkVisualStats(frame.ctx), before)
}

func BenchmarkColorPickerComponentTreeVisualCacheThemeInvalidation(b *testing.B) {
	frame := newColorPickerBenchmarkFrame()
	value := benchmarkColorPickerValue()
	frame.openPicker(value)
	before := colorPickerBenchmarkVisualStats(frame.ctx)
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()

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
	reportColorPickerBenchmarkStats(b, colorPickerBenchmarkVisualStats(frame.ctx), before)
}

func BenchmarkColorPickerVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	var cache colorPickerVisualCache
	cache.resolve(active, theme.ColorPickerRequest{})
	before := cache.stats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.resolve(active, theme.ColorPickerRequest{})
	}
	b.StopTimer()
	after := cache.stats()
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "colorpicker-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "colorpicker-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "colorpicker-cache-theme-invalidations/op")
}

func BenchmarkColorPickerVisualCacheMiss(b *testing.B) {
	active := testtheme.Behavior()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var cache colorPickerVisualCache
		cache.resolve(active, theme.ColorPickerRequest{})
	}
}

func BenchmarkColorPickerVisualCacheOverrideBypass(b *testing.B) {
	active := testtheme.Behavior()
	declaration := style.Style{}.Padding(1)
	request := theme.ColorPickerRequest{Override: &theme.ColorPickerOverride{Declaration: declaration}}
	var cache colorPickerVisualCache

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.resolve(active, request)
	}
	b.StopTimer()
	stats := cache.stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "colorpicker-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "colorpicker-cache-misses/op")
	b.ReportMetric(float64(b.N), "colorpicker-cache-override-bypasses/op")
}
