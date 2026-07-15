package candlestick

import (
	"fmt"
	"image"
	"math"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type axisTick struct {
	value float64
	label string
	pixel float32
}

type categoryTick struct {
	index int
	label string
	pixel float32
}

type chartGeometry struct {
	size          image.Point
	plot          image.Rectangle
	yScale        linearScale
	yTicks        []axisTick
	xTicks        []categoryTick
	categoryStart int
	categoryEnd   int
	bandWidth     float32
	candleWidth   float32
}

type chartSelection struct {
	index  int
	candle resolvedCandle
	pixelX float32
}

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := chartStateFor(ctx, w.key)
	restoreKey := frame.PushKey(ctx, w.key)
	defer restoreKey()
	enabled := gtx.Enabled() && !w.disabled
	activated, resetWindow := state.updateClicks(gtx, enabled)
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := resolveChartData(w, frame.ActiveTheme(ctx))
	if resetWindow && w.onDataWindowChange != nil {
		full := chart.FullDataWindow()
		if w.effectiveDataWindow() != full {
			w.onDataWindowChange(full)
		}
		activated = false
	}

	eventGtx := gtx
	if !enabled {
		eventGtx = eventGtx.Disabled()
	}
	eventGtx.Constraints = layout.Exact(size)
	semantic.EnabledOp(enabled).Add(eventGtx.Ops)
	semantic.DescriptionOp(w.semanticDescription(data)).Add(eventGtx.Ops)
	w.layoutContent(ctx, eventGtx, state, data, enabled, activated, size)
	state.addClickInput(eventGtx, size, enabled && (w.onDataClick != nil || w.onDataWindowChange != nil))
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutContent(ctx *frame.Context, gtx layout.Context, state *chartState, data chartData, enabled, activated bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.CandlestickChart
	style := candlestickStyleFor(activeTheme, !enabled)
	displayData := w.animatedData(ctx, gtx, state, data)
	yScale := w.resolveYScale(data)
	leftLabelWidth := w.measureYLabelWidth(ctx, gtx, yScale, max(size.X/2, 1))
	left := max(gtx.Dp(tokens.PlotPaddingLeft), leftLabelWidth+max(gtx.Dp(tokens.TickLabelGap), 0)+4)
	left = min(max(left, 0), max(size.X/2, 0))
	right := min(max(gtx.Dp(tokens.PlotPaddingRight), 0), max(size.X-left-1, 0))
	availableWidth := max(size.X-left-right, 0)
	yName := recordChartText(ctx, gtx, w.yAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)
	xName := recordChartText(ctx, gtx, w.xAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)
	top := max(gtx.Dp(tokens.PlotPaddingTop), 0)
	yNamePosition := image.Pt(left, top)
	if yName.dims.Size.Y > 0 {
		top += yName.dims.Size.Y + max(gtx.Dp(tokens.AxisNameGap), 0)
	}
	bottom := max(gtx.Dp(tokens.PlotPaddingBottom), 0)
	if xName.dims.Size.Y > 0 {
		bottom += xName.dims.Size.Y + max(gtx.Dp(tokens.AxisNameGap), 0)
	}
	plotLeft := min(left, max(size.X-1, 0))
	plotTop := min(top, max(size.Y-1, 0))
	plotBottom := min(max(size.Y-bottom, plotTop+1), size.Y)
	plotRight := min(max(size.X-right, plotLeft+1), size.X)
	geometry := w.resolveGeometry(data, size, image.Rect(plotLeft, plotTop, plotRight, plotBottom), yScale, gtx.Dp)
	geometry.xTicks = w.pruneCategoryTicks(ctx, gtx, geometry, style)

	state.updatePointer(gtx, enabled, geometry.plot, w.effectiveDataWindow(), w.onDataWindowChange)
	selectionEnabled := w.showTooltip || w.onDataClick != nil
	if !selectionEnabled {
		state.clearSelection()
	}
	selection := chartSelection{}
	selected := false
	if selectionEnabled {
		if index, ok := state.selectedIndex(geometry.categoryStart, geometry.categoryEnd, geometry.plot); ok {
			selection = w.resolveSelection(data, geometry, index)
			selected = selection.candle.valid
		}
	}
	if activated && selected && w.onDataClick != nil {
		w.onDataClick(w.publicSelection(selection))
	}
	tooltipVisible := enabled && w.showTooltip && selected
	if tooltipVisible {
		state.tooltipSelection = selection
	}
	tooltipProgress := float32(0)
	if enabled && w.showTooltip {
		tooltipProgress = state.tooltipTransition.Progress(gtx, tooltipVisible)
		if !tooltipVisible && tooltipProgress <= 0 {
			state.tooltipSelection = chartSelection{}
		}
	} else {
		state.tooltipTransition.Reset()
		state.tooltipSelection = chartSelection{}
	}
	tooltipPointer := state.pointer
	if tooltipVisible || tooltipProgress > 0 {
		tooltipPointer = state.tooltipTransition.Position(gtx, state.pointer)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeChartText(gtx, yName, yNamePosition)
	drawGrid(gtx, geometry, style, w.showGrid, tokens)
	w.drawMarkAreas(ctx, gtx, geometry, style)
	drawAxes(gtx, geometry, style, tokens)
	drawCandles(gtx, displayData, geometry, selection.index, selected, tokens)
	w.drawMarkLinesAndPoints(ctx, gtx, geometry, style)
	w.layoutAxisLabels(ctx, gtx, geometry, style)
	if selected && w.showTooltip && w.showCrosshair {
		w.drawCrosshair(ctx, gtx, geometry, selection, state.pointer, style, tokens)
	}
	if !data.extent.Valid {
		w.layoutEmpty(ctx, gtx, geometry, style)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, gtx, state.tooltipSelection, geometry.yScale.interval, tooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting())
	}
	if xName.dims.Size.X > 0 {
		placeChartText(gtx, xName, image.Pt(max(geometry.plot.Max.X-xName.dims.Size.X, 0), max(size.Y-xName.dims.Size.Y, 0)))
	}
	opacity.Pop()
	state.addPointerInput(gtx, geometry.plot, enabled && (selectionEnabled || w.onDataWindowChange != nil))
}

func (w Widget) resolveGeometry(data chartData, size image.Point, plot image.Rectangle, scale linearScale, dp func(unit.Dp) int) chartGeometry {
	geometry := chartGeometry{size: size, plot: plot, yScale: scale}
	for _, value := range scale.ticks {
		geometry.yTicks = append(geometry.yTicks, axisTick{value: value, label: w.yLabel(value, scale.interval), pixel: geometry.mapY(value)})
	}
	if len(data.candles) > 0 {
		geometry.categoryStart, geometry.categoryEnd = visibleCategoryRange(len(data.candles), w.effectiveDataWindow())
		visible := geometry.categoryEnd - geometry.categoryStart
		geometry.bandWidth = float32(plot.Dx()) / float32(visible)
		geometry.candleWidth = w.resolveCandleWidth(geometry.bandWidth, dp)
		timeFormat := w.timeAxisFormat(geometry.categoryStart, geometry.categoryEnd)
		for index := geometry.categoryStart; index < geometry.categoryEnd; index++ {
			geometry.xTicks = append(geometry.xTicks, categoryTick{index: index, label: w.axisCategoryLabel(index, timeFormat), pixel: geometry.categoryCenter(index)})
		}
	}
	return geometry
}

func (w Widget) resolveYScale(data chartData) linearScale {
	minimum, maximum := data.extent.Minimum, data.extent.Maximum
	if !w.hasYRange && !w.effectiveDataWindow().IsFull() {
		start, end := visibleCategoryRange(len(data.candles), w.effectiveDataWindow())
		visible := dataExtent{}
		for index := start; index < end; index++ {
			candle := data.candles[index]
			if !candle.valid {
				continue
			}
			visible.Include(candle.open)
			visible.Include(candle.close)
			visible.Include(candle.low)
			visible.Include(candle.high)
		}
		if visible.Valid {
			minimum, maximum = visible.Minimum, visible.Maximum
		}
	}
	if w.hasYRange {
		minimum, maximum = w.yMin, w.yMax
	}
	return newLinearScale(minimum, maximum, w.yTickCount, w.hasYRange)
}

func (w Widget) resolveCandleWidth(bandWidth float32, dp func(unit.Dp) int) float32 {
	if w.width > 0 {
		return float32(dp(w.width))
	}
	maximum := bandWidth
	if w.maxWidth > 0 {
		maximum = float32(dp(w.maxWidth))
	}
	minimum := float32(1)
	if w.minWidth > 0 {
		minimum = float32(dp(w.minWidth))
	}
	return max(min(bandWidth/2, maximum), minimum)
}

func (g chartGeometry) mapY(value float64) float32 {
	return float32(g.plot.Max.Y) - float32(g.yScale.ratio(value))*float32(g.plot.Dy())
}

func (g chartGeometry) categoryCenter(index int) float32 {
	return float32(g.plot.Min.X) + (float32(index-g.categoryStart)+.5)*g.bandWidth
}

func (w Widget) resolveSelection(data chartData, geometry chartGeometry, index int) chartSelection {
	if index < geometry.categoryStart || index >= geometry.categoryEnd || index >= len(data.candles) {
		return chartSelection{}
	}
	return chartSelection{index: index, candle: data.candles[index], pixelX: geometry.categoryCenter(index)}
}

func (w Widget) publicSelection(selection chartSelection) chart.Selection {
	candle := selection.candle
	seriesLabel := w.label
	if seriesLabel == "" {
		seriesLabel = "Candlestick"
	}
	return chart.Selection{
		Label: w.categoryLabel(selection.index),
		Index: selection.index,
		X:     float64(selection.index),
		Items: []chart.Datum{{
			SeriesKey: w.key, SeriesLabel: seriesLabel, X: float64(selection.index), Y: candle.close,
			Open: candle.open, Close: candle.close, Low: candle.low, High: candle.high, Color: candle.color,
		}},
	}
}

func (w Widget) semanticDescription(data chartData) string {
	label := w.label
	if label == "" {
		label = "Candlestick chart"
	}
	return fmt.Sprintf("%s, %d candles", label, len(data.candles))
}

func visibleCategoryRange(count int, window chart.DataWindow) (int, int) {
	if count <= 0 {
		return 0, 0
	}
	start := int(math.Floor(float64(window.Start) * float64(count)))
	end := int(math.Ceil(float64(window.End) * float64(count)))
	start = min(max(start, 0), count-1)
	end = min(max(end, start+1), count)
	return start, end
}
