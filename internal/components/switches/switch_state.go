package switches

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotSwitch = "switch"

func switchStateFor(ctx *frame.Context, key string) *switchState {
	key = frame.ClaimKey(ctx, state.KindSwitch, key)
	return frame.UseState[switchState](ctx, key, stateSlotSwitch)
}

type switchState struct {
	value    widget.Bool
	selected animation.FloatTransition
	focus    state.FocusAnimation
}

func (s *switchState) selection(gtx layout.Context, checked bool) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return s.selected.Value(gtx, target, switchSelectDuration, animation.EaseSmoothstep)
}

func (s *switchState) focusOpacity(gtx layout.Context, focused bool) float32 {
	return s.focus.Opacity(gtx, focused)
}

func (s *switchState) focusVisible(focused bool, history []widget.Press) bool {
	return s.focus.Visible(focused, history)
}
