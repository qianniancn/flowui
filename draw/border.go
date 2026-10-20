package draw

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// arcSampleStep is the maximum pixel distance between two points sampled on a
// corner arc. Smaller steps make corners smoother and produce more segments.
const arcSampleStep = 3

// RoundedBorder paints a rounded border inside rect.
func RoundedBorder(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
	RRectBorder(gtx, clip.UniformRRect(rect, radius), width, col)
}

// RoundedInsetStroke paints a uniform rounded stroke inside rect.
// radius is applied to the inset path, and inset controls its position inside rect.
func RoundedInsetStroke(gtx layout.Context, rect image.Rectangle, radius, width, inset int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	inset = max(inset, 0)
	borderRect := rect.Inset(inset)
	if borderRect.Empty() {
		return
	}
	radius = min(max(radius, 0), min(borderRect.Dx(), borderRect.Dy())/2)
	shape, offset := pixelAlignedStroke(clip.UniformRRect(borderRect, radius), width)
	if shape.Rect.Empty() {
		return
	}
	drawInsetStroke(gtx, shape, width, col, offset)
}

// RRectBorder paints a border inside a rounded rectangle.
func RRectBorder(gtx layout.Context, shape clip.RRect, width int, col color.NRGBA) {
	rect := shape.Rect
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	width = min(width, min(rect.Dx(), rect.Dy())/2)
	center := width / 2
	shape.Rect = rect.Inset(center)
	shape.NW = max(shape.NW-center, 0)
	shape.NE = max(shape.NE-center, 0)
	shape.SE = max(shape.SE-center, 0)
	shape.SW = max(shape.SW-center, 0)
	shape, offset := pixelAlignedStroke(shape, width)
	if shape.Rect.Empty() {
		return
	}
	drawInsetStroke(gtx, shape, width, col, offset)
}

// pixelAlignedStroke moves an odd-width stroke's centerline half a pixel inward
// so the stroke covers whole device pixels.
//
// A stroke is centered on its path. With an integer-width path inset by
// width/2, an even width covers [0, width] inside the box exactly, but an odd
// width straddles the pixel grid and paints two half-covered pixels — a 1px
// border then reads as a soft 2px line. Shrinking the path's far edges by one
// pixel and shifting the centerline by half a pixel lands the same stroke on
// the grid instead, so the border is one solid, fully covered pixel.
func pixelAlignedStroke(shape clip.RRect, width int) (clip.RRect, f32.Point) {
	if width%2 == 0 {
		return shape, f32.Point{}
	}
	if shape.Rect.Dx() > 1 && shape.Rect.Dy() > 1 {
		shape.Rect.Max = shape.Rect.Max.Sub(image.Pt(1, 1))
	}
	return shape, f32.Pt(0.5, 0.5)
}

