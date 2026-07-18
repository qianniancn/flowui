package togglebutton

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotToggleButton = "toggle-button"

type toggleButtonState struct {
	clickable            widget.Clickable
	focus                state.FocusAnimation
	backgroundTransition animation.ColorTransition
	scale                animation.FloatTransition
}

func toggleButtonStateFor(ctx *frame.Context, key string) *toggleButtonState {
	key = frame.ClaimKey(ctx, state.KindToggleButton, key)
	return frame.UseState[toggleButtonState](ctx, key, stateSlotToggleButton)
}

func (s *toggleButtonState) animateBackground(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, toggleButtonColorDuration, animation.EaseSmoothstep, motions...)
}

func (s *toggleButtonState) animateScale(gtx layout.Context, target float32, motions ...theme.MotionTheme) float32 {
	return s.scale.Value(gtx, target, toggleButtonScaleDuration, animation.EaseSmoothstep, motions...)
}
