package linechart

import (
	"image"
	"image/color"
	"math"
	"sort"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

const smoothSamplesPerSegment = 12

func seriesPixelSegments(series resolvedSeries, geometry chartGeometry) [][]f32.Point {
	segments := splitSmoothLine(series.points, series.connectNulls, func(point resolvedPoint) f32.Point {
		return f32.Pt(geometry.mapX(point.X), geometry.mapY(point.Y))
	})
	for index, segment := range segments {
		segment = visiblePixelSegment(segment, float32(geometry.plot.Min.X), float32(geometry.plot.Max.X))
		if series.sampling == SamplingMinMax {
			segment = minMaxPixelSample(segment, max(geometry.plot.Dx(), 1))
		}
		switch {
		case series.step != StepNone:
			segments[index] = steppedPoints(segment, series.step)
		case series.smooth > 0:
			segments[index] = sampledSmoothPoints(segment, series.smooth)
		}
	}
	return segments
}

func visiblePixelSegment(points []f32.Point, minimumX, maximumX float32) []f32.Point {
	if len(points) < 2 || !monotonicPointX(points) {
		return points
	}
	start := sort.Search(len(points), func(index int) bool { return points[index].X >= minimumX })
	if start > 0 {
		start--
	}
	end := sort.Search(len(points), func(index int) bool { return points[index].X > maximumX })
	if end < len(points) {
		end++
	}
	if start >= end {
		return nil
	}
	return points[start:end]
}

func minMaxPixelSample(points []f32.Point, pixelWidth int) []f32.Point {
	if len(points) <= pixelWidth*2 || pixelWidth <= 0 || !monotonicPointX(points) {
		return points
	}
	minimumX := points[0].X
	maximumX := points[len(points)-1].X
	if maximumX <= minimumX {
		return points
	}
	result := make([]f32.Point, 0, pixelWidth*2+2)
	result = append(result, points[0])
	for start := 1; start < len(points)-1; {
		bucket := int((points[start].X - minimumX) / (maximumX - minimumX) * float32(pixelWidth))
		end := start + 1
		for end < len(points)-1 {
			nextBucket := int((points[end].X - minimumX) / (maximumX - minimumX) * float32(pixelWidth))
			if nextBucket != bucket {
				break
			}
			end++
		}
		minimumIndex, maximumIndex := start, start
		for index := start + 1; index < end; index++ {
			if points[index].Y < points[minimumIndex].Y {
				minimumIndex = index
			}
			if points[index].Y > points[maximumIndex].Y {
				maximumIndex = index
			}
		}
		if minimumIndex < maximumIndex {
			result = append(result, points[minimumIndex], points[maximumIndex])
		} else if maximumIndex < minimumIndex {
			result = append(result, points[maximumIndex], points[minimumIndex])
		} else {
			result = append(result, points[minimumIndex])
		}
		start = end
	}
	result = append(result, points[len(points)-1])
	return result
}

func monotonicPointX(points []f32.Point) bool {
	for index := 1; index < len(points); index++ {
		if points[index].X < points[index-1].X {
			return false
		}
	}
	return true
}

func steppedPoints(points []f32.Point, mode StepMode) []f32.Point {
	if len(points) < 2 || mode == StepNone {
		return points
	}
	capacity := len(points)*2 - 1
	if mode == StepMiddle {
		capacity = len(points)*3 - 2
	}
	result := make([]f32.Point, 0, capacity)
	result = append(result, points[0])
	for index := 1; index < len(points); index++ {
		previous, current := points[index-1], points[index]
		switch mode {
		case StepStart:
			result = append(result, f32.Pt(previous.X, current.Y))
		case StepMiddle:
			middle := (previous.X + current.X) / 2
			result = append(result, f32.Pt(middle, previous.Y), f32.Pt(middle, current.Y))
		case StepEnd:
			result = append(result, f32.Pt(current.X, previous.Y))
		}
		result = append(result, current)
	}
	return result
}

func sampledSmoothPoints(points []f32.Point, smooth float32) []f32.Point {
	points = compactSmoothPoints(points)
	if len(points) < 2 || smooth <= 0 {
		return points
	}
	cubics := smoothCubics(points, smooth)
	result := make([]f32.Point, 0, len(cubics)*smoothSamplesPerSegment+1)
	result = append(result, points[0])
	from := points[0]
	for _, cubic := range cubics {
		for sample := 1; sample <= smoothSamplesPerSegment; sample++ {
			t := float32(sample) / smoothSamplesPerSegment
			result = append(result, cubicPoint(from, cubic.control0, cubic.control1, cubic.to, t))
		}
		from = cubic.to
	}
	return result
}

func cubicPoint(from, control0, control1, to f32.Point, t float32) f32.Point {
	inverse := 1 - t
	return f32.Pt(
		inverse*inverse*inverse*from.X+3*inverse*inverse*t*control0.X+3*inverse*t*t*control1.X+t*t*t*to.X,
		inverse*inverse*inverse*from.Y+3*inverse*inverse*t*control0.Y+3*inverse*t*t*control1.Y+t*t*t*to.Y,
	)
}

func drawLineSeriesSegment(gtx layout.Context, points []f32.Point, width float32, color color.NRGBA, style LineStyle) {
	if len(points) < 2 {
		return
	}
	switch style {
	case LineDashed:
		drawDashedPolyline(gtx, points, width, color, max(width*4, 6), max(width*2.5, 4))
	case LineDotted:
		drawDottedPolyline(gtx, points, width, color)
	default:
		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(points[0])
		for _, point := range points[1:] {
			path.LineTo(point)
		}
		stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, color)
		stroke.Pop()
	}
}

