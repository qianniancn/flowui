package candlestick

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
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
	yScale        chart.LinearScale
	yTicks        []axisTick
	xTicks        []categoryTick
	categoryStart int
	categoryEnd   int
	bandWidth     float32
	candleWidth   float32
	timeFormat    string
}

type geometryKey struct {
	generation      uint64
	themeGeneration uint64
	metric          unit.Metric
	size            image.Point
	plot            image.Rectangle
	window          chart.DataWindow
	yTickCount      int
	yMin            float64
	yMax            float64
	hasYRange       bool
	width           unit.Dp
	maxWidth        unit.Dp
	minWidth        unit.Dp
	formatY         uintptr
	formatTime      uintptr
	categoryCount   int
	timeCount       int
	scaleMinimum    float64
	scaleMaximum    float64
	scaleInterval   float64
}

type chartSelection struct {
	index  int
	candle resolvedCandle
	pixelX float32
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, visual theme.CandlestickVisual) layout.Dimensions {
	if w.animationDuration <= 0 {
		w.animationDuration = visual.Metrics.AnimationDuration
	}
	if w.updateAnimationDuration <= 0 {
		w.updateAnimationDuration = visual.Metrics.UpdateDuration
	}

	state := chartStateFor(ctx, w.key)
	restoreKey := core.PushKey(ctx, w.key)
	defer restoreKey()
	enabled := gtx.Enabled() && !w.disabled
	activated, resetWindow := state.updateClicks(gtx, enabled)
	active := core.ActiveTheme(ctx)
	metrics := visual.Metrics
	height := metrics.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := state.dataCache.resolve(w, active, visual)
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
	semantic.DescriptionOp(w.semanticDescription(ctx, data)).Add(eventGtx.Ops)
	w.layoutContent(ctx, eventGtx, state, data, visual, enabled, activated, size)
	state.addClickInput(eventGtx, size, enabled && (w.onDataClick != nil || w.onDataWindowChange != nil))
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutContent(ctx *core.Context, gtx layout.Context, state *chartState, data chartData, visual theme.CandlestickVisual, enabled, activated bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	metrics := visual.Metrics
	style := candlestickStyleFor(visual, !enabled)
	displayData := w.animatedData(ctx, gtx, state, data)
	yScale := w.resolveYScale(data)
	leftLabelWidth := w.measureYLabelWidth(ctx, gtx, yScale, visual, max(size.X/2, 1))
	left := max(gtx.Dp(metrics.PlotPaddingLeft), leftLabelWidth+max(gtx.Dp(metrics.TickLabelGap), 0)+4)
	left = min(max(left, 0), max(size.X/2, 0))
	right := min(max(gtx.Dp(metrics.PlotPaddingRight), 0), max(size.X-left-1, 0))
	availableWidth := max(size.X-left-right, 0)
	yName := recordChartText(ctx, gtx, w.yAxisLabel, metrics.AxisTextSize, metrics.EmphasisWeight, style.axisLabel, availableWidth)
	xName := recordChartText(ctx, gtx, w.xAxisLabel, metrics.AxisTextSize, metrics.EmphasisWeight, style.axisLabel, availableWidth)
	top := max(gtx.Dp(metrics.PlotPaddingTop), 0)
	yNamePosition := image.Pt(left, top)
	if yName.dims.Size.Y > 0 {
		top += yName.dims.Size.Y + max(gtx.Dp(metrics.AxisNameGap), 0)
	}
	bottom := max(gtx.Dp(metrics.PlotPaddingBottom), 0)
	if xName.dims.Size.Y > 0 {
		bottom += xName.dims.Size.Y + max(gtx.Dp(metrics.AxisNameGap), 0)
	}
	plotLeft := min(left, max(size.X-1, 0))
	plotTop := min(top, max(size.Y-1, 0))
	plotBottom := min(max(size.Y-bottom, plotTop+1), size.Y)
	plotRight := min(max(size.X-right, plotLeft+1), size.X)
	plot := image.Rect(plotLeft, plotTop, plotRight, plotBottom)
	geometry := state.geometryCache.Resolve(data.generation != 0, geometryKey{
		generation:      data.generation,
		themeGeneration: core.ThemeGeneration(ctx),
		metric:          gtx.Metric,
		size:            size,
		plot:            plot,
		window:          w.effectiveDataWindow(),
		yTickCount:      w.yTickCount,
		yMin:            w.yMin,
		yMax:            w.yMax,
		hasYRange:       w.hasYRange,
		width:           w.width,
		maxWidth:        w.maxWidth,
		minWidth:        w.minWidth,
		formatY:         chart.FuncID(w.formatY),
		formatTime:      chart.FuncID(w.formatTime),
		categoryCount:   len(w.categories),
		timeCount:       len(w.times),
		scaleMinimum:    yScale.Minimum,
		scaleMaximum:    yScale.Maximum,
		scaleInterval:   yScale.Interval,
	}, func() chartGeometry {
		return w.resolveGeometry(data, size, plot, yScale, gtx.Dp)
	})
	geometry.xTicks = w.pruneCategoryTicks(ctx, gtx, geometry, visual, style)

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
		w.onDataClick(w.publicSelectionFor(ctx, selection))
	}
	tooltipVisible := enabled && w.showTooltip && selected
	if tooltipVisible {
		state.tooltipSelection = selection
	}
	tooltipProgress := float32(0)
	if enabled && w.showTooltip {
		tooltipProgress = state.tooltipTransition.ProgressWithContext(ctx, gtx, tooltipVisible)
		if !tooltipVisible && tooltipProgress <= 0 {
			state.tooltipSelection = chartSelection{}
		}
	} else {
		state.tooltipTransition.Reset()
		state.tooltipSelection = chartSelection{}
	}
	tooltipPointer := state.pointer
	if tooltipVisible || tooltipProgress > 0 {
		tooltipPointer = state.tooltipTransition.PositionWithContext(ctx, gtx, state.pointer)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeChartText(gtx, yName, yNamePosition)
	drawGrid(gtx, geometry, style, w.showGrid, metrics)
	w.drawMarkAreas(ctx, gtx, geometry, visual, style)
	drawAxes(gtx, geometry, style, metrics)
	drawCandles(gtx, displayData, geometry, selection.index, selected, metrics)
	w.drawMarkLinesAndPoints(ctx, gtx, geometry, visual, style)
	w.layoutAxisLabels(ctx, gtx, geometry, visual, style)
	if selected && w.showTooltip && w.showCrosshair {
		w.drawCrosshair(ctx, gtx, geometry, selection, state.pointer, style, visual)
	}
	if !data.extent.Valid {
		w.layoutEmpty(ctx, gtx, geometry, visual, style)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, state.tooltipSelection, geometry.yScale.Interval, visual, chart.TooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting())
	}
	if xName.dims.Size.X > 0 {
		placeChartText(gtx, xName, image.Pt(max(geometry.plot.Max.X-xName.dims.Size.X, 0), max(size.Y-xName.dims.Size.Y, 0)))
	}
	opacity.Pop()
	state.addPointerInput(gtx, geometry.plot, enabled && (selectionEnabled || w.onDataWindowChange != nil))
}

