package toast

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkToastComponentTree(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	provider := New("bench", []Item{
		NewItem("one", "Saved").Description("A persisted notification").Timeout(0),
		NewItem("two", "Uploading").Loading(true).Timeout(0),
		NewItem("three", "Warning").Variant(Warning).Timeout(0),
	}).Placement(BottomEnd)
	now := time.Unix(1, 0)
	layoutToastFrame(ctx, router, provider, now, image.Pt(900, 600))
	baseline := provider.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layoutToastFrame(ctx, router, provider, now.Add(time.Duration(i)*time.Millisecond), image.Pt(900, 600))
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, provider.visualCache.stats(), baseline)
}

func BenchmarkToastVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	provider := New("bench", nil)
	provider.visualWithCache(ctx, Success, nil)
	baseline := provider.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.visualWithCache(ctx, Success, nil)
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, provider.visualCache.stats(), baseline)
}

func reportToastVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "toast-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "toast-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "toast-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "toast-visual-cache-key-invalidations/op")
}

func BenchmarkToastVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	provider := New("bench", nil)
	provider.visualWithCache(firstContext, Success, nil)
	baseline := provider.visualCache.stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index & 3 {
		case 0:
			provider.visualWithCache(firstContext, Success, nil)
		case 1:
			provider.visualWithCache(firstContext, Warning, nil)
		case 2:
			provider.visualWithCache(secondContext, Warning, nil)
		default:
			provider.visualWithCache(secondContext, Success, nil)
		}
	}
	b.StopTimer()
	reportToastVisualCacheStats(b, provider.visualCache.stats(), baseline)
}
