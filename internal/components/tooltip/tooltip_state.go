package tooltip

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTooltip = "tooltip"

const tooltipExclusiveGroup = "tooltip"

func tooltipStateFor(ctx *frame.Context, key string) (string, *tooltipState) {
	key = frame.ClaimKey(ctx, state.KindTooltip, key)
	value := frame.UseState[tooltipState](ctx, key, stateSlotTooltip)
	frame.RegisterExclusive(ctx, tooltipExclusiveGroup, key, value.closeForPeer)
	return key, value
}

type tooltipState struct {
	hovered bool
	focused bool
	active  bool
	open    bool
	showAt  time.Time
	hideAt  time.Time
	popup   PopupTransition
}

func (s *tooltipState) closeForPeer() {
	s.showAt = time.Time{}
	s.hideAt = time.Time{}
	s.open = false
}

func (s *tooltipState) update(gtx layout.Context, trigger TooltipTrigger, disabled bool, delay, closeDelay time.Duration) {
	s.updateEvents(gtx, trigger)
	active := s.hovered
	if trigger == TooltipFocus {
		active = s.focused
	}
	s.updateActive(gtx, active, disabled, delay, closeDelay)
}

func (s *tooltipState) updateActive(gtx layout.Context, active, disabled bool, delay, closeDelay time.Duration) {
	if disabled {
		s.hovered = false
		s.focused = false
		active = false
		s.showAt = time.Time{}
		s.hideAt = time.Time{}
		s.open = false
	}
	if active != s.active {
		s.active = active
		if active {
			s.hideAt = time.Time{}
			if s.open || s.popup.Value() > 0 || delay <= 0 {
				s.open = true
				s.showAt = time.Time{}
			} else {
				s.showAt = gtx.Now.Add(delay)
			}
		} else {
			s.showAt = time.Time{}
			if !s.open {
				s.hideAt = time.Time{}
			} else if closeDelay <= 0 {
				s.open = false
				s.hideAt = time.Time{}
			} else {
				s.hideAt = gtx.Now.Add(closeDelay)
			}
		}
	}
	s.resolveDeadline(gtx)
}

func (s *tooltipState) updateEvents(gtx layout.Context, trigger TooltipTrigger) {
	for {
		e, ok := gtx.Event(pointer.Filter{
			Target: s,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Cancel,
		})
		if !ok {
			break
		}
		if event, ok := e.(pointer.Event); ok {
			switch event.Kind {
			case pointer.Enter:
				s.hovered = true
			case pointer.Leave, pointer.Cancel:
				s.hovered = false
			}
		}
	}
	if trigger != TooltipFocus {
		return
	}
	for {
		e, ok := gtx.Event(key.FocusFilter{Target: s})
		if !ok {
			break
		}
		if event, ok := e.(key.FocusEvent); ok {
			s.focused = event.Focus
		}
	}
}

func (s *tooltipState) resolveDeadline(gtx layout.Context) {
	if !s.showAt.IsZero() {
		if !gtx.Now.Before(s.showAt) {
			s.open = s.active
			s.showAt = time.Time{}
		} else {
			gtx.Execute(op.InvalidateCmd{At: s.showAt})
		}
	}
	if !s.hideAt.IsZero() {
		if !gtx.Now.Before(s.hideAt) {
			s.open = s.active
			s.hideAt = time.Time{}
		} else {
			gtx.Execute(op.InvalidateCmd{At: s.hideAt})
		}
	}
}

func (s *tooltipState) addInput(gtx layout.Context, size image.Point, disabled bool) {
	if disabled || size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, s)
	pass.Pop()
	area.Pop()
}

func (s *tooltipState) progress(gtx layout.Context) float32 {
	return s.popup.Progress(gtx, s.open)
}

func (s *tooltipState) exiting() bool {
	return s.popup.Exiting()
}
