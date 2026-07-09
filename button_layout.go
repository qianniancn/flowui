package flowui

import (
	"gioui.org/font"
	"gioui.org/layout"
)

func (b ButtonWidget) layoutContent(ctx *Context, gtx layout.Context, style buttonStyle, child Widget) layout.Dimensions {
	if !b.loading {
		return child.Layout(ctx, gtx)
	}
	spinner := func(gtx layout.Context) layout.Dimensions {
		return drawButtonSpinner(gtx, buttonSpinnerSize(ctx.Theme, b.size), style.fg)
	}
	if b.iconOnly {
		return spinner(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(ctx.Theme.Components.Button.ContentGap),
	}.Layout(gtx,
		layout.Rigid(spinner),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		}),
	)
}

func (b ButtonWidget) styleChild(style buttonStyle) Widget {
	child := b.child
	text, ok := child.(TextWidget)
	if !ok {
		return child
	}
	if !text.hasColor {
		text = text.Color(style.fg)
	}
	if text.size == 0 {
		text = text.Size(style.textSize)
	}
	if text.weight == 0 {
		text = text.Weight(font.Medium)
	}
	return text
}
