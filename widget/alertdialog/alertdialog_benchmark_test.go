package alertdialog

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
)

type alertDialogBenchmarkFrame struct {
	ctx    *core.Context
	value  Widget
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newAlertDialogBenchmarkFrame(maximum image.Point) *alertDialogBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &alertDialogBenchmarkFrame{
		ctx:   core.New(nil, active, locale.LanguageEnglish),
		value: New("Delete project?", "This action cannot be undone. Review the project details before continuing.").Key("delete-project").Open(true).Footer(button.New("Delete")),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (frame *alertDialogBenchmarkFrame) layout() {
	frame.ops.Reset()
	core.BeginFrameWithViewport(frame.ctx, frame.gtx.Constraints.Max)
	frame.value.Layout(frame.ctx, frame.gtx)
	core.LayoutOverlays(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
}

func (frame *alertDialogBenchmarkFrame) cacheStats() (hits, misses, themeInvalidations, keyInvalidations uint64) {
	stats := frame.value.visualCache.stats()
	return stats.Hits, stats.Misses, stats.ThemeInvalidations, stats.KeyInvalidations
}

func reportAlertDialogCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "alertdialog-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "alertdialog-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "alertdialog-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "alertdialog-visual-cache-key-invalidations/op")
}

func BenchmarkAlertDialogLayoutSteadyState(b *testing.B) {
	frame := newAlertDialogBenchmarkFrame(image.Pt(480, 360))
	frame.layout()
	frame.gtx.Now = frame.gtx.Now.Add(250 * time.Millisecond)
	frame.layout()
	baseline := func() [4]uint64 {
		hits, misses, themeInvalidations, keyInvalidations := frame.cacheStats()
		return [4]uint64{hits, misses, themeInvalidations, keyInvalidations}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout()
	}
	b.StopTimer()
	afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations := frame.cacheStats()
	reportAlertDialogCacheStats(b, [4]uint64{afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations}, baseline)
}

func BenchmarkAlertDialogVisualCacheMiss(b *testing.B) {
	frame := newAlertDialogBenchmarkFrame(image.Pt(480, 360))
	frame.layout()
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
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout()
	}
	b.StopTimer()
	afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations := frame.cacheStats()
	reportAlertDialogCacheStats(
		b,
		[4]uint64{afterHits, afterMisses, afterThemeInvalidations, afterKeyInvalidations},
		[4]uint64{baselineHits, baselineMisses, baselineThemeInvalidations, baselineKeyInvalidations},
	)
}
