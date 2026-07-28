package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func layoutInputGroupDivider(ctx *frame.Context, gtx layout.Context, x, height, width int, style flowstyle.ResolvedStyle) {
	if width <= 0 || height <= 0 {
		return
	}
	dividerGtx := gtx
	dividerGtx.Constraints = layout.Exact(image.Pt(width, height))
	stack := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	layoutui.LayoutResolved(ctx, dividerGtx, style, nil)
	stack.Pop()
}
