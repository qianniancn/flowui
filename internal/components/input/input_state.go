package input

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
)

const inputTransitionDuration = 150 * time.Millisecond

type inputState struct {
	field.State
	ringWidth animation.FloatTransition
}

func inputStateFor(ctx *frame.Context, key string) *inputState {
	return frame.UseState[inputState](ctx, key, stateSlotInput)
}

func (s *inputState) RingWidth(gtx layout.Context, target unit.Dp) float32 {
	return s.ringWidth.Value(gtx, float32(target), inputTransitionDuration, animation.EaseSmoothstep)
}
