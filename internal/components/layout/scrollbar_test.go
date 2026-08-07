package layoutui

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func TestScrollbarOptionsAndState(t *testing.T) {
	ctx := newContext(nil)
	value := Scrollbar("body", Spacer(300, 20)).
		Horizontal().
		AlignEnd().
		StickToEnd().
		ScrollAnyAxis().
		Overlay().
		Disabled(true).
		Style(flowstyle.Style{}.Width(12))
	value.Layout(ctx, testLayoutContext())

	state := testComponentState[scrollbarState](ctx, "body", stateSlotScrollbar)
	if state == nil || state.list.Axis != layout.Horizontal || state.list.Alignment != layout.End {
		t.Fatalf("scrollbar state = %#v", state)
	}
	if !state.list.ScrollToEnd || !state.list.ScrollAnyAxis || !value.overlay || !value.disabled {
		t.Fatalf("scrollbar options = %#v state = %#v", value, state.list)
	}
	if value.customStyle.Resolve(flowstyle.StyleState{}).Box == nil {
		t.Fatal("scrollbar style was not retained")
	}
}

func TestScrollbarPassesDisabledContext(t *testing.T) {
	probe := new(enabledProbeWidget)
	Scrollbar("body", probe).Disabled(true).Layout(newContext(nil), testLayoutContext())
	if probe.enabled {
		t.Fatal("scrollbar child was laid out with enabled context")
	}
}

func TestScrollbarResolvesTrackAndThumbParts(t *testing.T) {
	track := color.NRGBA{R: 1, A: 0xff}
	thumb := color.NRGBA{G: 2, A: 0xff}
	custom := flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.Width(14).Background(flowstyle.SolidColor{Color: track})).
		Part(flowstyle.PartThumb, flowstyle.Style{}.Width(4).Background(flowstyle.SolidColor{Color: thumb}).Radius(2))

	style := resolvedScrollbarStyle(newContext(nil), testLayoutContext(), "scrollbar", new(widget.Scrollbar), false, custom)
	if style.Width() != 14 || style.Track.Color != track || style.Indicator.Color != thumb || style.Indicator.MinorWidth != 4 || style.Indicator.CornerRadius != 2 {
		t.Fatalf("parts = width %v track %#v thumb %#v/%v radius %v", style.Width(), style.Track.Color, style.Indicator.Color, style.Indicator.MinorWidth, style.Indicator.CornerRadius)
	}
}

func TestScrollbarOnlyReservesSpaceWhenScrollable(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := &scrollbarConstraintWidget{height: 20}
	value := Scrollbar("standard", probe)
	start := time.Unix(1, 0)

	layoutScrollbarFrame(ctx, router, value, start)
	if probe.constraints.Max.X != 100 {
		t.Fatalf("non-scrollable content max width = %d, want 100", probe.constraints.Max.X)
	}

	probe.height = 300
	layoutScrollbarFrame(ctx, router, value, start.Add(time.Millisecond))
	layoutScrollbarFrame(ctx, router, value, start.Add(2*time.Millisecond))
	if probe.constraints.Max.X != 86 {
		t.Fatalf("scrollable content max width = %d, want 86", probe.constraints.Max.X)
	}

	probe.height = 20
	layoutScrollbarFrame(ctx, router, value, start.Add(3*time.Millisecond))
	layoutScrollbarFrame(ctx, router, value, start.Add(4*time.Millisecond))
	if probe.constraints.Max.X != 100 {
		t.Fatalf("content max width after overflow = %d, want 100", probe.constraints.Max.X)
	}
}

func TestOverlayScrollbarNeverReservesSpace(t *testing.T) {
	for _, test := range []struct {
		name string
		bar  ScrollbarWidget
		want int
	}{
		{name: "short", bar: Scrollbar("short", &scrollbarConstraintWidget{height: 20}).Overlay(), want: 100},
		{name: "long", bar: Scrollbar("long", &scrollbarConstraintWidget{height: 300}).Overlay(), want: 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			probe := test.bar.child.(*scrollbarConstraintWidget)
			ctx := newContext(nil)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutScrollbarFrame(ctx, router, test.bar, start)
			layoutScrollbarFrame(ctx, router, test.bar, start.Add(time.Millisecond))
			if probe.constraints.Max.X != test.want {
				t.Fatalf("content max width = %d, want %d", probe.constraints.Max.X, test.want)
			}
		})
	}
}

type scrollbarConstraintWidget struct {
	height      int
	constraints layout.Constraints
}

func (w *scrollbarConstraintWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, w.height))}
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

func TestScrollbarDoesNotReuseDistanceAfterContentShrinks(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := &scrollbarConstraintWidget{height: 300}
	value := Scrollbar("body", probe)
	start := time.Unix(1, 0)
	layoutScrollbarFrame(ctx, router, value, start)
	layoutScrollbarFrame(ctx, router, value, start.Add(time.Millisecond))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(95, 80)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(95, 80)},
	)
	layoutScrollbarFrame(ctx, router, value, start.Add(2*time.Millisecond))

	probe.height = 20
	layoutScrollbarFrame(ctx, router, value, start.Add(3*time.Millisecond))
	state := testComponentState[scrollbarState](ctx, "body", stateSlotScrollbar)
	if state.list.Position.First != 0 || state.list.Position.Offset != 0 {
		t.Fatalf("short content position = %#v, want start", state.list.Position)
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
