package overlay

import (
	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// ClickArea is a pointer-only interaction surface for overlay blockers. Unlike
// widget.Clickable, it does not register keyboard focus or activation filters.
type ClickArea struct {
	click         gesture.Click
	requestClicks int
	pressed       bool
}

// Click queues a synthetic click for the next Clicked call.
func (a *ClickArea) Click() {
	a.requestClicks++
}

// Clicked processes pointer events and reports one completed click at a time.
func (a *ClickArea) Clicked(gtx layout.Context) bool {
	if a.requestClicks > 0 {
		a.requestClicks--
		return true
	}
	for {
		event, ok := a.click.Update(gtx.Source)
		if !ok {
			return false
		}
		if event.Kind == gesture.KindPress {
			a.pressed = true
		}
		if event.Kind == gesture.KindClick {
			return true
		}
	}
}

// TakePressed reports and clears whether Clicked observed a pointer press.
func (a *ClickArea) TakePressed() bool {
	pressed := a.pressed
	a.pressed = false
	return pressed
}

// Hovered reports whether the pointer is over the click area.
func (a *ClickArea) Hovered() bool {
	return a.click.Hovered()
}

// Layout registers a pointer-only click region around child.
func (a *ClickArea) Layout(gtx layout.Context, child layout.Widget) layout.Dimensions {
	for {
		_, ok := a.click.Update(gtx.Source)
		if !ok {
			break
		}
	}
	macro := op.Record(gtx.Ops)
	dims := child(gtx)
	call := macro.Stop()
	clipStack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	a.click.Add(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}
