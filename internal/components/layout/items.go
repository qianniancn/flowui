package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func LayoutItems(ctx *frame.Context, gtx layout.Context, horizontal bool, columnGap, rowGap int, children []layout.Widget) layout.Dimensions {
	columnGap = max(columnGap, 0)
	rowGap = max(rowGap, 0)
	if !horizontal {
		type trackedChild struct {
			dims      layout.Dimensions
			placement frame.OverlayPlacement
		}
		tracked := make([]trackedChild, len(children))
		flexChildren := make([]layout.FlexChild, 0, len(children))
		for index, child := range children {
			flexChildren = append(flexChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
					return child(gtx)
				})
				tracked[index] = trackedChild{dims: dims, placement: placement}
				return dims
			}))
		}
		gap := max(rowGap, 0)
		dims := layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, flexChildren...)
		y := 0
		for index, child := range tracked {
			child.placement.PlaceOffset(image.Pt(0, y))
			y += child.dims.Size.Y
			if index < len(tracked)-1 {
				y += gap
			}
		}
		return dims
	}

	rows := make([]itemRow, 0)
	maxWidth := gtx.Constraints.Max.X
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	var x, y, rowHeight, width int
	var row itemRow
	for _, child := range children {
		macro := op.Record(gtx.Ops)
		dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(childGtx)
		})
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
		row.children = append(row.children, itemChild{call: call, pos: image.Pt(x, y), placement: placement})
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
			child.placement.PlaceOffset(child.pos)
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
	call      op.CallOp
	pos       image.Point
	placement frame.OverlayPlacement
}
