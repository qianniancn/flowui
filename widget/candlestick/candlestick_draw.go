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
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
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

func candlestickStyleFor(visual theme.CandlestickVisual, disabled bool) chartStyle {
	opacity := float32(1)
	if disabled {
		opacity = visual.DisabledOpacity
	}
	return chartStyle{
		axis:      visual.Colors.Axis,
		axisLabel: visual.Colors.AxisLabel,
		grid:      visual.Colors.Grid,
		crosshair: visual.Colors.Crosshair,
		opacity:   opacity,
	}
}

func drawGrid(gtx layout.Context, geometry chartGeometry, style chartStyle, show bool, metrics theme.CandlestickMetrics) {
	if !show || geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(metrics.GridWidth), 1))
	for _, tick := range geometry.yTicks {
		drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), tick.pixel), f32.Pt(float32(geometry.plot.Max.X), tick.pixel), width, style.grid)
	}
}

func drawAxes(gtx layout.Context, geometry chartGeometry, style chartStyle, metrics theme.CandlestickMetrics) {
	if geometry.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(metrics.AxisWidth), 1))
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Min.Y)), f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)), width, style.axis)
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), float32(geometry.plot.Max.Y)), f32.Pt(float32(geometry.plot.Max.X), float32(geometry.plot.Max.Y)), width, style.axis)
}

func (w Widget) drawCrosshair(ctx *core.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, pointer f32.Point, style chartStyle, visual theme.CandlestickVisual) {
	if !selection.candle.valid || geometry.plot.Empty() {
		return
	}
	pointerY := min(max(pointer.Y, float32(geometry.plot.Min.Y)), float32(geometry.plot.Max.Y))
	metrics := visual.Metrics
	width := float32(max(gtx.Dp(metrics.CrosshairWidth), 1))
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	drawLine(gtx, f32.Pt(selection.pixelX, float32(geometry.plot.Min.Y)), f32.Pt(selection.pixelX, float32(geometry.plot.Max.Y)), width, style.crosshair)
	drawLine(gtx, f32.Pt(float32(geometry.plot.Min.X), pointerY), f32.Pt(float32(geometry.plot.Max.X), pointerY), width, style.crosshair)
	area.Pop()

	padding := max(gtx.Dp(metrics.CrosshairLabelPadding), 0)
	radius := max(gtx.Dp(metrics.CrosshairLabelRadius), 0)
	background := visual.Colors.CrosshairLabelBackground
	foreground := visual.Colors.CrosshairLabelForeground

	xLabel := recordChartText(ctx, gtx, w.categoryLabel(selection.index), metrics.AxisTextSize, metrics.TextWeight, foreground, max(geometry.plot.Dx()-padding*2, 1))
	xSize := xLabel.dims.Size.Add(image.Pt(padding*2, padding*2))
	x := int(math.Round(float64(selection.pixelX))) - xSize.X/2
	x = min(max(x, 0), max(geometry.size.X-xSize.X, 0))
	y := min(geometry.plot.Max.Y, max(geometry.size.Y-xSize.Y, 0))
	drawCrosshairLabel(gtx, xLabel, image.Pt(x, y), xSize, padding, radius, background)

	value := geometry.yScale.ValueAt(float64(float32(geometry.plot.Max.Y)-pointerY) / float64(max(geometry.plot.Dy(), 1)))
	yLabel := recordChartText(ctx, gtx, w.yLabel(value, geometry.yScale.Interval), metrics.AxisTextSize, metrics.TextWeight, foreground, max(geometry.plot.Min.X-padding*2, 1))
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

func drawCandles(gtx layout.Context, data chartData, geometry chartGeometry, selectedIndex int, selected bool, metrics theme.CandlestickMetrics) {
	if geometry.plot.Empty() || geometry.categoryEnd <= geometry.categoryStart {
		return
	}
	area := clip.Rect(geometry.plot).Push(gtx.Ops)
	for index := geometry.categoryStart; index < geometry.categoryEnd && index < len(data.candles); index++ {
		candle := data.candles[index]
		if !candle.valid || candle.color.A == 0 {
			continue
		}
		borderWidth := float32(max(gtx.Dp(metrics.BorderWidth), 1))
		if selected && index == selectedIndex {
			borderWidth = float32(max(gtx.Dp(metrics.EmphasisBorderWidth), 1))
		}
		center := geometry.categoryCenter(index)
		highY, lowY := geometry.mapY(candle.high), geometry.mapY(candle.low)
		drawLine(gtx, f32.Pt(center, highY), f32.Pt(center, lowY), max(float32(gtx.Dp(metrics.WickWidth)), borderWidth), candle.color)
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
	chart.StrokeLine(gtx, from, to, width, lineColor)
}

func (w Widget) layoutAxisLabels(ctx *core.Context, gtx layout.Context, geometry chartGeometry, visual theme.CandlestickVisual, style chartStyle) {
	metrics := visual.Metrics
	gap := max(gtx.Dp(metrics.TickLabelGap), 0)
	for _, tick := range geometry.yTicks {
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Min.X-gap, 1))
		placeChartText(gtx, label, image.Pt(max(geometry.plot.Min.X-gap-label.dims.Size.X, 0), int(math.Round(float64(tick.pixel)))-label.dims.Size.Y/2))
	}
	for _, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx(), 1))
		x := categoryTickLabelX(tick.pixel, label.dims.Size.X, geometry.plot)
		placeChartText(gtx, label, image.Pt(x, geometry.plot.Max.Y+gap))
	}
}

