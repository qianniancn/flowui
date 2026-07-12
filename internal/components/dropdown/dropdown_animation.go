package dropdown

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const dropdownPressDuration = 250 * time.Millisecond

func dropdownPressScale(gtx layout.Context, history []widget.Press, target float32, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	target = min(max(target, 0), 1)
	if target == 0 {
		target = 0.98
	}
	press := history[len(history)-1]
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), dropdownPressDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), dropdownPressDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
}

func dropdownTriggerScale(gtx layout.Context, history []widget.Press, activeTheme *theme.Theme, disabled bool) float32 {
	return dropdownPressScale(gtx, history, activeTheme.Components.Dropdown.TriggerPressedScale, disabled)
}
