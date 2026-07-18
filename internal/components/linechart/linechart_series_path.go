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

type linePixelSegment struct {
	points    []f32.Point
	stackedOn []f32.Point
}

func seriesPixelSegments(series resolvedSeries, geometry chartGeometry) []linePixelSegment {
	pointSegments := splitSmoothLine(series.points, series.connectNulls, func(point resolvedPoint) f32.Point {
		return f32.Pt(geometry.mapX(point.X), geometry.mapY(point.Y))
	})
	var baseSegments [][]f32.Point
	if series.area {
		baseline := min(max(float64(0), geometry.yScale.Minimum), geometry.yScale.Maximum)
		baseSegments = splitSmoothLine(series.points, series.connectNulls, func(point resolvedPoint) f32.Point {
			return f32.Pt(geometry.mapX(point.X), geometry.mapY(linePointStackBase(point, baseline)))
		})
	}
	segments := make([]linePixelSegment, 0, len(pointSegments))
	for index, points := range pointSegments {
		var stackedOn []f32.Point
		if series.area {
			stackedOn = baseSegments[index]
		}
		start, end := visiblePixelRange(points, float32(geometry.plot.Min.X), float32(geometry.plot.Max.X))
		points = points[start:end]
		if series.area {
			stackedOn = stackedOn[start:end]
		}
		if series.sampling == SamplingMinMax {
			if sample := minMaxPixelSampleIndices(points, max(geometry.plot.Dx(), 1)); sample != nil {
				points = selectPixelPoints(points, sample)
				if series.area {
					stackedOn = selectPixelPoints(stackedOn, sample)
				}
			}
		}
		if series.step != StepNone {
			points = steppedPoints(points, series.step)
			if series.area {
				stackedOn = steppedPoints(stackedOn, series.step)
			}
		} else {
			if series.smooth > 0 {
				points = sampledSmoothPoints(points, series.smooth)
			}
			if series.area && series.stackedOnSmooth > 0 {
				stackedOn = sampledSmoothPoints(stackedOn, series.stackedOnSmooth)
			}
		}
		segments = append(segments, linePixelSegment{points: points, stackedOn: stackedOn})
	}
	return segments
}

func visiblePixelSegment(points []f32.Point, minimumX, maximumX float32) []f32.Point {
	start, end := visiblePixelRange(points, minimumX, maximumX)
	return points[start:end]
}

func visiblePixelRange(points []f32.Point, minimumX, maximumX float32) (int, int) {
	if len(points) < 2 || !monotonicPointX(points) {
		return 0, len(points)
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
		return 0, 0
	}
	return start, end
}

func minMaxPixelSample(points []f32.Point, pixelWidth int) []f32.Point {
	indices := minMaxPixelSampleIndices(points, pixelWidth)
	if indices == nil {
		return points
	}
	return selectPixelPoints(points, indices)
}

func minMaxPixelSampleIndices(points []f32.Point, pixelWidth int) []int {
	if len(points) <= pixelWidth*2 || pixelWidth <= 0 || !monotonicPointX(points) {
		return nil
	}
	minimumX := points[0].X
	maximumX := points[len(points)-1].X
	if maximumX <= minimumX {
		return nil
	}
	result := make([]int, 0, pixelWidth*2+2)
	result = append(result, 0)
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
			result = append(result, minimumIndex, maximumIndex)
		} else if maximumIndex < minimumIndex {
			result = append(result, maximumIndex, minimumIndex)
		} else {
			result = append(result, minimumIndex)
		}
		start = end
	}
	result = append(result, len(points)-1)
	return result
}

func selectPixelPoints(points []f32.Point, indices []int) []f32.Point {
	selected := make([]f32.Point, len(indices))
	for index, sourceIndex := range indices {
		selected[index] = points[sourceIndex]
	}
	return selected
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
