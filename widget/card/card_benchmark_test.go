package card

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

func BenchmarkCardLayout(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(
		fixedWidget{size: image.Pt(80, 20)},
		fixedWidget{size: image.Pt(80, 20)},
		fixedWidget{size: image.Pt(80, 20)},
	).Variant(Secondary)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Ops:         new(op.Ops),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		value.Layout(ctx, gtx)
	}
}

func BenchmarkCardVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(nil).Variant(Secondary)
	value.visualWithCache(ctx, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		value.visualWithCache(ctx, nil)
	}
}

type cardBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	ops    op.Ops
	gtx    layout.Context
}

func newCardComponentTreeFrame(count int) *cardBenchmarkFrame {
	frame := &cardBenchmarkFrame{ctx: core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(640, 480)},
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		variant := Default
		if index%3 == 1 {
			variant = Secondary
		} else if index%3 == 2 {
			variant = Tertiary
		}
		frame.values[index] = New(
			fixedWidget{size: image.Pt(80, 20)},
			fixedWidget{size: image.Pt(80, 20)},
		).Variant(variant)
	}
	return frame
}

func (frame *cardBenchmarkFrame) layoutTree() layout.Dimensions {
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	children := make([]layout.FlexChild, len(frame.values))
	for index := range frame.values {
		value := frame.values[index]
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(frame.ctx, gtx)
		})
	}
	dimensions := layout.Flex{Axis: layout.Vertical}.Layout(frame.gtx, children...)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	return dimensions
}

func (frame *cardBenchmarkFrame) cacheStats() [4]uint64 {
	var result [4]uint64
	for _, value := range frame.values {
		stats := value.visualCache.stats()
		result[0] += stats.Hits
		result[1] += stats.Misses
		result[2] += stats.ThemeInvalidations
		result[3] += stats.KeyInvalidations
	}
	return result
}

func reportCardCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "card-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "card-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "card-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "card-visual-cache-key-invalidations/op")
}

func BenchmarkCardComponentTreeSteadyState(b *testing.B) {
	frame := newCardComponentTreeFrame(32)
	frame.layoutTree()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkCardComponentTreeVisualCacheHit(b *testing.B) {
	frame := newCardComponentTreeFrame(32)
	frame.layoutTree()
	baseline := frame.cacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
	b.StopTimer()
	after := frame.cacheStats()
	reportCardCacheStats(b, [4]uint64{after[0], after[1], after[2], after[3]}, [4]uint64{baseline[0], baseline[1], baseline[2], baseline[3]})
}

func BenchmarkCardComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newCardComponentTreeFrame(32)
	frame.layoutTree()
	baseline := frame.cacheStats()
	light := *testtheme.Behavior()
	dark := *testtheme.Behavior()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, light)
		} else {
			core.ReplaceTheme(frame.ctx, dark)
		}
		frame.layoutTree()
	}
	b.StopTimer()
	after := frame.cacheStats()
	reportCardCacheStats(b, [4]uint64{after[0], after[1], after[2], after[3]}, [4]uint64{baseline[0], baseline[1], baseline[2], baseline[3]})
}
