package input

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func (t TextAreaWidget) layoutFrame(ctx *frame.Context, gtx layout.Context, state *textAreaState, resolved flowstyle.ResolvedStyle, enabled bool, child layout.Widget) layout.Dimensions {
	return layoutFieldFrame(ctx, gtx, &state.State, resolved, enabled, child)
}
