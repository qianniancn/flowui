package ui

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestIsPrimaryPointerPress(t *testing.T) {
	tests := []struct {
		name  string
		event pointer.Event
		want  bool
	}{
		{name: "mouse primary", event: pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary}, want: true},
		{name: "touch", event: pointer.Event{Kind: pointer.Press, Source: pointer.Touch}, want: true},
		{name: "mouse secondary", event: pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary}},
		{name: "release", event: pointer.Event{Kind: pointer.Release, Source: pointer.Mouse}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsPrimaryPointerPress(test.event); got != test.want {
				t.Fatalf("IsPrimaryPointerPress() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNextPointerEventRoutesGioInput(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	var target struct{}
	gtx := layout.Context{Source: router.Source(), Ops: &ops}
	if _, ok := NextPointerEvent(gtx, &target, pointer.Press); ok {
		t.Fatal("unexpected pointer event in initial frame")
	}
	AddPointerArea(gtx, &target, image.Rect(0, 0, 40, 24), pointer.CursorDefault)
	router.Frame(&ops)
	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(12, 8),
	})
	ops = op.Ops{}
	gtx.Ops = &ops
	AddPointerArea(gtx, &target, image.Rect(0, 0, 40, 24), pointer.CursorDefault)
	eventValue, ok := NextPointerEvent(gtx, &target, pointer.Press)
	if !ok {
		t.Fatal("NextPointerEvent returned no event")
	}
	if eventValue.Position != f32.Pt(12, 8) || !IsPrimaryPointerPress(eventValue) {
		t.Fatalf("event = %#v, want primary press at (12, 8)", eventValue)
	}
	router.Frame(&ops)
}
