package barchart

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type recordedChartText struct {
	call op.CallOp
	dims layout.Dimensions
}

type recordedChartBlock struct {
	call op.CallOp
	dims layout.Dimensions
}

func recordChartText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, col color.NRGBA, maxWidth int) recordedChartText {
	if value == "" || maxWidth <= 0 {
		return recordedChartText{}
	}
	labelGtx := gtx
	labelGtx.Constraints.Min = image.Point{}
	labelGtx.Constraints.Max.X = min(maxWidth, labelGtx.Constraints.Max.X)
	macro := op.Record(gtx.Ops)
	label := material.Label(frame.ActiveTheme(ctx).Material, size, value)
	label.Color = col
	label.Font.Weight = weight
	label.MaxLines = 1
	dims := label.Layout(labelGtx)
	return recordedChartText{call: macro.Stop(), dims: dims}
}

func placeChartText(gtx layout.Context, value recordedChartText, position image.Point) {
	if value.dims.Size.X <= 0 || value.dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	value.call.Add(gtx.Ops)
	offset.Pop()
}

func placeChartBlock(gtx layout.Context, value recordedChartBlock, position image.Point) {
	if value.dims.Size.X <= 0 || value.dims.Size.Y <= 0 {
		return
	}
	offset := op.Offset(position).Push(gtx.Ops)
	value.call.Add(gtx.Ops)
	offset.Pop()
}
