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
	view        func(ViewSize) frame.Widget
	breakpoints []adaptiveBreakpoint
}

type adaptiveBreakpoint struct {
	minWidth unit.Dp
	view     func(ViewSize) frame.Widget
}

func Adaptive(view func(ViewSize) frame.Widget) AdaptiveWidget {
	return AdaptiveWidget{view: view}
}

func (a AdaptiveWidget) AtLeastWidth(dp int, view func(ViewSize) frame.Widget) AdaptiveWidget {
	a.breakpoints = append(a.breakpoints, adaptiveBreakpoint{minWidth: unit.Dp(max(dp, 0)), view: view})
	return a
}

func (a AdaptiveWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	size := ViewSize{
		Width:  gtx.Metric.PxToDp(gtx.Constraints.Max.X),
		Height: gtx.Metric.PxToDp(gtx.Constraints.Max.Y),
	}
	view := a.view
	selectedWidth := unit.Dp(-1)
	for _, breakpoint := range a.breakpoints {
		if size.Width >= breakpoint.minWidth && breakpoint.minWidth >= selectedWidth {
			selectedWidth = breakpoint.minWidth
			view = breakpoint.view
		}
	}
	child := view(size)
	prepareFieldAssociations(ctx, child)
	return child.Layout(ctx, gtx)
}
