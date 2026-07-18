package button

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotButton = "button"

func buttonStateFor(ctx *frame.Context, key string) *buttonState {
	return frame.UseState[buttonState](ctx, key, stateSlotButton)
}

func buttonAnimationScale(gtx layout.Context, history []widget.Press, theme *theme.Theme, size ButtonSize, disabled bool) float32 {
	target := buttonPressedScale(theme, size)
	return optionrow.PressScale(gtx, history, disabled, target, buttonPressInDuration, buttonPressOutDuration, theme.Motion)
}

func buttonPressedScale(theme *theme.Theme, size ButtonSize) float32 {
	switch size {
	case ButtonSmall:
		return theme.Components.Button.PressedScaleSmall
	case ButtonLarge:
		return theme.Components.Button.PressedScaleLarge
	default:
		return theme.Components.Button.PressedScaleMedium
	}
}

type buttonState struct {
	backgroundTransition animation.ColorTransition
	focus                state.FocusAnimation
}

func (s *buttonState) background(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, buttonColorDuration, animation.EaseSmoothstep, motions...)
}

func (s *buttonState) focusOpacity(gtx layout.Context, focused bool, motions ...theme.MotionTheme) float32 {
	return s.focus.Opacity(gtx, focused, motions...)
}
