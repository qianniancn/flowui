package state

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

const focusAnimationDuration = 100 * time.Millisecond

// FocusAnimation transitions focus-ring opacity.
type FocusAnimation struct {
	value float32
	from  float32
	to    float32
	at    time.Time
	ready bool
}

func (s *FocusAnimation) Opacity(gtx layout.Context, visible bool, motions ...theme.MotionTheme) float32 {
	if !visible {
		s.value = 0
		s.from = 0
		s.to = 0
		s.at = gtx.Now
		s.ready = true
		return 0
	}
	target := float32(1)
	if !s.ready {
		s.value = target
		s.from = target
		s.to = target
		s.at = gtx.Now
		s.ready = true
		return target
	}
	if target != s.to {
		s.from = s.value
		s.to = target
		s.at = gtx.Now
	}
	if s.from == s.to {
		s.value = s.to
		return s.value
	}
	duration := focusAnimationDuration
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if duration <= 0 {
		s.value = s.to
		s.from = s.to
		return s.value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.value = render.Lerp(s.from, s.to, progress)
	return s.value
}

func (s *FocusAnimation) TargetOpacity() float32 {
	return s.to
}
