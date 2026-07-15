package barchart

import (
	"fmt"
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

const maxTooltipSeries = 8

func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, anchor image.Rectangle, progress float32, exiting bool) {
	content := w.tooltipWidget(geometry, selection)
	if content == nil {
		return
	}
	placement := overlay.PopoverRightStart
	if geometry.horizontal {
		placement = overlay.PopoverRight
	}
	tooltip.NewPopup(content).
		Placement(placement).
		Offset(max(frame.ActiveTheme(ctx).Components.BarChart.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Layout(ctx, gtx, anchor)
}

func (w Widget) tooltipWidget(geometry chartGeometry, selection chartSelection) frame.Widget {
	if w.tooltipContent != nil {
		return w.tooltipContent(w.publicSelection(selection, geometry))
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutTooltipContent(ctx, gtx, geometry, selection)
	})
}

func (w Widget) layoutTooltipContent(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	chartTokens := activeTheme.Components.BarChart
	tooltipTokens := activeTheme.Components.Tooltip
	textColor := activeTheme.Palette.OverlayForegroundColor()
	markerSize := max(gtx.Dp(chartTokens.TooltipMarkerSize), 2)
	rowGap := max(gtx.Dp(chartTokens.TooltipRowGap), 0)
	limit := min(len(selection.entries), maxTooltipSeries)
	rows := make([]chart.TooltipRow, 0, limit+1)
	for index := 0; index < limit; index++ {
		entry := selection.entries[index]
		value := fmt.Sprintf("%s  %s", entry.series.label, w.yLabel(entry.bar.value, geometry.yScale.interval))
		rows = append(rows, chart.TooltipRow{Text: value, Color: entry.bar.color})
	}
	if len(selection.entries) > limit {
		rows = append(rows, chart.TooltipRow{Text: fmt.Sprintf("+%d series", len(selection.entries)-limit)})
	}
	return chart.LayoutTooltipRows(ctx, gtx, w.categoryLabel(selection.index), rows, tooltipTokens.TextSize, textColor, markerSize, rowGap, chart.TooltipMarkerSquare)
}

func barTooltipAnchor(pointer f32.Point) image.Rectangle {
	return chart.TooltipAnchor(pointer)
}
