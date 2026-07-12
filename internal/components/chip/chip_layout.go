package chip

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (c Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	style := chipStyleFor(activeTheme, c.color, c.variant)
	sizeStyle := chipSizeStyleFor(activeTheme, c.size)
	height := min(max(gtx.Dp(sizeStyle.height), gtx.Constraints.Min.Y), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		contentBackground := style.background
		if contentBackground.A == 0 {
			contentBackground = ctx.BackgroundColor()
		}
		restore := frame.PushColors(ctx, style.foreground, contentBackground)
		defer restore()
		contentDims = layout.Inset{
			Top:    sizeStyle.paddingY,
			Right:  sizeStyle.paddingX,
			Bottom: sizeStyle.paddingY,
			Left:   sizeStyle.paddingX,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return c.layoutContent(ctx, gtx, sizeStyle, style.foreground)
		})
	}()
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(style.radius), 0), min(size.X, size.Y)/2)
	if !rect.Empty() && style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	root := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	if c.label != "" {
		semantic.LabelOp(c.label).Add(gtx.Ops)
	}
	content.Add(gtx.Ops)
	root.Pop()
	return layout.Dimensions{Size: size, Baseline: contentDims.Baseline}
}

func (c Widget) layoutContent(ctx *frame.Context, gtx layout.Context, sizeStyle chipSizeStyle, foreground color.NRGBA) layout.Dimensions {
	gap := max(gtx.Dp(sizeStyle.contentGap), 0)
	gapCount := 0
	if c.startContent != nil {
		gapCount++
	}
	if c.endContent != nil {
		gapCount++
	}
	gapsWidth := gap * gapCount
	labelMinimum := min(gtx.Dp(sizeStyle.labelPaddingX)*2, max(gtx.Constraints.Max.X-gapsWidth, 0))
	accessoryWidth := max(gtx.Constraints.Max.X-gapsWidth-labelMinimum, 0)

	accessoryGtx := gtx
	accessoryGtx.Constraints.Min = image.Point{}
	accessoryGtx.Constraints.Max.X = accessoryWidth
	end := recordChipWidget(ctx, accessoryGtx, c.endContent)
	accessoryGtx.Constraints.Max.X = max(accessoryWidth-end.dims.Size.X, 0)
	start := recordChipWidget(ctx, accessoryGtx, c.startContent)

	labelGtx := gtx
	labelGtx.Constraints.Min = image.Point{}
	labelGtx.Constraints.Max.X = max(gtx.Constraints.Max.X-gapsWidth-start.dims.Size.X-end.dims.Size.X, 0)
	label := c.recordLabel(ctx, labelGtx, sizeStyle, foreground)

	width := start.dims.Size.X + label.dims.Size.X + end.dims.Size.X + gapsWidth
	height := max(gtx.Constraints.Min.Y, max(start.dims.Size.Y, max(label.dims.Size.Y, end.dims.Size.Y)))
	size := gtx.Constraints.Constrain(image.Pt(width, height))

	x := 0
	if c.startContent != nil {
		placeChipChild(gtx, start, image.Pt(x, max((size.Y-start.dims.Size.Y)/2, 0)))
		x += start.dims.Size.X + gap
	}
	labelPosition := image.Pt(x, max((size.Y-label.dims.Size.Y)/2, 0))
	placeChipChild(gtx, label, labelPosition)
	x += label.dims.Size.X
	if c.endContent != nil {
		x += gap
		placeChipChild(gtx, end, image.Pt(x, max((size.Y-end.dims.Size.Y)/2, 0)))
	}

	baseline := size.Y - labelPosition.Y - label.dims.Size.Y + label.dims.Baseline
	return layout.Dimensions{Size: size, Baseline: max(baseline, 0)}
}

type recordedChipChild struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func recordChipWidget(ctx *frame.Context, gtx layout.Context, child frame.Widget) recordedChipChild {
	if child == nil {
		return recordedChipChild{}
	}
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
	return recordedChipChild{call: macro.Stop(), dims: dims, placement: placement}
}

func (c Widget) recordLabel(ctx *frame.Context, gtx layout.Context, sizeStyle chipSizeStyle, foreground color.NRGBA) recordedChipChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layout.Inset{Left: sizeStyle.labelPaddingX, Right: sizeStyle.labelPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(frame.ActiveTheme(ctx).Material, sizeStyle.textSize, c.label)
			label.Color = foreground
			label.Font.Weight = font.Medium
			label.LineHeight = sizeStyle.lineHeight
			label.MaxLines = 1
			return label.Layout(gtx)
		})
	})
	return recordedChipChild{call: macro.Stop(), dims: dims, placement: placement}
}

func placeChipChild(gtx layout.Context, child recordedChipChild, position image.Point) {
	child.placement.PlaceOffset(position)
	stack := op.Offset(position).Push(gtx.Ops)
	child.call.Add(gtx.Ops)
	stack.Pop()
}
