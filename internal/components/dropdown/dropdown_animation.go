package dropdown

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const dropdownPressDuration = 250 * time.Millisecond

func dropdownPressScale(gtx layout.Context, history []widget.Press, target float32, disabled bool) float32 {
	target = min(max(target, 0), 1)
	if target == 0 {
		target = 0.98
	}
	return optionrow.PressScale(gtx, history, disabled, target, dropdownPressDuration, dropdownPressDuration)
}

func dropdownTriggerScale(gtx layout.Context, history []widget.Press, activeTheme *theme.Theme, disabled bool) float32 {
	return dropdownPressScale(gtx, history, activeTheme.Components.Dropdown.TriggerPressedScale, disabled)
}
