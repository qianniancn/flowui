package barchart

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/paint"
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

type columnLayout struct {
	offset float32
	width  float32
}

type chartGeometry struct {
	size          image.Point
	plot          image.Rectangle
	yScale        chart.LinearScale
	yTicks        []axisTick
	xTicks        []categoryTick
	bandWidth     float32
	categoryStart int
	categoryEnd   int
	horizontal    bool
	columnLayouts map[string]columnLayout
}

type chartSelectionEntry struct {
	series resolvedSeries
	bar    resolvedBar
}

type chartSelection struct {
	index   int
	pixelX  float32
	pixelY  float32
	entries []chartSelectionEntry
}

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := chartStateFor(ctx, w.key)
	state.beginLegendFrame()
	defer state.endLegendFrame()
	restoreKey := frame.PushKey(ctx, w.key)
	defer restoreKey()
	enabled := gtx.Enabled() && !w.disabled
	activated, resetWindow := state.updateClicks(gtx, enabled)

	tokens := frame.ActiveTheme(ctx).Components.BarChart
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := state.dataCache.resolve(w, frame.ActiveTheme(ctx), gtx.Metric)
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
	tokens := activeTheme.Components.BarChart
	style := barChartStyleFor(activeTheme, !enabled)
	displayData := w.animatedData(ctx, gtx, state, data)
	leftLabelWidth := w.measureYAxisLabelWidth(ctx, gtx, data, max(size.X/2, 1))
	if w.orientation == Horizontal {
		leftLabelWidth = w.measureCategoryLabelWidth(ctx, gtx, data, max(size.X/2, 1))
	}
	left := max(gtx.Dp(tokens.PlotPaddingLeft), leftLabelWidth+max(gtx.Dp(tokens.TickLabelGap), 0)+4)
	left = min(max(left, 0), max(size.X/2, 0))
	right := min(max(gtx.Dp(tokens.PlotPaddingRight), 0), max(size.X-left-1, 0))
	availableWidth := max(size.X-left-right, 0)

	legend := recordedChartBlock{}
	if w.legendVisible(data) {
		legend = w.recordLegend(ctx, gtx, state, data, style, availableWidth, enabled)
	}
	xAxisLabel, yAxisLabel := w.axisLabels()
	yName := recordChartText(ctx, gtx, yAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)
	xName := recordChartText(ctx, gtx, xAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)

	top := max(gtx.Dp(tokens.PlotPaddingTop), 0)
	legendPosition := image.Pt(left, top)
	if legend.dims.Size.Y > 0 {
		top += legend.dims.Size.Y + max(gtx.Dp(tokens.LegendGap), 0)
	}
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
	plot := image.Rect(plotLeft, plotTop, plotRight, plotBottom)

	geometry := w.resolveGeometry(data, size, plot)
	if geometry.horizontal {
		geometry.xTicks = w.pruneCategoryTicks(ctx, gtx, geometry, style)
	} else {
		geometry.xTicks = w.pruneXTicks(ctx, gtx, geometry, style)
	}
	state.updatePointer(gtx, enabled, plot, w.effectiveDataWindow(), geometry.horizontal, w.onDataWindowChange)
	selectionEnabled := w.showTooltip || w.onDataClick != nil
	if !selectionEnabled {
		state.clearSelection()
	}
	selection := chartSelection{}
	selected := false
	if selectionEnabled {
		selectedIndex, hasSelection := state.selectedIndex(geometry.categoryStart, geometry.categoryEnd, plot, geometry.horizontal)
		selected = hasSelection
		selection = w.resolveSelection(data, geometry, selectedIndex, hasSelection)
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
		tooltipProgress = state.tooltipTransition.Progress(gtx, tooltipVisible, frame.ActiveTheme(ctx).Motion)
		if !tooltipVisible && tooltipProgress <= 0 {
			state.tooltipSelection = chartSelection{}
		}
	} else {
		state.tooltipTransition.Reset()
		state.tooltipSelection = chartSelection{}
	}
	tooltipPointer := state.pointer
	if tooltipVisible || tooltipProgress > 0 {
		tooltipPointer = state.tooltipTransition.Position(gtx, state.pointer, frame.ActiveTheme(ctx).Motion)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeChartBlock(gtx, legend, legendPosition)
	placeChartText(gtx, yName, yNamePosition)
	drawChartGrid(gtx, geometry, style, w.showGrid, tokens)
	w.drawMarkAreas(ctx, gtx, geometry, style)
	drawChartAxes(gtx, geometry, style, tokens)
	if selected {
		drawCategoryHighlight(gtx, geometry, selection, style)
	}
	drawChartSeries(ctx, gtx, displayData, geometry, style, tokens)
	w.drawMarkLinesAndPoints(ctx, gtx, geometry, style)
	w.layoutAxisLabels(ctx, gtx, geometry, style)
	if !data.yExtent.Valid {
		w.layoutEmpty(ctx, gtx, geometry, style)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, gtx, geometry, state.tooltipSelection, chart.TooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting())
	}
	if xName.dims.Size.X > 0 {
		position := image.Pt(max(geometry.plot.Max.X-xName.dims.Size.X, 0), max(size.Y-xName.dims.Size.Y, 0))
		placeChartText(gtx, xName, position)
	}
	opacity.Pop()
	state.addPointerInput(gtx, plot, enabled && (selectionEnabled || w.onDataWindowChange != nil))
}

