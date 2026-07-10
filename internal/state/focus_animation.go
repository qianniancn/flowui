package state

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/render"
)

const focusAnimationDuration = 100 * time.Millisecond

// FocusAnimation tracks keyboard-visible focus and its opacity transition.
type FocusAnimation struct {
	value        float32
	from         float32
	to           float32
	at           time.Time
	ready        bool
	focused      bool
	pointerFocus bool
}

func (s *FocusAnimation) Visible(focused bool, history []widget.Press) bool {
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

func (s *FocusAnimation) Opacity(gtx layout.Context, visible bool) float32 {
	target := float32(0)
	if visible {
		target = 1
	}
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
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.at), focusAnimationDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.value = render.Lerp(s.from, s.to, progress)
	return s.value
}
