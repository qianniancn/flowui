package modal

import (
	"time"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotModal = "modal"

func modalStateFor(ctx *frame.Context, key string) *modalState {
	key = frame.ClaimKey(ctx, state.KindModal, key)
	return frame.UseState[modalState](ctx, key, stateSlotModal)
}

func hasVisibleModal(ctx *frame.Context, key string) bool {
	value, ok := frame.PeekState[modalState](ctx, key, stateSlotModal)
	return ok && value.visible()
}

func deleteModalState(ctx *frame.Context, key string) {
	frame.DeleteState(ctx, key, stateSlotModal)
}

type modalState struct {
	dismiss     [4]modalClickArea
	dialog      modalClickArea
	close       widget.Clickable
	closeFocus  state.FocusAnimation
	bodyList    layout.List
	outsideList layout.List
	focusStart  widget.Clickable
	focusTarget widget.Clickable
	focusEnd    widget.Clickable
	wasOpen     bool
	value       float32
	from        float32
	to          float32
	at          time.Time
	ready       bool
}

func (s *modalState) progress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	if !s.ready {
		s.value = 0
		s.from = 0
		s.to = 0
		s.at = gtx.Now
		s.ready = true
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
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.at), modalEnterDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.value = render.Lerp(s.from, s.to, progress)
	return s.value
}

func (s *modalState) visible() bool {
	return s.ready && s.value > 0
}

func (s *modalState) opening() bool {
	return s.to >= s.from
}

func (s *modalState) initialFocusTag() event.Tag {
	return &s.focusTarget
}

func (s *modalState) tabFocusTag(showClose bool) event.Tag {
	if showClose {
		return &s.close
	}
	return &s.focusTarget
}

func (s *modalState) endFocusTag() event.Tag {
	return &s.focusTarget
}

func (s *modalState) syncFocus(ctx *frame.Context, open bool) {
	if open && !s.wasOpen {
		frame.RequestFocus(ctx, s.initialFocusTag())
	}
	s.wasOpen = open
}

func (s *modalState) escapePressed(gtx layout.Context) bool {
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			return false
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			return true
		}
	}
}

type modalClickArea struct {
	click         gesture.Click
	requestClicks int
}

func (a *modalClickArea) Click() {
	a.requestClicks++
}

func (a *modalClickArea) Clicked(gtx layout.Context) bool {
	if a.requestClicks > 0 {
		a.requestClicks--
		return true
	}
	for {
		event, ok := a.click.Update(gtx.Source)
		if !ok {
			return false
		}
		if event.Kind == gesture.KindClick {
			return true
		}
	}
}

func (a *modalClickArea) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	for {
		_, ok := a.click.Update(gtx.Source)
		if !ok {
			break
		}
	}
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()
	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	a.click.Add(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}
