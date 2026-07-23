package selects

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotSelect = "select"

func selectStateFor(ctx *frame.Context, key string) *selectState {
	key = frame.ClaimKey(ctx, state.KindSelect, key)
	value := frame.UseStateWith(ctx, key, stateSlotSelect, func() *selectState {
		return &selectState{key: key}
	})
	frame.RegisterExclusive(ctx, "select", key, value.closeForPeer)
	return value
}

func activateSelect(ctx *frame.Context, value *selectState) {
	if value != nil && value.key != "" {
		frame.ActivateExclusive(ctx, "select", value.key)
	}
}

func releaseSelect(ctx *frame.Context, value *selectState) {
	if value != nil {
		frame.ReleaseExclusive(ctx, "select", value.key)
	}
}

type selectState struct {
	key                string
	trigger            widget.Clickable
	dismiss            [16]overlay.ClickArea
	dialog             overlay.ClickArea
	focus              state.FocusAnimation
	open               bool
	initialized        bool
	wasOpen            bool
	focusIntent        selectFocusIntent
	focusVisibleIntent bool
	triggerRect        image.Rectangle
	transition         animation.FloatTransition
	iconTransition     animation.FloatTransition
	binding            selectOpenBinding
	skipRestore        bool
	peerClosePending   bool
	dataVersion        uint64
	dataReady          bool
	cachedItems        []SelectItem
}

func (s *selectState) BeginFrame() {
	s.peerClosePending = false
}

func (s *selectState) itemsFor(widget SelectWidget) []SelectItem {
	if !widget.hasDataVersion {
		return widget.allItems()
	}
	if s.dataReady && s.dataVersion == widget.dataVersion {
		return s.cachedItems
	}
	s.cachedItems = widget.allItems()
	s.dataVersion = widget.dataVersion
	s.dataReady = true
	return s.cachedItems
}

type selectOpenBinding struct {
	controlled   bool
	open         bool
	onOpenChange func(bool)
}

type selectFocusIntent uint8

const (
	selectFocusNone selectFocusIntent = iota
	selectFocusSelected
	selectFocusFirst
	selectFocusLast
)

func (s *selectState) isOpen(widget SelectWidget) bool {
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

func (s *selectState) bind(widget SelectWidget) {
	s.binding = selectOpenBinding{
		controlled:   widget.hasOpen,
		open:         widget.open,
		onOpenChange: widget.onOpenChange,
	}
}

func (s *selectState) requestOpen(ctx *frame.Context, widget SelectWidget, open bool) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.skipRestore = false
		activateSelect(ctx, s)
	}
	if widget.hasOpen {
		if widget.open != open && widget.onOpenChange != nil {
			widget.onOpenChange(open)
		}
		if !widget.open {
			releaseSelect(ctx, s)
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
		releaseSelect(ctx, s)
	}
	return s.open
}

func (s *selectState) closeForPeer() {
	s.focusIntent = selectFocusNone
	s.focusVisibleIntent = false
	s.skipRestore = true
	s.peerClosePending = true
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

func (s *selectState) handleTrigger(ctx *frame.Context, gtx layout.Context, widget SelectWidget, open bool) bool {
	presses := state.SnapshotPresses(s.trigger.History())
	if widget.disabled {
		s.open = false
		s.focusIntent = selectFocusNone
		s.focusVisibleIntent = false
		return false
	}
	for s.trigger.Clicked(gtx) {
		if open {
			s.focusIntent = selectFocusNone
			s.focusVisibleIntent = false
			open = s.requestOpen(ctx, widget, false)
		} else {
			s.focusIntent = selectFocusSelected
			s.focusVisibleIntent = false
			open = s.requestOpen(ctx, widget, true)
		}
		frame.RequestFocusVisible(ctx, &s.trigger, presses.ClickFocusVisible(s.trigger.History()))
	}
	frame.FocusOnPress(ctx, &s.trigger, s.trigger.History(), presses.Active())
	return s.handleTriggerKeys(ctx, gtx, widget, open)
}

func (s *selectState) handleTriggerKeys(ctx *frame.Context, gtx layout.Context, widget SelectWidget, open bool) bool {
	filters := []event.Filter{
		key.Filter{Focus: &s.trigger, Name: key.NameDownArrow},
		key.Filter{Focus: &s.trigger, Name: key.NameUpArrow},
		key.Filter{Focus: &s.trigger, Name: key.NameReturn},
		key.Filter{Focus: &s.trigger, Name: key.NameEnter},
		key.Filter{Focus: &s.trigger, Name: key.NameSpace},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return open
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameUpArrow:
			if !open {
				s.focusIntent = selectFocusLast
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		case key.NameDownArrow:
			if !open {
				s.focusIntent = selectFocusFirst
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		default:
			if !open {
				s.focusIntent = selectFocusSelected
				s.focusVisibleIntent = true
				open = s.requestOpen(ctx, widget, true)
			}
		}
	}
}

func (s *selectState) handleOverlayEvents(ctx *frame.Context, gtx layout.Context, widget SelectWidget, open bool) bool {
	peerClosePending := s.peerClosePending
	s.peerClosePending = false
	for s.dialog.Clicked(gtx) {
	}
	if s.dialog.TakePressed() {
		frame.PreserveFocus(ctx)
	}
	dismissed := false
	for i := range s.dismiss {
		for s.dismiss[i].Clicked(gtx) {
			dismissed = true
		}
		dismissed = s.dismiss[i].TakePressed() || dismissed
	}
	if !open {
		return open
	}
	closed := dismissed
	escape := false
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			closed = true
			escape = true
		}
	}
	if peerClosePending {
		return open
	}
	if !closed {
		return open
	}
	s.focusIntent = selectFocusNone
	s.focusVisibleIntent = false
	if escape {
		frame.RequestFocus(ctx, &s.trigger)
	} else {
		s.skipRestore = true
	}
	open = s.requestOpen(ctx, widget, false)
	gtx.Execute(op.InvalidateCmd{})
	return open
}

func (s *selectState) observeOpen(ctx *frame.Context, open, restoreFocus bool) {
	if open && !s.wasOpen && s.focusIntent == selectFocusNone {
		s.focusIntent = selectFocusSelected
		s.focusVisibleIntent = false
	}
	if !open && s.wasOpen {
		s.focusIntent = selectFocusNone
		s.focusVisibleIntent = false
		if restoreFocus && !s.skipRestore {
			frame.RequestFocus(ctx, &s.trigger)
		}
		s.skipRestore = false
	}
	s.wasOpen = open
}

func (s *selectState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	duration := selectEnterDuration
	if !open {
		duration = selectExitDuration
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *selectState) iconProgress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	s.iconTransition.Initialize(0, gtx.Now)
	return s.iconTransition.Value(gtx, target, selectIndicatorDuration, animation.EaseSmoothstep, motions...)
}
