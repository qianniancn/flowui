package closebutton

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkCloseButtonLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("close")
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(64, 64)}}
	gtx.Now = time.Unix(0, int64(time.Millisecond))
	core.BeginFrame(ctx)
	value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		core.BeginFrame(ctx)
		value.Layout(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
	}
}

func BenchmarkCloseButtonVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("close")
	value.visualWithCache(ctx, nil)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportCloseButtonVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func reportCloseButtonVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "closebutton-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "closebutton-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "closebutton-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "closebutton-visual-cache-key-invalidations/op")
}

func BenchmarkCloseButtonVisualCacheInvalidation(b *testing.B) {
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	firstContext := core.New(nil, &first, locale.LanguageEnglish)
	secondContext := core.New(nil, &second, locale.LanguageEnglish)
	value := New("close")
	value.visualWithCache(firstContext, nil)
	baseline := value.visualCache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			value.visualWithCache(firstContext, nil)
		} else {
			value.visualWithCache(secondContext, nil)
		}
	}
	b.StopTimer()
	reportCloseButtonVisualCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkCloseButtonComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	values := make([]Widget, 128)
	for index := range values {
		values[index] = New("close-" + strconv.Itoa(index))
	}
	root := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, len(values))
		for _, value := range values {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return value.Layout(ctx, gtx)
			}))
		}
		return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx, children...)
	})
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(640, 1200)}}
	layoutCloseButtonBenchmarkFrame(ctx, gtx, root, time.Unix(1, 0))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutCloseButtonBenchmarkFrame(ctx, gtx, root, time.Unix(1, int64(i+1)*int64(time.Millisecond)))
	}
}

func layoutCloseButtonBenchmarkFrame(ctx *core.Context, gtx layout.Context, value core.Widget, now time.Time) {
	gtx.Now = now
	gtx.Ops.Reset()
	core.BeginFrame(ctx)
	value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
}
