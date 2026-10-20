package menu

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

func TestWrapSelectableTextOpensMenuAndKeepsHighlight(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	var sel widget.Selectable
	sel.SetText("Copy me")
	sel.SetCaret(0, 4)
	field := core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(200, 30))}
	})
	queue := func(core.SelectableEditAction) {}
	start := time.Unix(5, 0)

	var ops op.Ops
	frame := func(now time.Time, overlays bool) {
		gtx := layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 260)},
			Source:      router.Source(),
			Ops:         &ops,
			Now:         now,
		}
		core.BeginFrame(ctx)
		wrapSelectableText(ctx, gtx, "sel", field, &sel, queue)
		if overlays {
			core.LayoutOverlays(ctx, gtx)
		}
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(&ops)
	}

	frame(start, false)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 15),
	})
	frame(start.Add(time.Millisecond), true)
	frame(start.Add(150*time.Millisecond+time.Millisecond), true)
	if !core.HasTopOverlay(ctx) {
		t.Fatal("right-click did not register the selectable edit menu overlay")
	}
	if !sel.PaintSelectionWhenUnfocused {
		t.Fatal("open menu did not keep the selection highlight enabled")
	}

	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(290, 250)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(290, 250)},
	)
	frame(start.Add(200*time.Millisecond), true)
	frame(start.Add(350*time.Millisecond), true)
	if sel.PaintSelectionWhenUnfocused {
		t.Fatal("closed menu did not clear the selection highlight flag")
	}
}
