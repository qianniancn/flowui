package alert

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
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

type alertBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newAlertBenchmarkFrame(maximum image.Point) *alertBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &alertBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func newAlertComponentTreeFrame(count int) *alertBenchmarkFrame {
	frame := newAlertBenchmarkFrame(image.Pt(640, 480))
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New(
			"Alert "+strconv.Itoa(index),
			"Review the latest update before continuing.",
		).
			Status(Status(index % 5)).
			Content(text.New("Additional details are available.")).
			Action(button.New("Review").Key("alert-action-" + strconv.Itoa(index)).Size(button.Small))
	}
	return frame
}

func (frame *alertBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	dimensions := value.Layout(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
	return dimensions
}

func (frame *alertBenchmarkFrame) layoutTree() layout.Dimensions {
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

func (frame *alertBenchmarkFrame) cacheStats() (hits, misses, themeInvalidations, keyInvalidations uint64) {
	for index := range frame.values {
		stats := frame.values[index].visualCache.stats()
		hits += stats.Hits
		misses += stats.Misses
		themeInvalidations += stats.ThemeInvalidations
		keyInvalidations += stats.KeyInvalidations
	}
	return hits, misses, themeInvalidations, keyInvalidations
}

func reportAlertCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "alert-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "alert-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "alert-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "alert-visual-cache-key-invalidations/op")
}

func BenchmarkAlertLayoutSteadyState(b *testing.B) {
	frame := newAlertBenchmarkFrame(image.Pt(480, 200))
	value := New("Update available", "Review the latest update before continuing.").
		Content(text.New("Additional details are available.")).
		Action(button.New("Review").Size(button.Small))
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkAlertComponentTreeSteadyState(b *testing.B) {
	frame := newAlertComponentTreeFrame(64)
	frame.layoutTree()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkAlertComponentTreeVisualCacheHit(b *testing.B) {
	frame := newAlertComponentTreeFrame(64)
	frame.layoutTree()
	baselineHits, baselineMisses, baselineThemeInvalidations, baselineKeyInvalidations := frame.cacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
	b.StopTimer()
	afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations := frame.cacheStats()
	reportAlertCacheStats(
		b,
		[4]uint64{afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations},
		[4]uint64{baselineHits, baselineMisses, baselineThemeInvalidations, baselineKeyInvalidations},
	)
}

func BenchmarkAlertComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newAlertComponentTreeFrame(64)
	frame.layoutTree()
	baselineHits, baselineMisses, baselineThemeInvalidations, baselineKeyInvalidations := frame.cacheStats()
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
	afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations := frame.cacheStats()
	reportAlertCacheStats(
		b,
		[4]uint64{afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations},
		[4]uint64{baselineHits, baselineMisses, baselineThemeInvalidations, baselineKeyInvalidations},
	)
}
