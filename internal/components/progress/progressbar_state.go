package progress

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotProgressBar = "progress-bar"

func progressBarStateFor(ctx *frame.Context, key string) *progressBarState {
	key = frame.ClaimKey(ctx, state.KindProgressBar, key)
	return frame.UseState[progressBarState](ctx, key, stateSlotProgressBar)
}

type progressBarState struct {
	value animation.FloatTransition
}

func (s *progressBarState) progress(gtx layout.Context, target float32, indeterminate bool) float32 {
	if indeterminate {
		return target
	}
	return s.value.Value(gtx, target, progressBarValueDuration, animation.EaseSmoothstep)
}
