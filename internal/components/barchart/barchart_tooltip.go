package barchart

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	contentWidth := max(gtx.Constraints.Max.X, 1)
	title := recordChartText(ctx, gtx, w.categoryLabel(selection.index), tooltipTokens.TextSize, font.Medium, textColor, contentWidth)

	limit := min(len(selection.entries), maxTooltipSeries)
	rows := make([]recordedChartText, 0, limit+1)
	colors := make([]color.NRGBA, 0, limit+1)
	for index := 0; index < limit; index++ {
		entry := selection.entries[index]
		value := fmt.Sprintf("%s  %s", entry.series.label, w.yLabel(entry.bar.value, geometry.yScale.interval))
		rows = append(rows, recordChartText(ctx, gtx, value, tooltipTokens.TextSize, font.Normal, textColor, max(contentWidth-markerSize-rowGap, 1)))
		colors = append(colors, entry.bar.color)
	}
	if len(selection.entries) > limit {
		rows = append(rows, recordChartText(ctx, gtx, fmt.Sprintf("+%d series", len(selection.entries)-limit), tooltipTokens.TextSize, font.Normal, textColor, contentWidth))
		colors = append(colors, color.NRGBA{})
	}

	width := title.dims.Size.X
	height := title.dims.Size.Y
	if height > 0 && len(rows) > 0 {
		height += rowGap
	}
	for index, row := range rows {
		rowWidth := row.dims.Size.X
		if colors[index].A != 0 {
			rowWidth += markerSize + rowGap
		}
		width = max(width, rowWidth)
		height += row.dims.Size.Y
		if index < len(rows)-1 {
			height += rowGap
		}
	}
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	y := 0
	placeChartText(gtx, title, image.Pt(0, y))
	y += title.dims.Size.Y
	if title.dims.Size.Y > 0 && len(rows) > 0 {
		y += rowGap
	}
	for index, row := range rows {
		x := 0
		if colors[index].A != 0 {
			markerRect := image.Rect(x, y+(row.dims.Size.Y-markerSize)/2, x+markerSize, y+(row.dims.Size.Y-markerSize)/2+markerSize)
			paint.FillShape(gtx.Ops, colors[index], clip.UniformRRect(markerRect, 1).Op(gtx.Ops))
			x += markerSize + rowGap
		}
		placeChartText(gtx, row, image.Pt(x, y))
		y += row.dims.Size.Y + rowGap
	}
	return layout.Dimensions{Size: size}
}

func barTooltipAnchor(pointer f32.Point) image.Rectangle {
	position := pointer.Round()
	return image.Rectangle{Min: position, Max: position.Add(image.Pt(1, 1))}
}
