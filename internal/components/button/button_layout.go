package button

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type buttonPreparedContent struct {
	call  op.CallOp
	dims  layout.Dimensions
	width int
}

func (b ButtonWidget) prepareContent(ctx *frame.Context, gtx layout.Context) buttonPreparedContent {
	style := buttonSizeStyle(frame.ActiveTheme(ctx), b.size, b.iconOnly)
	height := min(gtx.Dp(style.height), gtx.Constraints.Max.Y)
	colors := buttonColors(frame.ActiveTheme(ctx), b.variant)
	style.fg = colors.fg
	style.bg = colors.bg
	if b.disabled {
		style.fg = frame.ActiveTheme(ctx).DisabledColor(style.fg)
		style.bg = frame.ActiveTheme(ctx).DisabledColor(style.bg)
	}
	if b.loading && !b.iconOnly && !b.fullWidth {
		style.inset = buttonLoadingInset(gtx, frame.ActiveTheme(ctx), b.size, style.inset)
	}
	child := b.styleChild(style)
	measureGtx := gtx
	if b.disabled || b.loading {
		measureGtx = measureGtx.Disabled()
	}
	measureGtx.Constraints.Min = image.Point{}
	measureGtx.Constraints.Max.Y = height
	macro := op.Record(gtx.Ops)
	dims := style.inset.Layout(measureGtx, func(gtx layout.Context) layout.Dimensions {
		restore := frame.PushColors(ctx, style.fg, style.bg)
		defer restore()
		return b.layoutContent(ctx, gtx, style, child)
	})
	call := macro.Stop()
	width := dims.Size.X
	if b.iconOnly {
		width = height
	}
	return buttonPreparedContent{call: call, dims: dims, width: width}
}

func (b ButtonWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style buttonStyle, child frame.Widget) layout.Dimensions {
	if !b.loading {
		return child.Layout(ctx, gtx)
	}
	spinner := func(gtx layout.Context) layout.Dimensions {
		return drawButtonSpinner(gtx, buttonSpinnerSize(frame.ActiveTheme(ctx), b.size), style.fg)
	}
	if b.iconOnly {
		return spinner(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(frame.ActiveTheme(ctx).Components.Button.ContentGap),
	}.Layout(gtx,
		layout.Rigid(spinner),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		}),
	)
}

func (b ButtonWidget) styleChild(style buttonStyle) frame.Widget {
	child := b.child
	text, ok := child.(text.Widget)
	if !ok {
		return child
	}
	text = text.DefaultColor(style.fg)
	text = text.DefaultSize(style.textSize)
	text = text.DefaultWeight(font.Medium)
	return text
}
