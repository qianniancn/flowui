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
	// Observation state: collected during layout, committed at frame boundaries.
	observed    map[event.Tag]bool // tag -> focused
	observedAny bool
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

// PressSnapshot records pointer history before a Clickable processes events.
type PressSnapshot struct {
	active  int
	count   int
	latest  widget.Press
	hasLast bool
}

// SnapshotPresses captures the pointer presses currently known to a Clickable.
func SnapshotPresses(history []widget.Press) PressSnapshot {
	snapshot := PressSnapshot{
		active: ActivePresses(history),
		count:  len(history),
	}
	if snapshot.count > 0 {
		snapshot.latest = history[snapshot.count-1]
		snapshot.hasLast = true
	}
	return snapshot
}

// Active reports how many pointer presses were active in the snapshot.
func (s PressSnapshot) Active() int {
	return s.active
}

// ClickFocusVisible reports whether a subsequent click was not pointer-driven.
func (s PressSnapshot) ClickFocusVisible(history []widget.Press) bool {
	if s.active > 0 || len(history) > s.count {
		return false
	}
	if len(history) == 0 || !s.hasLast {
		return true
	}
	latest := history[len(history)-1]
	return latest.Start == s.latest.Start && latest.Position == s.latest.Position
}

func (f *Focus) BeginFrame() {
	f.frame++
	f.pointerPress = false
	f.preserve = false
	f.target = nil
	f.observed = nil
	f.observedAny = false
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

// Observe records frame-local focus input and returns whether the widget should
// draw its focus ring. Persistent modality changes are deferred until
// CommitObservations. It is safe to call multiple times per frame for one tag.
func (f *Focus) Observe(tag event.Tag, focused bool) bool {
	if f.observed == nil {
		f.observed = make(map[event.Tag]bool)
	}
	f.observed[tag] = focused
	f.observedAny = true

	// Return the visibility based on current active state (read-only).
	if !focused {
		return false
	}
	// If this tag is active, return its origin.
	if f.active.tag == tag {
		return f.active.origin != FocusOriginPointer
	}
	// If pending matches this tag, it will be keyboard-visible.
	if f.pending.tag == tag {
		return f.pending.origin != FocusOriginPointer
	}
	// Otherwise default to keyboard-visible (unrequested focus).
	return true
}

// CommitObservations applies deferred state changes collected during Observe calls.
// Must be called once per frame after all layout passes, before ApplyFrameCommands.
func (f *Focus) CommitObservations() {
	if !f.observedAny {
		return
	}
	f.observedAny = false
	for tag, focused := range f.observed {
		if !focused {
			if f.active.tag == tag {
				f.active = focusTarget{}
			}
			continue
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
	}
}

// Visible is a deprecated alias that provides immediate-mode behavior for backward compatibility.
// New code should use Observe + CommitObservations instead.
func (f *Focus) Visible(tag event.Tag, focused bool) bool {
	// Record observation for potential CommitObservations call.
	if f.observed == nil {
		f.observed = make(map[event.Tag]bool)
	}
	f.observed[tag] = focused
	if focused {
		f.observedAny = true
	}

	// Apply state changes immediately (old immediate-mode behavior).
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
