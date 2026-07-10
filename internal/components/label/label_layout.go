package label

import (
	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (l LabelWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style labelStyle) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Label
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(l.text).
				Size(float32(theme.TextSize)).
				Weight(font.Medium).
				Color(style.text).
				Layout(ctx, gtx)
		}),
	}
	if l.required {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New("*").
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
