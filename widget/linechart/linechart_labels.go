package linechart

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
)

func (w Widget) layoutAxisLabels(ctx *core.Context, gtx layout.Context, geometry chartGeometry, style chartStyle, metrics theme.LineChartMetrics) {
	gap := max(gtx.Dp(metrics.TickLabelGap), 0)
	for _, tick := range geometry.yTicks {
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Min.X-gap, 0))
		position := image.Pt(
			max(geometry.plot.Min.X-gap-label.dims.Size.X, 0),
			int(tick.pixel)-label.dims.Size.Y/2,
		)
		placeChartText(gtx, label, position)
	}
	for _, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx()/2, 1))
		x := int(tick.pixel) - label.dims.Size.X/2
		x = min(max(x, 0), max(geometry.size.X-label.dims.Size.X, 0))
		position := image.Pt(x, geometry.plot.Max.Y+gap)
		placeChartText(gtx, label, position)
	}
}

func (w Widget) measureYAxisLabelWidth(ctx *core.Context, gtx layout.Context, data chartData, metrics theme.LineChartMetrics, labelColor color.NRGBA, maxWidth int) int {
	scale := w.resolveYScale(data)
	width := 0
	for _, value := range scale.Ticks {
		label := recordChartText(ctx, gtx, w.yLabel(value, scale.Interval), metrics.AxisTextSize, metrics.TextWeight, labelColor, maxWidth)
		width = max(width, label.dims.Size.X)
	}
	return width
}

func (w Widget) pruneXTicks(ctx *core.Context, gtx layout.Context, geometry chartGeometry, style chartStyle, metrics theme.LineChartMetrics) []axisTick {
	if len(geometry.xTicks) < 2 {
		return geometry.xTicks
	}
	type measuredTick struct {
		tick        axisTick
		left, right int
	}
	measured := make([]measuredTick, len(geometry.xTicks))
	for index, tick := range geometry.xTicks {
		label := recordChartText(ctx, gtx, tick.label, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx(), 1))
		left := int(tick.pixel) - label.dims.Size.X/2
		measured[index] = measuredTick{tick: tick, left: left, right: left + label.dims.Size.X}
	}
	gap := max(gtx.Dp(metrics.TickLabelGap), 4)
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
	if lastCandidate.tick.value != result[len(result)-1].tick.value {
		result = append(result, lastCandidate)
	}
	ticks := make([]axisTick, len(result))
	for index := range result {
		ticks[index] = result[index].tick
	}
	return ticks
}

func (w Widget) layoutEmpty(ctx *core.Context, gtx layout.Context, geometry chartGeometry, style chartStyle, metrics theme.LineChartMetrics) {
	value := w.emptyText
	if !w.hasEmptyText {
		value = ctx.Text(i18n.ChartNoData)
	}
	label := recordChartText(ctx, gtx, value, metrics.AxisTextSize, metrics.TextWeight, style.axisLabel, max(geometry.plot.Dx(), 1))
	position := image.Pt(
		geometry.plot.Min.X+max((geometry.plot.Dx()-label.dims.Size.X)/2, 0),
		geometry.plot.Min.Y+max((geometry.plot.Dy()-label.dims.Size.Y)/2, 0),
	)
	placeChartText(gtx, label, position)
}

func (w Widget) recordLegend(ctx *core.Context, gtx layout.Context, state *chartState, data chartData, style chartStyle, metrics theme.LineChartMetrics, hiddenOpacity float32, maxWidth int, enabled bool) recordedChartBlock {
	if maxWidth <= 0 || len(data.legend) == 0 {
		return recordedChartBlock{}
	}
	markerWidth := max(gtx.Dp(metrics.LegendMarkerWidth), 1)
	markerSize := max(gtx.Dp(metrics.LegendMarkerSize), 2)
	itemGap := max(gtx.Dp(metrics.LegendItemGap), 0)
	lineGap := max(gtx.Dp(metrics.LegendLineGap), 0)

	macro := op.Record(gtx.Ops)
	x, y, rowHeight, usedWidth := 0, 0, 0, 0
	for _, series := range data.legend {
		label := recordChartText(ctx, gtx, series.label, metrics.LegendTextSize, metrics.TextWeight, style.axisLabel, max(maxWidth-markerWidth, 1))
		itemWidth := markerWidth + label.dims.Size.X
		itemHeight := max(markerSize, label.dims.Size.Y)
		if x > 0 && x+itemWidth > maxWidth {
			y += rowHeight + lineGap
			x, rowHeight = 0, 0
		}
		item := state.legendItem(series.key)
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
					rect := image.Rectangle{Max: image.Pt(itemWidth, itemHeight)}
					paint.FillShape(gtx.Ops, style.legendHover, clip.UniformRRect(rect, min(itemHeight/2, 4)).Op(gtx.Ops))
				}
			}
			opacity := float32(1)
			if series.hidden {
				opacity = hiddenOpacity
			}
			fade := paint.PushOpacity(gtx.Ops, opacity)
			centerY := itemHeight / 2
			from := f32.Pt(0, float32(centerY))
			to := f32.Pt(float32(markerWidth), float32(centerY))
			drawChartLine(gtx, from, to, max(series.width, 2), series.color)
			paint.FillShape(gtx.Ops, series.color, clip.Ellipse(chart.PointRect(f32.Pt(float32(markerWidth/2), float32(centerY)), markerSize)).Op(gtx.Ops))
			placeChartText(gtx, label, image.Pt(markerWidth, (itemHeight-label.dims.Size.Y)/2))
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
	height := y + rowHeight
	return recordedChartBlock{call: macro.Stop(), dims: layout.Dimensions{Size: image.Pt(usedWidth, height)}}
}

func (w Widget) handleLegendClicks(gtx layout.Context, state *chartState, data chartData, enabled bool) bool {
	activated := false
	for _, series := range data.legend {
		for state.legendItem(series.key).Clicked(gtx) {
			activated = true
			if enabled && w.onLegendChange != nil {
				w.onLegendChange(series.key, !series.hidden)
			}
		}
	}
	return activated
}
