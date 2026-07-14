package layoutui

import (
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestScrollbarOptionsAndState(t *testing.T) {
	ctx := newContext(nil)
	value := Scrollbar("body", Spacer(300, 20)).
		Horizontal().
		AlignEnd().
		StickToEnd().
		ScrollAnyAxis().
		Overlay().
		Disabled(true)
	value.Layout(ctx, testLayoutContext())

	state := testComponentState[scrollbarState](ctx, "body", stateSlotScrollbar)
	if state == nil || state.list.Axis != layout.Horizontal || state.list.Alignment != layout.End {
		t.Fatalf("scrollbar state = %#v", state)
	}
	if !state.list.ScrollToEnd || !state.list.ScrollAnyAxis || !value.overlay || !value.disabled {
		t.Fatalf("scrollbar options = %#v state = %#v", value, state.list)
	}
}

func TestScrollbarPassesDisabledContext(t *testing.T) {
	probe := new(enabledProbeWidget)
	Scrollbar("body", probe).Disabled(true).Layout(newContext(nil), testLayoutContext())
	if probe.enabled {
		t.Fatal("scrollbar child was laid out with enabled context")
	}
}

func TestScrollbarThemeMatchesHeroUIThinStyle(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Scrollbar
	if tokens.TrackWidth != 10 || tokens.ThumbWidth != 6 || tokens.ContentGap != 4 || tokens.MinThumbLength != 32 || tokens.Radius != 3 || tokens.ThumbOpacity != .15 {
		t.Fatalf("scrollbar theme = %#v", tokens)
	}
	style := scrollbarStyleFor(&activeTheme, new(widget.Scrollbar), false)
	if style.Width() != 10 || style.Indicator.Color.A != 38 || style.Track.Color.A != 0 {
		t.Fatalf("scrollbar style width/color = %v/%#v", style.Width(), style.Indicator.Color)
	}
}

func TestScrollbarReservesContentGapOutsideOverlayMode(t *testing.T) {
	for _, test := range []struct {
		name string
		bar  ScrollbarWidget
		want int
	}{
		{name: "standard", bar: Scrollbar("standard", new(constraintWidget)), want: 86},
		{name: "overlay", bar: Scrollbar("overlay", new(constraintWidget)).Overlay(), want: 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe := test.bar.child.(*constraintWidget)
			test.bar.Layout(newContext(nil), layout.Context{
				Constraints: layout.Exact(image.Pt(100, 100)),
				Ops:         new(op.Ops),
			})
			if probe.constraints.Max.X != test.want {
				t.Fatalf("content max width = %d, want %d", probe.constraints.Max.X, test.want)
			}
		})
	}
}

func TestScrollbarViewportUsesListPosition(t *testing.T) {
	start, end := scrollbarViewport(layout.Position{
		First:      0,
		Offset:     150,
		OffsetLast: -50,
		Count:      1,
		Length:     300,
	}, 1, 100)
	if math.Abs(float64(start-.5)) > 1e-6 || math.Abs(float64(end-5.0/6.0)) > 1e-6 {
		t.Fatalf("viewport = %v..%v, want .5..%v", start, end, 5.0/6.0)
	}
	if start, end := scrollbarViewport(layout.Position{Length: 80}, 1, 100); start != 0 || end != 1 {
		t.Fatalf("non-scrollable viewport = %v..%v", start, end)
	}
}

func TestScrollbarTrackClickScrollsContent(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := Scrollbar("body", Spacer(100, 300))
	start := time.Unix(1, 0)
	layoutScrollbarFrame(ctx, router, value, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(95, 80)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(95, 80)},
	)
	layoutScrollbarFrame(ctx, router, value, start.Add(time.Millisecond))
	layoutScrollbarFrame(ctx, router, value, start.Add(2*time.Millisecond))

	state := testComponentState[scrollbarState](ctx, "body", stateSlotScrollbar)
	if state.list.Position.Offset <= 0 {
		t.Fatalf("scroll offset = %d, want positive after track click", state.list.Position.Offset)
	}
}

func layoutScrollbarFrame(ctx *frame.Context, router *input.Router, value ScrollbarWidget, now time.Time) layout.Dimensions {
	viewport := image.Pt(100, 100)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := value.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
