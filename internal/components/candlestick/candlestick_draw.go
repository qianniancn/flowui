package candlestick

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type chartStyle struct {
	axis      color.NRGBA
	axisLabel color.NRGBA
	grid      color.NRGBA
	crosshair color.NRGBA
	opacity   float32
}

type recordedChartText struct {
	call op.CallOp
	dims layout.Dimensions
}

func candlestickStyleFor(activeTheme *theme.Theme, disabled bool) chartStyle {
	grid := activeTheme.Palette.Border
	grid.A = byte(float32(grid.A) * .8)
	crosshair := activeTheme.Palette.MutedForeground
	crosshair.A = byte(float32(crosshair.A) * .75)
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return chartStyle{axis: activeTheme.Palette.Border, axisLabel: activeTheme.Palette.MutedForeground, grid: grid, crosshair: crosshair, opacity: opacity}
}

func drawGrid(gtx layout.Context, geometry chartGeometry, style chartStyle, show bool, tokens theme.CandlestickChartTheme) {
	if !show || geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.GridWidth), 1))
	for _, tick := range geometry.yTicks {
		drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), tick.pixel), f32.Pt(float32(geometry.plot.Max.X), tick.pixel), width, style.grid)
	}
}

func drawAxes(gtx layout.Context, geometry chartGeometry, style chartStyle, tokens theme.CandlestickChartTheme) {
	if geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.AxisWidth), 1))
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Min.Y)), f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)), width, style.axis)
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)), f32.Pt(float32(geometry.plot.Max.X), float32(geometry.plot.Max.Y)), width, style.axis)
}

func (w Widget) drawCrosshair(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, pointer f32.Point, style chartStyle, tokens theme.CandlestickChartTheme) {
	if !selection.candle.valid || geometry.plot.Empty() {
		return
	}
	pointerY := min(max(pointer.Y, float32(geometry.plot.Min.Y)), float32(geometry.plot.Max.Y))
	width := float32(max(gtx.Dp(tokens.CrosshairWidth), 1))
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	drawLine(gtx, f32.Pt(selection.pixelX, float32(geometry.plot.Min.Y)), f32.Pt(selection.pixelX, float32(geometry.plot.Max.Y)), width, style.crosshair)
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), pointerY), f32.Pt(float32(geometry.plot.Max.X), pointerY), width, style.crosshair)
	area.Pop()

	activeTheme := frame.ActiveTheme(ctx)
	tooltipTokens := activeTheme.Components.Tooltip
	padding := max(gtx.Dp(tooltipTokens.Padding), 0)
	radius := max(gtx.Dp(tooltipTokens.Radius), 0)
	background := activeTheme.Palette.MutedForeground
	foreground := activeTheme.Palette.Background

	xLabel := recordChartText(ctx, gtx, w.categoryLabel(selection.index), tokens.AxisTextSize, font.Normal, foreground, max(geometry.plot.Dx()-padding*2, 1))
	xSize := xLabel.dims.Size.Add(image.Pt(padding*2, padding*2))
	x := int(math.Round(float64(selection.pixelX))) - xSize.X/2
	x = min(max(x, 0), max(geometry.size.X-xSize.X, 0))
	y := min(geometry.plot.Max.Y, max(geometry.size.Y-xSize.Y, 0))
	drawCrosshairLabel(gtx, xLabel, image.Pt(x, y), xSize, padding, radius, background)

	value := geometry.yScale.valueAt(float64(float32(geometry.plot.Max.Y)-pointerY) / float64(max(geometry.plot.Dy(), 1)))
	yLabel := recordChartText(ctx, gtx, w.yLabel(value, geometry.yScale.interval), tokens.AxisTextSize, font.Normal, foreground, max(geometry.plot.Min.X-padding*2, 1))
	ySize := yLabel.dims.Size.Add(image.Pt(padding*2, padding*2))
	y = int(math.Round(float64(pointerY))) - ySize.Y/2
	y = min(max(y, geometry.plot.Min.Y), max(geometry.plot.Max.Y-ySize.Y, geometry.plot.Min.Y))
	drawCrosshairLabel(gtx, yLabel, image.Pt(max(geometry.plot.Min.X-ySize.X, 0), y), ySize, padding, radius, background)
}

func drawCrosshairLabel(gtx layout.Context, label recordedChartText, position, size image.Point, padding, radius int, background color.NRGBA) {
	if label.dims.Size.X <= 0 || label.dims.Size.Y <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Min: position, Max: position.Add(size)}
	radius = min(radius, min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	placeChartText(gtx, label, position.Add(image.Pt(padding, padding)))
}

