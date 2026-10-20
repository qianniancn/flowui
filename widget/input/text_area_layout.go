package input

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

func (t TextAreaWidget) layoutFrame(ctx *core.Context, gtx layout.Context, state *textAreaState, resolved style.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.State, resolved, enabled, child)
}
