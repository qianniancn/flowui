package render

import (
	"math"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// CheckPath returns a progressively drawn three-point checkmark path.
func CheckPath(ops *op.Ops, points [3]f32.Point, progress float32) clip.PathSpec {
	progress = min(max(progress, 0), 1)
	first, second, third := points[0], points[1], points[2]
	firstLen := pointDistance(first, second)
	secondLen := pointDistance(second, third)
	drawLen := (firstLen + secondLen) * progress

	var path clip.Path
	path.Begin(ops)
	path.MoveTo(first)

	// Guard against degenerate cases where points coincide.
	if firstLen == 0 {
		// First and second are the same; skip directly to second segment.
		if secondLen == 0 {
			// All three points coincide; draw nothing beyond the start point.
			return path.End()
		}
		path.LineTo(pointOnLine(second, third, progress))
		return path.End()
	}

	if drawLen <= firstLen {
		path.LineTo(pointOnLine(first, second, drawLen/firstLen))
		return path.End()
	}

	path.LineTo(second)

	if secondLen == 0 {
		// Second and third are the same; path ends at second.
		return path.End()
	}

	path.LineTo(pointOnLine(second, third, (drawLen-firstLen)/secondLen))
	return path.End()
}

func pointOnLine(from, to f32.Point, progress float32) f32.Point {
	return f32.Pt(Lerp(from.X, to.X, progress), Lerp(from.Y, to.Y, progress))
}

func pointDistance(from, to f32.Point) float32 {
	dx, dy := to.X-from.X, to.Y-from.Y
	return float32(math.Hypot(float64(dx), float64(dy)))
}
