package layoutui

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestScrollKeepsState(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll("body", Spacer(20, 200)).Horizontal().Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	state := testComponentState[layout.List](ctx, "body", stateSlotScroll)
	if state == nil {
		t.Fatal("missing scroll state")
	}
	if state.Axis != layout.Horizontal {
		t.Fatalf("axis = %v, want horizontal", state.Axis)
	}
}

func TestScrollPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	Scroll("body", probe).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("scroll child was laid out with enabled context")
	}
}

func TestScrollOptions(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll("body", Spacer(20, 200)).
		AlignMiddle().
		StickToEnd().
		ScrollAnyAxis().
		Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	state := testComponentState[layout.List](ctx, "body", stateSlotScroll)
	if state.Alignment != layout.Middle {
		t.Fatalf("alignment = %v, want middle", state.Alignment)
	}
	if !state.ScrollToEnd {
		t.Fatal("scroll-to-end was not enabled")
	}
	if !state.ScrollAnyAxis {
		t.Fatal("scroll-any-axis was not enabled")
	}
}

func TestScrollTrackClickUpdatesExistingListState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := Scroll("body", Spacer(100, 300))
	start := time.Unix(1, 0)
	layoutScrollFrame(ctx, router, value, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(95, 80)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(95, 80)},
	)
	layoutScrollFrame(ctx, router, value, start.Add(time.Millisecond))
	layoutScrollFrame(ctx, router, value, start.Add(2*time.Millisecond))

	state := testComponentState[layout.List](ctx, "body", stateSlotScroll)
	if state.Position.Offset <= 0 {
		t.Fatalf("scroll offset = %d, want positive after track click", state.Position.Offset)
	}
}

func layoutScrollFrame(ctx *frame.Context, router *input.Router, value ScrollWidget, now time.Time) layout.Dimensions {
	viewport := image.Pt(100, 100)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := value.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func TestScrollPropagatesContentOffset(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 100)
	scrollSize := image.Pt(100, 40)
	frame.BeginFrameWithViewport(ctx, viewport)
	state := ctx.ScrollState("body")
	frame.EndFrame(ctx)
	state.Position = layout.Position{Offset: 30, BeforeEnd: true}
	var got image.Rectangle
	probe := &overlayProbeWidget{
		key:    "scroll",
		size:   image.Pt(20, 100),
		anchor: image.Rect(0, 35, 20, 45),
		got:    &got,
	}
	gtx := layout.Context{Constraints: layout.Constraints{Max: scrollSize}, Ops: new(op.Ops)}
	frame.BeginFrameWithViewport(ctx, viewport)
	Scroll("body", probe).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	want := image.Rect(0, 5, 20, 15)
	if got != want {
		t.Fatalf("scroll anchor = %v, want %v", got, want)
	}

	frame.BeginFrameWithViewport(ctx, viewport)
	state = ctx.ScrollState("hidden-body")
	frame.EndFrame(ctx)
	state.Position = layout.Position{Offset: 30, BeforeEnd: true}
	called := false
	hidden := &overlayProbeWidget{
		key:    "hidden-scroll",
		size:   image.Pt(20, 100),
		anchor: image.Rect(0, 80, 20, 90),
		capture: func(image.Rectangle) {
			called = true
		},
	}
	gtx.Ops.Reset()
	frame.BeginFrameWithViewport(ctx, viewport)
	Scroll("hidden-body", hidden).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)
	if called {
		t.Fatal("content outside the local scroll viewport produced an overlay")
	}
}
