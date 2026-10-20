package nodegraph

import (
	"image"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type nodeGraphBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newNodeGraphBenchmarkFrame(maximum image.Point) *nodeGraphBenchmarkFrame {
	active := testtheme.Behavior()
	frame := &nodeGraphBenchmarkFrame{ctx: core.New(nil, active, locale.LanguageEnglish)}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *nodeGraphBenchmarkFrame) layout(value core.Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func nodeGraphBenchmarkGraph(index int) Graph {
	source := NewNode("source", "Source", Point{X: 24 + float32(index%3)*8, Y: 32}).
		WithSize(140, 88).
		Outputs(NewPort("out", "Out"))
	target := NewNode("target", "Target", Point{X: 280, Y: 48 + float32(index%2)*12}).
		WithSize(140, 88).
		Inputs(NewPort("in", "In"))
	return Graph{
		Nodes: []Node{source, target},
		Edges: []Edge{NewEdge("edge", NewEndpoint("source", "out"), NewEndpoint("target", "in"))},
	}
}

func nodeGraphBenchmarkValue(index int) Widget {
	return New("graph-"+strconv.Itoa(index), nodeGraphBenchmarkGraph(index)).
		Grid(false).
		Controls(true).
		Minimap(true)
}

type nodeGraphBenchmarkTreeWidget struct {
	values []Widget
}

func (t nodeGraphBenchmarkTreeWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(t.values))
	for _, value := range t.values {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return value.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Gap: 4}.Layout(gtx, children...)
}

func nodeGraphBenchmarkTree() nodeGraphBenchmarkTreeWidget {
	values := make([]Widget, 0, 24)
	for index := range 24 {
		values = append(values, nodeGraphBenchmarkValue(index))
	}
	return nodeGraphBenchmarkTreeWidget{values: values}
}

func reportNodeGraphVisualStats(b *testing.B, after, before NodeGraphVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "nodegraph-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "nodegraph-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "nodegraph-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "nodegraph-visual-cache-key-invalidations/op")
}

func BenchmarkNodeGraphProviderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	cache := nodeGraphVisualCache{}
	cache.cache.Resolve(active, struct{}{}, func() theme.NodeGraphVisual {
		provider, err := theme.ComponentProvider(active)
		if err != nil {
			b.Fatal(err)
		}
		visual := provider.NodeGraph(active, theme.NodeGraphRequest{})
		theme.ValidateNodeGraphVisual(visual)
		return visual
	})
	before := cache.Stats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.cache.Resolve(active, struct{}{}, func() theme.NodeGraphVisual {
			b.Helper()
			panic("node graph visual cache miss")
		})
	}
	b.StopTimer()
	reportNodeGraphVisualStats(b, cache.Stats(), before)
}

func BenchmarkNodeGraphLayoutSteadyState(b *testing.B) {
	frame := newNodeGraphBenchmarkFrame(image.Pt(640, 360))
	value := nodeGraphBenchmarkValue(0)
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkNodeGraphComponentTreeSteadyState(b *testing.B) {
	frame := newNodeGraphBenchmarkFrame(image.Pt(720, 24000))
	value := nodeGraphBenchmarkTree()
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkNodeGraphComponentTreeVisualCacheHit(b *testing.B) {
	frame := newNodeGraphBenchmarkFrame(image.Pt(720, 24000))
	value := nodeGraphBenchmarkTree()
	frame.layout(value)
	before := nodeGraphTreeVisualStats(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportNodeGraphVisualStats(b, nodeGraphTreeVisualStats(value), before)
}

func nodeGraphTreeVisualStats(value nodeGraphBenchmarkTreeWidget) NodeGraphVisualCacheStats {
	var total NodeGraphVisualCacheStats
	for _, item := range value.values {
		stats := item.visualCache.Stats()
		total.Hits += stats.Hits
		total.Misses += stats.Misses
		total.ThemeInvalidations += stats.ThemeInvalidations
		total.KeyInvalidations += stats.KeyInvalidations
	}
	return total
}

func BenchmarkNodeGraphComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newNodeGraphBenchmarkFrame(image.Pt(720, 24000))
	value := nodeGraphBenchmarkTree()
	frame.layout(value)
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()
	before := nodeGraphTreeVisualStats(value)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			core.ReplaceTheme(frame.ctx, first)
		} else {
			core.ReplaceTheme(frame.ctx, second)
		}
		frame.layout(value)
	}
	b.StopTimer()
	reportNodeGraphVisualStats(b, nodeGraphTreeVisualStats(value), before)
}

func BenchmarkNodeGraphStyledOverrideBypass(b *testing.B) {
	frame := newNodeGraphBenchmarkFrame(image.Pt(640, 360))
	value := (nodeGraphBenchmarkValue(0)).Style(style.Style{}.PaddingX(2))
	frame.layout(value)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}
