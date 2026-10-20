package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

type WrapWidget struct {
	children []Widget
	gap      unit.Dp
	lineGap  unit.Dp
	align    layout.Alignment
}

func Wrap(children ...Widget) WrapWidget {
	return WrapWidget{children: children}
}

func (w WrapWidget) Gap(dp int) WrapWidget {
	w.gap = unit.Dp(dp)
	w.lineGap = unit.Dp(dp)
	return w
}

// GapDp sets both wrap gaps while preserving fractional Dp values.
func (w WrapWidget) GapDp(dp unit.Dp) WrapWidget {
	w.gap = dp
	w.lineGap = dp
	return w
}

func (w WrapWidget) LineGap(dp int) WrapWidget {
	w.lineGap = unit.Dp(dp)
	return w
}

// LineGapDp is the dynamic-value counterpart to LineGap.
func (w WrapWidget) LineGapDp(dp unit.Dp) WrapWidget {
	w.lineGap = dp
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

func (w WrapWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, w.children...)
	gap := max(gtx.Dp(w.gap), 0)
	lineGap := max(gtx.Dp(w.lineGap), 0)
	var inlineChildren [16]wrapChild
	children := inlineChildren[:0]
	if len(w.children) > len(inlineChildren) {
		children = make([]wrapChild, 0, len(w.children))
	}
	maxWidth := gtx.Constraints.Max.X
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}

	var x, y, rowHeight, width, rowIndex int
	for _, child := range w.children {
		macro := op.Record(gtx.Ops)
		dims, placement := core.TrackWidgetPlacement(ctx, childGtx, child)
		call := macro.Stop()

		if x > 0 && x+gap+dims.Size.X > maxWidth {
			y += rowHeight + lineGap
			x = 0
			rowHeight = 0
			rowIndex++
		}
		if x > 0 {
			x += gap
		}
		children = append(children, wrapChild{
			call:      call,
			pos:       image.Pt(x, y),
			size:      dims.Size,
			row:       rowIndex,
			placement: placement,
		})
		x += dims.Size.X
		width = max(width, x)
		rowHeight = max(rowHeight, dims.Size.Y)
	}

	size := gtx.Constraints.Constrain(image.Pt(width, y+rowHeight))
	for start := 0; start < len(children); {
		end := start + 1
		for end < len(children) && children[end].row == children[start].row {
			end++
		}
		last := children[end-1]
		rowWidth := last.pos.X + last.size.X
		offset := alignmentOffset(w.align, size.X-rowWidth)
		for _, child := range children[start:end] {
			pos := child.pos.Add(image.Pt(offset, 0))
			child.placement.PlaceOffset(pos)
			trans := op.Offset(pos).Push(gtx.Ops)
			child.call.Add(gtx.Ops)
			trans.Pop()
		}
		start = end
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

type wrapChild struct {
	call      op.CallOp
	pos       image.Point
	size      image.Point
	row       int
	placement core.OverlayPlacement
}
