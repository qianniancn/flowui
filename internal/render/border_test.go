package render

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestDrawRoundedBorderHandlesNormalAndCollapsedInnerContours(t *testing.T) {
	gtx := layout.Context{Ops: new(op.Ops)}
	DrawRoundedBorder(gtx, image.Rect(0, 0, 100, 60), 24, 2, color.NRGBA{A: 255})
	DrawRoundedBorder(gtx, image.Rect(0, 0, 4, 4), 2, 4, color.NRGBA{A: 255})
}

func TestDrawRoundedBorderLeavesCenterOpen(t *testing.T) {
	window, err := headless.NewWindow(40, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	var ops op.Ops
	DrawRoundedBorder(layout.Context{Ops: &ops}, image.Rect(0, 0, 40, 40), 12, 3, color.NRGBA{R: 147, B: 234, A: 255})
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 40, 40))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := pixels.RGBAAt(1, 20).A; got == 0 {
		t.Fatal("border pixel was not drawn")
	}
	if got := pixels.RGBAAt(20, 20).A; got != 0 {
		t.Fatalf("border center alpha = %d, want 0", got)
	}
}

func TestDrawBottomBorderPaintsBottomEdge(t *testing.T) {
	window, err := headless.NewWindow(20, 20)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	var ops op.Ops
	DrawBottomBorder(layout.Context{Ops: &ops}, image.Rect(0, 0, 20, 20), 2, color.NRGBA{R: 147, B: 234, A: 255})
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 20, 20))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := pixels.RGBAAt(10, 19).A; got == 0 {
		t.Fatal("bottom border pixel was not drawn")
	}
	if got := pixels.RGBAAt(10, 0).A; got != 0 {
		t.Fatalf("top pixel alpha = %d, want 0", got)
	}
}
