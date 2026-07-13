package linechart

import (
	"math"

	"gioui.org/f32"
	"gioui.org/op/clip"
)

type cubicLineSegment struct {
	control0 f32.Point
	control1 f32.Point
	to       f32.Point
}

func drawSmoothChartPath(path *clip.Path, points []resolvedPoint, connectNulls bool, smooth float32, geometry chartGeometry) {
	segments := splitSmoothLine(points, connectNulls, func(point resolvedPoint) f32.Point {
		return f32.Pt(geometry.mapX(point.X), geometry.mapY(point.Y))
	})
	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		path.MoveTo(segment[0])
		for _, cubic := range smoothCubics(segment, smooth) {
			path.CubeTo(cubic.control0, cubic.control1, cubic.to)
		}
	}
}

func splitSmoothLine(points []resolvedPoint, connectNulls bool, transform func(resolvedPoint) f32.Point) [][]f32.Point {
	segments := make([][]f32.Point, 0, 1)
	current := make([]f32.Point, 0, len(points))
	flush := func() {
		if len(current) > 0 {
			segments = append(segments, current)
			current = nil
		}
	}
	for _, point := range points {
		if !point.valid {
			if !connectNulls {
				flush()
			}
			continue
		}
		current = append(current, transform(point))
	}
	flush()
	return segments
}

// smoothCubics follows the bounded non-monotone smoothing strategy used by
// ECharts 6.1.0. Each control point stays inside its adjacent point bounds.
func smoothCubics(points []f32.Point, smooth float32) []cubicLineSegment {
	points = compactSmoothPoints(points)
	if len(points) < 2 || smooth <= 0 {
		return nil
	}
	smooth = min(max(smooth, 0), 1)
	cubics := make([]cubicLineSegment, 0, len(points)-1)
	previous := points[0]
	control0 := previous
	for index := 1; index < len(points); index++ {
		current := points[index]
		control1 := current
		nextControl0 := current
		if index+1 < len(points) {
			next := points[index+1]
			previousLength := pointDistance(previous, current)
			nextLength := pointDistance(current, next)
			if previousLength > 0 && nextLength > 0 {
				ratioNext := nextLength / (nextLength + previousLength)
				vector := subtractPoint(next, previous)
				control1 = subtractPoint(current, scalePoint(vector, smooth*(1-ratioNext)))
				nextControl0 = addPoint(current, scalePoint(vector, smooth*ratioNext))

				nextControl0 = constrainPoint(nextControl0, current, next)
				adjusted := subtractPoint(nextControl0, current)
				control1 = subtractPoint(current, scalePoint(adjusted, previousLength/nextLength))
				control1 = constrainPoint(control1, previous, current)

				adjusted = subtractPoint(current, control1)
				nextControl0 = addPoint(current, scalePoint(adjusted, nextLength/previousLength))
			}
		}
		cubics = append(cubics, cubicLineSegment{control0: control0, control1: control1, to: current})
		control0 = nextControl0
		previous = current
	}
	return cubics
}

func compactSmoothPoints(points []f32.Point) []f32.Point {
	if len(points) < 2 {
		return points
	}
	compacted := make([]f32.Point, 0, len(points))
	for _, point := range points {
		if len(compacted) > 0 {
			dx := point.X - compacted[len(compacted)-1].X
			dy := point.Y - compacted[len(compacted)-1].Y
			if dx*dx+dy*dy < 0.5 {
				continue
			}
		}
		compacted = append(compacted, point)
	}
	return compacted
}

func pointDistance(first, second f32.Point) float32 {
	dx := second.X - first.X
	dy := second.Y - first.Y
	return float32(math.Hypot(float64(dx), float64(dy)))
}

func constrainPoint(value, first, second f32.Point) f32.Point {
	value.X = min(max(value.X, min(first.X, second.X)), max(first.X, second.X))
	value.Y = min(max(value.Y, min(first.Y, second.Y)), max(first.Y, second.Y))
	return value
}

func addPoint(first, second f32.Point) f32.Point {
	return f32.Pt(first.X+second.X, first.Y+second.Y)
}

func subtractPoint(first, second f32.Point) f32.Point {
	return f32.Pt(first.X-second.X, first.Y-second.Y)
}

func scalePoint(point f32.Point, scale float32) f32.Point {
	return f32.Pt(point.X*scale, point.Y*scale)
}