func (w Widget) axisLabels() (x, y string) {
	x, y = w.xAxisLabel, w.yAxisLabel
	if w.orientation == Horizontal {
		if w.hasCategoryAxisLabel {
			y = w.categoryAxisLabel
		}
		if w.hasValueAxisLabel {
			x = w.valueAxisLabel
		}
		return x, y
	}
	if w.hasCategoryAxisLabel {
		x = w.categoryAxisLabel
	}
	if w.hasValueAxisLabel {
		y = w.valueAxisLabel
	}
	return x, y
}

func (w Widget) resolveGeometry(data chartData, size image.Point, plot image.Rectangle) chartGeometry {
	yScale := w.resolveYScale(data)
	geometry := chartGeometry{size: size, plot: plot, yScale: yScale, horizontal: w.orientation == Horizontal}
	geometry.yTicks = make([]axisTick, 0, len(yScale.Ticks))
	for _, value := range yScale.Ticks {
		geometry.yTicks = append(geometry.yTicks, axisTick{value: value, label: w.yLabel(value, yScale.Interval), pixel: geometry.mapY(value)})
	}
	if data.categories > 0 {
		geometry.categoryStart, geometry.categoryEnd = chart.VisibleCategoryRange(data.categories, w.effectiveDataWindow())
		visibleCount := geometry.categoryEnd - geometry.categoryStart
		bandExtent := plot.Dx()
		if geometry.horizontal {
			bandExtent = plot.Dy()
		}
		geometry.bandWidth = float32(bandExtent) / float32(visibleCount)
		geometry.xTicks = make([]categoryTick, 0, visibleCount)
		for index := geometry.categoryStart; index < geometry.categoryEnd; index++ {
			geometry.xTicks = append(geometry.xTicks, categoryTick{index: index, pixel: geometry.categoryCenter(index)})
		}
	}
	barGap := float32(0.1)
	categoryGap := defaultCategoryGap(len(data.columns))
	if w.hasBarGap {
		barGap = w.barGap
	}
	if w.hasCategoryGap {
		categoryGap = w.categoryGap
	}
	geometry.columnLayouts = resolveColumnLayouts(data.columns, geometry.bandWidth, barGap, categoryGap)
	return geometry
}

func (w Widget) resolveYScale(data chartData) chart.LinearScale {
	minimum, maximum := data.yExtent.Minimum, data.yExtent.Maximum
	if w.hasYRange {
		minimum, maximum = w.yMin, w.yMax
	}
	return chart.NewLinearScale(minimum, maximum, w.yTickCount, w.includeZero && !w.hasYRange, w.hasYRange)
}

func (g chartGeometry) categoryCenter(index int) float32 {
	minimum := g.plot.Min.X
	if g.horizontal {
		minimum = g.plot.Min.Y
	}
	return float32(minimum) + (float32(index-g.categoryStart)+0.5)*g.bandWidth
}

