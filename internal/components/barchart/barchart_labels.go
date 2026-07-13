package barchart

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (w Widget) layoutAxisLabels(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	gap := max(gtx.Dp(tokens.TickLabelGap), 0)
	for _, tick := range geometry.yTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Min.X-gap, 0))
		position := image.Pt(max(geometry.plot.Min.X-gap-label.dims.Size.X, 0), int(tick.pixel)-label.dims.Size.Y/2)
		placeChartText(gtx, label, position)
	}
	for _, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
		x := int(tick.pixel) - label.dims.Size.X/2
		x = min(max(x, 0), max(geometry.size.X-label.dims.Size.X, 0))
		placeChartText(gtx, label, image.Pt(x, geometry.plot.Max.Y+gap))
	}
}

func (w Widget) measureYAxisLabelWidth(ctx *frame.Context, gtx layout.Context, data chartData, maxWidth int) int {
	scale := w.resolveYScale(data)
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	width := 0
	for _, value := range scale.ticks {
		label := recordChartText(ctx, gtx, w.yLabel(value, scale.interval), tokens.AxisTextSize, font.Normal, frame.ActiveTheme(ctx).Palette.MutedForeground, maxWidth)
		width = max(width, label.dims.Size.X)
	}
	return width
}

func (w Widget) pruneXTicks(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) []categoryTick {
	if len(geometry.xTicks) < 2 {
		return geometry.xTicks
	}
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	type measuredTick struct {
		tick        categoryTick
		left, right int
	}
	measured := make([]measuredTick, len(geometry.xTicks))
	for index, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
		left := int(tick.pixel) - label.dims.Size.X/2
		measured[index] = measuredTick{tick: tick, left: left, right: left + label.dims.Size.X}
	}
	gap := max(gtx.Dp(tokens.TickLabelGap), 4)
	result := []measuredTick{measured[0]}
	lastCandidate := measured[len(measured)-1]
	for _, candidate := range measured[1 : len(measured)-1] {
		previous := result[len(result)-1]
		if candidate.left >= previous.right+gap && candidate.right+gap <= lastCandidate.left {
			result = append(result, candidate)
		}
	}
	for len(result) > 1 && lastCandidate.left < result[len(result)-1].right+gap {
		result = result[:len(result)-1]
	}
	if lastCandidate.tick.index != result[len(result)-1].tick.index {
		result = append(result, lastCandidate)
	}
	ticks := make([]categoryTick, len(result))
	for index := range result {
		ticks[index] = result[index].tick
	}
	return ticks
}

func (w Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, style chartStyle) {
	value := w.emptyText
	if value == "" {
		value = "No data"
	}
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	label := recordChartText(ctx, gtx, value, tokens.AxisTextSize, font.Normal, style.axisLabel, max(geometry.plot.Dx(), 1))
	position := image.Pt(
		geometry.plot.Min.X+max((geometry.plot.Dx()-label.dims.Size.X)/2, 0),
		geometry.plot.Min.Y+max((geometry.plot.Dy()-label.dims.Size.Y)/2, 0),
	)
	placeChartText(gtx, label, position)
}

func (w Widget) recordLegend(ctx *frame.Context, gtx layout.Context, data chartData, style chartStyle, maxWidth int) recordedChartBlock {
	if maxWidth <= 0 || len(data.series) == 0 {
		return recordedChartBlock{}
	}
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	markerSize := max(gtx.Dp(tokens.LegendMarkerSize), 2)
	markerGap := max(gtx.Dp(tokens.LegendMarkerGap), 0)
	itemGap := max(gtx.Dp(tokens.LegendItemGap), 0)
	lineGap := max(gtx.Dp(tokens.LegendLineGap), 0)

	macro := op.Record(gtx.Ops)
	x, y, rowHeight, usedWidth := 0, 0, 0, 0
	for _, series := range data.series {
		label := recordChartText(ctx, gtx, series.label, tokens.LegendTextSize, font.Normal, style.axisLabel, max(maxWidth-markerSize-markerGap, 1))
		itemWidth := markerSize + markerGap + label.dims.Size.X
		itemHeight := max(markerSize, label.dims.Size.Y)
		if x > 0 && x+itemWidth > maxWidth {
			y += rowHeight + lineGap
			x, rowHeight = 0, 0
		}
		markerRect := image.Rect(x, y+(itemHeight-markerSize)/2, x+markerSize, y+(itemHeight-markerSize)/2+markerSize)
		radius := min(max(gtx.Dp(tokens.LegendMarkerRadius), 0), markerSize/2)
		paint.FillShape(gtx.Ops, series.color, clip.UniformRRect(markerRect, radius).Op(gtx.Ops))
		placeChartText(gtx, label, image.Pt(x+markerSize+markerGap, y+(itemHeight-label.dims.Size.Y)/2))
		x += itemWidth + itemGap
		usedWidth = max(usedWidth, min(x-itemGap, maxWidth))
		rowHeight = max(rowHeight, itemHeight)
	}
	height := y + rowHeight
	return recordedChartBlock{call: macro.Stop(), dims: layout.Dimensions{Size: image.Pt(usedWidth, height)}}
}
