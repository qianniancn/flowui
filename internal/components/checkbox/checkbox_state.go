package checkbox

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotCheckbox = "checkbox"

func checkboxStateFor(ctx *frame.Context, key string) *checkboxState {
	return frame.UseState[checkboxState](ctx, key, stateSlotCheckbox)
}

type checkboxState struct {
	SelectionAnimation
	focus state.FocusAnimation
}

// SelectionAnimation drives the shared checkbox fill and check-path progress.
type SelectionAnimation struct {
	transition animation.FloatTransition
}

func (s *checkboxState) selection(gtx layout.Context, checked bool) float32 {
	return s.SelectionAnimation.Progress(gtx, checked)
}

func (s *SelectionAnimation) Progress(gtx layout.Context, checked bool) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.transition.Value(gtx, target, checkboxSelectDuration, animation.EaseSmoothstep)
}

func (s *checkboxState) focusOpacity(gtx layout.Context, focused bool) float32 {
	return s.focus.Opacity(gtx, focused)
}

func (s *checkboxState) focusVisible(focused bool, history []widget.Press) bool {
	return s.focus.Visible(focused, history)
}
