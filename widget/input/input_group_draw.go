package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
)

func layoutInputGroupDivider(ctx *core.Context, gtx layout.Context, x, height, width int, resolved style.ResolvedStyle) {
	if width <= 0 || height <= 0 {
		return
	}
	dividerGtx := gtx
	dividerGtx.Constraints = layout.Exact(image.Pt(width, height))
	stack := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	layoutui.Resolved(ctx, dividerGtx, resolved, nil)
	stack.Pop()
}
