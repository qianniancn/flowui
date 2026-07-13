package barchart

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
)

const maxTooltipSeries = 8

func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, style chartStyle) {
	if len(selection.entries) == 0 {
		return
	}
	if w.tooltipContent != nil {
		w.drawCustomTooltip(ctx, gtx, geometry, selection, style)
		return
	}
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	padding := max(gtx.Dp(tokens.TooltipPadding), 0)
	markerSize := max(gtx.Dp(tokens.TooltipMarkerSize), 2)
	rowGap := max(gtx.Dp(tokens.TooltipRowGap), 0)
	maxWidth := min(max(gtx.Dp(tokens.TooltipMaxWidth), 80), max(geometry.size.X-padding*2, 1))
	contentWidth := max(maxWidth-padding*2, 1)
	title := recordChartText(ctx, gtx, w.categoryLabel(selection.index), tokens.TooltipTextSize, font.Medium, style.tooltipText, contentWidth)

	limit := min(len(selection.entries), maxTooltipSeries)
	rows := make([]recordedChartText, 0, limit+1)
	colors := make([][4]uint8, 0, limit+1)
	for index := 0; index < limit; index++ {
		entry := selection.entries[index]
		value := fmt.Sprintf("%s  %s", entry.series.label, w.yLabel(entry.bar.value, geometry.yScale.interval))
		row := recordChartText(ctx, gtx, value, tokens.TooltipTextSize, font.Normal, style.tooltipText, max(contentWidth-markerSize-rowGap, 1))
		rows = append(rows, row)
		colors = append(colors, [4]uint8{entry.bar.color.R, entry.bar.color.G, entry.bar.color.B, entry.bar.color.A})
	}
	if len(selection.entries) > limit {
		rows = append(rows, recordChartText(ctx, gtx, fmt.Sprintf("+%d series", len(selection.entries)-limit), tokens.TooltipTextSize, font.Normal, style.tooltipText, contentWidth))
		colors = append(colors, [4]uint8{})
	}

	panelWidth := title.dims.Size.X
	panelHeight := title.dims.Size.Y
	if panelHeight > 0 && len(rows) > 0 {
		panelHeight += rowGap
	}
	for index, row := range rows {
		rowWidth := row.dims.Size.X
		if colors[index][3] != 0 {
			rowWidth += markerSize + rowGap
		}
		panelWidth = max(panelWidth, rowWidth)
		panelHeight += row.dims.Size.Y
		if index < len(rows)-1 {
			panelHeight += rowGap
		}
	}
	panelWidth = min(panelWidth+padding*2, maxWidth)
	panelHeight += padding * 2
	panelSize := image.Pt(panelWidth, panelHeight)
	gap := max(gtx.Dp(tokens.TooltipGap), 0)
	position := barTooltipPosition(selection, geometry, panelSize, gap)

	offset := op.Offset(position).Push(gtx.Ops)
	drawBarTooltipSurface(ctx, gtx, panelSize, style)

	y := padding
	placeChartText(gtx, title, image.Pt(padding, y))
	y += title.dims.Size.Y
	if title.dims.Size.Y > 0 && len(rows) > 0 {
		y += rowGap
	}
	for index, row := range rows {
		x := padding
		if colors[index][3] != 0 {
			col := colors[index]
			markerRect := image.Rect(x, y+(row.dims.Size.Y-markerSize)/2, x+markerSize, y+(row.dims.Size.Y-markerSize)/2+markerSize)
			paint.FillShape(gtx.Ops, rgba(col), clip.UniformRRect(markerRect, 1).Op(gtx.Ops))
			x += markerSize + rowGap
		}
		placeChartText(gtx, row, image.Pt(x, y))
		y += row.dims.Size.Y + rowGap
	}
	offset.Pop()
}

func (w Widget) drawCustomTooltip(ctx *frame.Context, gtx layout.Context, geometry chartGeometry, selection chartSelection, style chartStyle) {
	content := w.tooltipContent(w.publicSelection(selection, geometry))
	if content == nil {
		return
	}
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	padding := max(gtx.Dp(tokens.TooltipPadding), 0)
	maxWidth := min(max(gtx.Dp(tokens.TooltipMaxWidth), 80), max(geometry.size.X-padding*2, 1))
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{Max: image.Pt(max(maxWidth-padding*2, 1), max(geometry.size.Y-padding*2, 1))}
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return content.Layout(ctx, contentGtx)
	})
	call := macro.Stop()
	panelSize := image.Pt(dims.Size.X+padding*2, dims.Size.Y+padding*2)
	gap := max(gtx.Dp(tokens.TooltipGap), 0)
	position := barTooltipPosition(selection, geometry, panelSize, gap)
	offset := op.Offset(position).Push(gtx.Ops)
	drawBarTooltipSurface(ctx, gtx, panelSize, style)
	offset.Pop()

	contentPosition := position.Add(image.Pt(padding, padding))
	placement.PlaceOffset(contentPosition)
	contentOffset := op.Offset(contentPosition).Push(gtx.Ops)
	call.Add(gtx.Ops)
	contentOffset.Pop()
}

func barTooltipPosition(selection chartSelection, geometry chartGeometry, panelSize image.Point, gap int) image.Point {
	position := image.Pt(int(selection.pixelX)+gap, geometry.plot.Min.Y+gap)
	if geometry.horizontal {
		position.Y = int(selection.pixelY) - panelSize.Y/2
	}
	if position.X+panelSize.X > geometry.size.X {
		position.X = int(selection.pixelX) - gap - panelSize.X
	}
	position.X = min(max(position.X, 0), max(geometry.size.X-panelSize.X, 0))
	position.Y = min(max(position.Y, 0), max(geometry.size.Y-panelSize.Y, 0))
	return position
}

func drawBarTooltipSurface(ctx *frame.Context, gtx layout.Context, panelSize image.Point, style chartStyle) {
	tokens := frame.ActiveTheme(ctx).Components.BarChart
	radius := min(max(gtx.Dp(tokens.TooltipRadius), 0), min(panelSize.X, panelSize.Y)/2)
	rect := image.Rectangle{Max: panelSize}
	render.DrawShadow(gtx, rect, render.RoundedShadowCorners(tokens.TooltipRadius, tokens.TooltipRadius, tokens.TooltipRadius, tokens.TooltipRadius), render.PopupShadow(frame.ActiveTheme(ctx).Palette.OverlayShadowColor()))
	paint.FillShape(gtx.Ops, style.tooltip, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	border := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: 1}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, style.tooltipBorder)
	border.Pop()
}

func rgba(value [4]uint8) (result color.NRGBA) {
	result.R, result.G, result.B, result.A = value[0], value[1], value[2], value[3]
	return result
}