func drawCandles(gtx layout.Context, data chartData, geometry chartGeometry, selectedIndex int, selected bool, tokens theme.CandlestickChartTheme) {
	if geometry.plot.Empty() || geometry.categoryEnd <= geometry.categoryStart {
		return
	}
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	for index := geometry.categoryStart; index < geometry.categoryEnd && index < len(data.candles); index++ {
		candle := data.candles[index]
		if !candle.valid || candle.color.A == 0 {
			continue
		}
		borderWidth := float32(max(gtx.Dp(tokens.BorderWidth), 1))
		if selected && index == selectedIndex {
			borderWidth = float32(max(gtx.Dp(tokens.EmphasisBorderWidth), 1))
		}
		center := geometry.categoryCenter(index)
		highY, lowY := geometry.mapY(candle.high), geometry.mapY(candle.low)
		drawLine(gtx, f32.Pt(center, highY), f32.Pt(center, lowY), max(float32(gtx.Dp(tokens.WickWidth)), borderWidth), candle.color)
		if geometry.candleWidth <= 1.3 {
			continue
		}
		left := int(math.Round(float64(center - geometry.candleWidth/2)))
		right := int(math.Round(float64(center + geometry.candleWidth/2)))
		if right <= left {
			right = left + 1
		}
		openY, closeY := geometry.mapY(candle.open), geometry.mapY(candle.close)
		top := int(math.Round(float64(min(openY, closeY))))
		bottom := int(math.Round(float64(max(openY, closeY))))
		if bottom <= top {
			y := float32(top) + .5
			drawLine(gtx, f32.Pt(float32(left), y), f32.Pt(float32(right), y), borderWidth, candle.color)
			continue
		}
		rect := image.Rect(left, top, right, bottom)
		paint.FillShape(gtx.Ops, candle.color, clip.Rect(rect).Op())
		path := clip.UniformRRect(rect, 0).Path(gtx.Ops)
		stroke := clip.Stroke{Path: path, Width: borderWidth}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, candle.color)
		stroke.Pop()
	}
	area.Pop()
}

func drawLine(gtx layout.Context, from, to f32.Point, width float32, lineColor color.NRGBA) {
	if width <= 0 || lineColor.A == 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, lineColor)
	stroke.Pop()
}

func (w Widget) layoutAxisLabels(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	gap := max(gtx.Dp(tokens.TickLabelGap), 0)
	for _, tick := range geometry.yTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Min.X-gap, 1))
		placeChartText(gtx, label, image.Pt(max(geometry.plot.Min.X-gap-label.dims.Size.X, 0), int(math.Round(float64(tick.pixel)))-label.dims.Size.Y/2))
	}
	for _, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
		x := categoryTickLabelX(tick.pixel, label.dims.Size.X, geometry.plot)
		placeChartText(gtx, label, image.Pt(x, geometry.plot.Max.Y+gap))
	}
}

func (w Widget) measureYLabelWidth(ctx *frame.Context, gtx layout.Context, scale linearScale, maxWidth int) int {
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	width := 0
	for _, value := range scale.ticks {
		label := recordChartText(ctx, gtx, w.yLabel(value, scale.interval), tokens.AxisTextSize, font.Normal, color.NRGBA{A: 0xff}, maxWidth)
		width = max(width, label.dims.Size.X)
	}
	return width
}

func (w Widget) pruneCategoryTicks(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) []categoryTick {
	if len(geometry.xTicks) <= 1 {
		return geometry.xTicks
	}
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	result := make([]categoryTick, 0, len(geometry.xTicks))
	lastRight := math.MinInt
	for index, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
		left := categoryTickLabelX(tick.pixel, label.dims.Size.X, geometry.plot)
		if index == 0 || left >= lastRight+16 {
			result = append(result, tick)
			lastRight = left + label.dims.Size.X
		}
	}
	return result
}

func categoryTickLabelX(pixel float32, width int, plot image.Rectangle) int {
	x := int(math.Round(float64(pixel))) - width/2
	return min(max(x, plot.Min.X), max(plot.Max.X-width, plot.Min.X))
}

func (w Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	value := w.emptyText
	if value == "" {
		value = "No data"
	}
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	label := recordChartText(ctx, gtx, value, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
	placeChartText(gtx, label, image.Pt(geometry.plot.Min.X+max((geometry.plot.Dx()-label.dims.Size.X)/2, 0), geometry.plot.Min.Y+max((geometry.plot.Dy()-label.dims.Size.Y)/2, 0)))
}

func recordChartText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, textColor color.NRGBA, maxWidth int) recordedChartText {
	if value == "" || maxWidth <= 0 {
		return recordedChartText{}
	}
	textGtx := gtx
	textGtx.Constraints.Min = image.Point{}
	textGtx.Constraints.Max.X = min(maxWidth, textGtx.Constraints.Max.X)
	macro := op.Record(gtx.Ops)
	label := material.Label(frame.ActiveTheme(ctx).Material, size, value)
	label.Color = textColor
	label.Font.Weight = weight
	label.MaxLines = 1
	dims := label.Layout(textGtx)
	return recordedChartText{call: macro.Stop(), dims: dims}
}

func placeChartText(gtx layout.Context, value recordedChartText, position image.Point) {
	if value.dims.Size.X <= 0 || value.dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	value.call.Add(gtx.Ops)
	offset.Pop()
}
