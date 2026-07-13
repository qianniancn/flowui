package progress

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type recordedMeterChild struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func (m MeterWidget) layout(ctx *frame.Context, gtx layout.Context, style meterStyle, sizeStyle meterSizeStyle, progress float32, valueText string) layout.Dimensions {
	output := valueText
	if !m.showValue {
		output = ""
	}
	if m.label == "" && output == "" && m.valueContent == nil {
		return m.layoutTrack(gtx, style, sizeStyle, progress)
	}
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(frame.ActiveTheme(ctx).Components.Meter.HeaderGap)}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutHeader(ctx, gtx, style, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return m.layoutTrack(gtx, style, sizeStyle, progress)
		}),
	)
}

func (m MeterWidget) layoutHeader(ctx *frame.Context, gtx layout.Context, style meterStyle, output string) layout.Dimensions {
	outputGtx := gtx
	outputGtx.Constraints.Min = image.Point{}
	outputChild := m.recordOutput(ctx, outputGtx, style, output)
	labelGtx := gtx
	labelGtx.Constraints.Min = image.Point{}
	labelGtx.Constraints.Max.X = max(labelGtx.Constraints.Max.X-outputChild.dims.Size.X, 0)
	labelChild := recordMeterWidget(ctx, labelGtx, func(gtx layout.Context) layout.Dimensions {
		if m.label == "" {
			return layout.Dimensions{}
		}
		return text.New(m.label).
			Size(float32(frame.ActiveTheme(ctx).Components.Meter.TextSize)).
			Weight(font.Medium).
			Color(style.label).
			Layout(ctx, gtx)
	})
	height := max(labelChild.dims.Size.Y, outputChild.dims.Size.Y)
	width := labelChild.dims.Size.X + outputChild.dims.Size.X
	if outputChild.dims.Size.X > 0 {
		width = gtx.Constraints.Max.X
	}
	size := gtx.Constraints.Constrain(image.Pt(width, height))
	placeMeterChild(gtx, labelChild, image.Pt(0, (size.Y-labelChild.dims.Size.Y)/2))
	placeMeterChild(gtx, outputChild, image.Pt(max(size.X-outputChild.dims.Size.X, 0), (size.Y-outputChild.dims.Size.Y)/2))
	return layout.Dimensions{Size: size}
}

func (m MeterWidget) recordOutput(ctx *frame.Context, gtx layout.Context, style meterStyle, output string) recordedMeterChild {
	return recordMeterWidget(ctx, gtx, func(gtx layout.Context) layout.Dimensions {
		restore := frame.PushColors(ctx, style.output, ctx.BackgroundColor())
		defer restore()
		if m.valueContent != nil {
			return m.valueContent.Layout(ctx, gtx)
		}
		if output == "" {
			return layout.Dimensions{}
		}
		return text.New(output).
			Size(float32(frame.ActiveTheme(ctx).Components.Meter.TextSize)).
			Weight(font.Medium).
			Color(style.output).
			Layout(ctx, gtx)
	})
}

func recordMeterWidget(ctx *frame.Context, gtx layout.Context, widget layout.Widget) recordedMeterChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return widget(gtx)
	})
	return recordedMeterChild{call: macro.Stop(), dims: dims, placement: placement}
}

func placeMeterChild(gtx layout.Context, child recordedMeterChild, position image.Point) {
	child.placement.PlaceOffset(position)
	offset := op.Offset(position).Push(gtx.Ops)
	child.call.Add(gtx.Ops)
	offset.Pop()
}

func (m MeterWidget) layoutTrack(gtx layout.Context, style meterStyle, sizeStyle meterSizeStyle, progress float32) layout.Dimensions {
	height := min(gtx.Dp(sizeStyle.height), gtx.Constraints.Max.Y)
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, height))
	drawLinearTrack(gtx, size, sizeStyle.radius, linearTrackStyle{track: style.track, fill: style.fill}, progress, false, !m.disabled)
	return layout.Dimensions{Size: size}
}
