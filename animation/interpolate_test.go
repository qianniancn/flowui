package animation

import (
	"image"
	"image/color"
	"math"
	"testing"

	"gioui.org/f32"
)

func TestInterpolationHelpers(t *testing.T) {
	if got := LerpFloat(10, 20, 0.25); got != 12.5 {
		t.Fatalf("LerpFloat = %v, want 12.5", got)
	}
	if got := LerpFloat(-math.MaxFloat32, math.MaxFloat32, 0.5); got != 0 {
		t.Fatalf("LerpFloat extreme midpoint = %v, want 0", got)
	}
	if got := LerpFloat64(-1e308, 1e308, 0.5); got != 0 {
		t.Fatalf("LerpFloat64 extreme midpoint = %v, want 0", got)
	}
	if got := LerpPoint(f32.Pt(0, 10), f32.Pt(20, 30), 0.25); got != (f32.Pt(5, 15)) {
		t.Fatalf("LerpPoint = %v, want (5,15)", got)
	}
	if got := LerpRect(image.Rect(0, 10, 20, 30), image.Rect(10, 30, 50, 70), 0.5); got != image.Rect(5, 20, 35, 50) {
		t.Fatalf("LerpRect = %v", got)
	}
}

func TestLerpColorPreservesHueThroughTransparency(t *testing.T) {
	red := color.NRGBA{R: 0xff, A: 0xff}
	if got := LerpColor(color.NRGBA{}, red, 0.5); got != (color.NRGBA{R: 0xff, A: 0x80}) {
		t.Fatalf("transparent to red = %#v", got)
	}
	if got := LerpColor(red, color.NRGBA{}, 0.5); got != (color.NRGBA{R: 0xff, A: 0x80}) {
		t.Fatalf("red to transparent = %#v", got)
	}
	if got := LerpColor(color.NRGBA{}, red, 2); got != red {
		t.Fatalf("overshooting color = %#v, want clamped red", got)
	}
}
