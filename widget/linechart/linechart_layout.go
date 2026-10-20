package linechart

import (
	"image"
	"math"

	"gioui.org/f32"
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

type chartGeometry struct {
	size   image.Point
	plot   image.Rectangle
	xScale chart.LinearScale
	yScale chart.LinearScale
	xTicks []axisTick
	yTicks []axisTick
}

type geometryKey struct {
	generation      uint64
	themeGeneration uint64
	metric          unit.Metric
	size            image.Point
	plot            image.Rectangle
	window          chart.DataWindow
	xTickCount      int
	yTickCount      int
	includeZero     bool
	xMin            float64
	xMax            float64
	hasXRange       bool
	yMin            float64
	yMax            float64
	hasYRange       bool
	categoryCount   int
	formatX         uintptr
	formatY         uintptr
}

type chartSelectionEntry struct {
	series resolvedSeries
	point  resolvedPoint
	pixel  f32.Point
}

type chartSelection struct {
	x       float64
	pixelX  float32
	entries []chartSelectionEntry
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, visual theme.LineChartVisual) layout.Dimensions {
	if w.animationDuration <= 0 {
		w.animationDuration = visual.Metrics.AnimationDuration
	}
	if w.updateAnimationDuration <= 0 {
		w.updateAnimationDuration = visual.Metrics.UpdateDuration
	}

	state := chartStateFor(ctx, w.key)
	state.beginLegendFrame()
	defer state.endLegendFrame()
	restoreKey := core.PushKey(ctx, w.key)
	defer restoreKey()
	enabled := gtx.Enabled() && !w.disabled
	activated, resetWindow := state.updateClicks(gtx, enabled)

	metrics := visual.Metrics
	height := metrics.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := state.dataCache.resolveWithVisual(w, core.ActiveTheme(ctx), visual, gtx.Metric)
	if w.handleLegendClicks(gtx, state, data, enabled) {
		activated = false
		resetWindow = false
	}
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

func (w Widget) layoutContent(ctx *core.Context, gtx layout.Context, state *chartState, data chartData, visual theme.LineChartVisual, enabled, activated bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	metrics := visual.Metrics
	chartStyle := lineChartStyleFor(visual, !enabled)
	displayData := w.animatedData(ctx, gtx, state, data)
	left := max(gtx.Dp(metrics.PlotPaddingLeft), w.measureYAxisLabelWidth(ctx, gtx, data, metrics, chartStyle.axisLabel, max(size.X/2, 1))+max(gtx.Dp(metrics.TickLabelGap), 0)+4)
	left = min(max(left, 0), max(size.X/2, 0))
	right := min(max(gtx.Dp(metrics.PlotPaddingRight), 0), max(size.X-left-1, 0))
	availableWidth := max(size.X-left-right, 0)

	legend := recordedChartBlock{}
	if w.legendVisible(data) {
		legend = w.recordLegend(ctx, gtx, state, data, chartStyle, metrics, visual.LegendHiddenOpacity, availableWidth, enabled)
	}
	yName := recordChartText(ctx, gtx, w.yAxisLabel, metrics.AxisTextSize, metrics.EmphasisWeight, chartStyle.axisLabel, availableWidth)
	xName := recordChartText(ctx, gtx, w.xAxisLabel, metrics.AxisTextSize, metrics.EmphasisWeight, chartStyle.axisLabel, availableWidth)

	top := max(gtx.Dp(metrics.PlotPaddingTop), 0)
	legendPosition := image.Pt(left, top)
	if legend.dims.Size.Y > 0 {
		top += legend.dims.Size.Y + max(gtx.Dp(metrics.LegendGap), 0)
	}
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
		xTickCount:      w.xTickCount,
		yTickCount:      w.yTickCount,
		includeZero:     w.includeZero,
		xMin:            w.xMin,
		xMax:            w.xMax,
		hasXRange:       w.hasXRange,
		yMin:            w.yMin,
		yMax:            w.yMax,
		hasYRange:       w.hasYRange,
		categoryCount:   len(w.categories),
		formatX:         chart.FuncID(w.formatX),
		formatY:         chart.FuncID(w.formatY),
	}, func() chartGeometry {
		return w.resolveGeometry(data, size, plot)
	})
	geometry.xTicks = w.pruneXTicks(ctx, gtx, geometry, chartStyle, metrics)
	visibleX := visibleXValues(data.xValues, geometry.xScale)
	state.updatePointer(gtx, enabled, plot, w.effectiveDataWindow(), w.onDataWindowChange)
	selectionEnabled := w.showTooltip || w.onDataClick != nil
	if !selectionEnabled {
		state.clearSelection()
	}
	selection := chartSelection{}
	selected := false
	if selectionEnabled {
		selectedX, hasSelection := state.selectedX(visibleX, geometry.xScale, plot)
		selected = hasSelection
		selection = w.resolveSelection(data, geometry, selectedX, hasSelection)
	}
	if activated && selected && w.onDataClick != nil {
		w.onDataClick(w.publicSelection(selection, geometry))
	}
	tooltipVisible := enabled && w.showTooltip && len(selection.entries) > 0
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

	opacity := paint.PushOpacity(gtx.Ops, chartStyle.opacity)
	placeChartBlock(gtx, legend, legendPosition)
	placeChartText(gtx, yName, yNamePosition)
	drawChartGrid(gtx, geometry, chartStyle, w.showGrid, metrics)
	w.drawMarkAreas(ctx, gtx, geometry, chartStyle, visual)
	drawChartAxes(gtx, geometry, chartStyle, metrics)
	drawChartSeries(ctx, gtx, displayData, geometry, visual)
	w.drawMarkLinesAndPoints(ctx, gtx, geometry, chartStyle, visual)
	w.layoutAxisLabels(ctx, gtx, geometry, chartStyle, metrics)
	if !data.yExtent.Valid {
		w.layoutEmpty(ctx, gtx, geometry, chartStyle, metrics)
	}
	if w.showTooltip {
		drawChartCrosshair(gtx, geometry, selection, chartStyle, metrics)
		drawChartSelection(gtx, selection, metrics, chartStyle.background)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, geometry, state.tooltipSelection, chart.TooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting(), visual)
	}
	if xName.dims.Size.X > 0 {
		position := image.Pt(max(geometry.plot.Max.X-xName.dims.Size.X, 0), max(size.Y-xName.dims.Size.Y, 0))
		placeChartText(gtx, xName, position)
	}
	opacity.Pop()
	state.addPointerInput(gtx, plot, enabled && (selectionEnabled || w.onDataWindowChange != nil))
}

