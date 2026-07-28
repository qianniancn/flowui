package input

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

func (i InputWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *inputState, resolved flowstyle.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.State, resolved, enabled, func(gtx layout.Context) layout.Dimensions {
		minWidth := gtx.Constraints.Min.X
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = min(minWidth, gtx.Constraints.Max.X)
			return child(gtx)
		})
	})
}

func layoutFieldFrame(ctx *frame.Context, gtx layout.Context, state *field.State, resolved flowstyle.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutui.LayoutInteractiveResolved(
		ctx,
		gtx,
		resolved,
		frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions { return child(gtx) }),
		func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
			dims := visual(gtx)
			state.AddPointer(gtx, dims.Size, !enabled)
			return dims
		},
	)
}
