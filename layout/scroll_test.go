package layout

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func TestScrollKeepsState(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll(Spacer(20, 200)).Key("body").Horizontal().Layout(ctx, layout.Context{
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

	Scroll(probe).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("scroll child was laid out with enabled context")
	}
}

func TestScrollOptions(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll(Spacer(20, 200)).
		Key("body").
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

func TestScrollOverlayKeepsContentWidth(t *testing.T) {
	probe := &scrollbarConstraintWidget{height: 300}
	var ops op.Ops

	Scroll(probe).
		Key("overlay").
		Overlay().
		Layout(newContext(nil), layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	if got := probe.constraints.Max.X; got != 100 {
		t.Fatalf("overlay content width = %d, want 100", got)
	}
}

func TestScrollReservesShadowOutsetAtViewportEdges(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	probe := new(constraintWidget)
	value := Scroll(Box(probe).Style(style.Style{}.
		FillWidth().
		Height(20).
		BoxShadow(0, 2, 4, 0, style.RGBA(0x00000080)),
	))
	start := time.Unix(1, 0)

	layoutScrollFrame(ctx, router, value, start)
	if probe.constraints.Max.X != 100 {
		t.Fatalf("initial child width = %d, want 100", probe.constraints.Max.X)
	}
	layoutScrollFrame(ctx, router, value, start.Add(time.Millisecond))

	// The shadow raster needs 10px on each horizontal edge. The item's 12px
	// vertical outsets are applied only at the viewport's first and last edge.
	if probe.constraints.Min.X != 80 || probe.constraints.Max.X != 80 {
		t.Fatalf("shadow-safe child width = %v, want exact 80", probe.constraints)
	}
	if probe.constraints.Min.Y != 20 || probe.constraints.Max.Y != 20 {
		t.Fatalf("shadow-safe child height = %v, want exact 20", probe.constraints)
	}
}

func TestTransientShadowsDoNotReserveVisualOutset(t *testing.T) {
	gtx := layout.Context{Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	resting := &style.PaintStyle{Shadows: []style.Shadow{{
		Blur: 5, Color: style.RGBA(0x3e63dd),
	}}}
	if boxVisualOutset(gtx, resting).Empty() {
		t.Fatal("resting shadow did not reserve visual outset")
	}
	transient := &style.PaintStyle{Shadows: []style.Shadow{{
		Blur: 5, Color: style.RGBA(0x3e63dd), Transient: true,
	}}}
	if outset := boxVisualOutset(gtx, transient); !outset.Empty() {
		t.Fatalf("transient shadow reserved visual outset %#v", outset)
	}
}

func TestVisualOutsetPreservesBottomBaseline(t *testing.T) {
	dims := layoutVisualOutset(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	}, core.VisualOutset{Top: 3, Bottom: 5}, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(20, 10), Baseline: 2}
	})

	if dims.Size != image.Pt(20, 18) || dims.Baseline != 7 {
		t.Fatalf("visual outset dimensions = %v, want size (20,18) and baseline 7", dims)
	}
}

func TestVisualOutsetStateSurvivesHiddenLayout(t *testing.T) {
	ctx := newContext(nil)
	visual := &VisualOutsetState{outset: core.VisualOutset{Top: 4, Right: 5, Bottom: 6, Left: 7}}
	list := new(layout.List)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         new(op.Ops),
	}

	core.BeginFrameWithViewport(ctx, image.Pt(100, 100))
	core.LayoutHidden(ctx, gtx, "hidden", core.WidgetFunc(func(ctx *core.Context, hiddenGtx layout.Context) layout.Dimensions {
		return layoutVisualOutsetList(ctx, hiddenGtx, list, 1, visual, func(layout.Context, int) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(20, 20)}
		})
	}))
	core.EndFrame(ctx)

	want := core.VisualOutset{Top: 4, Right: 5, Bottom: 6, Left: 7}
	if visual.outset != want {
		t.Fatalf("hidden layout changed visual outset = %#v, want %#v", visual.outset, want)
	}
}

func TestOverflowHiddenStopsVisualOutsetPropagation(t *testing.T) {
	ctx := newContext(nil)
	collector := new(core.VisualOverflowCollector)
	restore := core.PushVisualOverflowCollector(ctx, collector)
	child := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return VisualOutset(ctx, gtx, Spacer(10, 10), 8, 8, 8, 8)
	})
	Box(child).
		Style(style.Style{}.Width(20).Height(20).Overflow(style.OverflowHidden)).
		Layout(ctx, testLayoutContext())
	restore()

	if got := collector.Outset(); !got.Empty() {
		t.Fatalf("clipped descendant reported visual outset %#v", got)
	}
}

func TestScrollTrackClickUpdatesExistingListState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := Scroll(Spacer(100, 300)).Key("body")
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

func layoutScrollFrame(ctx *core.Context, router *input.Router, value ScrollWidget, now time.Time) layout.Dimensions {
	viewport := image.Pt(100, 100)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	core.BeginFrameWithViewport(ctx, viewport)
	dims := value.Layout(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func TestScrollPropagatesContentOffset(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(100, 100)
	scrollSize := image.Pt(100, 40)
	core.BeginFrameWithViewport(ctx, viewport)
	state := ctx.ScrollState("body")
	core.EndFrame(ctx)
	state.Position = layout.Position{Offset: 30, BeforeEnd: true}
	var got image.Rectangle
	probe := &overlayProbeWidget{
		key:    "scroll",
		size:   image.Pt(20, 100),
		anchor: image.Rect(0, 35, 20, 45),
		got:    &got,
	}
	gtx := layout.Context{Constraints: layout.Constraints{Max: scrollSize}, Ops: new(op.Ops)}
	core.BeginFrameWithViewport(ctx, viewport)
	Scroll(probe).Key("body").Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	want := image.Rect(0, 5, 20, 15)
	if got != want {
		t.Fatalf("scroll anchor = %v, want %v", got, want)
	}

	core.BeginFrameWithViewport(ctx, viewport)
	state = ctx.ScrollState("hidden-body")
	core.EndFrame(ctx)
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
	core.BeginFrameWithViewport(ctx, viewport)
	Scroll(hidden).Key("hidden-body").Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
	if called {
		t.Fatal("content outside the local scroll viewport produced an overlay")
	}
}
