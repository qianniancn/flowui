package button

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotButton = "button"

func buttonStateFor(ctx *frame.Context, key string) *buttonState {
	return frame.UseState[buttonState](ctx, key, stateSlotButton)
}

func buttonAnimationScale(gtx layout.Context, history []widget.Press, theme *theme.Theme, size ButtonSize, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	press := history[len(history)-1]
	target := buttonPressedScale(theme, size)
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), buttonPressInDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), buttonPressOutDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
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

func (s *buttonState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, buttonColorDuration, animation.EaseSmoothstep)
}

func (s *buttonState) focusOpacity(gtx layout.Context, focused bool) float32 {
	return s.focus.Opacity(gtx, focused)
}

func (s *buttonState) focusVisible(focused bool, history []widget.Press) bool {
	return s.focus.Visible(focused, history)
}

func (s *buttonState) prepareFocus(visible bool) {
	s.focus.Prepare(visible)
}
