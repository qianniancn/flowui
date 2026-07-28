package togglebutton

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
)

func (b ToggleButtonWidget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	selected := b.selected
	click := interact.Begin(ctx, gtx, b.key, nil, !b.disabled, true, func() {
		if b.onChange != nil {
			selected = !b.selected
			b.onChange(selected)
		}
	})
	styleState := click.StyleState
	styleState.Selected = selected
	styleState.Checked = selected
	style := b.resolveStyle(ctx, gtx, click.Key, styleState)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style.content, b.child)
		})
	})

	return layoutui.LayoutInteractiveResolved(ctx, gtx, style.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.SelectedOp(selected).Add(gtx.Ops)
			return visual(gtx)
		}, b.label)
	})
}
