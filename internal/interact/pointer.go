package interact

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// AddPointerArea registers a rectangular pointer target and its cursor.
func AddPointerArea(gtx layout.Context, target event.Tag, bounds image.Rectangle, cursor pointer.Cursor) {
	if target == nil || bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	cursor.Add(gtx.Ops)
	event.Op(gtx.Ops, target)
	area.Pop()
}

// NextPointerEvent returns the next pointer event matching target and kinds.
func NextPointerEvent(gtx layout.Context, target event.Tag, kinds pointer.Kind) (pointer.Event, bool) {
	if target == nil || kinds == 0 {
		return pointer.Event{}, false
	}
	for {
		value, ok := gtx.Event(pointer.Filter{Target: target, Kinds: kinds})
		if !ok {
			return pointer.Event{}, false
		}
		if eventValue, ok := value.(pointer.Event); ok {
			return eventValue, true
		}
	}
}

// IsPrimaryPointerPress reports whether event starts the primary action.
func IsPrimaryPointerPress(eventValue pointer.Event) bool {
	return eventValue.Kind == pointer.Press &&
		(eventValue.Source == pointer.Touch || eventValue.Buttons.Contain(pointer.ButtonPrimary))
}

// GrabPointer routes subsequent events for the pointer to target.
func GrabPointer(gtx layout.Context, target event.Tag, eventValue pointer.Event) {
	if target != nil {
		gtx.Execute(pointer.GrabCmd{Tag: target, ID: eventValue.PointerID})
	}
}
