package layoutui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
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
