package dock

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

func BenchmarkDockLayoutSteadyState(b *testing.B) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	router := newRouterForDockBenchmark()
	root := Split("workspace", Horizontal,
		Panel("left", new(dockProbe)),
		Split("editor", Vertical,
			Panel("document", new(dockProbe)),
			Panel("inspector", new(dockProbe)),
		).Ratio(.6),
	).Ratio(.25)
	value := New("workbench", root).KeepAlive(true)
	size := image.Pt(960, 640)
	when := time.Unix(1, 0)
	layoutDockFrame(ctx, router, value, when, size)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		layoutDockFrame(ctx, router, value, when, size)
	}
}

func BenchmarkDockControlledSnapshot(b *testing.B) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	router := newRouterForDockBenchmark()
	root := Split("workspace", Horizontal,
		Panel("left", new(dockProbe)), Panel("right", new(dockProbe)),
	)
	value := New("workbench", root).Snapshot(Snapshot{
		Version: SnapshotVersion,
		RootKey: "workspace",
		Ratios:  map[string]float32{"workspace": .7},
	})
	size := image.Pt(960, 640)
	when := time.Unix(1, 0)
	layoutDockFrame(ctx, router, value, when, size)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		layoutDockFrame(ctx, router, value, when, size)
	}
}

func newRouterForDockBenchmark() *input.Router {
	return new(input.Router)
}
