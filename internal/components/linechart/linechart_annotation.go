package linechart

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (w Widget) drawMarkAreas(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	if geometry.plot.Empty() {
		return
	}
	fallback := frame.ActiveTheme(ctx).Palette.Accent
	fallback.A = 0x18
	areaClip := clip.Rect(geometry.plot).Push(gtx.Ops)
	for _, mark := range w.markAreas {
		rect, ok := lineMarkAreaRect(mark, geometry)
		if !ok {
			continue
		}
		paint.FillShape(gtx.Ops, mark.ResolvedColor(fallback), clip.Rect(rect).Op())
	}
	areaClip.Pop()
	for _, mark := range w.markAreas {
		if mark.Label == "" {
			continue
		}
		rect, ok := lineMarkAreaRect(mark, geometry)
		if !ok {
			continue
		}
		label := recordChartText(ctx, gtx, mark.Label, frame.ActiveTheme(ctx).Components.LineChart.AxisTextSize, font.Medium, style.axisLabel, max(rect.Dx()-8, 1))
		placeChartText(gtx, label, image.Pt(rect.Min.X+4, rect.Min.Y+4))
	}
}

func (w Widget) drawMarkLinesAndPoints(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	if geometry.plot.Empty() {
		return
	}
	activeTheme := frame.ActiveTheme(ctx)
	lineFallback := activeTheme.Palette.Danger
	lineFallback.A = 0xc0
	for _, mark := range w.markLines {
		from, to, ok := lineMarkEndpoints(mark, geometry)
		if !ok {
			continue
		}
		width := float32(max(gtx.Dp(mark.ResolvedWidth(unit.Dp(1))), 1))
		color := mark.ResolvedColor(lineFallback)
		drawChartLine(gtx, from, to, width, color)
		if mark.Label != "" {
			label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.LineChart.AxisTextSize, font.Medium, color, max(geometry.plot.Dx()/2, 1))
			position := image.Pt(int(from.X)+4, int(from.Y)+4)
			if mark.Axis == chart.AxisY {
				position = image.Pt(max(geometry.plot.Max.X-label.dims.Size.X-4, geometry.plot.Min.X), max(int(from.Y)-label.dims.Size.Y-3, geometry.plot.Min.Y))
			}
			position = clampLineAnnotationLabel(position, label.dims.Size, geometry.plot)
			placeChartText(gtx, label, position)
		}
	}

	pointFallback := activeTheme.Palette.Warning
	for _, mark := range w.markPoints {
		if !geometry.xScale.contains(mark.X) || !geometry.yScale.contains(mark.Y) {
			continue
		}
		center := image.Pt(int(math.Round(float64(geometry.mapX(mark.X)))), int(math.Round(float64(geometry.mapY(mark.Y)))))
		size := max(gtx.Dp(mark.ResolvedSize(unit.Dp(9))), 3)
		color := mark.ResolvedColor(pointFallback)
		paint.FillShape(gtx.Ops, color, clip.Ellipse(chartPointRect(floatPoint(center), size)).Op(gtx.Ops))
		if mark.Label != "" {
			label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.LineChart.AxisTextSize, font.Medium, style.axisLabel, max(geometry.plot.Dx()/2, 1))
			position := clampLineAnnotationLabel(image.Pt(center.X+size/2+4, center.Y-label.dims.Size.Y/2), label.dims.Size, geometry.plot)
			placeChartText(gtx, label, position)
		}
	}
}

func clampLineAnnotationLabel(position, size image.Point, plot image.Rectangle) image.Point {
	position.X = min(max(position.X, plot.Min.X), max(plot.Max.X-size.X, plot.Min.X))
	position.Y = min(max(position.Y, plot.Min.Y), max(plot.Max.Y-size.Y, plot.Min.Y))
	return position
}

func lineMarkAreaRect(mark chart.MarkArea, geometry chartGeometry) (image.Rectangle, bool) {
	if mark.Axis == chart.AxisX {
		start := max(mark.Start, geometry.xScale.minimum)
		end := min(mark.End, geometry.xScale.maximum)
		if end <= start {
			return image.Rectangle{}, false
		}
		return image.Rect(int(math.Floor(float64(geometry.mapX(start)))), geometry.plot.Min.Y, int(math.Ceil(float64(geometry.mapX(end)))), geometry.plot.Max.Y), true
	}
	start := max(mark.Start, geometry.yScale.minimum)
	end := min(mark.End, geometry.yScale.maximum)
	if end <= start {
		return image.Rectangle{}, false
	}
	return image.Rect(geometry.plot.Min.X, int(math.Floor(float64(geometry.mapY(end)))), geometry.plot.Max.X, int(math.Ceil(float64(geometry.mapY(start))))), true
}

func lineMarkEndpoints(mark chart.MarkLine, geometry chartGeometry) (from, to f32.Point, ok bool) {
	if mark.Axis == chart.AxisX {
		if !geometry.xScale.contains(mark.Value) {
			return from, to, false
		}
		x := geometry.mapX(mark.Value)
		return f32.Pt(x, float32(geometry.plot.Min.Y)), f32.Pt(x, float32(geometry.plot.Max.Y)), true
	}
	if !geometry.yScale.contains(mark.Value) {
		return from, to, false
	}
	y := geometry.mapY(mark.Value)
	return f32.Pt(float32(geometry.plot.Min.X), y), f32.Pt(float32(geometry.plot.Max.X), y), true
}

func floatPoint(point image.Point) f32.Point {
	return f32.Pt(float32(point.X), float32(point.Y))
}
