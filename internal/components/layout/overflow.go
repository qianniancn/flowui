package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
)

// VisualOverflowDraw paints a decoration in the child's local coordinates.
type VisualOverflowDraw func(*frame.Context, layout.Context, image.Rectangle)

// LayoutVisualOverflow paints draw behind the child without changing its measured size.
// It escapes clips created by the child, while clips from ancestor widgets remain.
func LayoutVisualOverflow(ctx *frame.Context, gtx layout.Context, child frame.Widget, draw VisualOverflowDraw) layout.Dimensions {
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

// LayoutVisualOutset declares paint that may extend beyond child without
// changing its measured size. Scroll and other clipping containers use the
// declaration to reserve room within their viewport on subsequent frames.
func LayoutVisualOutset(ctx *frame.Context, gtx layout.Context, child frame.Widget, top, right, bottom, left unit.Dp) layout.Dimensions {
	frame.ReportVisualOverflow(ctx, frame.VisualOutset{
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
