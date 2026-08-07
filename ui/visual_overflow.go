package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
)

// VisualOverflowDraw paints a decoration in the child's local coordinates.
type VisualOverflowDraw = layoutui.VisualOverflowDraw

// LayoutVisualOverflow paints draw behind the child without changing its measured size.
// It escapes clips created by child, while clips from ancestor widgets remain.
func LayoutVisualOverflow(ctx *Context, gtx layout.Context, child Widget, draw VisualOverflowDraw) layout.Dimensions {
	return layoutui.LayoutVisualOverflow(ctx, gtx, child, draw)
}

// LayoutVisualOutset declares paint that may extend beyond child without
// changing its measured size. Clipping containers reserve the declared space
// within their viewport.
func LayoutVisualOutset(ctx *Context, gtx layout.Context, child Widget, top, right, bottom, left unit.Dp) layout.Dimensions {
	return layoutui.LayoutVisualOutset(ctx, gtx, child, top, right, bottom, left)
}
