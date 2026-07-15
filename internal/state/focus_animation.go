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
	pendingPress bool
	pendingAge   uint8
	historyLen   int
	lastPress    time.Time
	historyReady bool
}

// Prepare controls whether the next programmatic focus should be keyboard-visible.
func (s *FocusAnimation) Prepare(visible bool) {
	s.focused = false
	s.pointerFocus = !visible
	s.pendingPress = !visible
	s.pendingAge = 0
	if !visible {
		s.value = 0
		s.from = 0
		s.to = 0
		s.ready = true
	}
}

func (s *FocusAnimation) Visible(focused bool, history []widget.Press) bool {
	pointerPress := s.observePointerPress(history)
	if !focused {
		s.focused = false
		s.pointerFocus = false
		// A pointer-issued focus command is observed on the following frame.
		if s.pendingPress && !pointerPress {
			s.pendingAge++
			if s.pendingAge > 1 {
				s.pendingPress = false
				s.pendingAge = 0
			}
		}
		return false
	}
	if !s.focused {
		s.focused = true
		s.pointerFocus = s.pendingPress
		s.pendingPress = false
		s.pendingAge = 0
	} else if pointerPress {
		s.pointerFocus = true
		s.pendingPress = false
		s.pendingAge = 0
	}
	return !s.pointerFocus
}

func (s *FocusAnimation) observePointerPress(history []widget.Press) bool {
	length := len(history)
	var last time.Time
	if length > 0 {
		last = history[length-1].Start
	}
	pointerPress := s.historyReady && length > 0 && (length > s.historyLen || last != s.lastPress)
	if !s.historyReady && length > 0 {
		pointerPress = true
	}
	s.historyLen = length
	s.lastPress = last
	s.historyReady = true
	if pointerPress {
		s.pendingPress = true
		s.pendingAge = 0
	}
	return pointerPress
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

func (s *FocusAnimation) PointerOrigin() bool {
	return s.pointerFocus
}

func (s *FocusAnimation) TargetOpacity() float32 {
	return s.to
}
