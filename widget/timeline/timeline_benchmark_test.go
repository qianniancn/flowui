package timeline

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
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type timelineBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newTimelineBenchmarkFrame(maximum image.Point) *timelineBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &timelineBenchmarkFrame{
		ctx: core.New(nil, active, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *timelineBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

type timelineBenchmarkProbe struct {
	size image.Point
}

func (p timelineBenchmarkProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func timelineBenchmarkItem(index int) Item {
	probe := timelineBenchmarkProbe{size: image.Pt(96+index%3*12, 20+index%2*4)}
	return Item{
		Key:     "item",
		Title:   probe,
		Content: probe,
		Color:   Color(index % int(ColorCustom)),
	}
}

func timelineBenchmarkValue() Widget {
	items := make([]Item, 0, 4)
	for index := range 4 {
		items = append(items, timelineBenchmarkItem(index))
	}
	return New(items).Mode(ModeAlternate).TitleWidth(120)
}

type timelineBenchmarkTreeWidget struct {
	values []Widget
}

func (t timelineBenchmarkTreeWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(t.values))
	for _, value := range t.values {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx, children...)
}

func timelineBenchmarkTree() timelineBenchmarkTreeWidget {
	values := make([]Widget, 0, 32)
	for index := range 32 {
		value := timelineBenchmarkValue()
		if index%2 == 0 {
			value = value.Orientation(OrientationHorizontal)
		}
		values = append(values, value)
	}
	return timelineBenchmarkTreeWidget{values: values}
}

func reportTimelineVisualStats(b *testing.B, after, before TimelineVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "timeline-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "timeline-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "timeline-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "timeline-visual-cache-key-invalidations/op")
}

func BenchmarkTimelineProviderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	cache := timelineVisualCache{}
	cache.cache.Resolve(active, struct{}{}, func() theme.TimelineVisual {
		provider, err := theme.ComponentProvider(active)
		if err != nil {
			b.Fatal(err)
		}
		visual := provider.Timeline(active, theme.TimelineRequest{})
		theme.ValidateTimelineVisual(visual)
		return visual
	})
	before := cache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.cache.Resolve(active, struct{}{}, func() theme.TimelineVisual {
			b.Helper()
			panic("timeline visual cache miss")
		})
	}
	b.StopTimer()
	reportTimelineVisualStats(b, cache.Stats(), before)
}

func BenchmarkTimelineLayoutSteadyState(b *testing.B) {
	frame := newTimelineBenchmarkFrame(image.Pt(640, 360))
	value := timelineBenchmarkValue()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkTimelineComponentTreeSteadyState(b *testing.B) {
	frame := newTimelineBenchmarkFrame(image.Pt(720, 24000))
	value := timelineBenchmarkTree()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkTimelineComponentTreeVisualCacheHit(b *testing.B) {
	frame := newTimelineBenchmarkFrame(image.Pt(720, 24000))
	value := timelineBenchmarkTree()
	frame.layout(value)
	before := timelineTreeVisualStats(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportTimelineVisualStats(b, timelineTreeVisualStats(value), before)
}

func timelineTreeVisualStats(value timelineBenchmarkTreeWidget) TimelineVisualCacheStats {
	var total TimelineVisualCacheStats
	for _, item := range value.values {
		stats := item.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func BenchmarkTimelineComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newTimelineBenchmarkFrame(image.Pt(720, 24000))
	value := timelineBenchmarkTree()
	frame.layout(value)
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	before := timelineTreeVisualStats(value)

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
	reportTimelineVisualStats(b, timelineTreeVisualStats(value), before)
}

func BenchmarkTimelineStyledOverrideBypass(b *testing.B) {
	frame := newTimelineBenchmarkFrame(image.Pt(640, 360))
	value := timelineBenchmarkValue()
	styled := (value).Style(style.Style{}.PaddingX(2))
	frame.layout(styled)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(styled)
	}
}
