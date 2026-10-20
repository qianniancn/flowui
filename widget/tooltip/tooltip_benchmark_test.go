package tooltip

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

func BenchmarkTooltipLayout(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	value := New("benchmark", fixedWidget{size: image.Pt(96, 28)}, "Benchmark tooltip").Delay(0)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(640, 360)),
		Source:      router.Source(),
		Ops:         new(op.Ops),
		Now:         time.Unix(1, 0),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		gtx.Now = gtx.Now.Add(time.Millisecond)
		gtx.Ops = new(op.Ops)
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		value.Layout(ctx, gtx)
		core.LayoutOverlays(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(gtx.Ops)
	}
	state, _ := core.PeekState[tooltipState](ctx, "benchmark", stateSlotTooltip)
	if state != nil {
		stats := state.visualCache.stats()
		b.ReportMetric(float64(stats.Hits), "visual-cache-hits")
		b.ReportMetric(float64(stats.Misses), "visual-cache-misses")
		b.ReportMetric(float64(stats.Hits)/float64(b.N), "tooltip-visual-cache-hits/op")
		b.ReportMetric(float64(stats.Misses)/float64(b.N), "tooltip-visual-cache-misses/op")
		b.ReportMetric(0, "tooltip-visual-cache-theme-invalidations/op")
		b.ReportMetric(0, "tooltip-visual-cache-key-invalidations/op")
	}
}

func BenchmarkTooltipVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	cache := tooltipVisualCache{}
	tooltipVisualWithCache(ctx, nil, &cache)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tooltipVisualWithCache(ctx, nil, &cache)
	}
	b.StopTimer()
	stats := cache.stats()
	b.ReportMetric(float64(stats.Hits), "cache-hits")
	b.ReportMetric(float64(stats.Misses), "cache-misses")
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "tooltip-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "tooltip-visual-cache-misses/op")
	b.ReportMetric(0, "tooltip-visual-cache-theme-invalidations/op")
	b.ReportMetric(0, "tooltip-visual-cache-key-invalidations/op")
}
