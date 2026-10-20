package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

type GridWidget struct {
	children       []Widget
	columns        int
	minColumnWidth unit.Dp
	columnGap      unit.Dp
	rowGap         unit.Dp
}

func Grid(columns int, children ...Widget) GridWidget {
	if columns < 1 {
		panic("flowui: grid columns must be positive")
	}
	return GridWidget{
		children: children,
		columns:  columns,
	}
}

func AutoGrid(minColumnWidth int, children ...Widget) GridWidget {
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

// GapDp sets both grid gaps without converting a runtime Dp value to int.
func (g GridWidget) GapDp(dp unit.Dp) GridWidget {
	g.columnGap = dp
	g.rowGap = dp
	return g
}

func (g GridWidget) ColumnGap(dp int) GridWidget {
	g.columnGap = unit.Dp(dp)
	return g
}

// ColumnGapDp is the dynamic-value counterpart to ColumnGap.
func (g GridWidget) ColumnGapDp(dp unit.Dp) GridWidget {
	g.columnGap = dp
	return g
}

func (g GridWidget) RowGap(dp int) GridWidget {
	g.rowGap = unit.Dp(dp)
	return g
}

// RowGapDp is the dynamic-value counterpart to RowGap.
func (g GridWidget) RowGapDp(dp unit.Dp) GridWidget {
	g.rowGap = dp
	return g
}

func (g GridWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, g.children...)
	if len(g.children) == 0 {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Point{})}
	}

	columnGap := max(gtx.Dp(g.columnGap), 0)
	rowGap := max(gtx.Dp(g.rowGap), 0)
	columns := g.columnsFor(gtx, columnGap)
	cellWidth := gridCellWidth(gtx.Constraints.Max.X, columns, columnGap)
	childGtx := gtx
	childGtx.Constraints.Min.Y = 0
	childGtx.Constraints.Min.X = cellWidth
	childGtx.Constraints.Max.X = cellWidth

	children := make([]gridChild, 0, len(g.children))
	var y int
	for rowStart := 0; rowStart < len(g.children); rowStart += columns {
		rowEnd := min(rowStart+columns, len(g.children))
		rowHeight := 0
		for i := rowStart; i < rowEnd; i++ {
			macro := op.Record(gtx.Ops)
			dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return g.children[i].Layout(ctx, childGtx)
			})
			call := macro.Stop()
			rowHeight = max(rowHeight, dims.Size.Y)
			col := i - rowStart
			children = append(children, gridChild{
				call:      call,
				pos:       image.Pt(col*(cellWidth+columnGap), y),
				placement: placement,
			})
		}
		y += rowHeight
		if rowEnd < len(g.children) {
			y += rowGap
		}
	}

	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, y))
	for _, child := range children {
		child.placement.PlaceOffset(child.pos)
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
	call      op.CallOp
	pos       image.Point
	placement core.OverlayPlacement
}
