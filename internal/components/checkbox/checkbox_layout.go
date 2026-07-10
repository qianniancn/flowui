package checkbox

import (
	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (c CheckboxWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style checkboxStyle) layout.Dimensions {
	control := func(gtx layout.Context) layout.Dimensions {
		return drawCheckbox(gtx, frame.ActiveTheme(ctx), style)
	}
	if c.label == "" {
		return control(gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(frame.ActiveTheme(ctx).Components.Checkbox.LabelGap),
	}.Layout(gtx,
		layout.Rigid(control),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(c.label).
				Size(float32(frame.ActiveTheme(ctx).Typography.ControlSize)).
				Color(style.fg).
				Weight(font.Medium).
				Layout(ctx, gtx)
		}),
	)
}
