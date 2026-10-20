package statusbar

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

func BenchmarkStatusBarLayout(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(nil, nil)
	gtx := layout.Context{
		Constraints: layout.Constraints{Min: image.Pt(640, 0), Max: image.Pt(640, 100)},
		Ops:         new(op.Ops),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		value.Layout(ctx, gtx)
	}
}

func BenchmarkStatusBarVisualCacheHit(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	value := New(nil, nil)
	value.visualWithCache(ctx, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		value.visualWithCache(ctx, nil)
	}
}

type statusBarBenchmarkFrame struct {
	ctx    *core.Context
	values []Widget
	ops    op.Ops
	gtx    layout.Context
}

type benchmarkStatusBarChild struct{}

func (benchmarkStatusBarChild) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(96, 16))}
}

func newStatusBarComponentTreeFrame(count int) *statusBarBenchmarkFrame {
	frame := &statusBarBenchmarkFrame{ctx: core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Min: image.Pt(640, 0), Max: image.Pt(640, 480)},
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	frame.values = make([]Widget, count)
	for index := range frame.values {
		frame.values[index] = New(benchmarkStatusBarChild{}, benchmarkStatusBarChild{})
	}
	return frame
}

func (frame *statusBarBenchmarkFrame) layoutTree() layout.Dimensions {
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

func (frame *statusBarBenchmarkFrame) cacheStats() [4]uint64 {
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

func reportStatusBarCacheStats(b *testing.B, after, before [4]uint64) {
	if b.N == 0 {
		return
	}
	b.ReportMetric(float64(after[0]-before[0])/float64(b.N), "statusbar-visual-cache-hits/op")
	b.ReportMetric(float64(after[1]-before[1])/float64(b.N), "statusbar-visual-cache-misses/op")
	b.ReportMetric(float64(after[2]-before[2])/float64(b.N), "statusbar-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after[3]-before[3])/float64(b.N), "statusbar-visual-cache-key-invalidations/op")
}

func BenchmarkStatusBarComponentTreeSteadyState(b *testing.B) {
	frame := newStatusBarComponentTreeFrame(16)
	frame.layoutTree()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
}

func BenchmarkStatusBarComponentTreeVisualCacheHit(b *testing.B) {
	frame := newStatusBarComponentTreeFrame(16)
	frame.layoutTree()
	baseline := frame.cacheStats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layoutTree()
	}
	b.StopTimer()
	after := frame.cacheStats()
	reportStatusBarCacheStats(b, [4]uint64{after[0], after[1], after[2], after[3]}, [4]uint64{baseline[0], baseline[1], baseline[2], baseline[3]})
}

func BenchmarkStatusBarComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newStatusBarComponentTreeFrame(16)
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
	reportStatusBarCacheStats(b, [4]uint64{after[0], after[1], after[2], after[3]}, [4]uint64{baseline[0], baseline[1], baseline[2], baseline[3]})
}
