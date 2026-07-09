package flowui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
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
