package candlestick

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, selection chartSelection, interval float64, anchor image.Rectangle, progress float32, exiting bool) {
	content := w.tooltipWidget(selection, interval)
	if content == nil {
		return
	}
	tooltip.NewPopup(content).
		Placement(overlay.PopoverRightStart).
		Offset(max(frame.ActiveTheme(ctx).Components.CandlestickChart.TooltipGap, 0)).
		TransformMotion(false).
		Progress(progress).
		Exiting(exiting).
		Layout(ctx, gtx, anchor)
}

func (w Widget) tooltipWidget(selection chartSelection, interval float64) frame.Widget {
	if w.tooltipContent != nil {
		return w.tooltipContent(w.publicSelection(selection))
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return w.layoutTooltipContent(ctx, gtx, selection, interval)
	})
}

func (w Widget) layoutTooltipContent(ctx *frame.Context, gtx layout.Context, selection chartSelection, interval float64) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	chartTokens := activeTheme.Components.CandlestickChart
	tooltipTokens := activeTheme.Components.Tooltip
	textColor := activeTheme.Palette.OverlayForegroundColor()
	gap := max(gtx.Dp(chartTokens.TooltipRowGap), 0)
	values := w.candlestickTooltipRows(selection, interval)
	rows := make([]chart.TooltipRow, len(values))
	for index, value := range values {
		rows[index].Text = value
	}
	return chart.LayoutTooltipRows(ctx, gtx, w.categoryLabel(selection.index), rows, tooltipTokens.TextSize, textColor, 0, gap, chart.TooltipMarkerNone)
}

func (w Widget) candlestickTooltipRows(selection chartSelection, interval float64) []string {
	return []string{
		fmt.Sprintf("Open  %s", w.yLabel(selection.candle.open, interval)),
		fmt.Sprintf("High  %s", w.yLabel(selection.candle.high, interval)),
		fmt.Sprintf("Low  %s", w.yLabel(selection.candle.low, interval)),
		fmt.Sprintf("Close  %s", w.yLabel(selection.candle.close, interval)),
	}
}
