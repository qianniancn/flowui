package avatar

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

func BenchmarkAvatarLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("AB").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
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

func BenchmarkAvatarVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New("AB").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
	value.visualWithCache(ctx, nil)
	baseline := value.visualCache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportAvatarCacheStats(b, value.visualCache.Stats(), baseline)
}

func BenchmarkAvatarVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("AB").Color(ColorSuccess).Variant(VariantSoft).Size(SizeMedium)
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
	reportAvatarCacheStats(b, value.visualCache.Stats(), AvatarVisualCacheStats{})
}

func BenchmarkAvatarComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New("AB").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(ctx, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(512, 64)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := avatarCacheStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		gtx.Now = time.Unix(0, int64(i+1)*int64(time.Millisecond))
		gtx.Ops.Reset()
		layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	}
	b.StopTimer()
	reportAvatarCacheStats(b, avatarCacheStats(values), baseline)
}

func BenchmarkAvatarComponentTreeThemeSwitch(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	active := firstContext
	values := make([]Widget, 128)
	children := make([]layout.FlexChild, len(values))
	for index := range values {
		values[index] = New("AB").Size(SizeSmall)
		index := index
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return values[index].Layout(active, gtx)
		})
	}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(512, 64)}}
	gtx.Now = time.Unix(0, 0)
	gtx.Ops.Reset()
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
	baseline := avatarCacheStats(values)
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
	reportAvatarCacheStats(b, avatarCacheStats(values), baseline)
}

func avatarCacheStats(values []Widget) AvatarVisualCacheStats {
	var total AvatarVisualCacheStats
	for _, value := range values {
		stats := value.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func reportAvatarCacheStats(b *testing.B, after, before AvatarVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "avatar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "avatar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "avatar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "avatar-visual-cache-key-invalidations/op")
}
