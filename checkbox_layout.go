package flowui

import (
	"gioui.org/font"
	"gioui.org/layout"
)

func (c CheckboxWidget) layoutContent(ctx *Context, gtx layout.Context, style checkboxStyle) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return drawCheckbox(gtx, ctx.Theme, style)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(ctx.Theme.Components.Checkbox.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(c.label).
				Size(float32(ctx.Theme.Typography.ControlSize)).
				Color(style.fg).
				Weight(font.Medium).
				Layout(ctx, gtx)
		}),
	)
}
