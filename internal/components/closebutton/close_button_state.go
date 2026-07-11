package closebutton

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotCloseButton = "close-button"

type closeButtonState struct {
	focus      state.FocusAnimation
	bg         color.NRGBA
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

func closeButtonStateFor(ctx *frame.Context, key string) *closeButtonState {
	return frame.UseState[closeButtonState](ctx, key, stateSlotCloseButton)
}

func (s *closeButtonState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	if !s.bgReady {
		s.bg = target
		s.bgFrom = target
		s.bgTo = target
		s.bgAt = gtx.Now
		s.bgReady = true
		return target
	}
	if target != s.bgTo {
		s.bgFrom = s.bg
		s.bgTo = target
		s.bgAt = gtx.Now
	}
	if s.bgFrom == s.bgTo {
		s.bg = s.bgTo
		return s.bg
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.bgAt), closeButtonColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.bg = render.LerpColor(s.bgFrom, s.bgTo, progress)
	return s.bg
}

func closeButtonPressedScale(pressedScale float32) float32 {
	if pressedScale <= 0 || pressedScale > 1 {
		return 0.93
	}
	return pressedScale
}

func (s *closeButtonState) scale(gtx layout.Context, target float32) float32 {
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
	progress := closeButtonEaseOutQuart(render.Progress(gtx.Now.Sub(s.scaleAt), closeButtonScaleDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.scaleValue = render.Lerp(s.scaleFrom, s.scaleTo, progress)
	return s.scaleValue
}

func closeButtonEaseOutQuart(progress float32) float32 {
	progress = min(max(progress, 0), 1)
	inverse := 1 - progress
	return 1 - inverse*inverse*inverse*inverse
}
