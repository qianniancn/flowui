package dropdown

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	stateSlotDropdown     = "dropdown"
	dropdownExclusive     = "dropdown"
	dropdownEnterDuration = 150 * time.Millisecond
	dropdownExitDuration  = 100 * time.Millisecond
	dropdownLongPress     = 500 * time.Millisecond
)

type dropdownState struct {
	key            string
	trigger        widget.Clickable
	triggerFocus   state.FocusAnimation
	longPressTag   struct{}
	touchTracking  bool
	pointerID      pointer.ID
	pointerStart   f32.Point
	pointerAt      time.Time
	longPressMoved bool
	dismiss        [16]overlay.ClickArea
	dialog         overlay.ClickArea
	open           bool
	initialized    bool
	wasOpen        bool
	triggerRect    image.Rectangle
	focusFirst     bool
	focusLast      bool
	focusVisible   bool
	skipRestore    bool
	transition     animation.FloatTransition
	binding        dropdownBinding
}

type dropdownBinding struct {
	controlled   bool
	open         bool
	onOpenChange func(bool)
}

func dropdownStateFor(ctx *frame.Context, key string) *dropdownState {
	key = frame.ClaimKey(ctx, state.KindDropdown, key)
	value := frame.UseStateWith(ctx, key, stateSlotDropdown, func() *dropdownState {
		return &dropdownState{key: key}
	})
	frame.RegisterExclusive(ctx, dropdownExclusive, key, value.closeForPeer)
	return value
}

func (s *dropdownState) bind(widget Widget) {
	s.binding = dropdownBinding{controlled: widget.hasOpen, open: widget.open, onOpenChange: widget.onOpenChange}
}

func (s *dropdownState) isOpen(widget Widget) bool {
	if !s.initialized {
		if widget.hasDefaultOpen {
			s.open = widget.defaultOpen
		}
		s.initialized = true
	}
	if widget.hasOpen {
		return widget.open
	}
	return s.open
}

func (s *dropdownState) requestOpen(ctx *frame.Context, widget Widget, open bool) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.skipRestore = false
		frame.ActivateExclusive(ctx, dropdownExclusive, s.key)
	}
	if widget.hasOpen {
		if widget.open != open && widget.onOpenChange != nil {
			widget.onOpenChange(open)
		}
		if !widget.open {
			frame.ReleaseExclusive(ctx, dropdownExclusive, s.key)
		}
		return widget.open
	}
	if s.open != open {
		s.open = open
		if widget.onOpenChange != nil {
			widget.onOpenChange(open)
		}
	}
	if !s.open {
		frame.ReleaseExclusive(ctx, dropdownExclusive, s.key)
	}
	return s.open
}

func (s *dropdownState) closeForPeer() {
	s.skipRestore = true
	if s.binding.controlled {
		if s.binding.open && s.binding.onOpenChange != nil {
			s.binding.onOpenChange(false)
		}
		return
	}
	if s.open {
		s.open = false
		if s.binding.onOpenChange != nil {
			s.binding.onOpenChange(false)
		}
	}
}

func (s *dropdownState) observeOpen(open bool) {
	wasOpen := s.wasOpen
	if !open && wasOpen {
		s.touchTracking = false
		s.focusFirst = false
		s.focusLast = false
	}
	if open && !wasOpen && !s.focusFirst && !s.focusLast {
		s.focusFirst = true
		s.focusVisible = false
	}
	s.wasOpen = open
}

func (s *dropdownState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := dropdownExitDuration
	if open {
		target = 1
		duration = dropdownEnterDuration
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}
