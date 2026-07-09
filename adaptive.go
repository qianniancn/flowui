package flowui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

type ViewSize struct {
	Width  unit.Dp
	Height unit.Dp
}

type AdaptiveWidget struct {
	view func(ViewSize) Widget
}

func Adaptive(view func(ViewSize) Widget) AdaptiveWidget {
	return AdaptiveWidget{view: view}
}

func (a AdaptiveWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	size := ViewSize{
		Width:  gtx.Metric.PxToDp(gtx.Constraints.Max.X),
		Height: gtx.Metric.PxToDp(gtx.Constraints.Max.Y),
	}
	return a.view(size).Layout(ctx, gtx)
}
