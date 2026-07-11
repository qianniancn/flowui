package togglebutton

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotToggleButton = "toggle-button"

type toggleButtonState struct {
	clickable  widget.Clickable
	focus      state.FocusAnimation
	background color.NRGBA
	bgFrom     color.NRGBA
	bgTo       color.NRGBA
	bgAt       time.Time
	bgReady    bool
	scaleValue float32
	scaleFrom  float32
	scaleTo    float32
	scaleAt    time.Time
	scaleReady bool
}

func toggleButtonStateFor(ctx *frame.Context, key string) *toggleButtonState {
	key = frame.ClaimKey(ctx, state.KindToggleButton, key)
	return frame.UseState[toggleButtonState](ctx, key, stateSlotToggleButton)
}

func (s *toggleButtonState) animateBackground(gtx layout.Context, target color.NRGBA) color.NRGBA {
	if !s.bgReady {
		s.background = target
		s.bgFrom = target
		s.bgTo = target
		s.bgAt = gtx.Now
		s.bgReady = true
		return target
	}
	if target != s.bgTo {
		s.bgFrom = s.background
		s.bgTo = target
		s.bgAt = gtx.Now
	}
	if s.bgFrom == s.bgTo {
		s.background = s.bgTo
		return s.background
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.bgAt), toggleButtonColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.background = render.LerpColor(s.bgFrom, s.bgTo, progress)
	return s.background
}

func (s *toggleButtonState) animateScale(gtx layout.Context, target float32) float32 {
	if !s.scaleReady {
		s.scaleValue = target
		s.scaleFrom = target
		s.scaleTo = target
		s.scaleAt = gtx.Now
		s.scaleReady = true
		return target
	}
	if target != s.scaleTo {
		s.scaleFrom = s.scaleValue
		s.scaleTo = target
		s.scaleAt = gtx.Now
	}
	if s.scaleFrom == s.scaleTo {
		s.scaleValue = s.scaleTo
		return s.scaleValue
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.scaleAt), toggleButtonScaleDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.scaleValue = render.Lerp(s.scaleFrom, s.scaleTo, progress)
	return s.scaleValue
}
