package togglebutton

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
)

func (b Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ToggleOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	selected := b.selected
	click := core.Begin(ctx, gtx, b.key, nil, !b.disabled, true, func() {
		if b.onChange != nil {
			selected = !b.selected
			b.onChange(selected)
		}
	})
	componentState := toggleButtonStateFor(ctx, click.Key)
	styleState := click.State
	styleState.Selected = selected
	styleState.Checked = selected
	style := b.resolveStyle(ctx, gtx, click.Key, styleState, override, &componentState.visualCache)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, style.content, b.child)
		})
	})

	return layoutui.InteractiveResolved(ctx, gtx, style.root, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.SelectedOp(selected).Add(gtx.Ops)
			return visual(gtx)
		}, b.label)
	})
}
