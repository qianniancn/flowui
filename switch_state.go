package flowui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type switchState struct {
	value         widget.Bool
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

func (s *switchState) selection(gtx layout.Context, checked bool) float32 {
	target := float32(0)
	if checked {
		target = 1
	}
	return switchAnimateFloat(gtx, target, &s.selected, &s.selectedFrom, &s.selectedTo, &s.selectedAt, &s.selectedReady, switchSelectDuration)
}

func (s *switchState) focusOpacity(gtx layout.Context, focused bool) float32 {
	target := float32(0)
	if focused {
		target = 1
	}
	return switchAnimateFloat(gtx, target, &s.focus, &s.focusFrom, &s.focusTo, &s.focusAt, &s.focusReady, switchFocusDuration)
}

func (s *switchState) focusVisible(focused bool, history []widget.Press) bool {
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

func switchAnimateFloat(gtx layout.Context, target float32, value, from, to *float32, at *time.Time, ready *bool, duration time.Duration) float32 {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(*at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = lerp(*from, *to, progress)
	return *value
}
