package ui

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
)

// VisualOverflowDraw paints a decoration in the child's local coordinates.
type VisualOverflowDraw = layoutui.VisualOverflowDraw

// LayoutVisualOverflow paints draw behind the child without changing its measured size.
// It escapes clips created by child, while clips from ancestor widgets remain.
func LayoutVisualOverflow(ctx *Context, gtx layout.Context, child Widget, draw VisualOverflowDraw) layout.Dimensions {
	return layoutui.LayoutVisualOverflow(ctx, gtx, child, draw)
}
