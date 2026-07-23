package overlay

import (
	"image"
	"math"
	"slices"
	"testing"

	"gioui.org/f32"
)

func TestAffineRectBounds(t *testing.T) {
	tests := []struct {
		name      string
		rect      image.Rectangle
		transform f32.Affine2D
		want      image.Rectangle
	}{
		{
			name:      "identity",
			rect:      image.Rect(10, 20, 40, 60),
			transform: f32.AffineId(),
			want:      image.Rect(10, 20, 40, 60),
		},
		{
			name:      "fractional translation rounds outwards",
			rect:      image.Rect(10, 20, 40, 60),
			transform: f32.AffineId().Offset(f32.Pt(0.25, -0.75)),
			want:      image.Rect(10, 19, 41, 60),
		},
		{
			name:      "scale around origin",
			rect:      image.Rect(10, 20, 30, 40),
			transform: f32.AffineId().Scale(f32.Pt(20, 30), f32.Pt(0.5, 0.5)),
			want:      image.Rect(15, 25, 25, 35),
		},
		{
			name:      "negative scale",
			rect:      image.Rect(1, 2, 5, 8),
			transform: f32.NewAffine2D(-2, 0, 7, 0, -0.5, 6),
			want:      image.Rect(-3, 2, 5, 5),
		},
		{
			name:      "quarter turn",
			rect:      image.Rect(0, 0, 10, 20),
			transform: f32.AffineId().Rotate(f32.Point{}, float32(math.Pi/2)),
			want:      image.Rect(-20, 0, 0, 10),
		},
		{
			name:      "empty rectangle",
			rect:      image.Rect(4, 5, 4, 9),
			transform: f32.AffineId().Offset(f32.Pt(10, 20)),
			want:      image.Rectangle{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := AffineRectBounds(test.rect, test.transform); got != test.want {
				t.Fatalf("AffineRectBounds() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestDismissRectsExcluding(t *testing.T) {
	tests := []struct {
		name       string
		bounds     image.Rectangle
		exclusions []image.Rectangle
	}{
		{
			name:   "no exclusions",
			bounds: image.Rect(3, 4, 13, 14),
		},
		{
			name:       "single interior exclusion",
			bounds:     image.Rect(3, 4, 13, 14),
			exclusions: []image.Rectangle{image.Rect(6, 7, 10, 12)},
		},
		{
			name:   "overlapping exclusions",
			bounds: image.Rect(0, 0, 12, 10),
			exclusions: []image.Rectangle{
				image.Rect(2, 2, 8, 7),
				image.Rect(6, 4, 11, 9),
			},
		},
		{
			name:   "partially outside exclusions",
			bounds: image.Rect(5, 5, 15, 15),
			exclusions: []image.Rectangle{
				image.Rect(0, 0, 8, 9),
				image.Rect(12, 13, 20, 20),
			},
		},
		{
			name:       "empty exclusion",
			bounds:     image.Rect(1, 2, 6, 8),
			exclusions: []image.Rectangle{image.Rect(3, 3, 3, 7)},
		},
		{
			name:       "exclusion covers bounds",
			bounds:     image.Rect(5, 6, 9, 10),
			exclusions: []image.Rectangle{image.Rect(0, 0, 20, 20)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := DismissRectsExcluding(test.bounds, test.exclusions...)
			assertDismissPartition(t, test.bounds, test.exclusions, got)
		})
	}
}

func TestDismissRectsExcludingEmptyBounds(t *testing.T) {
	if got := DismissRectsExcluding(image.Rectangle{}); len(got) != 0 {
		t.Fatalf("DismissRectsExcluding(empty) = %v, want no rectangles", got)
	}
}

func assertDismissPartition(t *testing.T, bounds image.Rectangle, exclusions, areas []image.Rectangle) {
	t.Helper()

	covered := make(map[image.Point]int, bounds.Dx()*bounds.Dy())
	for _, area := range areas {
		if area.Empty() {
			t.Errorf("partition contains empty rectangle %v", area)
			continue
		}
		if area.Intersect(bounds) != area {
			t.Errorf("partition rectangle %v lies outside bounds %v", area, bounds)
		}
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				point := image.Pt(x, y)
				covered[point]++
			}
		}
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			point := image.Pt(x, y)
			want := 1
			if slices.ContainsFunc(exclusions, point.In) {
				want = 0
			}
			if covered[point] != want {
				t.Errorf("coverage at %v = %d, want %d", point, covered[point], want)
			}
		}
	}
}
