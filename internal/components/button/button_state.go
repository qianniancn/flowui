package button

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
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
	bg           color.NRGBA
	bgFrom       color.NRGBA
	bgTo         color.NRGBA
	bgStart      time.Time
	bgReady      bool
	focus        float32
	focusFrom    float32
	focusTo      float32
	focusAt      time.Time
	focusReady   bool
	focused      bool
	pointerFocus bool
}

func (s *buttonState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	if !s.bgReady {
		s.bg = target
		s.bgFrom = target
		s.bgTo = target
		s.bgStart = gtx.Now
		s.bgReady = true
		return target
	}
	if target != s.bgTo {
		s.bgFrom = s.bg
		s.bgTo = target
		s.bgStart = gtx.Now
	}
	if s.bgFrom == s.bgTo {
		s.bg = s.bgTo
		return s.bg
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.bgStart), buttonColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.bg = render.LerpColor(s.bgFrom, s.bgTo, progress)
	return s.bg
}

func (s *buttonState) focusOpacity(gtx layout.Context, focused bool) float32 {
	target := float32(0)
	if focused {
		target = 1
	}
	if !s.focusReady {
		s.focus = target
		s.focusFrom = target
		s.focusTo = target
		s.focusAt = gtx.Now
		s.focusReady = true
		return target
	}
	if target != s.focusTo {
		s.focusFrom = s.focus
		s.focusTo = target
		s.focusAt = gtx.Now
	}
	if s.focusFrom == s.focusTo {
		s.focus = s.focusTo
		return s.focus
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.focusAt), buttonFocusDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.focus = render.Lerp(s.focusFrom, s.focusTo, progress)
	return s.focus
}

func (s *buttonState) focusVisible(focused bool, history []widget.Press) bool {
	if !focused {
		s.focused = false
		s.pointerFocus = false
		return false
	}
	if !s.focused {
		s.focused = true
		s.pointerFocus = len(history) > 0
	}
	return !s.pointerFocus
}
