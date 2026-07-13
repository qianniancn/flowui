package barchart

import (
	"fmt"
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/paint"
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
	yScale        linearScale
	yTicks        []axisTick
	xTicks        []categoryTick
	bandWidth     float32
	columnLayouts map[string]columnLayout
}

type chartSelectionEntry struct {
	series resolvedSeries
	bar    resolvedBar
}

type chartSelection struct {
	index   int
	pixelX  float32
	entries []chartSelectionEntry
}

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := chartStateFor(ctx, w.key)
	enabled := gtx.Enabled() && !w.disabled
	state.requestPointerFocus(ctx, gtx, enabled)

	tokens := frame.ActiveTheme(ctx).Components.BarChart
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := resolveChartData(w, frame.ActiveTheme(ctx), gtx.Dp)

	eventGtx := gtx
	if !enabled {
		eventGtx = eventGtx.Disabled()
	}
	eventGtx.Constraints = layout.Exact(size)
	return state.root.Layout(eventGtx, func(gtx layout.Context) layout.Dimensions {
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		semantic.DescriptionOp(w.semanticDescription(data)).Add(gtx.Ops)
		w.layoutContent(ctx, gtx, state, data, enabled, size)
		return layout.Dimensions{Size: size}
	})
}

func (w Widget) layoutContent(ctx *frame.Context, gtx layout.Context, state *chartState, data chartData, enabled bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.BarChart
	style := barChartStyleFor(activeTheme, !enabled)
	left := max(gtx.Dp(tokens.PlotPaddingLeft), w.measureYAxisLabelWidth(ctx, gtx, data, max(size.X/2, 1))+max(gtx.Dp(tokens.TickLabelGap), 0)+4)
	left = min(max(left, 0), max(size.X/2, 0))
	right := min(max(gtx.Dp(tokens.PlotPaddingRight), 0), max(size.X-left-1, 0))
	availableWidth := max(size.X-left-right, 0)

	legend := recordedChartBlock{}
	if w.legendVisible(data) {
		legend = w.recordLegend(ctx, gtx, data, style, availableWidth)
	}
	yName := recordChartText(ctx, gtx, w.yAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)
	xName := recordChartText(ctx, gtx, w.xAxisLabel, tokens.AxisTextSize, font.Medium, style.axisLabel, availableWidth)

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
	geometry.xTicks = w.pruneXTicks(ctx, gtx, geometry, style)
	state.updatePointer(gtx, enabled)
	if w.showTooltip {
		state.updateKeyboard(ctx, gtx, data.categories, enabled)
	} else {
		state.clearSelection()
	}
	focused := gtx.Focused(&state.root)
	selection := chartSelection{}
	selected := false
	if w.showTooltip {
		selectedIndex, hasSelection := state.selectedIndex(data.categories, plot, focused)
		selected = hasSelection
		selection = w.resolveSelection(data, geometry, selectedIndex, hasSelection)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeChartBlock(gtx, legend, legendPosition)
	placeChartText(gtx, yName, yNamePosition)
	drawChartGrid(gtx, geometry, style, w.showGrid, tokens)
	drawChartAxes(gtx, geometry, style, tokens)
	if selected {
		drawCategoryHighlight(gtx, geometry, selection, style)
	}
	drawChartSeries(ctx, gtx, data, geometry, style, tokens)
	w.layoutAxisLabels(ctx, gtx, geometry, style)
	if !data.yExtent.valid {
		w.layoutEmpty(ctx, gtx, geometry, style)
	}
	if w.showTooltip && selected {
		w.drawTooltip(ctx, gtx, geometry, selection, style)
	}
	if xName.dims.Size.X > 0 {
		position := image.Pt(max(geometry.plot.Max.X-xName.dims.Size.X, 0), max(size.Y-xName.dims.Size.Y, 0))
		placeChartText(gtx, xName, position)
	}
	opacity.Pop()

	focusVisible := frame.FocusVisible(ctx, &state.root, focused)
	focusOpacity := state.focus.Opacity(gtx, focusVisible && enabled)
	drawChartFocus(gtx, size, style.focus, focusOpacity, tokens)
	state.addPointerInput(gtx, plot, enabled && w.showTooltip)
}

func (w Widget) resolveGeometry(data chartData, size image.Point, plot image.Rectangle) chartGeometry {
	yScale := w.resolveYScale(data)
	geometry := chartGeometry{size: size, plot: plot, yScale: yScale}
	geometry.yTicks = make([]axisTick, 0, len(yScale.ticks))
	for _, value := range yScale.ticks {
		geometry.yTicks = append(geometry.yTicks, axisTick{value: value, label: w.yLabel(value, yScale.interval), pixel: geometry.mapY(value)})
	}
	if data.categories > 0 {
		geometry.bandWidth = float32(plot.Dx()) / float32(data.categories)
		geometry.xTicks = make([]categoryTick, 0, data.categories)
		for index := 0; index < data.categories; index++ {
			geometry.xTicks = append(geometry.xTicks, categoryTick{index: index, label: w.categoryLabel(index), pixel: geometry.categoryCenter(index)})
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

func (w Widget) resolveYScale(data chartData) linearScale {
	minimum, maximum := data.yExtent.minimum, data.yExtent.maximum
	if w.hasYRange {
		minimum, maximum = w.yMin, w.yMax
	}
	return newLinearScale(minimum, maximum, w.yTickCount, w.includeZero && !w.hasYRange, w.hasYRange)
}

func (g chartGeometry) categoryCenter(index int) float32 {
	return float32(g.plot.Min.X) + (float32(index)+0.5)*g.bandWidth
}

func (g chartGeometry) mapY(value float64) float32 {
	return float32(g.plot.Max.Y) - float32(g.yScale.ratio(value))*float32(g.plot.Dy())
}

func (w Widget) resolveSelection(data chartData, geometry chartGeometry, index int, selected bool) chartSelection {
	if !selected || index < 0 || index >= data.categories {
		return chartSelection{}
	}
	selection := chartSelection{index: index, pixelX: geometry.categoryCenter(index)}
	for _, series := range data.series {
		bar := series.values[index]
		if bar.valid {
			selection.entries = append(selection.entries, chartSelectionEntry{series: series, bar: bar})
		}
	}
	return selection
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
