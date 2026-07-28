package linechart

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/components/chart"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawChartGrid(gtx layout.Context, geometry chartGeometry, style chartStyle, show bool, tokens theme.LineChartTheme) {
	if !show || geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.GridWidth), 1))
	for _, tick := range geometry.yTicks {
		drawChartLine(gtx, f32.Pt(float32(geometry.plot.Min.X), tick.pixel), f32.Pt(float32(geometry.plot.Max.X), tick.pixel), width, style.grid)
	}
}

func drawChartAxes(gtx layout.Context, geometry chartGeometry, style chartStyle, tokens theme.LineChartTheme) {
	if geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.AxisWidth), 1))
	drawChartLine(
		gtx,
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Min.Y)),
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)),
		width,
		style.axis,
	)
	drawChartLine(
		gtx,
		f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)),
		f32.Pt(float32(geometry.plot.Max.X), float32(geometry.plot.Max.Y)),
		width,
		style.axis,
	)
}

func drawChartSeries(ctx *frame.Context, gtx layout.Context, data chartData, geometry chartGeometry, _ chartStyle, tokens theme.LineChartTheme) {
	if geometry.plot.Empty() {
		return
	}
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	for _, series := range data.series {
		drawOneChartSeries(ctx, gtx, series, geometry, tokens)
	}
	area.Pop()
}

func drawOneChartSeries(ctx *frame.Context, gtx layout.Context, series resolvedSeries, geometry chartGeometry, tokens theme.LineChartTheme) {
	segments := seriesPixelSegments(series, geometry)
	if series.area {
		for _, segment := range segments {
			if len(segment.points) == 0 || len(segment.stackedOn) == 0 {
				continue
			}
			var area clip.Path
			area.Begin(gtx.Ops)
			area.MoveTo(segment.points[0])
			for _, point := range segment.points[1:] {
				area.LineTo(point)
			}
			for index := len(segment.stackedOn) - 1; index >= 0; index-- {
				area.LineTo(segment.stackedOn[index])
			}
			area.Close()
			paint.FillShape(gtx.Ops, series.areaColor, clip.Outline{Path: area.End()}.Op())
		}
	}
	for _, segment := range segments {
		drawLineSeriesSegment(gtx, segment.points, series.width, series.color, series.lineStyle)
	}

	validCount := countValidPoints(series.points)
	showPoints := series.showPoints
	if showPoints && !series.pointsSet {
		showPoints = validCount <= max(geometry.plot.Dx()/8, 24)
	}
	if !showPoints {
		return
	}
	pointSize := series.pointSize
	if pointSize <= 0 {
		pointSize = gtx.Dp(tokens.PointSize)
	}
	pointSize = max(pointSize, 2)
	for _, point := range series.points {
		if !point.valid {
			continue
		}
		drawChartHollowPoint(gtx, f32.Pt(geometry.mapX(point.X), geometry.mapY(point.Y)), pointSize, series.color, ctx.BackgroundColor())
	}
}

func countValidPoints(points []resolvedPoint) int {
	count := 0
	for _, point := range points {
		if point.valid {
			count++
		}
	}
	return count
}

func walkLine(points []resolvedPoint, connectNulls bool, visit func(resolvedPoint, bool)) {
	started := false
	for _, point := range points {
		if !point.valid {
			if !connectNulls {
				started = false
			}
			continue
		}
		visit(point, !started)
		started = true
	}
}

func drawChartSelection(ctx *frame.Context, gtx layout.Context, selection chartSelection, tokens theme.LineChartTheme) {
	if len(selection.entries) == 0 {
		return
	}
	pointSize := max(gtx.Dp(tokens.HoverPointSize), 4)
	for _, entry := range selection.entries {
		drawChartEmphasisPoint(gtx, entry.pixel, pointSize, entry.series.color, ctx.BackgroundColor())
	}
}

func drawChartCrosshair(gtx layout.Context, geometry chartGeometry, selection chartSelection, style chartStyle, tokens theme.LineChartTheme) {
	if len(selection.entries) == 0 {
		return
	}
	width := float32(max(gtx.Dp(tokens.CrosshairWidth), 1))
	drawChartLine(
		gtx,
		f32.Pt(selection.pixelX, float32(geometry.plot.Min.Y)),
		f32.Pt(selection.pixelX, float32(geometry.plot.Max.Y)),
		width,
		style.crosshair,
	)
}

func drawChartLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	chart.StrokeLine(gtx, from, to, width, col)
}

func drawChartHollowPoint(gtx layout.Context, center f32.Point, diameter int, col, background color.NRGBA) {
	outer := chart.PointRect(center, diameter)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(outer).Op(gtx.Ops))
	innerSize := max(diameter-4, 2)
	inner := chart.PointRect(center, innerSize)
	paint.FillShape(gtx.Ops, background, clip.Ellipse(inner).Op(gtx.Ops))
}

func drawChartEmphasisPoint(gtx layout.Context, center f32.Point, diameter int, col, background color.NRGBA) {
	outer := chart.PointRect(center, diameter+4)
	paint.FillShape(gtx.Ops, background, clip.Ellipse(outer).Op(gtx.Ops))
	inner := chart.PointRect(center, diameter)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(inner).Op(gtx.Ops))
}
