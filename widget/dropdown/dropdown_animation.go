package dropdown

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/optionrow"
)

func dropdownPressScale(gtx layout.Context, history []widget.Press, target float32, disabled bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target = min(max(target, 0), 1)
	if target == 0 {
		target = 0.98
	}
	return optionrow.PressScale(gtx, history, disabled, target, duration, duration, motions...)
}

func dropdownTriggerScale(gtx layout.Context, history []widget.Press, visual theme.DropdownVisual, disabled bool) float32 {
	return dropdownPressScale(gtx, history, visual.TriggerPressedScale, disabled, visual.PressDuration, visual.Motion)
}
