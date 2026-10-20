package description

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkDescriptionLayout(b *testing.B) {
	themeValue := testtheme.Behavior()
	ctx := core.New(nil, themeValue, locale.LanguageEnglish)
	value := New("Supporting text")
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Ops:         &ops,
	}
	value.Layout(ctx, gtx)
	ops.Reset()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		ops.Reset()
		value.Layout(ctx, gtx)
	}
}

func BenchmarkDescriptionVisualCacheHit(b *testing.B) {
	themeValue := testtheme.Behavior()
	ctx := core.New(nil, themeValue, locale.LanguageEnglish)
	value := New("Supporting text")
	value.visualWithCache(ctx, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		value.visualWithCache(ctx, nil)
	}
	b.StopTimer()
	stats := value.visualCache.stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "description-visual-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "description-visual-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/float64(b.N), "description-visual-cache-theme-invalidations/op")
}
