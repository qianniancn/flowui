package render

import (
	"image"
	"image/color"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestLinearGradientNormalizesStopsWithoutMutatingInput(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	input := []GradientStop{{Offset: .8, Color: blue}, {Offset: .2, Color: red}}
	brush := LinearGradient(input...)
	if input[0].Offset != .8 || input[1].Offset != .2 {
		t.Fatal("linear gradient mutated its input stops")
	}
	if len(brush.stops) != 4 {
		t.Fatalf("normalized stops = %d, want 4", len(brush.stops))
	}
	if brush.stops[0].Offset != 0 || brush.stops[1].Offset != .2 || brush.stops[2].Offset != .8 || brush.stops[3].Offset != 1 {
		t.Fatalf("normalized offsets = %#v", brush.stops)
	}
}

func TestLinearGradientClampsInvalidStops(t *testing.T) {
	brush := LinearGradient(
		GradientStop{Offset: float32(math.NaN()), Color: color.NRGBA{R: 1}},
		GradientStop{Offset: float32(math.Inf(1)), Color: color.NRGBA{R: 2}},
	)
	if brush.stops[0].Offset != 0 || brush.stops[len(brush.stops)-1].Offset != 1 {
		t.Fatalf("invalid stop offsets = %#v", brush.stops)
	}
}

func TestBrushValueSemanticsAndSampling(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 127}
	base := LinearGradient(
		GradientStop{Offset: 0, Color: red},
		GradientStop{Offset: 1, Color: blue},
	)
	angled := base.Angle(90)
	if base.angle != 180 || angled.angle != 90 {
		t.Fatalf("brush angles = %v and %v, want 180 and 90", base.angle, angled.angle)
	}
	if got := base.ColorAt(.5); got != (color.NRGBA{R: 128, B: 128, A: 191}) {
		t.Fatalf("midpoint color = %#v", got)
	}
	if got := SolidBrush(red).ColorAt(.5); got != red {
		t.Fatalf("solid sample = %#v, want %#v", got, red)
	}
}

func TestBrushAngleNormalizesAndRejectsInvalidValues(t *testing.T) {
	brush := LinearGradient(
		GradientStop{Color: color.NRGBA{R: 255}},
		GradientStop{Offset: 1, Color: color.NRGBA{B: 255}},
	)
	if got := brush.Angle(-90).angle; got != 270 {
		t.Fatalf("negative angle = %v, want 270", got)
	}
	if got := brush.Angle(450).angle; got != 90 {
		t.Fatalf("wrapped angle = %v, want 90", got)
	}
	if got := brush.Angle(float32(math.NaN())).angle; got != 180 {
		t.Fatalf("NaN angle = %v, want default 180", got)
	}
}

func TestSingleGradientStopBecomesSolidBrush(t *testing.T) {
	col := color.NRGBA{R: 10, G: 20, B: 30, A: 200}
	brush := LinearGradient(GradientStop{Offset: .5, Color: col})
	if brush.kind != brushSolid || brush.ColorAt(0) != col || brush.ColorAt(1) != col {
		t.Fatalf("single-stop brush = %#v", brush)
	}
}

func TestLinearGradientPreservesHardStops(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	brush := LinearGradient(
		GradientStop{Offset: 0, Color: red},
		GradientStop{Offset: .5, Color: red},
		GradientStop{Offset: .5, Color: blue},
		GradientStop{Offset: 1, Color: blue},
	)
	if got := brush.ColorAt(.5); got != blue {
		t.Fatalf("hard stop sample = %#v, want %#v", got, blue)
	}
}

func TestLinearGradientLineUsesCSSAngles(t *testing.T) {
	rect := image.Rect(0, 0, 100, 50)
	tests := []struct {
		angle     float32
		wantStart f32.Point
		wantEnd   f32.Point
	}{
		{angle: 0, wantStart: f32.Pt(50, 50), wantEnd: f32.Pt(50, 0)},
		{angle: 90, wantStart: f32.Pt(0, 25), wantEnd: f32.Pt(100, 25)},
		{angle: 180, wantStart: f32.Pt(50, 0), wantEnd: f32.Pt(50, 50)},
		{angle: 270, wantStart: f32.Pt(100, 25), wantEnd: f32.Pt(0, 25)},
	}
	for _, test := range tests {
		start, end := linearGradientLine(rect, test.angle)
		if !nearPoint(start, test.wantStart) || !nearPoint(end, test.wantEnd) {
			t.Fatalf("angle %v line = %v to %v, want %v to %v", test.angle, start, end, test.wantStart, test.wantEnd)
		}
	}
}

func TestDrawBrushHandlesSolidGradientAndEmptyBounds(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	DrawBrush(gtx, image.Rect(0, 0, 100, 40), 8, SolidBrush(red))
	DrawBrush(gtx, image.Rect(0, 0, 100, 40), 8, LinearGradient(
		GradientStop{Color: red},
		GradientStop{Offset: .5, Color: blue},
		GradientStop{Offset: 1, Color: red},
	))
	DrawBrush(gtx, image.Rectangle{}, 0, Brush{})
}

func nearPoint(got, want f32.Point) bool {
	return math.Abs(float64(got.X-want.X)) < .001 && math.Abs(float64(got.Y-want.Y)) < .001
}
