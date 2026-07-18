package barchart

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
		rect, ok := barMarkAreaRect(mark, geometry)
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
		rect, ok := barMarkAreaRect(mark, geometry)
		if !ok {
			continue
		}
		label := recordChartText(ctx, gtx, mark.Label, frame.ActiveTheme(ctx).Components.BarChart.AxisTextSize, font.Medium, style.axisLabel, max(rect.Dx()-8, 1))
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
		from, to, ok := barMarkEndpoints(mark, geometry)
		if !ok {
			continue
		}
		width := float32(max(gtx.Dp(mark.ResolvedWidth(unit.Dp(1))), 1))
		color := mark.ResolvedColor(lineFallback)
		drawChartLine(gtx, from, to, width, color)
		if mark.Label != "" {
			label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.BarChart.AxisTextSize, font.Medium, color, max(geometry.plot.Dx()/2, 1))
			position := image.Pt(int(from.X)+4, int(from.Y)+4)
			if (!geometry.horizontal && mark.Axis == chart.AxisY) || (geometry.horizontal && mark.Axis == chart.AxisX) {
				position = image.Pt(max(geometry.plot.Max.X-label.dims.Size.X-4, geometry.plot.Min.X), max(int(from.Y)-label.dims.Size.Y-3, geometry.plot.Min.Y))
			}
			position = chart.ClampLabelPosition(position, label.dims.Size, geometry.plot)
			placeChartText(gtx, label, position)
		}
	}

	pointFallback := activeTheme.Palette.Warning
	for _, mark := range w.markPoints {
		if mark.X < float64(geometry.categoryStart) || mark.X >= float64(geometry.categoryEnd) || !geometry.yScale.Contains(mark.Y) {
			continue
		}
		center := f32.Pt(barCategoryX(geometry, mark.X), geometry.mapY(mark.Y))
		if geometry.horizontal {
			center = f32.Pt(geometry.mapY(mark.Y), barCategoryX(geometry, mark.X))
		}
		color := mark.ResolvedColor(pointFallback)
		size, custom := chart.LayoutMarkPointContent(ctx, gtx, mark, center, geometry.plot, unit.Dp(9), pointFallback)
		if !custom {
			paint.FillShape(gtx.Ops, color, clip.Ellipse(chart.PointRect(center, size)).Op(gtx.Ops))
		}
		if mark.Label != "" {
			label := recordChartText(ctx, gtx, mark.Label, activeTheme.Components.BarChart.AxisTextSize, font.Medium, style.axisLabel, max(geometry.plot.Dx()/2, 1))
			position := chart.ClampLabelPosition(image.Pt(int(center.X)+size/2+4, int(center.Y)-label.dims.Size.Y/2), label.dims.Size, geometry.plot)
			placeChartText(gtx, label, position)
		}
	}
}

func barMarkAreaRect(mark chart.MarkArea, geometry chartGeometry) (image.Rectangle, bool) {
	if mark.Axis == chart.AxisX {
		start := max(mark.Start, float64(geometry.categoryStart))
		end := min(mark.End, float64(geometry.categoryEnd-1))
		if end <= start {
			return image.Rectangle{}, false
		}
		if geometry.horizontal {
			return image.Rect(geometry.plot.Min.X, int(math.Floor(float64(barCategoryX(geometry, start)))), geometry.plot.Max.X, int(math.Ceil(float64(barCategoryX(geometry, end))))), true
		}
		return image.Rect(int(math.Floor(float64(barCategoryX(geometry, start)))), geometry.plot.Min.Y, int(math.Ceil(float64(barCategoryX(geometry, end)))), geometry.plot.Max.Y), true
	}
	start := max(mark.Start, geometry.yScale.Minimum)
	end := min(mark.End, geometry.yScale.Maximum)
	if end <= start {
		return image.Rectangle{}, false
	}
	if geometry.horizontal {
		return image.Rect(int(math.Floor(float64(geometry.mapY(start)))), geometry.plot.Min.Y, int(math.Ceil(float64(geometry.mapY(end)))), geometry.plot.Max.Y), true
	}
	return image.Rect(geometry.plot.Min.X, int(math.Floor(float64(geometry.mapY(end)))), geometry.plot.Max.X, int(math.Ceil(float64(geometry.mapY(start))))), true
}

func barMarkEndpoints(mark chart.MarkLine, geometry chartGeometry) (from, to f32.Point, ok bool) {
	if mark.Axis == chart.AxisX {
		if mark.Value < float64(geometry.categoryStart) || mark.Value >= float64(geometry.categoryEnd) {
			return from, to, false
		}
		x := barCategoryX(geometry, mark.Value)
		if geometry.horizontal {
			return f32.Pt(float32(geometry.plot.Min.X), x), f32.Pt(float32(geometry.plot.Max.X), x), true
		}
		return f32.Pt(x, float32(geometry.plot.Min.Y)), f32.Pt(x, float32(geometry.plot.Max.Y)), true
	}
	if !geometry.yScale.Contains(mark.Value) {
		return from, to, false
	}
	y := geometry.mapY(mark.Value)
	if geometry.horizontal {
		return f32.Pt(y, float32(geometry.plot.Min.Y)), f32.Pt(y, float32(geometry.plot.Max.Y)), true
	}
	return f32.Pt(float32(geometry.plot.Min.X), y), f32.Pt(float32(geometry.plot.Max.X), y), true
}

func barCategoryX(geometry chartGeometry, value float64) float32 {
	minimum := geometry.plot.Min.X
	if geometry.horizontal {
		minimum = geometry.plot.Min.Y
	}
	return float32(minimum) + (float32(value)-float32(geometry.categoryStart)+0.5)*geometry.bandWidth
}
