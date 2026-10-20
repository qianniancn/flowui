package chip

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkChipLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("Complete").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(320, 80)}}
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

func BenchmarkChipVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("Complete").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	value.visualWithCache(ctx, nil)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportChipCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkChipVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("Complete").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
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
	reportChipCacheStats(b, value.visualCache.Stats(), computed.ProviderCacheStats{})
}

func BenchmarkChipComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New("A").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(ctx, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(1024, 40)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := chipCacheStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	}
	b.StopTimer()
	reportChipCacheStats(b, chipCacheStats(values), baseline)
}

func BenchmarkChipComponentTreeThemeSwitch(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	active := firstContext
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New("A").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(active, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(1024, 40)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := chipCacheStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if i%2 == 0 {
			active = firstContext
		} else {
			active = secondContext
		}
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	}
	b.StopTimer()
	reportChipCacheStats(b, chipCacheStats(values), baseline)
}

func chipCacheStats(values []Widget) ChipVisualCacheStats {
	var total ChipVisualCacheStats
	for _, value := range values {
		stats := value.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportChipCacheStats(b *testing.B, after, before ChipVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "chip-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "chip-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "chip-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "chip-visual-cache-key-invalidations/op")
}
