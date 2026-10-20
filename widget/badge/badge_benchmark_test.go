package badge

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkBadgeLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(&benchmarkBadgeAnchor{size: image.Pt(120, 36)}, "99+").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(320, 160)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	value.Layout(ctx, gtx)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		value.Layout(ctx, gtx)
	}
}

func BenchmarkBadgeVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(nil, "99+").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	value.visualWithCache(ctx, nil)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportBadgeCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkBadgeVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New(nil, "99+").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	value.visualWithCache(firstContext, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 7 {
		case 0, 1:
			value.visualWithCache(firstContext, nil)
		case 2, 3:
			value.visualWithCache(secondContext, nil)
		case 4, 5:
			value.Color(ColorDanger).visualWithCache(secondContext, nil)
		case 6, 7:
			value.visualWithCache(secondContext, nil)
		}
	}
	b.StopTimer()
	reportBadgeCacheStats(b, value.visualCache.Stats(), BadgeVisualCacheStats{})
}

func BenchmarkBadgeComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New(&benchmarkBadgeAnchor{size: image.Pt(32, 24)}, "7").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(ctx, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(4096, 40)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := badgeCacheStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	}
	b.StopTimer()
	reportBadgeCacheStats(b, badgeCacheStats(values), baseline)
}

func BenchmarkBadgeComponentTreeThemeSwitch(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	active := firstContext
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New(&benchmarkBadgeAnchor{size: image.Pt(32, 24)}, "7").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(active, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(4096, 40)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := badgeCacheStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if i%2 == 1 {
			active = secondContext
		} else {
			active = firstContext
		}
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	}
	b.StopTimer()
	reportBadgeCacheStats(b, badgeCacheStats(values), baseline)
}

func badgeCacheStats(values []Widget) BadgeVisualCacheStats {
	var total BadgeVisualCacheStats
	for _, value := range values {
		stats := value.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportBadgeCacheStats(b *testing.B, after, before BadgeVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "badge-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "badge-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "badge-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "badge-visual-cache-key-invalidations/op")
}

type benchmarkBadgeAnchor struct{ size image.Point }

func (a *benchmarkBadgeAnchor) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(a.size)}
}
