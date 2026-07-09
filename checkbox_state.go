package flowui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type checkboxState struct {
	selected      float32
	selectedFrom  float32
	selectedTo    float32
	selectedAt    time.Time
	selectedReady bool
	focus         float32
	focusFrom     float32
	focusTo       float32
	focusAt       time.Time
	focusReady    bool
	focused       bool
	pointerFocus  bool
}

func (s *checkboxState) selection(gtx layout.Context, checked bool) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	if !s.selectedReady {
		s.selected = target
		s.selectedFrom = target
		s.selectedTo = target
		s.selectedAt = gtx.Now
		s.selectedReady = true
		return target
	}
	if target != s.selectedTo {
		s.selectedFrom = s.selected
		s.selectedTo = target
		s.selectedAt = gtx.Now
	}
	if s.selectedFrom == s.selectedTo {
		s.selected = s.selectedTo
		return s.selected
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(s.selectedAt), checkboxSelectDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.selected = lerp(s.selectedFrom, s.selectedTo, progress)
	return s.selected
}

func (s *checkboxState) focusOpacity(gtx layout.Context, focused bool) float32 {
	target := float32(0)
	if focused {
		target = 1
	}
	if !s.focusReady {
		s.focus = target
		s.focusFrom = target
		s.focusTo = target
		s.focusAt = gtx.Now
		s.focusReady = true
		return target
	}
	if target != s.focusTo {
		s.focusFrom = s.focus
		s.focusTo = target
		s.focusAt = gtx.Now
	}
	if s.focusFrom == s.focusTo {
		s.focus = s.focusTo
		return s.focus
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(s.focusAt), checkboxFocusDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.focus = lerp(s.focusFrom, s.focusTo, progress)
	return s.focus
}

func (s *checkboxState) focusVisible(focused bool, history []widget.Press) bool {
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
