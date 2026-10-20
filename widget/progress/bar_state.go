package progress

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotProgressBar = "progress-bar"

func progressBarStateFor(ctx *core.Context, key string) *progressBarState {
	return core.UseState[progressBarState](ctx, key, stateSlotProgressBar)
}

type progressBarState struct {
	value            animation.FloatTransition
	barStyleCache    progressBarVisualCache
	circleStyleCache progressCircleVisualCache
}

func (s *progressBarState) progress(gtx layout.Context, target float32, indeterminate bool, duration time.Duration, motion theme.MotionTheme) float32 {
	if indeterminate {
		return target
	}
	return s.value.Value(gtx, target, duration, animation.EaseSmoothstep, motion)
}
