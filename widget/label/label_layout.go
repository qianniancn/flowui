package label

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	textui "github.com/qianniancn/flowui/widget/text"
)

type labelResolvedStyle struct {
	root     style.ResolvedStyle
	required style.ResolvedStyle
}

func (l Widget) layoutContent(
	ctx *core.Context,
	gtx layout.Context,
	resolved labelResolvedStyle,
	requiredMarkGap unit.Dp,
) layout.Dimensions {
	textStyle := style.TextDeclaration(resolved.root.Text)
	if resolved.required.Text != nil {
		textStyle = textStyle.TextColor(resolved.required.Text.Color)
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return textui.New(l.text).
				Style(style.TextDeclaration(resolved.root.Text)).
				Layout(ctx, gtx)
		}),
	}
	if l.required {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return textui.New("*").
				Style(textStyle).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Baseline,
		Gap:       gtx.Dp(requiredMarkGap),
	}.Layout(gtx, children...)
}
