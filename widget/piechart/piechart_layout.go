package piechart

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

type chartGeometry struct {
	size        image.Point
	area        image.Rectangle
	center      f32.Point
	innerRadius float32
	outerRadius float32
}

type geometryKey struct {
	generation      uint64
	themeGeneration uint64
	metric          unit.Metric
	size            image.Point
	area            image.Rectangle
	innerRadius     float32
	outerRadius     float32
}

type chartStyle struct {
	text      color.NRGBA
	mutedText color.NRGBA
	empty     color.NRGBA
	opacity   float32
}

type recordedText struct {
	call op.CallOp
	dims layout.Dimensions
}

type recordedBlock struct {
	call op.CallOp
	dims layout.Dimensions
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, visual theme.PieChartVisual) layout.Dimensions {
	state := chartStateFor(ctx, w.key)
	state.beginLegendFrame()
	defer state.endLegendFrame()
	restoreKey := core.PushKey(ctx, w.key)
	defer restoreKey()

	enabled := gtx.Enabled() && !w.disabled
	activated := state.updateClick(gtx, enabled)
	active := core.ActiveTheme(ctx)
	metrics := visual.Metrics
	height := metrics.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := state.dataCache.resolve(w, active, visual)
	if w.handleLegendClicks(gtx, state, data, enabled) {
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
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutContent(ctx *core.Context, gtx layout.Context, state *chartState, data chartData, visual theme.PieChartVisual, enabled, activated bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	metrics := visual.Metrics
	style := pieChartStyle(visual, enabled)
	left := min(max(gtx.Dp(metrics.PlotPaddingLeft), 0), size.X/2)
	right := min(max(gtx.Dp(metrics.PlotPaddingRight), 0), size.X-left)
	availableWidth := max(size.X-left-right, 0)

	legend := recordedBlock{}
	if w.legendVisible(data) {
		legend = w.recordLegend(ctx, gtx, state, data, visual, availableWidth, enabled)
	}
	top := max(gtx.Dp(metrics.PlotPaddingTop), 0)
	legendPosition := image.Pt(left+max((availableWidth-legend.dims.Size.X)/2, 0), top)
	if legend.dims.Size.Y > 0 {
		top += legend.dims.Size.Y + max(gtx.Dp(metrics.LegendGap), 0)
	}
	bottom := max(gtx.Dp(metrics.PlotPaddingBottom), 0)
	area := image.Rect(left, top, max(size.X-right, left), max(size.Y-bottom, top))
	geometry := state.geometryCache.Resolve(data.generation != 0, geometryKey{
		generation:      data.generation,
		themeGeneration: core.ThemeGeneration(ctx),
		metric:          gtx.Metric,
		size:            size,
		area:            area,
		innerRadius:     w.innerRadius,
		outerRadius:     w.outerRadius,
	}, func() chartGeometry {
		return w.resolveGeometry(size, area)
	})
	displayData := w.animatedData(ctx, gtx, state, data)

	selectionEnabled := w.showTooltip || w.onDataClick != nil
	state.updatePointer(gtx, enabled && selectionEnabled)
	selectedIndex := -1
	if selectionEnabled && state.hovered {
		selectedIndex = hitTestPie(displayData, geometry, state.pointer)
	}
	if activated && selectedIndex >= 0 && w.onDataClick != nil {
		w.onDataClick(w.publicSelection(displayData.slices[selectedIndex]))
	}

	tooltipVisible := enabled && w.showTooltip && selectedIndex >= 0
	if tooltipVisible {
		state.tooltipSlice = displayData.slices[selectedIndex]
	}
	tooltipProgress := float32(0)
	if enabled && w.showTooltip {
		tooltipProgress = state.tooltipTransition.ProgressWithContext(ctx, gtx, tooltipVisible)
		if !tooltipVisible && tooltipProgress <= 0 {
			state.tooltipSlice = resolvedSlice{}
		}
	} else {
		state.clearTooltip()
	}
	tooltipPointer := state.pointer
	if tooltipVisible || tooltipProgress > 0 {
		tooltipPointer = state.tooltipTransition.PositionWithContext(ctx, gtx, state.pointer)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeBlock(gtx, legend, legendPosition)
	hoveredKey := ""
	if selectedIndex >= 0 {
		hoveredKey = displayData.slices[selectedIndex].key
	}
	if hasVisibleSector(data.slices) {
		drawPieSlices(gtx, displayData.slices, geometry, hoveredKey, float32(gtx.Dp(metrics.EmphasisSize)))
		if w.showLabels {
			w.drawLabels(ctx, gtx, data.slices, geometry, style, metrics)
		}
	} else {
		drawEmptyPie(gtx, geometry, style.empty)
		w.drawEmptyText(ctx, gtx, geometry, style, metrics)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, state.tooltipSlice, visual, chart.TooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting())
	}
	opacity.Pop()

	interactionArea := geometry.interactionArea(float32(gtx.Dp(metrics.EmphasisSize)))
	state.addPointerInput(gtx, interactionArea, enabled && selectionEnabled)
	state.addClickInput(gtx, interactionArea, enabled && w.onDataClick != nil)
}

func (w Widget) resolveGeometry(size image.Point, area image.Rectangle) chartGeometry {
	area = area.Intersect(image.Rectangle{Max: size})
	center := f32.Pt(float32(area.Min.X+area.Dx()/2), float32(area.Min.Y+area.Dy()/2))
	availableRadius := float32(min(area.Dx(), area.Dy())) / 2
	return chartGeometry{
		size:        size,
		area:        area,
		center:      center,
		innerRadius: availableRadius * w.innerRadius,
		outerRadius: availableRadius * w.outerRadius,
	}
}

func (g chartGeometry) interactionArea(extra float32) image.Rectangle {
	radius := max(g.outerRadius+max(extra, 0), 0)
	return image.Rect(
		int(math.Floor(float64(g.center.X-radius))),
		int(math.Floor(float64(g.center.Y-radius))),
		int(math.Ceil(float64(g.center.X+radius))),
		int(math.Ceil(float64(g.center.Y+radius))),
	).Intersect(image.Rectangle{Max: g.size})
}

func hitTestPie(data chartData, geometry chartGeometry, point f32.Point) int {
	delta := point.Sub(geometry.center)
	radius := float32(math.Hypot(float64(delta.X), float64(delta.Y)))
	if radius < geometry.innerRadius || radius > geometry.outerRadius {
		return -1
	}
	angle := float32(math.Atan2(float64(delta.Y), float64(delta.X)))
	for index, slice := range data.slices {
		if slice.sweep() <= 1e-5 || slice.radiusRatio <= 0 || radius > sliceOuterRadius(slice, geometry) {
			continue
		}
		distance := normalizeAngle((angle - slice.startAngle) * data.dir)
		if distance <= slice.sweep()+1e-5 {
			return index
		}
	}
	return -1
}

func normalizeAngle(value float32) float32 {
	result := float32(math.Mod(float64(value), fullCircle))
	if result < 0 {
		result += float32(fullCircle)
	}
	return result
}

func (w Widget) publicSelection(slice resolvedSlice) chart.Selection {
	return chart.Selection{
		Label: slice.label,
		Index: slice.index,
		X:     float64(slice.index),
		Items: []chart.Datum{{
			SeriesKey:   slice.key,
			SeriesLabel: slice.label,
			X:           float64(slice.index),
			Y:           slice.value,
			Percent:     slice.percent,
			Color:       slice.color,
		}},
	}
}

func (w Widget) semanticDescription(ctx *core.Context, data chartData) string {
	label := w.label
	if label == "" {
		label = ctx.Text(i18n.ChartPieDefaultLabel)
	}
	key := i18n.ChartPieDescriptionMany
	if len(data.slices) == 1 {
		key = i18n.ChartPieDescriptionOne
	}
	return ctx.Textf(key, label, len(data.slices))
}

func (w Widget) handleLegendClicks(gtx layout.Context, state *chartState, data chartData, enabled bool) bool {
	activated := false
	for _, slice := range data.legend {
		for state.legendItem(slice.key).Clicked(gtx) {
			activated = true
			if enabled && w.onLegendChange != nil {
				w.onLegendChange(slice.key, !slice.hidden)
			}
		}
	}
	return activated
}

func (w Widget) recordLegend(ctx *core.Context, gtx layout.Context, state *chartState, data chartData, visual theme.PieChartVisual, maxWidth int, enabled bool) recordedBlock {
	if maxWidth <= 0 || len(data.legend) == 0 {
		return recordedBlock{}
	}
	metrics := visual.Metrics
	chartStyle := pieChartStyle(visual, enabled)
	markerSize := max(gtx.Dp(metrics.LegendMarkerSize), 2)
	markerGap := max(gtx.Dp(metrics.LegendMarkerGap), 0)
	itemGap := max(gtx.Dp(metrics.LegendItemGap), 0)
	lineGap := max(gtx.Dp(metrics.LegendLineGap), 0)

	macro := op.Record(gtx.Ops)
	x, y, rowHeight, usedWidth := 0, 0, 0, 0
	for _, slice := range data.legend {
		label := recordText(ctx, gtx, slice.label, metrics.LegendTextSize, metrics.TextWeight, chartStyle.text, max(maxWidth-markerSize-markerGap, 1))
		itemWidth := markerSize + markerGap + label.dims.Size.X
		itemHeight := max(markerSize, label.dims.Size.Y)
		if x > 0 && x+itemWidth > maxWidth {
			y += rowHeight + lineGap
			x, rowHeight = 0, 0
		}
		item := state.legendItem(slice.key)
		itemGtx := gtx
		if !enabled || w.onLegendChange == nil {
			itemGtx = itemGtx.Disabled()
		}
		itemGtx.Constraints = layout.Exact(image.Pt(itemWidth, itemHeight))
		itemMacro := op.Record(gtx.Ops)
		item.Layout(itemGtx, func(gtx layout.Context) layout.Dimensions {
			if enabled && w.onLegendChange != nil {
				pointer.CursorPointer.Add(gtx.Ops)
				if item.Hovered() {
					paint.FillShape(gtx.Ops, visual.Colors.LegendHover, clip.UniformRRect(image.Rectangle{Max: image.Pt(itemWidth, itemHeight)}, min(itemHeight/2, 4)).Op(gtx.Ops))
				}
			}
			itemOpacity := float32(1)
			if slice.hidden {
				itemOpacity = visual.LegendHiddenOpacity
			}
			fade := paint.PushOpacity(gtx.Ops, itemOpacity)
			markerRect := image.Rect(0, (itemHeight-markerSize)/2, markerSize, (itemHeight-markerSize)/2+markerSize)
			radius := min(max(gtx.Dp(metrics.LegendMarkerRadius), 0), markerSize/2)
			paint.FillShape(gtx.Ops, slice.color, clip.UniformRRect(markerRect, radius).Op(gtx.Ops))
			placeText(gtx, label, image.Pt(markerSize+markerGap, (itemHeight-label.dims.Size.Y)/2))
			fade.Pop()
			return layout.Dimensions{Size: image.Pt(itemWidth, itemHeight)}
		})
		itemCall := itemMacro.Stop()
		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		itemCall.Add(gtx.Ops)
		offset.Pop()
		x += itemWidth + itemGap
		usedWidth = max(usedWidth, min(x-itemGap, maxWidth))
		rowHeight = max(rowHeight, itemHeight)
	}
	return recordedBlock{call: macro.Stop(), dims: layout.Dimensions{Size: image.Pt(usedWidth, y+rowHeight)}}
}

func (w Widget) drawEmptyText(ctx *core.Context, gtx layout.Context, geometry chartGeometry, style chartStyle, metrics theme.PieChartMetrics) {
	value := w.emptyText
	if !w.hasEmptyText {
		value = ctx.Text(i18n.ChartNoData)
	}
	label := recordText(ctx, gtx, value, metrics.LabelTextSize, metrics.TextWeight, style.mutedText, max(geometry.area.Dx(), 1))
	position := image.Pt(
		int(geometry.center.X)-label.dims.Size.X/2,
		int(geometry.center.Y)-label.dims.Size.Y/2,
	)
	placeText(gtx, label, position)
}

func (w Widget) drawTooltip(ctx *core.Context, slice resolvedSlice, visual theme.PieChartVisual, anchor image.Rectangle, progress float32, exiting bool) {
	content := w.tooltipWidget(slice, visual)
	if content == nil {
		return
	}
	tooltip.NewPopup(content).
		Placement(core.PopoverRightStart).
		Offset(max(visual.Metrics.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Overlay(ctx, core.FullKey(ctx, "tooltip:"+w.key), anchor)
}

func (w Widget) tooltipWidget(slice resolvedSlice, visual theme.PieChartVisual) core.Widget {
	selection := w.publicSelection(slice)
	if w.tooltipContent != nil {
		return w.tooltipContent(selection)
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		metrics := visual.Metrics
		textColor := visual.Colors.OverlayForeground
		markerSize := max(gtx.Dp(metrics.TooltipMarkerSize), 2)
		gap := max(gtx.Dp(metrics.TooltipRowGap), 0)
		contentWidth := max(gtx.Constraints.Max.X, 1)
		title := recordText(ctx, gtx, slice.label, metrics.TooltipTextSize, metrics.EmphasisWeight, textColor, contentWidth)
		value := fmt.Sprintf("%s  %.2f%%", formatPieNumber(slice.value), slice.percent)
		row := recordText(ctx, gtx, value, metrics.TooltipTextSize, metrics.TextWeight, textColor, max(contentWidth-markerSize-gap, 1))
		width := max(title.dims.Size.X, markerSize+gap+row.dims.Size.X)
		height := title.dims.Size.Y + gap + row.dims.Size.Y
		placeText(gtx, title, image.Point{})
		y := title.dims.Size.Y + gap
		marker := image.Rect(0, y+(row.dims.Size.Y-markerSize)/2, markerSize, y+(row.dims.Size.Y-markerSize)/2+markerSize)
		paint.FillShape(gtx.Ops, slice.color, clip.UniformRRect(marker, 1).Op(gtx.Ops))
		placeText(gtx, row, image.Pt(markerSize+gap, y))
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
	})
}

func recordText(ctx *core.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, textColor color.NRGBA, maxWidth int) recordedText {
	if value == "" || maxWidth <= 0 {
		return recordedText{}
	}
	textGtx := gtx
	textGtx.Constraints.Min = image.Point{}
	textGtx.Constraints.Max.X = maxWidth
	macro := op.Record(gtx.Ops)
	label := material.Label(core.ActiveMaterial(ctx), size, value)
	label.Font.Weight = weight
	label.Color = textColor
	dims := label.Layout(textGtx)
	return recordedText{call: macro.Stop(), dims: dims}
}

func placeText(gtx layout.Context, value recordedText, position image.Point) {
	if value.dims.Size.X <= 0 || value.dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	value.call.Add(gtx.Ops)
	offset.Pop()
}

func placeBlock(gtx layout.Context, value recordedBlock, position image.Point) {
	if value.dims.Size.X <= 0 || value.dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	value.call.Add(gtx.Ops)
	offset.Pop()
}

func pieChartStyle(visual theme.PieChartVisual, enabled bool) chartStyle {
	opacity := float32(1)
	if !enabled {
		opacity = visual.DisabledOpacity
	}
	return chartStyle{text: visual.Colors.Text, mutedText: visual.Colors.MutedText, empty: visual.Colors.Empty, opacity: opacity}
}

func formatPieNumber(value float64) string {
	return fmt.Sprintf("%g", value)
}
