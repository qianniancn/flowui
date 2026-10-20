package dropdown

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

const (
	stateSlotDropdown = "dropdown"
	dropdownExclusive = "dropdown"
)

type dropdownState struct {
	key              string
	trigger          widget.Clickable
	triggerFocus     animation.FocusAnimation
	longPressTag     struct{}
	contextTag       struct{}
	touchTracking    bool
	pointerID        pointer.ID
	pointerStart     f32.Point
	pointerAt        time.Time
	longPressMoved   bool
	dismiss          [16]core.ClickArea
	dialog           core.ClickArea
	arrow            core.ClickArea
	open             bool
	initialized      bool
	wasOpen          bool
	triggerRect      image.Rectangle
	contextAnchor    image.Rectangle
	hasContextAnchor bool
	focusFirst       bool
	focusLast        bool
	focusVisible     bool
	skipRestore      bool
	hoverOpenAt      time.Time
	hoverCloseAt     time.Time
	menuHovered      bool
	openSource       OpenChangeSource
	transition       animation.FloatTransition
	binding          dropdownBinding
	visualCache      dropdownVisualCache
}

type dropdownBinding struct {
	controlled        bool
	open              bool
	onOpenChangeEvent func(OpenChangeEvent)
}

func dropdownStateFor(ctx *core.Context, key string) *dropdownState {
	key = core.ClaimKey(ctx, core.KindDropdown, key)
	value := core.UseStateWith(ctx, key, stateSlotDropdown, func() *dropdownState {
		return &dropdownState{key: key}
	})
	core.RegisterExclusive(ctx, dropdownExclusive, key, value.closeForPeer)
	return value
}

func (s *dropdownState) bind(widget Widget) {
	s.binding = dropdownBinding{
		controlled:        widget.hasOpen,
		open:              widget.open,
		onOpenChangeEvent: widget.onOpenChangeEvent,
	}
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

func (s *dropdownState) requestOpen(ctx *core.Context, widget Widget, open bool) bool {
	return s.requestOpenFrom(ctx, widget, open, OpenChangeProgrammatic)
}

func (s *dropdownState) requestOpenFrom(ctx *core.Context, widget Widget, open bool, source OpenChangeSource) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.openSource = source
		s.skipRestore = false
		core.ActivateExclusive(ctx, dropdownExclusive, s.key)
	} else if source != OpenChangeContextMenu {
		// A subsequent open requested by the model or another trigger must not
		// inherit a pointer anchor from a previously opened context menu.
		s.openSource = source
	}
	if widget.hasOpen {
		if widget.open != open && widget.onOpenChangeEvent != nil {
			widget.onOpenChangeEvent(OpenChangeEvent{Open: open, Source: source})
		}
		if !widget.open {
			core.ReleaseExclusive(ctx, dropdownExclusive, s.key)
		}
		return widget.open
	}
	if s.open != open {
		s.open = open
		if widget.onOpenChangeEvent != nil {
			widget.onOpenChangeEvent(OpenChangeEvent{Open: open, Source: source})
		}
	}
	if !s.open {
		core.ReleaseExclusive(ctx, dropdownExclusive, s.key)
	}
	return s.open
}

func (s *dropdownState) closeForPeer() {
	s.skipRestore = true
	if s.binding.controlled {
		if s.binding.open && s.binding.onOpenChangeEvent != nil {
			s.binding.onOpenChangeEvent(OpenChangeEvent{Open: false, Source: OpenChangePeer})
		}
		return
	}
	if s.open {
		s.open = false
		if s.binding.onOpenChangeEvent != nil {
			s.binding.onOpenChangeEvent(OpenChangeEvent{Open: false, Source: OpenChangePeer})
		}
	}
}

func (s *dropdownState) observeOpen(open bool) {
	wasOpen := s.wasOpen
	if !open && wasOpen {
		s.touchTracking = false
		s.focusFirst = false
		s.focusLast = false
		// Keep the last context anchor for the exit animation, but make the
		// next non-context open use the trigger instead of stale pointer data.
		s.openSource = OpenChangeProgrammatic
	}
	if open && !wasOpen && !s.focusFirst && !s.focusLast {
		s.focusFirst = true
		s.focusVisible = false
	}
	s.wasOpen = open
}

func (s *dropdownState) progressFor(gtx layout.Context, open bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := exit
	if open {
		target = 1
		duration = enter
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}