func (g chartGeometry) mapY(value float64) float32 {
	if g.horizontal {
		return float32(g.plot.Min.X) + float32(g.yScale.Ratio(value))*float32(g.plot.Dx())
	}
	return float32(g.plot.Max.Y) - float32(g.yScale.Ratio(value))*float32(g.plot.Dy())
}

func (w Widget) resolveSelection(data chartData, geometry chartGeometry, index int, selected bool) chartSelection {
	if !selected || index < geometry.categoryStart || index >= geometry.categoryEnd || index >= data.categories {
		return chartSelection{}
	}
	selection := chartSelection{index: index}
	if geometry.horizontal {
		selection.pixelX = float32(geometry.plot.Min.X) + float32(geometry.plot.Dx())/2
		selection.pixelY = geometry.categoryCenter(index)
	} else {
		selection.pixelX = geometry.categoryCenter(index)
		selection.pixelY = float32(geometry.plot.Min.Y) + float32(geometry.plot.Dy())/2
	}
	for _, series := range data.series {
		bar := series.values[index]
		if bar.valid {
			selection.entries = append(selection.entries, chartSelectionEntry{series: series, bar: bar})
		}
	}
	return selection
}

func (w Widget) publicSelection(selection chartSelection, _ chartGeometry) chart.Selection {
	result := chart.Selection{
		Label: w.categoryLabel(selection.index),
		Index: selection.index,
		X:     float64(selection.index),
		Items: make([]chart.Datum, 0, len(selection.entries)),
	}
	for _, entry := range selection.entries {
		result.Items = append(result.Items, chart.Datum{
			SeriesKey:   entry.series.key,
			SeriesLabel: entry.series.label,
			X:           float64(selection.index),
			Y:           entry.bar.value,
			Color:       entry.bar.color,
		})
	}
	return result
}

func (w Widget) semanticDescription(data chartData) string {
	label := w.label
	if label == "" {
		label = "Bar chart"
	}
	return fmt.Sprintf("%s, %d series, %d categories", label, len(data.series), data.categories)
}

func defaultCategoryGap(columnCount int) float32 {
	return float32(max(35-columnCount*4, 15)) / 100
}

func resolveColumnLayouts(columns []barColumn, bandWidth, barGap, categoryGap float32) map[string]columnLayout {
	result := make(map[string]columnLayout, len(columns))
	if len(columns) == 0 || bandWidth <= 0 {
		return result
	}
	widths := make([]float32, len(columns))
	fixed := make([]bool, len(columns))
	for index, column := range columns {
		if column.width > 0 {
			widths[index] = max(column.width, 1)
			if column.maxWidth > 0 {
				widths[index] = min(widths[index], column.maxWidth)
			}
			fixed[index] = true
		}
	}
	contentWidth := max(bandWidth*(1-categoryGap), 0)
	for {
		fixedWidth := float32(0)
		autoFactor := float32(0)
		for index, width := range widths {
			factor := float32(1)
			if index < len(widths)-1 {
				factor += barGap
			}
			if fixed[index] {
				fixedWidth += width * factor
			} else {
				autoFactor += factor
			}
		}
		if autoFactor == 0 {
			break
		}
		autoWidth := max((contentWidth-fixedWidth)/autoFactor, 0)
		constrained := false
		for index, column := range columns {
			if fixed[index] || column.maxWidth <= 0 || autoWidth <= column.maxWidth {
				continue
			}
			widths[index] = max(column.maxWidth, 1)
			fixed[index] = true
			constrained = true
		}
		if constrained {
			continue
		}
		for index := range widths {
			if !fixed[index] {
				widths[index] = max(autoWidth, 1)
			}
		}
		break
	}
	widthSum := float32(0)
	for index, width := range widths {
		widthSum += width
		if index < len(widths)-1 {
			widthSum += width * barGap
		}
	}
	offset := -widthSum / 2
	for index, column := range columns {
		result[column.id] = columnLayout{offset: offset, width: widths[index]}
		offset += widths[index] * (1 + barGap)
	}
	return result
}
