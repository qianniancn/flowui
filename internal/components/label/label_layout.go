package label

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

func (l LabelWidget) layoutContent(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle, required color.NRGBA) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Label
	textStyle := flowstyle.TextDeclaration(resolved.Text)
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(l.text).
				Style(textStyle).
				Layout(ctx, gtx)
		}),
	}
	if l.required {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New("*").
				Style(textStyle.
					TextColor(flowstyle.SolidColor{Color: required}),
				).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Baseline,
		Gap:       gtx.Dp(theme.RequiredMarkOffset),
	}.Layout(gtx, children...)
}
