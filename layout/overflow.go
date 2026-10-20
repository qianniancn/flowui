package layout

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
)

// VisualOverflowDraw paints a decoration in the child's local coordinates.
type VisualOverflowDraw func(*Context, layout.Context, image.Rectangle)

// VisualOverflow paints draw behind the child without changing its measured size.
// It escapes clips created by the child, while clips from ancestor widgets remain.
func VisualOverflow(ctx *Context, gtx layout.Context, child Widget, draw VisualOverflowDraw) layout.Dimensions {
	if child == nil {
		return layout.Dimensions{}
	}
	macro := op.Record(gtx.Ops)
	dims := child.Layout(ctx, gtx)
	content := macro.Stop()
	if draw != nil && dims.Size.X > 0 && dims.Size.Y > 0 {
		drawGtx := gtx
		drawGtx.Constraints = layout.Exact(dims.Size)
		draw(ctx, drawGtx, image.Rectangle{Max: dims.Size})
	}
	content.Add(gtx.Ops)
	return dims
}

// VisualOutset declares paint that may extend beyond child without
// changing its measured size. Scroll and other clipping containers use the
// declaration to reserve room within their viewport on subsequent frames.
func VisualOutset(ctx *Context, gtx layout.Context, child Widget, top, right, bottom, left unit.Dp) layout.Dimensions {
	core.ReportVisualOverflow(ctx, core.VisualOutset{
		Top:    max(gtx.Dp(top), 0),
		Right:  max(gtx.Dp(right), 0),
		Bottom: max(gtx.Dp(bottom), 0),
		Left:   max(gtx.Dp(left), 0),
	})
	if child == nil {
		return layout.Dimensions{}
	}
	return child.Layout(ctx, gtx)
}
