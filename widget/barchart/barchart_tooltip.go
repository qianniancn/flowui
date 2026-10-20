package barchart

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

const maxTooltipSeries = 8

func (w Widget) drawTooltip(ctx *core.Context, geometry chartGeometry, selection chartSelection, anchor image.Rectangle, progress float32, exiting bool, metrics theme.BarChartMetrics, foreground color.NRGBA) {
	content := w.tooltipWidget(geometry, selection, metrics, foreground)
	if content == nil {
		return
	}
	placement := core.PopoverRightStart
	if geometry.horizontal {
		placement = core.PopoverRight
	}
	tooltip.NewPopup(content).
		Placement(placement).
		Offset(max(metrics.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Overlay(ctx, core.FullKey(ctx, "tooltip:"+w.key), anchor)
}

func (w Widget) tooltipWidget(geometry chartGeometry, selection chartSelection, metrics theme.BarChartMetrics, foreground color.NRGBA) core.Widget {
	if w.tooltipContent != nil {
		return w.tooltipContent(w.publicSelection(selection, geometry))
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutTooltipContent(ctx, gtx, geometry, selection, metrics, foreground)
	})
}

func (w Widget) layoutTooltipContent(ctx *core.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, metrics theme.BarChartMetrics, textColor color.NRGBA) layout.Dimensions {
	markerSize := max(gtx.Dp(metrics.TooltipMarkerSize), 2)
	rowGap := max(gtx.Dp(metrics.TooltipRowGap), 0)
	limit := min(len(selection.entries), maxTooltipSeries)
	rows := make([]chart.TooltipRow, 0, limit+1)
	for index := range limit {
		entry := selection.entries[index]
		value := fmt.Sprintf("%s  %s", entry.series.label, w.yLabel(entry.bar.value, geometry.yScale.Interval))
		rows = append(rows, chart.TooltipRow{Text: value, Color: entry.bar.color})
	}
	if len(selection.entries) > limit {
		rows = append(rows, chart.TooltipRow{Text: ctx.Textf(i18n.ChartMoreSeries, len(selection.entries)-limit)})
	}
	return chart.LayoutTooltipRows(ctx, gtx, w.categoryLabel(selection.index), rows, metrics.TooltipTextSize, textColor, markerSize, rowGap, chart.TooltipMarkerSquare)
}