func drawDashedPolyline(gtx layout.Context, points []f32.Point, width float32, color color.NRGBA, dash, gap float32) {
	drawing := true
	remaining := dash
	for index := 1; index < len(points); index++ {
		from, to := points[index-1], points[index]
		dx, dy := to.X-from.X, to.Y-from.Y
		length := float32(math.Hypot(float64(dx), float64(dy)))
		if length <= 0 {
			continue
		}
		position := float32(0)
		for position < length {
			step := min(remaining, length-position)
			if drawing {
				startRatio := position / length
				endRatio := (position + step) / length
				drawChartLine(gtx,
					f32.Pt(from.X+dx*startRatio, from.Y+dy*startRatio),
					f32.Pt(from.X+dx*endRatio, from.Y+dy*endRatio),
					width, color,
				)
			}
			position += step
			remaining -= step
			if remaining <= 1e-4 {
				drawing = !drawing
				if drawing {
					remaining = dash
				} else {
					remaining = gap
				}
			}
		}
	}
}

func drawDottedPolyline(gtx layout.Context, points []f32.Point, width float32, color color.NRGBA) {
	spacing := max(width*2.75, 5)
	remaining := float32(0)
	for index := 1; index < len(points); index++ {
		from, to := points[index-1], points[index]
		dx, dy := to.X-from.X, to.Y-from.Y
		length := float32(math.Hypot(float64(dx), float64(dy)))
		if length <= 0 {
			continue
		}
		for position := remaining; position <= length; position += spacing {
			ratio := position / length
			center := f32.Pt(from.X+dx*ratio, from.Y+dy*ratio)
			radius := max(width/2, 1)
			rect := image.Rect(int(center.X-radius), int(center.Y-radius), int(center.X+radius), int(center.Y+radius))
			paint.FillShape(gtx.Ops, color, clip.Ellipse(rect).Op(gtx.Ops))
		}
		used := length - remaining
		remaining = spacing - float32(math.Mod(float64(max(used, 0)), float64(spacing)))
		if remaining == spacing {
			remaining = 0
		}
	}
}
