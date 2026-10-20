package tree

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
	"github.com/qianniancn/flowui/theme"
)

type treeCheckboxBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newTreeCheckboxBenchmarkFrame(maximum image.Point) *treeCheckboxBenchmarkFrame {
	activeTheme := *testtheme.Behavior()
	frame := &treeCheckboxBenchmarkFrame{
		ctx: core.New(nil, &activeTheme, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *treeCheckboxBenchmarkFrame) layout(value Widget) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkTreeCheckboxCompositeLayout(b *testing.B) {
	frame := newTreeCheckboxBenchmarkFrame(image.Pt(720, 360))
	items := make([]Item, 0, 16)
	expanded := make([]string, 0, 16)
	checked := make([]string, 0, 16)
	for rootIndex := range 16 {
		rootKey := "root-" + strconv.Itoa(rootIndex)
		root := Item{Key: rootKey, Label: "Root " + strconv.Itoa(rootIndex)}
		expanded = append(expanded, rootKey)
		for childIndex := range 8 {
			childKey := rootKey + "-child-" + strconv.Itoa(childIndex)
			root.Children = append(root.Children, Item{Key: childKey, Label: "Child " + strconv.Itoa(childIndex)})
			if childIndex == 2 {
				checked = append(checked, childKey)
			}
		}
		items = append(items, root)
	}
	value := New("checkbox-tree", "", items).
		Checkable(true).
		ExpandedKeys(expanded).
		CheckedKeys(checked).
		DataVersion(1).
		MaxHeight(320)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	state := treePeekState(frame.ctx, value.key)
	if state == nil {
		b.Fatal("tree state was not initialized")
	}
	stats := state.checkboxVisual.Stats()
	b.ReportMetric(float64(stats.Hits)/float64(b.N), "checkbox-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/float64(b.N), "checkbox-cache-misses/op")
}

func treeBenchmarkValue() Widget {
	items := make([]Item, 0, 16)
	expanded := make([]string, 0, 16)
	checked := make([]string, 0, 16)
	for rootIndex := range 16 {
		rootKey := "root-" + strconv.Itoa(rootIndex)
		root := Item{Key: rootKey, Label: "Root " + strconv.Itoa(rootIndex)}
		expanded = append(expanded, rootKey)
		for childIndex := range 8 {
			childKey := rootKey + "-child-" + strconv.Itoa(childIndex)
			root.Children = append(root.Children, Item{Key: childKey, Label: "Child " + strconv.Itoa(childIndex)})
			if childIndex == 2 {
				checked = append(checked, childKey)
			}
		}
		items = append(items, root)
	}
	return New("tree-benchmark", "", items).
		Checkable(true).
		ExpandedKeys(expanded).
		CheckedKeys(checked).
		DataVersion(1).
		MaxHeight(320)
}

func treeVisualStats(frame *treeCheckboxBenchmarkFrame, key string) treeVisualCacheStats {
	state := treePeekState(frame.ctx, key)
	if state == nil {
		return treeVisualCacheStats{}
	}
	return state.visualCache.Stats()
}

func reportTreeVisualStats(b *testing.B, after, before treeVisualCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), "tree-visual-cache-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), "tree-visual-cache-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), "tree-visual-cache-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), "tree-visual-cache-key-invalidations/op")
}

func BenchmarkTreeProviderVisualCacheHit(b *testing.B) {
	active := testtheme.Behavior()
	cache := treeVisualCache{}
	request := theme.TreeRequest{Variant: theme.TreeVariantDefault, Size: theme.TreeSizeMedium}
	cache.resolve(active, request)
	before := cache.Stats()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.resolve(active, request)
	}
	b.StopTimer()
	reportTreeVisualStats(b, cache.Stats(), before)
}

func BenchmarkTreeComponentTreeSteadyState(b *testing.B) {
	frame := newTreeCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := treeBenchmarkValue()
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkTreeComponentTreeVisualCacheHit(b *testing.B) {
	frame := newTreeCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := treeBenchmarkValue()
	frame.layout(value)
	before := treeVisualStats(frame, value.key)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
	b.StopTimer()
	reportTreeVisualStats(b, treeVisualStats(frame, value.key), before)
}

func BenchmarkTreeComponentTreeVisualCacheMiss(b *testing.B) {
	frame := newTreeCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := treeBenchmarkValue()
	frame.layout(value)
	before := treeVisualStats(frame, value.key)
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()

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
	reportTreeVisualStats(b, treeVisualStats(frame, value.key), before)
}

func BenchmarkTreeComponentTreeVisualCacheKeyInvalidation(b *testing.B) {
	frame := newTreeCheckboxBenchmarkFrame(image.Pt(720, 360))
	value := treeBenchmarkValue()
	alternate := value.Variant(VariantSurface).Size(SizeSmall)
	frame.layout(value)
	before := treeVisualStats(frame, value.key)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; b.Loop(); index++ {
		if index&1 == 0 {
			frame.layout(value)
		} else {
			frame.layout(alternate)
		}
	}
	b.StopTimer()
	reportTreeVisualStats(b, treeVisualStats(frame, value.key), before)
}
