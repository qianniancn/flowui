package gantt

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type ganttBenchmarkFrame struct {
	ctx *core.Context
	ops op.Ops
	gtx layout.Context
}

func newGanttBenchmarkFrame(maximum image.Point) *ganttBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &ganttBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Ops:         &frame.ops, Now: time.Unix(1, 0),
	}
	return frame
}

func (f *ganttBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	return dimensions
}

func ganttBenchmarkValue(index int) Widget {
	base := ganttDate(1).AddDate(0, 0, index%8)
	tasks := []Task{
		NewTask("plan-"+strconv.Itoa(index), "Plan "+strconv.Itoa(index), base, base.Add(3*24*time.Hour)).Group("planning").Progress(.65),
		NewTask("ship-"+strconv.Itoa(index), "Ship "+strconv.Itoa(index), base.Add(3*24*time.Hour), base.Add(6*24*time.Hour)).Group("delivery").DependsOn("plan-" + strconv.Itoa(index)),
		NewMilestone("review-"+strconv.Itoa(index), "Review "+strconv.Itoa(index), base.Add(7*24*time.Hour)).DependsOn("ship-" + strconv.Itoa(index)),
	}
	return New("gantt-"+strconv.Itoa(index), tasks).Animation(false).Legend(true).TaskLabels(true)
}

func ganttBenchmarkTree() []Widget {
	values := make([]Widget, 32)
	for index := range values {
		values[index] = ganttBenchmarkValue(index)
	}
	return values
}

func layoutGanttBenchmarkTree(frame *ganttBenchmarkFrame, values []Widget) {
	children := make([]layout.FlexChild, len(values))
	for index, value := range values {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(frame.ctx, gtx)
		})
	}
	layout.Flex{Axis: layout.Vertical}.Layout(frame.gtx, children...)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
}

func reportGanttVisualStats(b *testing.B, after, before GanttVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "gantt-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "gantt-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "gantt-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "gantt-visual-cache-key-invalidations/op")
}

func BenchmarkGanttProviderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	cache := ganttVisualCache{}
	cache.cache.Resolve(active, struct{}{}, func() theme.GanttVisual {
		provider, err := theme.ComponentProvider(active)
		if err != nil {
			b.Fatal(err)
		}
		visual := provider.Gantt(active, theme.GanttRequest{})
		theme.ValidateGanttVisual(visual)
		return visual
	})
	before := cache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.cache.Resolve(active, struct{}{}, func() theme.GanttVisual {
			b.Helper()
			panic("gantt visual cache miss")
		})
	}
	b.StopTimer()
	reportGanttVisualStats(b, cache.Stats(), before)
}

func BenchmarkGanttLayoutSteadyState(b *testing.B) {
	frame := newGanttBenchmarkFrame(image.Pt(720, 480))
	value := ganttBenchmarkValue(0)
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkGanttComponentTreeSteadyState(b *testing.B) {
	frame := newGanttBenchmarkFrame(image.Pt(720, 24000))
	values := ganttBenchmarkTree()
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	layoutGanttBenchmarkTree(frame, values)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.ops.Reset()
		core.BeginFrame(frame.ctx)
		layoutGanttBenchmarkTree(frame, values)
	}
}

func ganttTreeVisualStats(values []Widget) GanttVisualCacheStats {
	var total GanttVisualCacheStats
	for _, value := range values {
		stats := value.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func BenchmarkGanttComponentTreeVisualCacheHit(b *testing.B) {
	frame := newGanttBenchmarkFrame(image.Pt(720, 24000))
	values := ganttBenchmarkTree()
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	layoutGanttBenchmarkTree(frame, values)
	before := ganttTreeVisualStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.ops.Reset()
		core.BeginFrame(frame.ctx)
		layoutGanttBenchmarkTree(frame, values)
	}
	b.StopTimer()
	reportGanttVisualStats(b, ganttTreeVisualStats(values), before)
}

func BenchmarkGanttComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newGanttBenchmarkFrame(image.Pt(720, 24000))
	values := ganttBenchmarkTree()
	frame.ops.Reset()
	core.BeginFrame(frame.ctx)
	layoutGanttBenchmarkTree(frame, values)
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	before := ganttTreeVisualStats(values)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, first)
		} else {
			core.ReplaceTheme(frame.ctx, second)
		}
		frame.ops.Reset()
		core.BeginFrame(frame.ctx)
		layoutGanttBenchmarkTree(frame, values)
	}
	b.StopTimer()
	reportGanttVisualStats(b, ganttTreeVisualStats(values), before)
}

func BenchmarkGanttStyledOverrideBypass(b *testing.B) {
	frame := newGanttBenchmarkFrame(image.Pt(720, 480))
	styled := (ganttBenchmarkValue(0)).Style(style.Style{}.PaddingX(2))
	frame.layout(styled)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(styled)
	}
}
