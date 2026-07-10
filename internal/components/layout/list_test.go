package layoutui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestListKeepsState(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	List("items", 2, func(i int) frame.Widget {
		return text.New("item")
	}).Gap(8).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	state := testComponentState[layout.List](ctx, "items", stateSlotList)
	if state == nil {
		t.Fatal("missing list state")
	}
	if state.Axis != layout.Vertical {
		t.Fatalf("axis = %v, want vertical", state.Axis)
	}
	if state.Gap != 8 {
		t.Fatalf("gap = %d, want 8", state.Gap)
	}
}

func TestListDisabled(t *testing.T) {
	l := List("items", 1, func(int) frame.Widget {
		return text.New("item")
	}).Disabled(true)

	if !l.disabled {
		t.Fatal("list was not disabled")
	}
}

func TestListPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	List("items", 1, func(int) frame.Widget {
		return probe
	}).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("list item was laid out with enabled context")
	}
}

func TestListScrollOptions(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	List("items", 1, func(int) frame.Widget {
		return Spacer(20, 20)
	}).AlignEnd().StickToEnd().ScrollAnyAxis().Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	state := testComponentState[layout.List](ctx, "items", stateSlotList)
	if state.Alignment != layout.End {
		t.Fatalf("alignment = %v, want end", state.Alignment)
	}
	if !state.ScrollToEnd {
		t.Fatal("scroll-to-end was not enabled")
	}
	if !state.ScrollAnyAxis {
		t.Fatal("scroll-any-axis was not enabled")
	}
}

func TestListPropagatesScrollAndCrossAxisAlignment(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 100)
	listSize := image.Pt(100, 25)
	frame.BeginFrameWithViewport(ctx, viewport)
	state := ctx.ListState("items")
	frame.EndFrame(ctx)
	state.Position = layout.Position{First: 1, Offset: 5, BeforeEnd: true}
	got := make(map[int]image.Rectangle)
	gtx := layout.Context{Constraints: layout.Constraints{Max: listSize}, Ops: new(op.Ops)}
	frame.BeginFrameWithViewport(ctx, viewport)

	List("items", 5, func(index int) frame.Widget {
		width := 20
		if index == 2 {
			width = 10
		}
		return &overlayProbeWidget{
			key:    "list-" + string(rune('0'+index)),
			size:   image.Pt(width, 10),
			anchor: image.Rect(0, 0, width, 10),
			capture: func(anchor image.Rectangle) {
				got[index] = anchor
			},
		}
	}).Gap(2).AlignEnd().Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	want := image.Rect(10, 7, 20, 17)
	if got[2] != want {
		t.Fatalf("list item anchor = %v, want %v", got[2], want)
	}
	if _, visible := got[0]; visible {
		t.Fatal("offscreen leading list item produced an overlay")
	}
	if _, visible := got[4]; visible {
		t.Fatal("offscreen trailing list item produced an overlay")
	}
}
