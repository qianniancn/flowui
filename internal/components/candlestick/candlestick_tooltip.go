package candlestick

import (
	"fmt"
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
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
	contentWidth := max(gtx.Constraints.Max.X, 1)
	title := recordChartText(ctx, gtx, w.categoryLabel(selection.index), tooltipTokens.TextSize, font.Medium, textColor, contentWidth)
	values := w.candlestickTooltipRows(selection, interval)
	rows := make([]recordedChartText, len(values))
	width, height := title.dims.Size.X, title.dims.Size.Y
	if title.dims.Size.Y > 0 && len(rows) > 0 {
		height += gap
	}
	for index, value := range values {
		rows[index] = recordChartText(ctx, gtx, value, tooltipTokens.TextSize, font.Normal, textColor, contentWidth)
		width = max(width, rows[index].dims.Size.X)
		height += rows[index].dims.Size.Y
		if index < len(rows)-1 {
			height += gap
		}
	}
	placeChartText(gtx, title, image.Point{})
	y := title.dims.Size.Y
	if title.dims.Size.Y > 0 && len(rows) > 0 {
		y += gap
	}
	for _, row := range rows {
		placeChartText(gtx, row, image.Pt(0, y))
		y += row.dims.Size.Y + gap
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
}

func (w Widget) candlestickTooltipRows(selection chartSelection, interval float64) []string {
	return []string{
		fmt.Sprintf("Open  %s", w.yLabel(selection.candle.open, interval)),
		fmt.Sprintf("High  %s", w.yLabel(selection.candle.high, interval)),
		fmt.Sprintf("Low  %s", w.yLabel(selection.candle.low, interval)),
		fmt.Sprintf("Close  %s", w.yLabel(selection.candle.close, interval)),
	}
}

func tooltipAnchor(pointerPosition f32.Point) image.Rectangle {
	position := pointerPosition.Round()
	return image.Rectangle{Min: position, Max: position.Add(image.Pt(1, 1))}
}
