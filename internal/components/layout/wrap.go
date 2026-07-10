package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type WrapWidget struct {
	children []frame.Widget
	gap      unit.Dp
	lineGap  unit.Dp
	align    layout.Alignment
}

func Wrap(children ...frame.Widget) WrapWidget {
	return WrapWidget{children: children}
}

func (w WrapWidget) Gap(dp int) WrapWidget {
	w.gap = unit.Dp(dp)
	w.lineGap = unit.Dp(dp)
	return w
}

func (w WrapWidget) LineGap(dp int) WrapWidget {
	w.lineGap = unit.Dp(dp)
	return w
}

func (w WrapWidget) AlignStart() WrapWidget {
	w.align = layout.Start
	return w
}

func (w WrapWidget) AlignMiddle() WrapWidget {
	w.align = layout.Middle
	return w
}

func (w WrapWidget) AlignEnd() WrapWidget {
	w.align = layout.End
	return w
}

func (w WrapWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, w.children...)
	gap := gtx.Dp(w.gap)
	lineGap := gtx.Dp(w.lineGap)
	rows := make([]wrapRow, 0)
	maxWidth := gtx.Constraints.Max.X
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	var x, y, rowHeight, width int
	var row wrapRow
	for _, child := range w.children {
		macro := op.Record(gtx.Ops)
		dims := child.Layout(ctx, childGtx)
		call := macro.Stop()

		if x > 0 && x+gap+dims.Size.X > maxWidth {
			row.width = x
			rows = append(rows, row)
			y += rowHeight + lineGap
			x = 0
			rowHeight = 0
			row = wrapRow{}
		}
		if x > 0 {
			x += gap
		}
		row.children = append(row.children, wrapChild{
			call: call,
			pos:  image.Pt(x, y),
		})
		x += dims.Size.X
		width = max(width, x)
		rowHeight = max(rowHeight, dims.Size.Y)
	}
	if len(row.children) > 0 {
		row.width = x
		rows = append(rows, row)
	}

	size := gtx.Constraints.Constrain(image.Pt(width, y+rowHeight))
	for _, row := range rows {
		offset := alignmentOffset(w.align, size.X-row.width)
		for _, child := range row.children {
			trans := op.Offset(child.pos.Add(image.Pt(offset, 0))).Push(gtx.Ops)
			child.call.Add(gtx.Ops)
			trans.Pop()
		}
	}
	return layout.Dimensions{Size: size}
}

func alignmentOffset(align layout.Alignment, free int) int {
	if free <= 0 {
		return 0
	}
	switch align {
	case layout.Middle:
		return free / 2
	case layout.End:
		return free
	default:
		return 0
	}
}

type wrapRow struct {
	children []wrapChild
	width    int
}

type wrapChild struct {
	call op.CallOp
	pos  image.Point
}
