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
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

type chartGeometry struct {
	size        image.Point
	area        image.Rectangle
	center      f32.Point
	innerRadius float32
	outerRadius float32
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

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := chartStateFor(ctx, w.key)
	state.beginLegendFrame()
	defer state.endLegendFrame()
	restoreKey := frame.PushKey(ctx, w.key)
	defer restoreKey()

	enabled := gtx.Enabled() && !w.disabled
	activated := state.updateClick(gtx, enabled)
	tokens := frame.ActiveTheme(ctx).Components.PieChart
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	data := state.dataCache.resolve(w, frame.ActiveTheme(ctx))
	if w.handleLegendClicks(gtx, state, data, enabled) {
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
	return layout.Dimensions{Size: size}
}

func (w Widget) layoutContent(ctx *frame.Context, gtx layout.Context, state *chartState, data chartData, enabled, activated bool, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.PieChart
	style := pieChartStyle(activeTheme.Palette.Foreground, activeTheme.Palette.MutedForeground, activeTheme.Palette.SurfaceSecondary, activeTheme.DisabledOpacityValue(), enabled)
	left := min(max(gtx.Dp(tokens.PlotPaddingLeft), 0), size.X/2)
	right := min(max(gtx.Dp(tokens.PlotPaddingRight), 0), size.X-left)
	availableWidth := max(size.X-left-right, 0)

	legend := recordedBlock{}
	if w.legendVisible(data) {
		legend = w.recordLegend(ctx, gtx, state, data, style, availableWidth, enabled)
	}
	top := max(gtx.Dp(tokens.PlotPaddingTop), 0)
	legendPosition := image.Pt(left+max((availableWidth-legend.dims.Size.X)/2, 0), top)
	if legend.dims.Size.Y > 0 {
		top += legend.dims.Size.Y + max(gtx.Dp(tokens.LegendGap), 0)
	}
	bottom := max(gtx.Dp(tokens.PlotPaddingBottom), 0)
	geometry := w.resolveGeometry(size, image.Rect(left, top, max(size.X-right, left), max(size.Y-bottom, top)))
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
		tooltipProgress = state.tooltipTransition.Progress(gtx, tooltipVisible)
		if !tooltipVisible && tooltipProgress <= 0 {
			state.tooltipSlice = resolvedSlice{}
		}
	} else {
		state.clearTooltip()
	}
	tooltipPointer := state.pointer
	if tooltipVisible || tooltipProgress > 0 {
		tooltipPointer = state.tooltipTransition.Position(gtx, state.pointer)
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	placeBlock(gtx, legend, legendPosition)
	hoveredKey := ""
	if selectedIndex >= 0 {
		hoveredKey = displayData.slices[selectedIndex].key
	}
	if hasVisibleSector(data.slices) {
		drawPieSlices(gtx, displayData.slices, geometry, hoveredKey, float32(gtx.Dp(tokens.EmphasisSize)))
		if w.showLabels {
			w.drawLabels(ctx, gtx, data.slices, geometry, style)
		}
	} else {
		drawEmptyPie(gtx, geometry, style.empty)
		w.drawEmptyText(ctx, gtx, geometry, style)
	}
	if tooltipVisible || tooltipProgress > 0 {
		w.drawTooltip(ctx, gtx, state.tooltipSlice, chart.TooltipAnchor(tooltipPointer), tooltipProgress, state.tooltipTransition.Exiting())
	}
	opacity.Pop()

	interactionArea := geometry.interactionArea(float32(gtx.Dp(tokens.EmphasisSize)))
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

func (w Widget) semanticDescription(data chartData) string {
	label := w.label
	if label == "" {
		label = "Pie chart"
	}
	sliceLabel := "slices"
	if len(data.slices) == 1 {
		sliceLabel = "slice"
	}
	return fmt.Sprintf("%s, %d %s", label, len(data.slices), sliceLabel)
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

func (w Widget) recordLegend(ctx *frame.Context, gtx layout.Context, state *chartState, data chartData, style chartStyle, maxWidth int, enabled bool) recordedBlock {
	if maxWidth <= 0 || len(data.legend) == 0 {
		return recordedBlock{}
	}
	tokens := frame.ActiveTheme(ctx).Components.PieChart
	activeTheme := frame.ActiveTheme(ctx)
	markerSize := max(gtx.Dp(tokens.LegendMarkerSize), 2)
	markerGap := max(gtx.Dp(tokens.LegendMarkerGap), 0)
	itemGap := max(gtx.Dp(tokens.LegendItemGap), 0)
	lineGap := max(gtx.Dp(tokens.LegendLineGap), 0)

	macro := op.Record(gtx.Ops)
	x, y, rowHeight, usedWidth := 0, 0, 0, 0
	for _, slice := range data.legend {
		label := recordText(ctx, gtx, slice.label, tokens.LegendTextSize, font.Normal, style.text, max(maxWidth-markerSize-markerGap, 1))
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
					paint.FillShape(gtx.Ops, activeTheme.Palette.SurfaceHover, clip.UniformRRect(image.Rectangle{Max: image.Pt(itemWidth, itemHeight)}, min(itemHeight/2, 4)).Op(gtx.Ops))
				}
			}
			itemOpacity := float32(1)
			if slice.hidden {
				itemOpacity = .38
			}
			fade := paint.PushOpacity(gtx.Ops, itemOpacity)
			markerRect := image.Rect(0, (itemHeight-markerSize)/2, markerSize, (itemHeight-markerSize)/2+markerSize)
			radius := min(max(gtx.Dp(tokens.LegendMarkerRadius), 0), markerSize/2)
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

func (w Widget) drawEmptyText(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	value := w.emptyText
	if value == "" {
		value = "No data"
	}
	tokens := frame.ActiveTheme(ctx).Components.PieChart
	label := recordText(ctx, gtx, value, tokens.LabelTextSize, font.Normal, style.mutedText, max(geometry.area.Dx(), 1))
	position := image.Pt(
		int(geometry.center.X)-label.dims.Size.X/2,
		int(geometry.center.Y)-label.dims.Size.Y/2,
	)
	placeText(gtx, label, position)
}

func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, slice resolvedSlice, anchor image.Rectangle, progress float32, exiting bool) {
	content := w.tooltipWidget(slice)
	if content == nil {
		return
	}
	tooltip.NewPopup(content).
		Placement(overlay.PopoverRightStart).
		Offset(max(frame.ActiveTheme(ctx).Components.PieChart.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Layout(ctx, gtx, anchor)
}

func (w Widget) tooltipWidget(slice resolvedSlice) frame.Widget {
	selection := w.publicSelection(slice)
	if w.tooltipContent != nil {
		return w.tooltipContent(selection)
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		activeTheme := frame.ActiveTheme(ctx)
		chartTokens := activeTheme.Components.PieChart
		tooltipTokens := activeTheme.Components.Tooltip
		textColor := activeTheme.Palette.OverlayForegroundColor()
		markerSize := max(gtx.Dp(chartTokens.TooltipMarkerSize), 2)
		gap := max(gtx.Dp(chartTokens.TooltipRowGap), 0)
		contentWidth := max(gtx.Constraints.Max.X, 1)
		title := recordText(ctx, gtx, slice.label, tooltipTokens.TextSize, font.Medium, textColor, contentWidth)
		value := fmt.Sprintf("%s  %.2f%%", formatPieNumber(slice.value), slice.percent)
		row := recordText(ctx, gtx, value, tooltipTokens.TextSize, font.Normal, textColor, max(contentWidth-markerSize-gap, 1))
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

func recordText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, textColor color.NRGBA, maxWidth int) recordedText {
	if value == "" || maxWidth <= 0 {
		return recordedText{}
	}
	textGtx := gtx
	textGtx.Constraints.Min = image.Point{}
	textGtx.Constraints.Max.X = maxWidth
	macro := op.Record(gtx.Ops)
	label := material.Label(frame.ActiveTheme(ctx).Material, size, value)
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

func pieChartStyle(text, muted, empty color.NRGBA, disabledOpacity float32, enabled bool) chartStyle {
	empty.A = byte(float32(empty.A) * .72)
	opacity := float32(1)
	if !enabled {
		opacity = disabledOpacity
	}
	return chartStyle{text: text, mutedText: muted, empty: empty, opacity: opacity}
}

func formatPieNumber(value float64) string {
	return fmt.Sprintf("%g", value)
}
