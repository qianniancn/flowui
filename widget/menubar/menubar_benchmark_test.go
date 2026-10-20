package menubar

import (
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func BenchmarkMenubarLayout(b *testing.B) {
	benchmarkMenubarLayout(b)
}

func BenchmarkMenubarLayoutSteadyState(b *testing.B) {
	benchmarkMenubarLayout(b)
}

func benchmarkMenubarLayout(b *testing.B) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(1, 0)
	layoutMenubarFrame(ctx, router, widget, now)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		layoutMenubarFrame(ctx, router, widget, now)
	}
}

func BenchmarkMenubarVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	cache := &menubarVisualCache{}
	widget := Widget{}
	widget.visualWithCache(ctx, nil, cache)
	before := cache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		widget.visualWithCache(ctx, nil, cache)
	}
	b.StopTimer()
	reportMenubarVisualCacheStats(b, cache.stats(), before)
}

func BenchmarkMenubarVisualCacheInvalidation(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	cache := &menubarVisualCache{}
	widget := Widget{}
	widget.visualWithCache(ctx, nil, cache)
	override := &theme.MenubarOverride{Declaration: style.Style{}.Padding(2)}
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index % 8 {
		case 0, 1:
			core.ReplaceTheme(ctx, *first)
			widget.compact = false
			widget.visualWithCache(ctx, nil, cache)
		case 2, 3:
			core.ReplaceTheme(ctx, *second)
			widget.compact = true
			widget.visualWithCache(ctx, nil, cache)
		case 4, 5:
			widget.compact = false
			widget.visualWithCache(ctx, override, cache)
		case 6, 7:
			widget.compact = false
			widget.visualWithCache(ctx, nil, cache)
		}
	}
	b.StopTimer()
	reportMenubarVisualCacheStats(b, cache.stats(), computed.ProviderCacheStats{})
}

func BenchmarkMenubarVisualOverrideBypass(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	widget := Widget{}
	override := &theme.MenubarOverride{Declaration: style.Style{}.Padding(2)}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		widget.visualWithCache(ctx, override, nil)
	}
}

func reportMenubarVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "menubar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "menubar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "menubar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "menubar-visual-cache-key-invalidations/op")
}

func BenchmarkMenubarComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	items := make([]Item, 0, 32)
	for index := 0; index < cap(items); index++ {
		items = append(items, NewMenu("item-"+strconv.Itoa(index), "Item "+strconv.Itoa(index), nil))
	}
	widget := New("benchmark", items)
	now := time.Unix(0, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutMenubarFrame(ctx, router, widget, now.Add(time.Duration(index+1)*time.Millisecond))
	}
}
