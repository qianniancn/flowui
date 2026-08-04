package ui

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/interact"
)

// AddPointerArea registers a rectangular pointer target and its cursor.
// bounds are in the current layout coordinate space.
func AddPointerArea(gtx layout.Context, target event.Tag, bounds image.Rectangle, cursor pointer.Cursor) {
	interact.AddPointerArea(gtx, target, bounds, cursor)
}

// NextPointerEvent returns the next pointer event matching target and kinds.
func NextPointerEvent(gtx layout.Context, target event.Tag, kinds pointer.Kind) (pointer.Event, bool) {
	return interact.NextPointerEvent(gtx, target, kinds)
}

// IsPrimaryPointerPress reports whether event starts the primary action.
func IsPrimaryPointerPress(eventValue pointer.Event) bool {
	return interact.IsPrimaryPointerPress(eventValue)
}

// GrabPointer routes subsequent events for the pointer to target.
func GrabPointer(gtx layout.Context, target event.Tag, eventValue pointer.Event) {
	interact.GrabPointer(gtx, target, eventValue)
}
