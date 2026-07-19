package closebutton

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotCloseButton = "close-button"

type State struct {
	focus                state.FocusAnimation
	backgroundTransition animation.ColorTransition
	scaleTransition      animation.FloatTransition
}

type closeButtonState = State

func closeButtonStateFor(ctx *frame.Context, key string) *closeButtonState {
	return frame.UseState[closeButtonState](ctx, key, stateSlotCloseButton)
}

func (s *State) background(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, closeButtonColorDuration, animation.EaseSmoothstep, motions...)
}

func closeButtonPressedScale(pressedScale float32) float32 {
	if pressedScale <= 0 || pressedScale > 1 {
		return 0.93
	}
	return pressedScale
}

func (s *State) scale(gtx layout.Context, target float32, motions ...theme.MotionTheme) float32 {
	return s.scaleTransition.Value(gtx, target, closeButtonScaleDuration, animation.EaseQuarticOut, motions...)
}
