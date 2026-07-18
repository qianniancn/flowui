package candlestick

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
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	for _, mark := range w.markAreas {
		rect, ok := candlestickMarkAreaRect(mark, geometry)
		if ok {
			paint.FillShape(gtx.Ops, mark.ResolvedColor(fallback), clip.Rect(rect).Op())
		}
	}
	area.Pop()

	for _, mark := range w.markAreas {
		if mark.Label == "" {
			continue
		}
		rect, ok := candlestickMarkAreaRect(mark, geometry)
		if !ok {
			continue
		}
		label := recordChartText(ctx, gtx, mark.Label, frame.ActiveTheme(ctx).Components.CandlestickChart.AxisTextSize, font.Medium, style.axisLabel, max(rect.Dx()-8, 1))
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
		from, to, ok := candlestickMarkEndpoints(mark, geometry)
		if !ok {
			continue
		}
		lineColor := mark.ResolvedColor(lineFallback)
		drawLine(gtx, from, to, float32(max(gtx.Dp(mark.ResolvedWidth(unit.Dp(1))), 1)), lineColor)
		if mark.Label == "" {
			continue
		}
		label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.CandlestickChart.AxisTextSize, font.Medium, lineColor, max(geometry.plot.Dx()/2, 1))
		position := image.Pt(int(from.X)+4, int(from.Y)+4)
		if mark.Axis == chart.AxisY {
			position = image.Pt(max(geometry.plot.Max.X-label.dims.Size.X-4, geometry.plot.Min.X), max(int(from.Y)-label.dims.Size.Y-3, geometry.plot.Min.Y))
		}
		placeChartText(gtx, label, chart.ClampLabelPosition(position, label.dims.Size, geometry.plot))
	}

	pointFallback := activeTheme.Palette.Warning
	for _, mark := range w.markPoints {
		if mark.X < float64(geometry.categoryStart) || mark.X >= float64(geometry.categoryEnd) || !geometry.yScale.Contains(mark.Y) {
			continue
		}
		center := f32.Pt(candlestickCategoryX(geometry, mark.X), geometry.mapY(mark.Y))
		pointColor := mark.ResolvedColor(pointFallback)
		size, custom := chart.LayoutMarkPointContent(ctx, gtx, mark, center, geometry.plot, unit.Dp(9), pointFallback)
		if !custom {
			paint.FillShape(gtx.Ops, pointColor, clip.Ellipse(chart.PointRect(center, size)).Op(gtx.Ops))
		}
		if mark.Label != "" {
			label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.CandlestickChart.AxisTextSize, font.Medium, style.axisLabel, max(geometry.plot.Dx()/2, 1))
			position := image.Pt(int(center.X)+size/2+4, int(center.Y)-label.dims.Size.Y/2)
			placeChartText(gtx, label, chart.ClampLabelPosition(position, label.dims.Size, geometry.plot))
		}
	}
}

func candlestickMarkAreaRect(mark chart.MarkArea, geometry chartGeometry) (image.Rectangle, bool) {
	if mark.Axis == chart.AxisX {
		start := max(mark.Start, float64(geometry.categoryStart))
		end := min(mark.End, float64(geometry.categoryEnd-1))
		if end <= start {
			return image.Rectangle{}, false
		}
		return image.Rect(int(math.Floor(float64(candlestickCategoryX(geometry, start)))), geometry.plot.Min.Y, int(math.Ceil(float64(candlestickCategoryX(geometry, end)))), geometry.plot.Max.Y), true
	}
	start := max(mark.Start, geometry.yScale.Minimum)
	end := min(mark.End, geometry.yScale.Maximum)
	if end <= start {
		return image.Rectangle{}, false
	}
	return image.Rect(geometry.plot.Min.X, int(math.Floor(float64(geometry.mapY(end)))), geometry.plot.Max.X, int(math.Ceil(float64(geometry.mapY(start))))), true
}

func candlestickMarkEndpoints(mark chart.MarkLine, geometry chartGeometry) (from, to f32.Point, ok bool) {
	if mark.Axis == chart.AxisX {
		if mark.Value < float64(geometry.categoryStart) || mark.Value >= float64(geometry.categoryEnd) {
			return from, to, false
		}
		x := candlestickCategoryX(geometry, mark.Value)
		return f32.Pt(x, float32(geometry.plot.Min.Y)), f32.Pt(x, float32(geometry.plot.Max.Y)), true
	}
	if !geometry.yScale.Contains(mark.Value) {
		return from, to, false
	}
	y := geometry.mapY(mark.Value)
	return f32.Pt(float32(geometry.plot.Min.X), y), f32.Pt(float32(geometry.plot.Max.X), y), true
}

func candlestickCategoryX(geometry chartGeometry, value float64) float32 {
	return float32(geometry.plot.Min.X) + (float32(value)-float32(geometry.categoryStart)+.5)*geometry.bandWidth
}
