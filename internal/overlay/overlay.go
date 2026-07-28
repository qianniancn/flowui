package overlay

import (
	"image"
	"math"

	"gioui.org/f32"
)

// AffineRectBounds returns the smallest integer rectangle that contains rect
// after transform is applied. Fractional edges are rounded outwards so the
// result is safe to use as a pointer hit region.
func AffineRectBounds(rect image.Rectangle, transform f32.Affine2D) image.Rectangle {
	if rect.Empty() {
		return image.Rectangle{}
	}

	points := [4]f32.Point{
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Max.Y))),
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Max.Y))),
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, point := range points[1:] {
		minX = min(minX, point.X)
		maxX = max(maxX, point.X)
		minY = min(minY, point.Y)
		maxY = max(maxY, point.Y)
	}
	return image.Rect(
		int(math.Floor(float64(minX))),
		int(math.Floor(float64(minY))),
		int(math.Ceil(float64(maxX))),
		int(math.Ceil(float64(maxY))),
	)
}

func DismissRects(bounds, excluded image.Rectangle) [4]image.Rectangle {
	excluded = excluded.Intersect(bounds)
	if excluded.Empty() {
		return [4]image.Rectangle{bounds}
	}
	return [4]image.Rectangle{
		image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, excluded.Min.Y),
		image.Rect(bounds.Min.X, excluded.Max.Y, bounds.Max.X, bounds.Max.Y),
		image.Rect(bounds.Min.X, excluded.Min.Y, excluded.Min.X, excluded.Max.Y),
		image.Rect(excluded.Max.X, excluded.Min.Y, bounds.Max.X, excluded.Max.Y),
	}
}

// DismissRectsExcluding partitions bounds into non-overlapping rectangles
// outside every excluded rectangle.
func DismissRectsExcluding(bounds image.Rectangle, excluded ...image.Rectangle) []image.Rectangle {
	if bounds.Empty() {
		return nil
	}
	// Use double buffering to avoid allocating a new slice on each exclusion.
	current := []image.Rectangle{bounds}
	next := make([]image.Rectangle, 0, 8) // Pre-allocate for common cases

	for _, exclusion := range excluded {
		next = next[:0] // Reuse backing array
		for _, area := range current {
			intersection := exclusion.Intersect(area)
			if intersection.Empty() {
				next = append(next, area)
				continue
			}
			for _, part := range DismissRects(area, intersection) {
				if !part.Empty() {
					next = append(next, part)
				}
			}
		}
		// Swap buffers
		current, next = next, current
	}
	return current
}
