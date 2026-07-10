package button

import (
	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

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
