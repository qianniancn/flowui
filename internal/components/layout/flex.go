package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type CenterWidget struct {
	child frame.Widget
}

func Center(child frame.Widget) CenterWidget {
	return CenterWidget{child: child}
}

func (c CenterWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, c.child)
	return layoutTrackedDirection(ctx, gtx, layout.Center, func(gtx layout.Context) layout.Dimensions {
		return c.child.Layout(ctx, gtx)
	})
}

type ColumnWidget struct {
	children []frame.Widget
	gap      unit.Dp
	align    layout.Alignment
}

func Column(children ...frame.Widget) ColumnWidget {
	return ColumnWidget{children: children}
}

func (c ColumnWidget) Gap(dp int) ColumnWidget {
	c.gap = unit.Dp(dp)
	return c
}

func (c ColumnWidget) AlignStart() ColumnWidget {
	c.align = layout.Start
	return c
}

func (c ColumnWidget) AlignMiddle() ColumnWidget {
	c.align = layout.Middle
	return c
}

func (c ColumnWidget) AlignEnd() ColumnWidget {
	c.align = layout.End
	return c
}

func (c ColumnWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, c.children...)
	return flexLayout(ctx, gtx, layout.Vertical, c.gap, c.align, c.children)
}

type RowWidget struct {
	children []frame.Widget
	gap      unit.Dp
	align    layout.Alignment
}

func Row(children ...frame.Widget) RowWidget {
	return RowWidget{children: children}
}

func (r RowWidget) Gap(dp int) RowWidget {
	r.gap = unit.Dp(dp)
	return r
}

func (r RowWidget) AlignStart() RowWidget {
	r.align = layout.Start
	return r
}

func (r RowWidget) AlignMiddle() RowWidget {
	r.align = layout.Middle
	return r
}

func (r RowWidget) AlignEnd() RowWidget {
	r.align = layout.End
	return r
}

func (r RowWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, r.children...)
	return flexLayout(ctx, gtx, layout.Horizontal, r.gap, r.align, r.children)
}

type FlexWidget struct {
	child  frame.Widget
	weight float32
}

func Expanded(child frame.Widget) FlexWidget {
	return FlexWidget{
		child:  child,
		weight: 1,
	}
}

func Flexible(weight float32, child frame.Widget) FlexWidget {
	return FlexWidget{
		child:  child,
		weight: weight,
	}
}

func (f FlexWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, f.child)
	return f.child.Layout(ctx, gtx)
}

func flexLayout(ctx *frame.Context, gtx layout.Context, axis layout.Axis, gap unit.Dp, align layout.Alignment, widgets []frame.Widget) layout.Dimensions {
	type childPlacement struct {
		dims      layout.Dimensions
		placement frame.OverlayPlacement
	}
	placements := make([]childPlacement, len(widgets))
	children := make([]layout.FlexChild, 0, len(widgets))
	for index, child := range widgets {
		index := index
		if flex, ok := child.(FlexWidget); ok {
			children = append(children, layout.Flexed(flex.weight, func(gtx layout.Context) layout.Dimensions {
				dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
					return flex.child.Layout(ctx, gtx)
				})
				placements[index] = childPlacement{dims: dims, placement: placement}
				return dims
			}))
			continue
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return child.Layout(ctx, gtx)
			})
			placements[index] = childPlacement{dims: dims, placement: placement}
			return dims
		}))
	}
	gapPx := max(gtx.Dp(gap), 0)
	dims := layout.Flex{
		Axis:      axis,
		Gap:       gapPx,
		Alignment: align,
	}.Layout(gtx, children...)

	maxCross := axisCrossMinimum(gtx.Constraints, axis)
	maxBaseline := 0
	for _, child := range placements {
		maxCross = max(maxCross, axis.Convert(child.dims.Size).Y)
		maxBaseline = max(maxBaseline, child.dims.Size.Y-child.dims.Baseline)
	}
	main := 0
	for index, child := range placements {
		size := axis.Convert(child.dims.Size)
		cross := 0
		switch align {
		case layout.End:
			cross = maxCross - size.Y
		case layout.Middle:
			cross = (maxCross - size.Y) / 2
		case layout.Baseline:
			if axis == layout.Horizontal {
				cross = maxBaseline - (child.dims.Size.Y - child.dims.Baseline)
			}
		}
		child.placement.PlaceOffset(axis.Convert(image.Pt(main, cross)))
		main += size.X
		if index < len(placements)-1 {
			main += gapPx
		}
	}
	return dims
}

func axisCrossMinimum(constraints layout.Constraints, axis layout.Axis) int {
	if axis == layout.Horizontal {
		return constraints.Min.Y
	}
	return constraints.Min.X
}
