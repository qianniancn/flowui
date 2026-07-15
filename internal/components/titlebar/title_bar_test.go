package titlebar

import (
	"image"
	"testing"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestTitleBarFillsWidthAndLimitsMoveRegion(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "main.go - FlowUI", fixedMenu{size: image.Pt(120, 32)})

	dims := layoutTitleBarFrame(ctx, router, bar, time.Unix(1, 0), image.Pt(600, 35))
	if dims.Size != image.Pt(600, 35) {
		t.Fatalf("title bar size = %v, want (600,35)", dims.Size)
	}
	for _, point := range []f32.Point{f32.Pt(240, 1), f32.Pt(240, 16), f32.Pt(240, 33)} {
		if action, ok := router.ActionAt(point); !ok || action != system.ActionMove {
			t.Fatalf("title move action at %v = %v, found %v", point, action, ok)
		}
	}
	for _, point := range []f32.Point{f32.Pt(20, 16), f32.Pt(580, 16)} {
		if action, ok := router.ActionAt(point); ok && action == system.ActionMove {
			t.Fatalf("non-title point %v was marked movable", point)
		}
	}
	semantics := router.AppendSemantics(nil)
	for _, label := range []string{"Minimize", "Maximize", "Close"} {
		if !hasSemanticLabel(semantics, label) {
			t.Fatalf("missing %q control semantics", label)
		}
	}
}

func TestTitleBarTracksMaximizedWindowState(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "FlowUI", nil)

	layoutTitleBarFrame(ctx, router, bar, time.Unix(2, 0), image.Pt(500, 35))
	state, ok := frame.PeekState[titleBarState](ctx, "workspace", stateSlotTitleBar)
	if !ok {
		t.Fatal("missing title bar state")
	}
	if state.decorations.Maximized {
		t.Fatal("windowed title bar reported maximized")
	}

	frame.UpdateWindowConfig(ctx, app.Config{Mode: app.Maximized})
	layoutTitleBarFrame(ctx, router, bar, time.Unix(2, 0).Add(time.Millisecond), image.Pt(500, 35))
	if !state.decorations.Maximized {
		t.Fatal("maximized title bar did not switch to restore state")
	}
}

func TestTitleBarThemeAndLabels(t *testing.T) {
	tokens := theme.DefaultTheme().Components.TitleBar
	if tokens.Height != 35 || tokens.ControlWidth != 46 || tokens.IconSize != 12 || tokens.CloseHover.A == 0 {
		t.Fatalf("title bar theme = %#v", tokens)
	}
	ctx := frame.New(nil, nil, locale.LanguageChinese)
	if got := controlLabel(ctx, system.ActionMaximize, false); got != "最大化" {
		t.Fatalf("maximize label = %q", got)
	}
	if got := controlLabel(ctx, system.ActionMaximize, true); got != "还原" {
		t.Fatalf("restore label = %q", got)
	}
	for _, action := range titleBarActions {
		if len(controlIcon(action, false)) == 0 {
			t.Fatalf("missing icon for %v", action)
		}
	}
}

func layoutTitleBarFrame(ctx *frame.Context, router *input.Router, bar Widget, now time.Time, size image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(size),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := bar.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

type fixedMenu struct {
	size image.Point
}

func (m fixedMenu) Layout(*frame.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: m.size}
}

func hasSemanticLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || hasSemanticLabel(node.Children, label) {
			return true
		}
	}
	return false
}
