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
		Overlay(probe).Align(AlignBottomEnd),
	).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)

	want := image.Rect(60, 30, 80, 40)
	if got != want {
		t.Fatalf("stack anchor = %v, want %v", got, want)
	}
}
