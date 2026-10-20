package popover

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
	"github.com/qianniancn/flowui/widget/text"
)

type popoverBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newPopoverBenchmarkFrame(maximum image.Point) *popoverBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &popoverBenchmarkFrame{
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

func (frame *popoverBenchmarkFrame) layout(value Widget) {
	frame.ops.Reset()
	core.BeginFrameWithViewport(frame.ctx, frame.gtx.Constraints.Max)
	value.Layout(frame.ctx, frame.gtx)
	core.LayoutOverlays(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
}

func (frame *popoverBenchmarkFrame) visualCacheStats() PopoverVisualCacheStats {
	state, ok := core.PeekState[popoverState](frame.ctx, "help", stateSlotPopover)
	if !ok || state == nil {
		return PopoverVisualCacheStats{}
	}
	return state.visualCache.stats()
}

func reportPopoverCacheStats(b *testing.B, stats PopoverVisualCacheStats) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "popover-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "popover-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "popover-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/float64(b.N), "popover-visual-cache-key-invalidations/op")
}

func popoverCacheStatsDelta(after, before PopoverVisualCacheStats) PopoverVisualCacheStats {
	return PopoverVisualCacheStats{
		Hits:               after.Hits - before.Hits,
		Misses:             after.Misses - before.Misses,
		ThemeInvalidations: after.ThemeInvalidations - before.ThemeInvalidations,
		KeyInvalidations:   after.KeyInvalidations - before.KeyInvalidations,
	}
}

// BenchmarkPopoverLayoutSteadyState measures a settled popover overlay with a
// heading and text body using the neutral behavior provider.
func BenchmarkPopoverLayoutSteadyState(b *testing.B) {
	frame := newPopoverBenchmarkFrame(image.Pt(480, 360))
	value := New(Spacer(120, 32), text.New("Review the current workspace settings before applying changes.")).Key("help").Open(true).
		Heading("Settings").
		Arrow(true)
	frame.layout(value)
	frame.gtx.Now = frame.gtx.Now.Add(popoverEnterDuration)
	frame.layout(value)
	baseline := frame.visualCacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout(value)
	}
	b.StopTimer()
	reportPopoverCacheStats(b, popoverCacheStatsDelta(frame.visualCacheStats(), baseline))
}

// BenchmarkPopoverVisualCacheMiss measures the provider path when the active
// theme changes on every frame, separately exposing invalidation cost.
func BenchmarkPopoverVisualCacheMiss(b *testing.B) {
	frame := newPopoverBenchmarkFrame(image.Pt(480, 360))
	value := New(Spacer(120, 32), text.New("Body")).Key("help").Open(true)
	frame.layout(value)
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()
	baseline := frame.visualCacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout(value)
	}
	b.StopTimer()
	reportPopoverCacheStats(b, popoverCacheStatsDelta(frame.visualCacheStats(), baseline))
}
