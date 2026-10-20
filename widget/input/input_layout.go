package input

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

func (i Widget) layoutFrame(ctx *core.Context, gtx layout.Context, state *inputState, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.State, resolved, enabled, func(gtx layout.Context) layout.Dimensions {
		minWidth := gtx.Constraints.Min.X
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = min(minWidth, gtx.Constraints.Max.X)
			return child(gtx)
		})
	})
}

// layoutFieldFrame hosts a resolved field box and its interaction area.
//
// The frame used to take a draw callback so a replaced visual could wrap the
// default one. Both of the widget's call sites passed nil, and the third source
// of a callback -- the inputVisual.Draw slot -- was a guard whose only possible
// value was nil, so the parameter moved through three functions without ever
// being non-nil. It is gone, and the frame renders through layout's ordinary
// resolved-interaction entry point like the compound group already did.
func layoutFieldFrame(ctx *core.Context, gtx layout.Context, state *field.State, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutui.InteractiveResolved(
		ctx,
		gtx,
		resolved,
		core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions { return child(gtx) }),
		func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			dims := visual(gtx)
			state.AddPointer(gtx, dims.Size, !enabled)
			return dims
		},
	)
}
