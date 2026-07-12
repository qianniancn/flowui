package alert

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/flowui-icons-lucide"
)

func (a Widget) layout(ctx *frame.Context, gtx layout.Context, style alertStyle) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Alert
	gtx.Constraints.Min.X = gtx.Constraints.Max.X

	macro := op.Record(gtx.Ops)
	contentDims := layout.Inset{
		Top:    tokens.PaddingY,
		Right:  tokens.PaddingX,
		Bottom: tokens.PaddingY,
		Left:   tokens.PaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 3)
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutIndicator(ctx, gtx, style)
		}))
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.layoutContent(ctx, gtx, style)
		}))
		if a.action != nil {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				restore := frame.PushColors(ctx, style.foreground, style.background)
				defer restore()
				return a.action.Layout(ctx, gtx)
			}))
		}
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Start,
			Gap:       gtx.Dp(tokens.Gap),
		}.Layout(gtx, children...)
	})
	content := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(tokens.Radius), 0), min(size.X, size.Y)/2)
	drawAlertSurface(gtx, activeTheme, rect, radius, style.background)
	root := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	if a.title != "" {
		semantic.LabelOp(a.title).Add(gtx.Ops)
	}
	if a.description != "" && a.content == nil {
		semantic.DescriptionOp(a.description).Add(gtx.Ops)
	}
	content.Add(gtx.Ops)
	root.Pop()
	return layout.Dimensions{Size: size}
}

func (a Widget) layoutIndicator(ctx *frame.Context, gtx layout.Context, style alertStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Alert
	return layout.Inset{
		Top:    tokens.IndicatorPadding,
		Right:  tokens.IndicatorPadding,
		Bottom: tokens.IndicatorPadding,
		Left:   tokens.IndicatorPadding,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.indicator != nil {
			restore := frame.PushColors(ctx, style.indicator, style.background)
			defer restore()
			return a.indicator.Layout(ctx, gtx)
		}
		size := gtx.Dp(tokens.IconSize)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(size, size))
		return icon.Layout(alertIcon(a.status), iconGtx, style.indicator)
	})
}

func (a Widget) layoutContent(ctx *frame.Context, gtx layout.Context, style alertStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Alert
	children := make([]layout.FlexChild, 0, 2)
	if a.title != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(frame.ActiveTheme(ctx).Material, tokens.TitleSize, a.title)
			label.Color = style.title
			label.Font.Weight = font.Medium
			label.LineHeight = tokens.TitleLineHeight
			return layoutAlertLineBox(gtx, tokens.TitleLineHeight, label.Layout)
		}))
	}
	if a.content != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			restore := frame.PushColors(ctx, style.description, style.background)
			defer restore()
			return a.content.Layout(ctx, gtx)
		}))
	} else if a.description != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Label(frame.ActiveTheme(ctx).Material, tokens.DescriptionSize, a.description)
			label.Color = style.description
			label.LineHeight = tokens.DescriptionLineHeight
			return layoutAlertLineBox(gtx, tokens.DescriptionLineHeight, label.Layout)
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start}.Layout(gtx, children...)
}

func layoutAlertLineBox(gtx layout.Context, lineHeight unit.Sp, child layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints.Min.Y = 0
	dims := child(childGtx)
	call := macro.Stop()
	height := min(max(dims.Size.Y, gtx.Sp(lineHeight)), gtx.Constraints.Max.Y)
	offset := op.Offset(image.Pt(0, max((height-dims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	dims.Size.Y = height
	return dims
}

func alertIcon(status Status) []byte {
	switch status {
	case StatusSuccess:
		return lucide.CircleCheck
	case StatusWarning:
		return lucide.TriangleAlert
	case StatusDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
