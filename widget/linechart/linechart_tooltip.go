package linechart

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

const maxTooltipSeries = 8

func (w Widget) drawTooltip(ctx *core.Context, geometry chartGeometry, selection chartSelection, anchor image.Rectangle, progress float32, exiting bool, visual theme.LineChartVisual) {
	content := w.tooltipWidget(geometry, selection, visual)
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

func (w Widget) tooltipWidget(geometry chartGeometry, selection chartSelection, visual theme.LineChartVisual) core.Widget {
	if w.tooltipContent != nil {
		return w.tooltipContent(w.publicSelection(selection, geometry))
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutTooltipContent(ctx, gtx, geometry, selection, visual)
	})
}

func (w Widget) layoutTooltipContent(ctx *core.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, visual theme.LineChartVisual) layout.Dimensions {
	textColor := visual.Colors.OverlayForeground
	markerSize := max(gtx.Dp(visual.Metrics.TooltipMarkerSize), 2)
	rowGap := max(gtx.Dp(visual.Metrics.TooltipRowGap), 0)
	limit := min(len(selection.entries), maxTooltipSeries)
	rows := make([]chart.TooltipRow, 0, limit+1)
	for index := range limit {
		entry := selection.entries[index]
		value := fmt.Sprintf("%s  %s", entry.series.label, w.yLabel(entry.point.rawY, geometry.yScale.Interval))
		rows = append(rows, chart.TooltipRow{Text: value, Color: entry.series.color})
	}
	if len(selection.entries) > limit {
		rows = append(rows, chart.TooltipRow{Text: ctx.Textf(i18n.ChartMoreSeries, len(selection.entries)-limit)})
	}
	return chart.LayoutTooltipRows(ctx, gtx, w.xLabel(selection.x, geometry.xScale.Interval), rows, visual.Metrics.TooltipTextSize, textColor, markerSize, rowGap, chart.TooltipMarkerCircle)
}
