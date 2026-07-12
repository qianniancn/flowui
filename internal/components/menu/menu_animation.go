package menu

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const menuItemPressDuration = 250 * time.Millisecond

func menuItemScale(gtx layout.Context, history []widget.Press, activeTheme *theme.Theme, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	target := min(max(activeTheme.Components.Menu.PressedScale, 0), 1)
	if target == 0 {
		target = 0.98
	}
	press := history[len(history)-1]
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), menuItemPressDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), menuItemPressDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
}