func (w Widget) measureYLabelWidth(ctx *core.Context, gtx layout.Context, scale chart.LinearScale, visual theme.CandlestickVisual, maxWidth int) int {
	metrics := visual.Metrics
	width := 0
	for _, value := range scale.Ticks {
		label := recordChartText(ctx, gtx, w.yLabel(value, scale.Interval), metrics.AxisTextSize, metrics.TextWeight, visual.Colors.AxisLabel, maxWidth)
		width = max(width, label.dims.Size.X)
	}
	return width
}

func (w Widget) pruneCategoryTicks(ctx *core.Context, gtx layout.Context, geometry chartGeometry, visual theme.CandlestickVisual, style chartStyle) []categoryTick {
	if len(geometry.xTicks) <= 1 {
		w.resolveCategoryTickLabels(geometry, geometry.xTicks)
		return geometry.xTicks
	}
	metrics := visual.Metrics
	result := make([]categoryTick, 0, len(geometry.xTicks))
	lastRight := math.MinInt
	lastWidth := 0
	for index, tick := range geometry.xTicks {
		if index > 0 && int(tick.pixel)-lastWidth/2 < lastRight+16 {
			continue
		}
		tick.label = w.categoryTickLabel(geometry, tick)
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx(), 1))
		left := categoryTickLabelX(tick.pixel, label.dims.Size.X, geometry.plot)
		if index == 0 || left >= lastRight+16 {
			result = append(result, tick)
			lastRight = left + label.dims.Size.X
			lastWidth = label.dims.Size.X
		}
	}
	return result
}

func (w Widget) categoryTickLabel(geometry chartGeometry, tick categoryTick) string {
	if tick.label != "" {
		return tick.label
	}
	return w.axisCategoryLabel(tick.index, geometry.timeFormat)
}

func (w Widget) resolveCategoryTickLabels(geometry chartGeometry, ticks []categoryTick) {
	for index := range ticks {
		ticks[index].label = w.categoryTickLabel(geometry, ticks[index])
	}
}

func categoryTickLabelX(pixel float32, width int, plot image.Rectangle) int {
	x := int(math.Round(float64(pixel))) - width/2
	return min(max(x, plot.Min.X), max(plot.Max.X-width, plot.Min.X))
}

func (w Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, geometry chartGeometry, visual theme.CandlestickVisual, style chartStyle) {
	value := w.emptyText
	if !w.hasEmptyText {
		value = ctx.Text(i18n.ChartNoData)
	}
	metrics := visual.Metrics
	label := recordChartText(ctx, gtx, value, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx(), 1))
	placeChartText(gtx, label, image.Pt(geometry.plot.Min.X+max((geometry.plot.Dx()-label.dims.Size.X)/2, 0), geometry.plot.Min.Y+max((geometry.plot.Dy()-label.dims.Size.Y)/2, 0)))
}

func recordChartText(ctx *core.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, textColor color.NRGBA, maxWidth int) recordedChartText {
	call, dims := chart.RecordText(ctx, gtx, value, size, weight, textColor, maxWidth)
	return recordedChartText{call: call, dims: dims}
}

func placeChartText(gtx layout.Context, value recordedChartText, position image.Point) {
	chart.PlaceRecorded(gtx, value.call, value.dims, position)
}
