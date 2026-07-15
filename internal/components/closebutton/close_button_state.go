package closebutton

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotCloseButton = "close-button"

type closeButtonState struct {
	focus                state.FocusAnimation
	backgroundTransition animation.ColorTransition
	scaleTransition      animation.FloatTransition
}

func closeButtonStateFor(ctx *frame.Context, key string) *closeButtonState {
	return frame.UseState[closeButtonState](ctx, key, stateSlotCloseButton)
}

func (s *closeButtonState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, closeButtonColorDuration, animation.EaseSmoothstep)
}

func closeButtonPressedScale(pressedScale float32) float32 {
	if pressedScale <= 0 || pressedScale > 1 {
		return 0.93
	}
	return pressedScale
}

func (s *closeButtonState) scale(gtx layout.Context, target float32) float32 {
	return s.scaleTransition.Value(gtx, target, closeButtonScaleDuration, animation.EaseQuarticOut)
}
