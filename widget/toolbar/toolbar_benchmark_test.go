package toolbar

import (
	"image"
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
	"github.com/qianniancn/flowui/widget/button"
)

func BenchmarkToolbarLayoutSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	toolbar := New(
		probe{size: image.Pt(20, 10)},
		Separator(),
		probe{size: image.Pt(30, 12)},
	).Attached(true)
	layoutToolbarFrame(ctx, router, toolbar, time.Unix(0, 0))
	baseline := toolbar.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutToolbarFrame(ctx, router, toolbar, time.Unix(int64(index), 0))
	}
	b.StopTimer()
	reportToolbarVisualCacheStats(b, toolbar.visualCache.stats(), baseline)
}

func BenchmarkToolbarComponentTreeSteadyState(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	children := make([]core.Widget, 0, 128)
	for index := 0; index < cap(children); index++ {
		children = append(children, button.New("Action").Key("toolbar-benchmark-"+strconv.Itoa(index)).Size(button.Small))
	}
	toolbar := New(children...).Attached(true)
	now := time.Unix(0, 0)
	layoutToolbarFrame(ctx, router, toolbar, now)
	baseline := toolbar.visualCache.stats()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		layoutToolbarFrame(ctx, router, toolbar, now.Add(time.Duration(index+1)*time.Millisecond))
	}
	b.StopTimer()
	reportToolbarVisualCacheStats(b, toolbar.visualCache.stats(), baseline)
}

func BenchmarkToolbarVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	toolbar := New()
	toolbar.visualWithCache(ctx, nil)
	baseline := toolbar.visualCache.stats()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		toolbar.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	reportToolbarVisualCacheStats(b, toolbar.visualCache.stats(), baseline)
}

func reportToolbarVisualCacheStats(b *testing.B, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "toolbar-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "toolbar-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "toolbar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "toolbar-visual-cache-key-invalidations/op")
}

func BenchmarkToolbarVisualCacheInvalidation(b *testing.B) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	toolbar := New()
	toolbar.visualWithCache(firstContext, nil)
	override := &theme.ToolbarOverride{Declaration: style.Style{}.Padding(2)}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		switch index % 12 {
		case 0:
			toolbar.visualWithCache(firstContext, nil)
		case 1:
			toolbar.visualWithCache(firstContext, nil)
		case 2:
			toolbar.visualWithCache(secondContext, nil)
		case 3:
			toolbar.visualWithCache(secondContext, nil)
		case 4, 5:
			toolbar.Orientation(Vertical).visualWithCache(secondContext, nil)
		case 6, 7:
			toolbar.Attached(true).visualWithCache(secondContext, nil)
		case 8, 9:
			toolbar.Attached(true).visualWithCache(secondContext, override)
		case 10, 11:
			toolbar.Attached(true).visualWithCache(secondContext, nil)
		}
	}
	b.StopTimer()
	reportToolbarVisualCacheStats(b, toolbar.visualCache.stats(), computed.ProviderCacheStats{})
}
