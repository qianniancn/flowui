package layoutui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func TestStackOverlayDoesNotChangeSize(t *testing.T) {
	var ops op.Ops

	dims := Stack(
		Stacked(Spacer(80, 40)),
		Overlay(Spacer(120, 90)).Align(AlignTopEnd),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(80, 40) {
		t.Fatalf("stack size = %v, want (80,40)", dims.Size)
	}
}

func TestStackOffsetDoesNotChangeSize(t *testing.T) {
	var ops op.Ops

	dims := Stack(
		Stacked(Spacer(80, 40)).Offset(100, -50),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(80, 40) {
		t.Fatalf("stack size = %v, want (80,40)", dims.Size)
	}
}

func TestStackOverlayDoesNotSetBaseline(t *testing.T) {
	var ops op.Ops

	dims := Stack(
		Overlay(baselineWidget{size: image.Pt(10, 10), baseline: 2}),
		Stacked(baselineWidget{size: image.Pt(20, 20), baseline: 5}),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 200)},
		Ops:         &ops,
	})

	if dims.Baseline != 5 {
		t.Fatalf("stack baseline = %d, want 5", dims.Baseline)
	}
}

func TestStackExpandedLayerUsesLargestStackedLayer(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	Stack(
		Stacked(Spacer(80, 40)),
		Overlay(child).Expanded(),
	).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 200)},
		Ops:         &ops,
	})

	if child.constraints.Min != image.Pt(80, 40) {
		t.Fatalf("expanded constraints = %v, want min (80,40)", child.constraints)
	}
}

func TestStackPropagatesAlignedLayerPosition(t *testing.T) {
	probe := &overlayProbeWidget{
		key:    "stack",
		size:   image.Pt(20, 10),
		anchor: image.Rect(0, 0, 20, 10),
	}
	var got image.Rectangle
	probe.got = &got
	ctx := newContext(nil)
	frame.BeginFrameWithViewport(ctx, image.Pt(100, 100))
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 100)}, Ops: new(op.Ops)}
	Stack(
		Stacked(Spacer(80, 40)),
		Overlay(probe).Align(AlignBottomEnd).Offset(-4, 5),
	).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)

	want := image.Rect(56, 35, 76, 45)
	if got != want {
		t.Fatalf("stack anchor = %v, want %v", got, want)
	}
}

type baselineWidget struct {
	size     image.Point
	baseline int
}

func (w baselineWidget) Layout(_ *frame.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: w.size, Baseline: w.baseline}
}
