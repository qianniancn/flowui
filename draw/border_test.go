package draw

import (
	"gioui.org/op/paint"
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

func TestDrawRoundedBorderHandlesNormalAndCollapsedInnerContours(t *testing.T) {
	gtx := layout.Context{Ops: new(op.Ops)}
	RoundedBorder(gtx, image.Rect(0, 0, 100, 60), 24, 2, color.NRGBA{A: 255})
	RoundedBorder(gtx, image.Rect(0, 0, 4, 4), 2, 4, color.NRGBA{A: 255})
}

func TestDrawRoundedBorderLeavesCenterOpen(t *testing.T) {
	window, err := headless.NewWindow(40, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	var ops op.Ops
	RoundedBorder(layout.Context{Ops: &ops}, image.Rect(0, 0, 40, 40), 12, 3, color.NRGBA{R: 147, B: 234, A: 255})
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
	BottomBorder(layout.Context{Ops: &ops}, image.Rect(0, 0, 20, 20), 2, color.NRGBA{R: 147, B: 234, A: 255})
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

// A dashed border must paint some columns of the top edge and leave others
// blank. Counting both proves the pattern ran rather than degrading to solid.
func TestDrawDashedRRectBorderAlternatesAlongTheEdge(t *testing.T) {
	window, err := headless.NewWindow(80, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	const (
		width  = 80
		height = 40
		radius = 6
		on     = 4
		off    = 3
	)
	var ops op.Ops
	rect := image.Rect(0, 0, width, height)
	DashedRRectBorder(
		layout.Context{Ops: &ops},
		clip.UniformRRect(rect, radius),
		1, on, off,
		color.NRGBA{R: 0xff, A: 255},
	)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(rect)
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}

	// Walk the straight part of the top edge, away from the corner arcs.
	painted, blank := 0, 0
	for x := radius + 2; x < width-radius-2; x++ {
		found := false
		for y := range radius {
			if pixels.RGBAAt(x, y).A != 0 {
				found = true
				break
			}
		}
		if found {
			painted++
		} else {
			blank++
		}
	}
	if painted == 0 {
		t.Fatal("dashed border painted nothing along the top edge")
	}
	if blank == 0 {
		t.Fatal("dashed border left no gaps; it degraded to a solid stroke")
	}
}

// A zero gap has no meaningful pattern, so it must fall back to a solid line
// instead of dividing by zero.
func TestDrawDashedRRectBorderFallsBackToSolidWithoutAGap(t *testing.T) {
	window, err := headless.NewWindow(60, 30)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	var ops op.Ops
	rect := image.Rect(0, 0, 60, 30)
	DashedRRectBorder(
		layout.Context{Ops: &ops},
		clip.UniformRRect(rect, 4),
		1, 4, 0,
		color.NRGBA{B: 0xff, A: 255},
	)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(rect)
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}

	blank := 0
	for x := 6; x < 54; x++ {
		found := false
		for y := range 4 {
			if pixels.RGBAAt(x, y).A != 0 {
				found = true
				break
			}
		}
		if !found {
			blank++
		}
	}
	if blank != 0 {
		t.Fatalf("solid fallback left %d blank columns, want 0", blank)
	}
}

// A one-pixel border must cover whole device pixels. A stroke centered on the
// pixel grid instead paints two half-covered pixels per edge, which reads as a
// soft two-pixel line; this is the difference the pixel alignment exists for.
func TestDrawOnePixelBorderCoversWholePixels(t *testing.T) {
	window, err := headless.NewWindow(40, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	const (
		width  = 40
		height = 40
	)
	border := color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 255}
	background := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	var ops op.Ops
	paint.Fill(&ops, background)
	RRectBorder(
		layout.Context{Ops: &ops},
		clip.UniformRRect(image.Rect(0, 0, width, height), 6),
		1,
		border,
	)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, width, height))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}

	// Sample straight edges away from the corner arcs: the outer pixel carries the
	// border, the next one is clean background. A straddling stroke would leave
	// both outer pixels at roughly half coverage.
	near := func(got color.RGBA, want color.NRGBA) bool {
		const tolerance = 8
		return abs(int(got.R)-int(want.R)) <= tolerance &&
			abs(int(got.G)-int(want.G)) <= tolerance &&
			abs(int(got.B)-int(want.B)) <= tolerance
	}
	edges := []image.Point{
		{width / 2, 0},          // top
		{width / 2, height - 1}, // bottom
		{0, height / 2},         // left
		{width - 1, height / 2}, // right
	}
	inside := []image.Point{
		{width / 2, 1},
		{width / 2, height - 2},
		{1, height / 2},
		{width - 2, height / 2},
	}
	for index, point := range edges {
		if got := pixels.RGBAAt(point.X, point.Y); !near(got, border) {
			t.Fatalf("border pixel at %v = %v, want a whole pixel of %v (inside pixel %v = %v)",
				point, got, border, inside[index], pixels.RGBAAt(inside[index].X, inside[index].Y))
		}
		if got := pixels.RGBAAt(inside[index].X, inside[index].Y); !near(got, background) {
			t.Fatalf("pixel inside the border at %v = %v, want background %v", inside[index], got, background)
		}
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// A collapsed rectangle must not panic while the outline is generated.
func TestDrawDashedRRectBorderHandlesDegenerateRect(t *testing.T) {
	gtx := layout.Context{Ops: new(op.Ops)}
	DashedRRectBorder(gtx, clip.UniformRRect(image.Rect(0, 0, 4, 4), 2), 4, 4, 3, color.NRGBA{A: 255})
	DashedRRectBorder(gtx, clip.UniformRRect(image.Rect(0, 0, 0, 0), 0), 1, 4, 3, color.NRGBA{A: 255})
	DashedRRectBorder(gtx, clip.UniformRRect(image.Rect(0, 0, 40, 20), 0), 0, 4, 3, color.NRGBA{A: 255})
}