func drawInsetStroke(gtx layout.Context, shape clip.RRect, width int, col color.NRGBA, offset f32.Point) {
	if offset != (f32.Point{}) {
		stack := op.Affine(f32.Affine2D{}.Offset(offset)).Push(gtx.Ops)
		defer stack.Pop()
	}
	stroke := clip.Stroke{Path: shape.Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

// BottomBorder paints a straight border along the bottom edge of rect.
// The border is drawn inside the rectangle, matching RRectBorder.
func BottomBorder(gtx layout.Context, rect image.Rectangle, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	width = min(width, rect.Dy())
	paint.FillShape(gtx.Ops, col, clip.Rect{
		Min: image.Point{X: rect.Min.X, Y: rect.Max.Y - width},
		Max: rect.Max,
	}.Op())
}

// DashedRRectBorder paints a patterned border inside a rounded rectangle.
//
// Gio's clip.Stroke has no dash support, so the outline is walked by arc
// length and the "on" runs are collected into one path, then stroked once.
// on and off are pixel lengths; a non-positive off degrades to a solid border.
func DashedRRectBorder(gtx layout.Context, shape clip.RRect, width, on, off int, col color.NRGBA) {
	if off <= 0 {
		RRectBorder(gtx, shape, width, col)
		return
	}
	rect := shape.Rect
	if rect.Empty() || width <= 0 || on <= 0 || col.A == 0 {
		return
	}
	width = min(width, min(rect.Dx(), rect.Dy())/2)
	center := width / 2
	shape.Rect = rect.Inset(center)
	shape.NW = max(shape.NW-center, 0)
	shape.NE = max(shape.NE-center, 0)
	shape.SE = max(shape.SE-center, 0)
	shape.SW = max(shape.SW-center, 0)
	shape, offset := pixelAlignedStroke(shape, width)
	rect = shape.Rect
	if rect.Empty() || rect.Dx() <= 0 || rect.Dy() <= 0 {
		return
	}

	limit := min(rect.Dx(), rect.Dy()) / 2
	clamp := func(value int) int { return min(max(value, 0), limit) }
	points := rrectOutline(rect, [4]int{
		clamp(shape.NW), clamp(shape.NE), clamp(shape.SE), clamp(shape.SW),
	})
	if offset != (f32.Point{}) {
		for index, point := range points {
			points[index] = point.Add(offset)
		}
	}
	dashStroke(gtx, points, float32(on), float32(off), float32(width), col)
}

// rrectOutline returns the clockwise outline of a rounded rectangle as a
// closed polyline. A corner with radius zero contributes its vertex only.
func rrectOutline(rect image.Rectangle, radii [4]int) []f32.Point {
	minX, minY := float32(rect.Min.X), float32(rect.Min.Y)
	maxX, maxY := float32(rect.Max.X), float32(rect.Max.Y)
	rtl, rtr := float32(radii[0]), float32(radii[1])
	rbr, rbl := float32(radii[2]), float32(radii[3])

	points := make([]f32.Point, 0, 40)
	add := func(x, y float32) {
		point := f32.Pt(x, y)
		if last := len(points) - 1; last >= 0 && points[last] == point {
			return
		}
		points = append(points, point)
	}
	// Angles are radians and grow clockwise in Gio's y-down space.
	arc := func(cx, cy, r, from, to float32) {
		if r <= 0 {
			add(cx, cy)
			return
		}
		steps := max(int(math.Ceil(float64(r)*(float64(to-from))/arcSampleStep)), 1)
		for index := 0; index <= steps; index++ {
			angle := from + (to-from)*float32(index)/float32(steps)
			add(
				cx+r*float32(math.Cos(float64(angle))),
				cy+r*float32(math.Sin(float64(angle))),
			)
		}
	}

	add(minX+rtl, minY)
	add(maxX-rtr, minY)
	arc(maxX-rtr, minY+rtr, rtr, -math.Pi/2, 0)
	add(maxX, maxY-rbr)
	arc(maxX-rbr, maxY-rbr, rbr, 0, math.Pi/2)
	add(minX+rbl, maxY)
	arc(minX+rbl, maxY-rbl, rbl, math.Pi/2, math.Pi)
	add(minX, minY+rtl)
	arc(minX+rtl, minY+rtl, rtl, math.Pi, 3*math.Pi/2)
	add(minX+rtl, minY)
	return points
}

type dashSegment struct {
	from, to f32.Point
}

// dashStroke walks a closed polyline and strokes every "on" run of the
// on/off period, using the arc length travelled so far to keep the pattern
// continuous across corners.
func dashStroke(gtx layout.Context, points []f32.Point, on, off, width float32, col color.NRGBA) {
	period := on + off
	if len(points) < 2 || period <= 0 {
		return
	}

	var segments []dashSegment
	travelled := float32(0)
	for index := 0; index+1 < len(points); index++ {
		from, to := points[index], points[index+1]
		deltaX, deltaY := to.X-from.X, to.Y-from.Y
		length := float32(math.Hypot(float64(deltaX), float64(deltaY)))
		if length <= 0 {
			continue
		}
		position := float32(0)
		for position < length {
			phase := float32(math.Mod(float64(travelled+position), float64(period)))
			if phase < on {
				run := min(on-phase, length-position)
				segments = append(segments, dashSegment{
					from: lerpPoint(from, to, position/length),
					to:   lerpPoint(from, to, (position+run)/length),
				})
				position += run
				continue
			}
			position += min(period-phase, length-position)
		}
		travelled += length
	}
	if len(segments) == 0 {
		return
	}

	var path clip.Path
	path.Begin(gtx.Ops)
	for _, segment := range segments {
		path.MoveTo(segment.from)
		path.LineTo(segment.to)
	}
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func lerpPoint(from, to f32.Point, t float32) f32.Point {
	return f32.Pt(from.X+(to.X-from.X)*t, from.Y+(to.Y-from.Y)*t)
}
