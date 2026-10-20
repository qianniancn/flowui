package radiogroup

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
)

type radioBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newRadioBenchmarkFrame(maximum image.Point) *radioBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &radioBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *radioBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkRadioGroupLayoutSteadyState(b *testing.B) {
	frame := newRadioBenchmarkFrame(image.Pt(320, 120))
	value := New([]RadioItem{
		{Key: "basic", Label: "Basic", Description: "For small teams"},
		{Key: "pro", Label: "Pro", Description: "For growing teams"},
	}).Key("plan").SelectedKey("basic")
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkRadioGroupComponentTreeSteadyState(b *testing.B) {
	frame := newRadioBenchmarkFrame(image.Pt(640, 900))
	items := make([]RadioItem, 128)
	for index := range items {
		items[index] = RadioItem{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
		if index%7 == 0 {
			items[index].Description = "Supporting text"
		}
		if index%11 == 0 {
			items[index].Disabled = true
		}
	}
	value := New(items).Key("plans").SelectedKey("item-0")
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkRadioGroupVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	var cache VisualCache
	cache.Resolve(active, RadioPrimary)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = cache.Resolve(active, RadioPrimary)
	}
	b.StopTimer()
	stats := cache.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "radiogroup-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "radiogroup-visual-cache-misses/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "radiogroup-visual-cache-key-invalidations/op")
	b.ReportMetric(0, "radiogroup-visual-cache-theme-invalidations/op")
}

func BenchmarkRadioGroupVisualCacheMiss(b *testing.B) {
	active := testtheme.Behavior()
	b.ReportAllocs()
	for index := 0; b.Loop(); index++ {
		var cache VisualCache
		variant := RadioPrimary
		if index&1 != 0 {
			variant = RadioSecondary
		}
		_ = cache.Resolve(active, variant)
	}
	b.StopTimer()
	b.ReportMetric(0, "radiogroup-visual-cache-hits/op")
	b.ReportMetric(1, "radiogroup-visual-cache-misses/op")
	b.ReportMetric(0, "radiogroup-visual-cache-key-invalidations/op")
	b.ReportMetric(0, "radiogroup-visual-cache-theme-invalidations/op")
}