func (w Widget) resolveGeometry(data chartData, size image.Point, plot image.Rectangle) chartGeometry {
	xScale := w.resolveXScale(data)
	yScale := w.resolveYScale(data)
	geometry := chartGeometry{size: size, plot: plot, xScale: xScale, yScale: yScale}
	geometry.xTicks = w.resolveXTicks(geometry)
	geometry.yTicks = make([]axisTick, 0, len(yScale.Ticks))
	for _, value := range yScale.Ticks {
		geometry.yTicks = append(geometry.yTicks, axisTick{
			value: value,
			label: w.yLabel(value, yScale.Interval),
			pixel: geometry.mapY(value),
		})
	}
	return geometry
}

func (w Widget) resolveXScale(data chartData) chart.LinearScale {
	minimum, maximum := data.xExtent.Minimum, data.xExtent.Maximum
	if w.hasXRange {
		minimum, maximum = w.xMin, w.xMax
	}
	window := w.effectiveDataWindow()
	if maximum > minimum && !window.IsFull() {
		span := maximum - minimum
		base := minimum
		minimum = base + float64(window.Start)*span
		maximum = base + float64(window.End)*span
	}
	fixed := w.hasXRange || len(w.categories) > 0 || !window.IsFull()
	return chart.NewLinearScale(minimum, maximum, w.xTickCount, false, fixed)
}

