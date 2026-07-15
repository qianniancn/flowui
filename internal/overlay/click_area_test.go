package overlay

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

func TestClickAreaTracksPointerWithoutEnteringFocusOrder(t *testing.T) {
	router := new(input.Router)
	area := new(ClickArea)
	button := new(widget.Clickable)
	layoutClickAreaTestFrame(router, area, button)

	router.MoveFocus(key.FocusForward)
	if !router.Source().Focused(button) {
		t.Fatal("keyboard focus did not skip the pointer-only click area")
	}
	if router.Source().Focused(area) {
		t.Fatal("pointer-only click area entered keyboard focus order")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Move,
		Source:   pointer.Mouse,
		Position: f32.Pt(10, 10),
	})
	if area.Clicked(clickAreaTestContext(router, new(op.Ops))) {
		t.Fatal("pointer move was reported as a completed click")
	}
	if !area.Hovered() {
		t.Fatal("pointer hover was not reported")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(10, 10),
	})
	if area.Clicked(clickAreaTestContext(router, new(op.Ops))) {
		t.Fatal("pointer press was reported as a completed click")
	}
	if !area.TakePressed() {
		t.Fatal("pointer press was not reported")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(10, 10),
	})
	if !area.Clicked(clickAreaTestContext(router, new(op.Ops))) {
		t.Fatal("pointer release did not complete the click")
	}
}

func layoutClickAreaTestFrame(router *input.Router, area *ClickArea, button *widget.Clickable) {
	var ops op.Ops
	gtx := clickAreaTestContext(router, &ops)
	area.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(40, 40)}
	})
	offset := op.Offset(image.Pt(60, 0)).Push(gtx.Ops)
	button.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(20, 20)}
	})
	offset.Pop()
	router.Frame(&ops)
}

func clickAreaTestContext(router *input.Router, ops *op.Ops) layout.Context {
	return layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         ops,
	}
}
