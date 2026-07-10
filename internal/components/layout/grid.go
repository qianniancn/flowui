package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type GridWidget struct {
	children       []frame.Widget
	columns        int
	minColumnWidth unit.Dp
	columnGap      unit.Dp
	rowGap         unit.Dp
}

func Grid(columns int, children ...frame.Widget) GridWidget {
	if columns < 1 {
		panic("flowui: grid columns must be positive")
	}
	return GridWidget{
		children: children,
		columns:  columns,
	}
}

func AutoGrid(minColumnWidth int, children ...frame.Widget) GridWidget {
	return GridWidget{
		children:       children,
		columns:        1,
		minColumnWidth: unit.Dp(minColumnWidth),
	}
}

func (g GridWidget) Gap(dp int) GridWidget {
	gap := unit.Dp(dp)
	g.columnGap = gap
	g.rowGap = gap
	return g
}

func (g GridWidget) ColumnGap(dp int) GridWidget {
	g.columnGap = unit.Dp(dp)
	return g
}

func (g GridWidget) RowGap(dp int) GridWidget {
	g.rowGap = unit.Dp(dp)
	return g
}

func (g GridWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, g.children...)
	if len(g.children) == 0 {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Point{})}
	}

	columnGap := gtx.Dp(g.columnGap)
	rowGap := gtx.Dp(g.rowGap)
	columns := g.columnsFor(gtx, columnGap)
	cellWidth := gridCellWidth(gtx.Constraints.Max.X, columns, columnGap)
	childGtx := gtx
	childGtx.Constraints.Min.X = cellWidth
	childGtx.Constraints.Max.X = cellWidth

	children := make([]gridChild, 0, len(g.children))
	var y int
	for rowStart := 0; rowStart < len(g.children); rowStart += columns {
		rowEnd := min(rowStart+columns, len(g.children))
		rowHeight := 0
		for i := rowStart; i < rowEnd; i++ {
			macro := op.Record(gtx.Ops)
			dims := g.children[i].Layout(ctx, childGtx)
			call := macro.Stop()
			rowHeight = max(rowHeight, dims.Size.Y)
			col := i - rowStart
			children = append(children, gridChild{
				call: call,
				pos:  image.Pt(col*(cellWidth+columnGap), y),
			})
		}
		y += rowHeight
		if rowEnd < len(g.children) {
			y += rowGap
		}
	}

	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, y))
	for _, child := range children {
		trans := op.Offset(child.pos).Push(gtx.Ops)
		child.call.Add(gtx.Ops)
		trans.Pop()
	}
	return layout.Dimensions{Size: size}
}

func (g GridWidget) columnsFor(gtx layout.Context, gap int) int {
	if g.minColumnWidth <= 0 {
		return g.columns
	}
	minWidth := max(gtx.Dp(g.minColumnWidth), 1)
	columns := (gtx.Constraints.Max.X + gap) / (minWidth + gap)
	return max(columns, 1)
}

func gridCellWidth(width, columns, gap int) int {
	space := width - gap*(columns-1)
	if space < 0 {
		return 0
	}
	return space / columns
}

type gridChild struct {
	call op.CallOp
	pos  image.Point
}
