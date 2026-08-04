package ui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestVisualOverflowKeepsChildDimensions(t *testing.T) {
	want := layout.Dimensions{Size: image.Pt(80, 32), Baseline: 7}
	childCalls := 0
	drawCalls := 0
	child := WidgetFunc(func(*Context, layout.Context) layout.Dimensions {
		childCalls++
		return want
	})
	draw := func(_ *Context, gtx layout.Context, bounds image.Rectangle) {
		drawCalls++
		if bounds != (image.Rectangle{Max: want.Size}) {
			t.Fatalf("overflow bounds = %v, want %v", bounds, image.Rectangle{Max: want.Size})
		}
		if gtx.Constraints != layout.Exact(want.Size) {
			t.Fatalf("overflow constraints = %v, want exact child size", gtx.Constraints)
		}
	}

	dims := LayoutVisualOverflow(nil, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         new(op.Ops),
	}, child, draw)
	if dims != want {
		t.Fatalf("dimensions = %v, want %v", dims, want)
	}
	if childCalls != 1 || drawCalls != 1 {
		t.Fatalf("calls = child %d, draw %d; want 1 each", childCalls, drawCalls)
	}
}

func TestVisualOverflowHandlesNilValues(t *testing.T) {
	gtx := layout.Context{Ops: new(op.Ops)}
	if dims := LayoutVisualOverflow(nil, gtx, nil, nil); dims != (layout.Dimensions{}) {
		t.Fatalf("nil child dimensions = %v", dims)
	}
	want := layout.Dimensions{Size: image.Pt(12, 8)}
	dims := LayoutVisualOverflow(nil, gtx, WidgetFunc(func(*Context, layout.Context) layout.Dimensions {
		return want
	}), nil)
	if dims != want {
		t.Fatalf("nil draw dimensions = %v, want %v", dims, want)
	}
}
