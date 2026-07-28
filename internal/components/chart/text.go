package chart

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func RecordText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, textColor color.NRGBA, maxWidth int) (op.CallOp, layout.Dimensions) {
	if value == "" || maxWidth <= 0 {
		return op.CallOp{}, layout.Dimensions{}
	}
	textGtx := gtx
	textGtx.Constraints.Min = image.Point{}
	textGtx.Constraints.Max.X = min(maxWidth, textGtx.Constraints.Max.X)
	macro := op.Record(gtx.Ops)
	label := material.Label(frame.ActiveMaterial(ctx), size, value)
	label.Color = textColor
	label.Font.Weight = weight
	label.MaxLines = 1
	dims := label.Layout(textGtx)
	return macro.Stop(), dims
}

func PlaceRecorded(gtx layout.Context, call op.CallOp, dims layout.Dimensions, position image.Point) {
	if dims.Size.X <= 0 || dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
}

type TooltipMarker uint8

const (
	TooltipMarkerNone TooltipMarker = iota
	TooltipMarkerCircle
	TooltipMarkerSquare
)

type TooltipRow struct {
	Text  string
	Color color.NRGBA
}

type recordedTooltipRow struct {
	call  op.CallOp
	dims  layout.Dimensions
	color color.NRGBA
}

func LayoutTooltipRows(ctx *frame.Context, gtx layout.Context, title string, rows []TooltipRow, textSize unit.Sp, textColor color.NRGBA, markerSize, rowGap int, marker TooltipMarker) layout.Dimensions {
	contentWidth := max(gtx.Constraints.Max.X, 1)
	titleCall, titleDims := RecordText(ctx, gtx, title, textSize, font.Medium, textColor, contentWidth)
	recorded := make([]recordedTooltipRow, len(rows))
	width, height := titleDims.Size.X, titleDims.Size.Y
	if height > 0 && len(rows) > 0 {
		height += rowGap
	}
	for index, row := range rows {
		maxWidth := contentWidth
		if row.Color.A != 0 && marker != TooltipMarkerNone {
			maxWidth = max(contentWidth-markerSize-rowGap, 1)
		}
		call, dims := RecordText(ctx, gtx, row.Text, textSize, font.Normal, textColor, maxWidth)
		recorded[index] = recordedTooltipRow{call: call, dims: dims, color: row.Color}
		rowWidth := dims.Size.X
		if row.Color.A != 0 && marker != TooltipMarkerNone {
			rowWidth += markerSize + rowGap
		}
		width = max(width, rowWidth)
		height += dims.Size.Y
		if index < len(rows)-1 {
			height += rowGap
		}
	}

	PlaceRecorded(gtx, titleCall, titleDims, image.Point{})
	y := titleDims.Size.Y
	if titleDims.Size.Y > 0 && len(recorded) > 0 {
		y += rowGap
	}
	for _, row := range recorded {
		x := 0
		if row.color.A != 0 && marker != TooltipMarkerNone {
			drawTooltipMarker(gtx, image.Pt(markerSize, markerSize), image.Pt(0, y+(row.dims.Size.Y-markerSize)/2), marker, row.color)
			x = markerSize + rowGap
		}
		PlaceRecorded(gtx, row.call, row.dims, image.Pt(x, y))
		y += row.dims.Size.Y + rowGap
	}
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
}

func drawTooltipMarker(gtx layout.Context, size, position image.Point, marker TooltipMarker, col color.NRGBA) {
	rect := image.Rectangle{Min: position, Max: position.Add(size)}
	switch marker {
	case TooltipMarkerCircle:
		paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
	case TooltipMarkerSquare:
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, 1).Op(gtx.Ops))
	}
}
