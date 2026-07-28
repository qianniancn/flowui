package barchart

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/chart"
	"github.com/qianniancn/flowui/internal/frame"
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
	call, dims := chart.RecordText(ctx, gtx, value, size, weight, col, maxWidth)
	return recordedChartText{call: call, dims: dims}
}

func placeChartText(gtx layout.Context, value recordedChartText, position image.Point) {
	chart.PlaceRecorded(gtx, value.call, value.dims, position)
}

func placeChartBlock(gtx layout.Context, value recordedChartBlock, position image.Point) {
	chart.PlaceRecorded(gtx, value.call, value.dims, position)
}
