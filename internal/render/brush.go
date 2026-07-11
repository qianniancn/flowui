package render

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

type brushKind uint8

const (
	brushNone brushKind = iota
	brushSolid
	brushLinearGradient
)

// GradientStop describes one color along a gradient. Offset is normalized to
// the inclusive range from 0 to 1 when a brush is constructed.
type GradientStop struct {
	Offset float32
	Color  color.NRGBA
}

// Brush is an immutable fill style. Its zero value draws nothing.
type Brush struct {
	kind  brushKind
	color color.NRGBA
	stops []GradientStop
	angle float32
}

func SolidBrush(col color.NRGBA) Brush {
	return Brush{kind: brushSolid, color: col}
}

// LinearGradient constructs a gradient from two or more color stops. A single
// stop is treated as a solid brush, while an empty stop list draws nothing.
func LinearGradient(stops ...GradientStop) Brush {
	normalized := normalizeGradientStops(stops)
	switch len(normalized) {
	case 0:
		return Brush{}
	case 1:
		return SolidBrush(normalized[0].Color)
	default:
		return Brush{
			kind:  brushLinearGradient,
			stops: normalized,
			angle: 180,
		}
	}
}

// Angle returns a copy using CSS angle semantics: 0 degrees points upward,
// 90 degrees points right, and 180 degrees points downward.
func (b Brush) Angle(degrees float32) Brush {
	if b.kind != brushLinearGradient || math.IsNaN(float64(degrees)) || math.IsInf(float64(degrees), 0) {
		return b
	}
	b.angle = normalizeDegrees(degrees)
	return b
}

// ColorAt samples a brush at a normalized offset. It is primarily useful for
// propagating an approximate background color through nested UI contexts.
func (b Brush) ColorAt(offset float32) color.NRGBA {
	if b.kind == brushSolid {
		return b.color
	}
	if b.kind != brushLinearGradient || len(b.stops) == 0 {
		return color.NRGBA{}
	}
	if math.IsNaN(float64(offset)) {
		offset = 0
	}
	offset = min(max(offset, 0), 1)
	previous := b.stops[0]
	for index := 1; index < len(b.stops); index++ {
		next := b.stops[index]
		if offset < next.Offset {
			span := next.Offset - previous.Offset
			if span <= 0 {
				return next.Color
			}
			return lerpNRGBA(previous.Color, next.Color, (offset-previous.Offset)/span)
		}
		previous = next
	}
	return b.stops[len(b.stops)-1].Color
}

// DrawBrush fills rect with brush and clips it to the supplied corner radius.
func DrawBrush(gtx layout.Context, rect image.Rectangle, radius int, brush Brush) {
	if rect.Empty() || brush.kind == brushNone {
		return
	}
	radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
	shape := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	defer shape.Pop()

	switch brush.kind {
	case brushSolid:
		paint.Fill(gtx.Ops, brush.color)
	case brushLinearGradient:
		start, end := linearGradientLine(rect, brush.angle)
		paintLinearGradient(gtx, start, end, brush.stops, float32(math.Hypot(float64(rect.Dx()), float64(rect.Dy()))))
	}
}

// PaintLinearGradient paints the current clip with a two-color gradient. It is
// the low-level path for components that provide their own non-rectangular clip.
func PaintLinearGradient(gtx layout.Context, start f32.Point, startColor color.NRGBA, end f32.Point, endColor color.NRGBA) {
	paint.LinearGradientOp{
		Stop1:  start,
		Color1: startColor,
		Stop2:  end,
		Color2: endColor,
	}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func normalizeGradientStops(stops []GradientStop) []GradientStop {
	if len(stops) == 0 {
		return nil
	}
	normalized := append([]GradientStop(nil), stops...)
	for index := range normalized {
		offset := normalized[index].Offset
		if math.IsNaN(float64(offset)) || math.IsInf(float64(offset), -1) {
			offset = 0
		} else if math.IsInf(float64(offset), 1) {
			offset = 1
		}
		normalized[index].Offset = min(max(offset, 0), 1)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].Offset < normalized[j].Offset
	})
	if len(normalized) == 1 {
		return normalized
	}
	if normalized[0].Offset > 0 {
		normalized = append([]GradientStop{{Color: normalized[0].Color}}, normalized...)
	}
	last := normalized[len(normalized)-1]
	if last.Offset < 1 {
		normalized = append(normalized, GradientStop{Offset: 1, Color: last.Color})
	}
	return normalized
}

func normalizeDegrees(degrees float32) float32 {
	degrees = float32(math.Mod(float64(degrees), 360))
	if degrees < 0 {
		degrees += 360
	}
	return degrees
}

func linearGradientLine(rect image.Rectangle, angle float32) (f32.Point, f32.Point) {
	radians := float64(normalizeDegrees(angle)) * math.Pi / 180
	direction := f32.Pt(float32(math.Sin(radians)), -float32(math.Cos(radians)))
	center := f32.Pt(
		float32(rect.Min.X+rect.Max.X)/2,
		float32(rect.Min.Y+rect.Max.Y)/2,
	)
	halfExtent := float32(math.Abs(float64(direction.X))*float64(rect.Dx())/2 + math.Abs(float64(direction.Y))*float64(rect.Dy())/2)
	delta := direction.Mul(halfExtent)
	return center.Sub(delta), center.Add(delta)
}

func paintLinearGradient(gtx layout.Context, start, end f32.Point, stops []GradientStop, crossExtent float32) {
	if len(stops) < 2 {
		return
	}
	if len(stops) == 2 {
		PaintLinearGradient(gtx, start, stops[0].Color, end, stops[1].Color)
		return
	}
	direction := end.Sub(start)
	length := float32(math.Hypot(float64(direction.X), float64(direction.Y)))
	if length <= 0 {
		paint.Fill(gtx.Ops, stops[len(stops)-1].Color)
		return
	}
	overlap := direction.Mul(1 / length)
	perpendicular := f32.Pt(-direction.Y/length, direction.X/length).Mul(max(crossExtent, 1))
	for index := 1; index < len(stops); index++ {
		previous := stops[index-1]
		next := stops[index]
		if next.Offset <= previous.Offset {
			continue
		}
		segmentStart := start.Add(direction.Mul(previous.Offset))
		segmentEnd := start.Add(direction.Mul(next.Offset))
		clipEnd := segmentEnd
		if index < len(stops)-1 {
			clipEnd = clipEnd.Add(overlap)
		}
		segmentClip := gradientSegmentClip(gtx, segmentStart, clipEnd, perpendicular)
		PaintLinearGradient(gtx, segmentStart, previous.Color, segmentEnd, next.Color)
		segmentClip.Pop()
	}
}

func gradientSegmentClip(gtx layout.Context, start, end, perpendicular f32.Point) clip.Stack {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(start.Sub(perpendicular))
	path.LineTo(end.Sub(perpendicular))
	path.LineTo(end.Add(perpendicular))
	path.LineTo(start.Add(perpendicular))
	path.Close()
	return clip.Outline{Path: path.End()}.Op().Push(gtx.Ops)
}

func lerpNRGBA(start, end color.NRGBA, progress float32) color.NRGBA {
	progress = min(max(progress, 0), 1)
	channel := func(a, b byte) byte {
		return byte(float32(a) + (float32(b)-float32(a))*progress + .5)
	}
	return color.NRGBA{
		R: channel(start.R, end.R),
		G: channel(start.G, end.G),
		B: channel(start.B, end.B),
		A: channel(start.A, end.A),
	}
}
