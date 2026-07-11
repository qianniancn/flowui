package input

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
)

const inputTransitionDuration = 150 * time.Millisecond

type inputState struct {
	field.State
	ringWidth      float32
	ringWidthFrom  float32
	ringWidthTo    float32
	ringWidthAt    time.Time
	ringWidthReady bool
}

func inputStateFor(ctx *frame.Context, key string) *inputState {
	return frame.UseState[inputState](ctx, key, stateSlotInput)
}

func (s *inputState) RingWidth(gtx layout.Context, target unit.Dp) float32 {
	targetValue := float32(target)
	if !s.ringWidthReady {
		s.ringWidth = targetValue
		s.ringWidthFrom = targetValue
		s.ringWidthTo = targetValue
		s.ringWidthAt = gtx.Now
		s.ringWidthReady = true
		return targetValue
	}
	if targetValue != s.ringWidthTo {
		s.ringWidthFrom = s.ringWidth
		s.ringWidthTo = targetValue
		s.ringWidthAt = gtx.Now
	}
	if s.ringWidthFrom == s.ringWidthTo {
		s.ringWidth = s.ringWidthTo
		return s.ringWidth
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.ringWidthAt), inputTransitionDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.ringWidth = render.Lerp(s.ringWidthFrom, s.ringWidthTo, progress)
	return s.ringWidth
}
