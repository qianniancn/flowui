package progress

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotProgressBar = "progress-bar"

func progressBarStateFor(ctx *frame.Context, key string) *progressBarState {
	key = frame.ClaimKey(ctx, state.KindProgressBar, key)
	return frame.UseState[progressBarState](ctx, key, stateSlotProgressBar)
}

type progressBarState struct {
	value      float32
	valueFrom  float32
	valueTo    float32
	valueAt    time.Time
	valueReady bool
}

func (s *progressBarState) progress(gtx layout.Context, target float32, indeterminate bool) float32 {
	if indeterminate {
		return target
	}
	if !s.valueReady {
		s.value = target
		s.valueFrom = target
		s.valueTo = target
		s.valueAt = gtx.Now
		s.valueReady = true
		return target
	}
	if target != s.valueTo {
		s.valueFrom = s.value
		s.valueTo = target
		s.valueAt = gtx.Now
	}
	if s.valueFrom == s.valueTo {
		s.value = s.valueTo
		return s.value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.valueAt), progressBarValueDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.value = render.Lerp(s.valueFrom, s.valueTo, progress)
	return s.value
}
