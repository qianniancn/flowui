package dropdown

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const (
	stateSlotDropdown     = "dropdown"
	dropdownExclusive     = "dropdown"
	dropdownEnterDuration = 150 * time.Millisecond
	dropdownExitDuration  = 100 * time.Millisecond
	dropdownLongPress     = 500 * time.Millisecond
)

type dropdownState struct {
	key                string
	trigger            widget.Clickable
	triggerFocus       state.FocusAnimation
	prepareButtonFocus func(bool)
	longPressTag       struct{}
	touchTracking      bool
	pointerID          pointer.ID
	pointerStart       f32.Point
	pointerAt          time.Time
	longPressMoved     bool
	dismiss            [16]overlay.ClickArea
	dialog             overlay.ClickArea
	open               bool
	initialized        bool
	wasOpen            bool
	triggerRect        image.Rectangle
	focusFirst         bool
	focusLast          bool
	focusVisible       bool
	skipRestore        bool
	progressValue      float32
	progressFrom       float32
	progressTo         float32
	progressAt         time.Time
	progressReady      bool
	binding            dropdownBinding
}

func (s *dropdownState) prepareTriggerFocus(visible bool) {
	s.triggerFocus.Prepare(visible)
	if s.prepareButtonFocus != nil {
		s.prepareButtonFocus(visible)
	}
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

func (s *dropdownState) progress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	duration := dropdownExitDuration
	if open {
		target = 1
		duration = dropdownEnterDuration
	}
	if !s.progressReady {
		s.progressAt = gtx.Now
		s.progressReady = true
	}
	if target != s.progressTo {
		s.progressFrom = s.progressValue
		s.progressTo = target
		s.progressAt = gtx.Now
	}
	if s.progressFrom == s.progressTo {
		s.progressValue = s.progressTo
		return s.progressValue
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.progressAt), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.progressValue = render.Lerp(s.progressFrom, s.progressTo, progress)
	return s.progressValue
}
