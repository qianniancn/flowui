package togglebutton

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/interaction"
)

func (b ToggleButtonWidget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	click := interaction.BeginClick(ctx, gtx, b.key, nil, !b.disabled, true, func() {
		if b.onChange != nil {
			b.onChange(!b.selected)
		}
	})
	styleState := click.StyleState
	styleState.Selected = b.selected
	styleState.Checked = b.selected
	style := b.resolveStyle(ctx, gtx, click.Key, styleState)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style.content, b.child)
		})
	})

	return layoutui.LayoutInteractiveResolved(ctx, gtx, style.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.SelectedOp(b.selected).Add(gtx.Ops)
			return visual(gtx)
		}, b.label)
	})
}
