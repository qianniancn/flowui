package menu

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const menuItemPressDuration = 250 * time.Millisecond

func menuItemScale(gtx layout.Context, history []widget.Press, activeTheme *theme.Theme, disabled bool) float32 {
	target := min(max(activeTheme.Components.Menu.PressedScale, 0), 1)
	if target == 0 {
		target = 0.98
	}
	return optionrow.PressScale(gtx, history, disabled, target, menuItemPressDuration, menuItemPressDuration, activeTheme.Motion)
}
