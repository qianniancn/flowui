package tabs

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkTabsLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{
		{Key: "tab-0", Label: "Overview"}, {Key: "tab-1", Label: "Activity"},
		{Key: "tab-2", Label: "Settings"},
	}).Fit().Separators(true)
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(480, 160)}}
	var router input.Router
	value.Layout(ctx, gtx)
	baseline := value.visualCache.cache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		gtx.Source = router.Source()
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		value.Layout(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(gtx.Ops)
	}
	b.StopTimer()
	reportTabsVisualCacheStats(b, value.visualCache.cache.Stats(), baseline)
}

func reportTabsVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "tabs-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "tabs-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "tabs-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "tabs-visual-cache-key-invalidations/op")
}

func BenchmarkTabsVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{{Key: "tab-0", Label: "Overview"}})
	value.visual(ctx)
	baseline := value.visualCache.cache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value.visual(ctx)
	}
	b.StopTimer()
	reportTabsVisualCacheStats(b, value.visualCache.cache.Stats(), baseline)
}

func BenchmarkTabsVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("benchmark-tabs", "tab-0", []TabItem{{Key: "tab-0", Label: "Overview"}})
	value.visual(firstContext)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			value.visual(firstContext)
		case 1:
			value.visual(secondContext)
		case 2:
			value.Size(Small).visual(secondContext)
		default:
			value.visual(secondContext)
		}
	}
	b.StopTimer()
	reportTabsVisualCacheStats(b, value.visualCache.cache.Stats(), computed.ProviderCacheStats{})
}
