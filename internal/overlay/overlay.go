package overlay

import (
	"image"
)

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