func (w Widget) resolveGeometry(data chartData, size image.Point, plot image.Rectangle, scale chart.LinearScale, dp func(unit.Dp) int) chartGeometry {
	geometry := chartGeometry{size: size, plot: plot, yScale: scale}
	for _, value := range scale.Ticks {
		geometry.yTicks = append(geometry.yTicks, axisTick{value: value, label: w.yLabel(value, scale.Interval), pixel: geometry.mapY(value)})
	}
	if len(data.candles) > 0 {
		geometry.categoryStart, geometry.categoryEnd = chart.VisibleCategoryRange(len(data.candles), w.effectiveDataWindow())
		visible := geometry.categoryEnd - geometry.categoryStart
		geometry.bandWidth = float32(plot.Dx()) / float32(visible)
		geometry.candleWidth = w.resolveCandleWidth(geometry.bandWidth, dp)
		timeFormat := w.timeAxisFormat(geometry.categoryStart, geometry.categoryEnd)
		geometry.timeFormat = timeFormat
		for index := geometry.categoryStart; index < geometry.categoryEnd; index++ {
			geometry.xTicks = append(geometry.xTicks, categoryTick{index: index, pixel: geometry.categoryCenter(index)})
		}
	}
	return geometry
}

func (w Widget) resolveYScale(data chartData) chart.LinearScale {
	minimum, maximum := data.extent.Minimum, data.extent.Maximum
	if !w.hasYRange && !w.effectiveDataWindow().IsFull() {
		start, end := chart.VisibleCategoryRange(len(data.candles), w.effectiveDataWindow())
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
	return chart.NewLinearScale(minimum, maximum, w.yTickCount, false, w.hasYRange)
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
	return float32(g.plot.Max.Y) - float32(g.yScale.Ratio(value))*float32(g.plot.Dy())
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
	return w.publicSelectionFor(nil, selection)
}

func (w Widget) publicSelectionFor(ctx *core.Context, selection chartSelection) chart.Selection {
	candle := selection.candle
	seriesLabel := w.label
	if seriesLabel == "" {
		seriesLabel = ctx.Text(i18n.ChartCandlestickDefaultLabel)
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

func (w Widget) semanticDescription(ctx *core.Context, data chartData) string {
	label := w.label
	if label == "" {
		label = ctx.Text(i18n.ChartCandlestickDefaultLabel)
	}
	return ctx.Textf(i18n.ChartCandlestickDescription, label, len(data.candles))
}