func (w Widget) resolveYScale(data chartData) chart.LinearScale {
	yMinimum, yMaximum := data.yExtent.Minimum, data.yExtent.Maximum
	if w.hasYRange {
		yMinimum, yMaximum = w.yMin, w.yMax
	}
	return chart.NewLinearScale(yMinimum, yMaximum, w.yTickCount, w.includeZero && !w.hasYRange, w.hasYRange)
}

func (w Widget) resolveXTicks(geometry chartGeometry) []axisTick {
	if len(w.categories) == 0 {
		ticks := make([]axisTick, 0, len(geometry.xScale.Ticks))
		for _, value := range geometry.xScale.Ticks {
			ticks = append(ticks, axisTick{value: value, label: w.xLabel(value, geometry.xScale.Interval), pixel: geometry.mapX(value)})
		}
		return ticks
	}
	maximumLabels := max(geometry.plot.Dx()/64, 2)
	step := max(int(math.Ceil(float64(len(w.categories))/float64(maximumLabels))), 1)
	ticks := make([]axisTick, 0, maximumLabels+1)
	appendIndex := func(index int) {
		value := float64(index)
		if !geometry.xScale.Contains(value) {
			return
		}
		ticks = append(ticks, axisTick{value: value, label: w.xLabel(value, 1), pixel: geometry.mapX(value)})
	}
	for index := 0; index < len(w.categories); index += step {
		appendIndex(index)
	}
	last := len(w.categories) - 1
	if last >= 0 && (len(ticks) == 0 || ticks[len(ticks)-1].value != float64(last)) {
		appendIndex(last)
	}
	return ticks
}

func (g chartGeometry) mapX(value float64) float32 {
	return float32(g.plot.Min.X) + float32(g.xScale.Ratio(value))*float32(g.plot.Dx())
}

func (g chartGeometry) mapY(value float64) float32 {
	return float32(g.plot.Max.Y) - float32(g.yScale.Ratio(value))*float32(g.plot.Dy())
}

func (w Widget) resolveSelection(data chartData, geometry chartGeometry, selectedX float64, selected bool) chartSelection {
	if !selected || !geometry.xScale.Contains(selectedX) {
		return chartSelection{}
	}
	selection := chartSelection{x: selectedX, pixelX: geometry.mapX(selectedX)}
	for _, series := range data.series {
		for _, point := range series.points {
			if !point.valid || point.X != selectedX || !geometry.yScale.Contains(point.Y) {
				continue
			}
			selection.entries = append(selection.entries, chartSelectionEntry{
				series: series,
				point:  point,
				pixel:  f32.Pt(selection.pixelX, geometry.mapY(point.Y)),
			})
			break
		}
	}
	return selection
}

func (w Widget) publicSelection(selection chartSelection, geometry chartGeometry) chart.Selection {
	index := -1
	if len(w.categories) > 0 {
		candidate := int(math.Round(selection.x))
		if candidate >= 0 && candidate < len(w.categories) && math.Abs(selection.x-float64(candidate)) < 1e-9 {
			index = candidate
		}
	}
	result := chart.Selection{
		Label: w.xLabel(selection.x, geometry.xScale.Interval),
		Index: index,
		X:     selection.x,
		Items: make([]chart.Datum, 0, len(selection.entries)),
	}
	for _, entry := range selection.entries {
		result.Items = append(result.Items, chart.Datum{
			SeriesKey:   entry.series.key,
			SeriesLabel: entry.series.label,
			X:           entry.point.X,
			Y:           entry.point.rawY,
			Color:       entry.series.color,
		})
	}
	return result
}

func (w Widget) semanticDescription(ctx *core.Context, data chartData) string {
	label := w.label
	if label == "" {
		label = ctx.Text(i18n.ChartLineDefaultLabel)
	}
	return ctx.Textf(i18n.ChartLineDescription, label, len(data.series))
}
