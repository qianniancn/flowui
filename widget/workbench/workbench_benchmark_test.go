package workbench

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/dock"
)

func BenchmarkWorkbenchSnapshot(b *testing.B) {
	state := benchmarkWorkbenchState()
	b.ReportAllocs()
	var snapshot Snapshot
	b.ResetTimer()
	for b.Loop() {
		snapshot = state.Snapshot()
	}
	b.StopTimer()
	if len(snapshot.Groups) != 4 {
		b.Fatalf("snapshot groups = %d, want 4", len(snapshot.Groups))
	}
}

func BenchmarkWorkbenchSnapshotClone(b *testing.B) {
	snapshot := benchmarkWorkbenchState().Snapshot()
	b.ReportAllocs()
	var clone Snapshot
	b.ResetTimer()
	for b.Loop() {
		clone = snapshot.Clone()
	}
	b.StopTimer()
	if len(clone.Groups) != len(snapshot.Groups) {
		b.Fatalf("clone groups = %d, want %d", len(clone.Groups), len(snapshot.Groups))
	}
}

func BenchmarkWorkbenchCommandsFor(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	controller := NewController(benchmarkWorkbenchState())
	b.ReportAllocs()
	var commands []command.Command
	b.ResetTimer()
	for b.Loop() {
		commands = controller.CommandsFor(ctx)
	}
	b.StopTimer()
	if len(commands) != 6 {
		b.Fatalf("commands = %d, want 6", len(commands))
	}
}

func BenchmarkWorkbenchCommandScopeShortcutDispatch(b *testing.B) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	controller := NewController(benchmarkWorkbenchState())
	router := new(input.Router)
	scope := controller.CommandScopeFor(ctx, nil)
	workbenchBenchmarkFrame(ctx, router, scope, time.Unix(1, 0))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		router.Queue(key.Event{Name: key.NameTab, Modifiers: key.ModShortcut, State: key.Press})
		workbenchBenchmarkFrame(ctx, router, scope, time.Unix(1, 0))
	}
	b.StopTimer()
	if controller.State().ActiveTab("group-0") == "" {
		b.Fatal("shortcut dispatch lost active tab")
	}
}

func benchmarkWorkbenchState() State {
	groups := make([]Group, 4)
	for groupIndex := range groups {
		groups[groupIndex].Key = "group-" + string(rune('0'+groupIndex))
		groups[groupIndex].Tabs = make([]Tab, 16)
		for tabIndex := range groups[groupIndex].Tabs {
			groups[groupIndex].Tabs[tabIndex] = Tab{
				Key:      groups[groupIndex].Key + "-tab-" + string(rune('a'+tabIndex)),
				Closable: true,
			}
		}
		groups[groupIndex].ActiveKey = groups[groupIndex].Tabs[0].Key
	}
	state := NewState(groups)
	state.SetDockSnapshot(benchmarkDockSnapshot())
	return state
}

func benchmarkDockSnapshot() dock.Snapshot {
	return dock.Snapshot{
		Version:      dock.SnapshotVersion,
		RootKey:      "root",
		Ratios:       map[string]float32{"root": .5, "left": .4, "right": .6},
		Collapsed:    map[string]dock.CollapseState{"left": {First: true}},
		MaximizedKey: "group-0",
	}
}

func workbenchBenchmarkFrame(ctx *core.Context, router *input.Router, value core.Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(800, 600)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	dimensions := value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dimensions
}
