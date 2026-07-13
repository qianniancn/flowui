package piechart

import (
	"image"
	"image/color"
	"math"
	"sort"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func drawPieSlices(gtx layout.Context, slices []resolvedSlice, geometry chartGeometry, hoveredKey string, emphasis float32) {
	for _, slice := range slices {
		if slice.key != hoveredKey {
			drawPieSlice(gtx, slice, geometry, geometry.outerRadius)
		}
	}
	for _, slice := range slices {
		if slice.key == hoveredKey {
			drawPieSlice(gtx, slice, geometry, geometry.outerRadius+max(emphasis, 0))
			return
		}
	}
}

func drawPieSlice(gtx layout.Context, slice resolvedSlice, geometry chartGeometry, outerRadius float32) {
	if slice.color.A == 0 || slice.sweep() <= 1e-5 || outerRadius <= 0 {
		return
	}
	path := pieSectorPath(gtx, geometry.center, geometry.innerRadius, outerRadius, slice.startAngle, slice.endAngle)
	stack := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, slice.color)
	stack.Pop()
}

func drawEmptyPie(gtx layout.Context, geometry chartGeometry, fill color.NRGBA) {
	if geometry.outerRadius <= 0 || fill.A == 0 {
		return
	}
	path := pieSectorPath(gtx, geometry.center, geometry.innerRadius, geometry.outerRadius, 0, float32(fullCircle))
	stack := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, fill)
	stack.Pop()
}

func pieSectorPath(gtx layout.Context, center f32.Point, innerRadius, outerRadius, startAngle, endAngle float32) clip.PathSpec {
	outerStart := piePoint(center, outerRadius, startAngle)
	sweep := endAngle - startAngle
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(outerStart)
	path.ArcTo(center, center, sweep)
	if innerRadius > 0 {
		path.LineTo(piePoint(center, innerRadius, endAngle))
		path.ArcTo(center, center, -sweep)
	} else {
		path.LineTo(center)
	}
	path.Close()
	return path.End()
}

func piePoint(center f32.Point, radius, angle float32) f32.Point {
	return f32.Pt(
		center.X+float32(math.Cos(float64(angle)))*radius,
		center.Y+float32(math.Sin(float64(angle)))*radius,
	)
}

func hasVisibleSector(slices []resolvedSlice) bool {
	for _, slice := range slices {
		if slice.sweep() > 1e-5 {
			return true
		}
	}
	return false
}

type pieLabel struct {
	slice    resolvedSlice
	text     recordedText
	right    bool
	anchor   f32.Point
	bend     f32.Point
	end      f32.Point
	centerY  float32
	position image.Point
}

func (w Widget) drawLabels(ctx *frame.Context, gtx layout.Context, slices []resolvedSlice, geometry chartGeometry, style chartStyle) {
	tokens := frame.ActiveTheme(ctx).Components.PieChart
	line1 := float32(max(gtx.Dp(tokens.LabelLineLength), 0))
	line2 := float32(max(gtx.Dp(tokens.LabelLineLength2), 0))
	labelGap := float32(max(gtx.Dp(tokens.LabelGap), 0))
	labels := make([]pieLabel, 0, len(slices))
	for _, slice := range slices {
		if slice.sweep() <= 1e-5 {
			continue
		}
		mid := (slice.startAngle + slice.endAngle) / 2
		anchor := piePoint(geometry.center, geometry.outerRadius+1, mid)
		bend := piePoint(geometry.center, geometry.outerRadius+line1, mid)
		right := float32(math.Cos(float64(mid))) >= 0
		label := recordText(ctx, gtx, slice.label, tokens.LabelTextSize, font.Normal, style.text, max(geometry.area.Dx()/2, 1))
		labels = append(labels, pieLabel{slice: slice, text: label, right: right, anchor: anchor, bend: bend, centerY: bend.Y})
	}
	resolveLabelSide(labels, true, geometry.area, line2, labelGap)
	resolveLabelSide(labels, false, geometry.area, line2, labelGap)
	for _, label := range labels {
		drawPieLine(gtx, label.anchor, label.bend, float32(max(gtx.Dp(tokens.LabelLineWidth), 1)), label.slice.color)
		drawPieLine(gtx, label.bend, label.end, float32(max(gtx.Dp(tokens.LabelLineWidth), 1)), label.slice.color)
		placeText(gtx, label.text, label.position)
	}
}

func resolveLabelSide(labels []pieLabel, right bool, area image.Rectangle, lineLength, gap float32) {
	indexes := make([]int, 0, len(labels))
	for index := range labels {
		if labels[index].right == right {
			indexes = append(indexes, index)
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool { return labels[indexes[i]].centerY < labels[indexes[j]].centerY })
	previousBottom := float32(area.Min.Y)
	for _, index := range indexes {
		halfHeight := float32(labels[index].text.dims.Size.Y) / 2
		labels[index].centerY = max(labels[index].centerY, previousBottom+halfHeight)
		previousBottom = labels[index].centerY + halfHeight + 2
	}
	nextTop := float32(area.Max.Y)
	for position := len(indexes) - 1; position >= 0; position-- {
		index := indexes[position]
		halfHeight := float32(labels[index].text.dims.Size.Y) / 2
		labels[index].centerY = min(labels[index].centerY, nextTop-halfHeight)
		nextTop = labels[index].centerY - halfHeight - 2
	}
	for _, index := range indexes {
		direction := float32(1)
		if !right {
			direction = -1
		}
		labels[index].bend.Y = labels[index].centerY
		labels[index].end = f32.Pt(labels[index].bend.X+direction*lineLength, labels[index].centerY)
		x := labels[index].end.X + gap
		if !right {
			x = labels[index].end.X - gap - float32(labels[index].text.dims.Size.X)
		}
		x = min(max(x, float32(area.Min.X)), float32(max(area.Max.X-labels[index].text.dims.Size.X, area.Min.X)))
		y := labels[index].centerY - float32(labels[index].text.dims.Size.Y)/2
		y = min(max(y, float32(area.Min.Y)), float32(max(area.Max.Y-labels[index].text.dims.Size.Y, area.Min.Y)))
		labels[index].position = image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
	}
}

func drawPieLine(gtx layout.Context, from, to f32.Point, width float32, lineColor color.NRGBA) {
	if width <= 0 || lineColor.A == 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stack := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, lineColor)
	stack.Pop()
}
