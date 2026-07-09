package flowui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestAspectRatioFitsWithinConstraints(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := AspectRatio(16.0/9.0, child).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(320, 180) {
		t.Fatalf("aspect size = %v, want (320,180)", dims.Size)
	}
	if child.constraints.Min != image.Pt(320, 180) || child.constraints.Max != image.Pt(320, 180) {
		t.Fatalf("child constraints = %v, want exact (320,180)", child.constraints)
	}
}

func TestAspectRatioUsesHeightWhenWidthIsTooTall(t *testing.T) {
	child := &constraintWidget{}
	var ops op.Ops

	dims := AspectRatio(16.0/9.0, child).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 100)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(178, 100) {
		t.Fatalf("aspect size = %v, want (178,100)", dims.Size)
	}
}
