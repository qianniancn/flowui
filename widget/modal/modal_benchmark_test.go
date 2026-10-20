package modal

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
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

type modalBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newModalBenchmarkFrame(maximum image.Point) *modalBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &modalBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (frame *modalBenchmarkFrame) layout(value Widget) {
	frame.ops.Reset()
	core.BeginFrameWithViewport(frame.ctx, frame.gtx.Constraints.Max)
	value.Layout(frame.ctx, frame.gtx)
	core.LayoutOverlays(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
}

func (frame *modalBenchmarkFrame) visualCacheStats() ModalVisualCacheStats {
	state, ok := core.PeekState[modalState](frame.ctx, "settings", stateSlotModal)
	if !ok || state == nil {
		return ModalVisualCacheStats{}
	}
	return state.visualCache.stats()
}

// BenchmarkModalLayoutSteadyState measures a settled modal overlay with a
// header, text body, and footer action using the neutral behavior provider.
func BenchmarkModalLayoutSteadyState(b *testing.B) {
	frame := newModalBenchmarkFrame(image.Pt(480, 360))
	value := New("Settings", text.New("Review the current workspace settings before applying changes.")).Key("settings").Open(true).
		Footer(button.New("Apply"))
	frame.layout(value)
	frame.gtx.Now = frame.gtx.Now.Add(modalEnterDuration)
	frame.layout(value)
	baseline := frame.visualCacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout(value)
	}
	b.StopTimer()
	stats := frame.visualCacheStats()
	if b.N > 0 {
		b.ReportMetric(float64(stats.Hits-baseline.Hits)/float64(b.N), "modal-visual-cache-hits/op")
		b.ReportMetric(float64(stats.Misses-baseline.Misses)/float64(b.N), "modal-visual-cache-misses/op")
		b.ReportMetric(float64(stats.ThemeInvalidations-baseline.ThemeInvalidations)/float64(b.N), "modal-visual-cache-theme-invalidations/op")
		b.ReportMetric(float64(stats.KeyInvalidations-baseline.KeyInvalidations)/float64(b.N), "modal-visual-cache-key-invalidations/op")
	}
}

// BenchmarkModalVisualCacheMiss measures the provider path when the active
// theme changes on every frame. It keeps invalidation cost visible separately
// from the steady-state component tree benchmark above.
func BenchmarkModalVisualCacheMiss(b *testing.B) {
	frame := newModalBenchmarkFrame(image.Pt(480, 360))
	value := New("Settings", text.New("Body")).Key("settings").Open(true)
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
	stats := frame.visualCacheStats()
	if b.N > 0 {
		b.ReportMetric(float64(stats.Hits-baseline.Hits)/float64(b.N), "modal-visual-cache-hits/op")
		b.ReportMetric(float64(stats.Misses-baseline.Misses)/float64(b.N), "modal-visual-cache-misses/op")
		b.ReportMetric(float64(stats.ThemeInvalidations-baseline.ThemeInvalidations)/float64(b.N), "modal-visual-cache-theme-invalidations/op")
		b.ReportMetric(float64(stats.KeyInvalidations-baseline.KeyInvalidations)/float64(b.N), "modal-visual-cache-key-invalidations/op")
	}
}
