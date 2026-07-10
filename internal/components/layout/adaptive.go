package layoutui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ViewSize struct {
	Width  unit.Dp
	Height unit.Dp
}

type AdaptiveWidget struct {
	view func(ViewSize) frame.Widget
}

func Adaptive(view func(ViewSize) frame.Widget) AdaptiveWidget {
	return AdaptiveWidget{view: view}
}

func (a AdaptiveWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	size := ViewSize{
		Width:  gtx.Metric.PxToDp(gtx.Constraints.Max.X),
		Height: gtx.Metric.PxToDp(gtx.Constraints.Max.Y),
	}
	child := a.view(size)
	prepareFieldAssociations(ctx, child)
	return child.Layout(ctx, gtx)
}
