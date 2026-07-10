package layoutui

import (
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
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
	children := make([]layout.FlexChild, 0, len(widgets))
	for _, child := range widgets {
		if flex, ok := child.(FlexWidget); ok {
			children = append(children, layout.Flexed(flex.weight, func(gtx layout.Context) layout.Dimensions {
				return flex.child.Layout(ctx, gtx)
			}))
			continue
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      axis,
		Gap:       gtx.Dp(gap),
		Alignment: align,
	}.Layout(gtx, children...)
}
