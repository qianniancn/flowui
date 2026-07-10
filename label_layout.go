package flowui

import (
	"gioui.org/font"
	"gioui.org/layout"
)

func (l LabelWidget) layoutContent(ctx *Context, gtx layout.Context, style labelStyle) layout.Dimensions {
	theme := ctx.Theme.Components.Label
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(l.text).
				Size(float32(theme.TextSize)).
				Weight(font.Medium).
				Color(style.text).
				Layout(ctx, gtx)
		}),
	}
	if l.required {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text("*").
				Size(float32(theme.TextSize)).
				Weight(font.Medium).
				Color(style.required).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Baseline,
		Gap:       gtx.Dp(theme.RequiredMarkOffset),
	}.Layout(gtx, children...)
}
