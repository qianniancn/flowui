package state

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

// Focus tracks frame-local focus commands and the pointer catcher used to clear
// focus when a user presses outside a focused widget.
type Focus struct {
	catcher      focusCatcher
	pointerPress bool
	preserve     bool
	target       event.Tag
	pending      focusTarget
	active       focusTarget
	frame        uint64
}

type focusCatcher struct{}

type FocusOrigin uint8

const (
	FocusOriginKeyboard FocusOrigin = iota
	FocusOriginPointer
)

type focusTarget struct {
	tag         event.Tag
	origin      FocusOrigin
	requestedAt uint64
	applied     bool
}

func (f *Focus) BeginFrame() {
	f.frame++
	f.pointerPress = false
	f.preserve = false
	f.target = nil
}

func (f *Focus) Request(tag event.Tag, origin FocusOrigin) {
	f.target = tag
	f.pending = focusTarget{tag: tag, origin: origin, requestedAt: f.frame}
}

func (f *Focus) OnPress(tag event.Tag, history []widget.Press, before int) {
	if ActivePresses(history) > before {
		f.Request(tag, FocusOriginPointer)
	}
}

func (f *Focus) Visible(tag event.Tag, focused bool) bool {
	if !focused {
		if f.active.tag == tag {
			f.active = focusTarget{}
		}
		return false
	}
	if f.pending.tag == tag {
		f.active = f.pending
		f.pending = focusTarget{}
	} else if f.active.tag != tag {
		f.active = focusTarget{tag: tag, origin: FocusOriginKeyboard}
		if f.pending.applied && f.pending.requestedAt < f.frame {
			f.pending = focusTarget{}
		}
	}
	return f.active.origin != FocusOriginPointer
}

func (f *Focus) Preserve() {
	f.preserve = true
}

func (f *Focus) ApplyFrameCommands(gtx layout.Context) {
	f.updatePointerPress(gtx)
	if f.target != nil {
		gtx.Execute(key.FocusCmd{Tag: f.target})
		if f.pending.tag == f.target {
			f.pending.applied = true
		}
	} else if f.pointerPress && !f.preserve {
		gtx.Execute(key.FocusCmd{})
		f.pending = focusTarget{}
		f.active = focusTarget{}
	}
	f.addCatcher(gtx)
}

func (f *Focus) updatePointerPress(gtx layout.Context) {
	for {
		e, ok := gtx.Event(pointer.Filter{
			Target: &f.catcher,
			Kinds:  pointer.Press,
		})
		if !ok {
			return
		}
		if _, ok := e.(pointer.Event); ok {
			f.pointerPress = true
		}
	}
}

func (f *Focus) addCatcher(gtx layout.Context) {
	const edge = 1 << 20
	stack := clip.Rect(image.Rect(-edge, -edge, edge, edge)).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &f.catcher)
	pass.Pop()
	stack.Pop()
}

func ActivePresses(history []widget.Press) int {
	var n int
	for _, press := range history {
		if press.End.IsZero() && !press.Cancelled {
			n++
		}
	}
	return n
}
