package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
)

func LayoutItems(gtx layout.Context, horizontal bool, columnGap, rowGap int, children []layout.Widget) layout.Dimensions {
	if !horizontal {
		flexChildren := make([]layout.FlexChild, 0, len(children))
		for _, child := range children {
			flexChildren = append(flexChildren, layout.Rigid(child))
		}
		return layout.Flex{Axis: layout.Vertical, Gap: rowGap}.Layout(gtx, flexChildren...)
	}

	rows := make([]itemRow, 0)
	maxWidth := gtx.Constraints.Max.X
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	var x, y, rowHeight, width int
	var row itemRow
	for _, child := range children {
		macro := op.Record(gtx.Ops)
		dims := child(childGtx)
		call := macro.Stop()

		if x > 0 && x+columnGap+dims.Size.X > maxWidth {
			rows = append(rows, row)
			y += rowHeight + rowGap
			x = 0
			rowHeight = 0
			row = itemRow{}
		}
		if x > 0 {
			x += columnGap
		}
		row.children = append(row.children, itemChild{call: call, pos: image.Pt(x, y)})
		x += dims.Size.X
		width = max(width, x)
		rowHeight = max(rowHeight, dims.Size.Y)
	}
	if len(row.children) > 0 {
		rows = append(rows, row)
	}

	size := gtx.Constraints.Constrain(image.Pt(width, y+rowHeight))
	for _, row := range rows {
		for _, child := range row.children {
			trans := op.Offset(child.pos).Push(gtx.Ops)
			child.call.Add(gtx.Ops)
			trans.Pop()
		}
	}
	return layout.Dimensions{Size: size}
}

type itemRow struct {
	children []itemChild
}

type itemChild struct {
	call op.CallOp
	pos  image.Point
}
