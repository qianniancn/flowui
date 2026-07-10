package state

import (
	"image"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestFocusPreserveLastsOnlyForCurrentFrame(t *testing.T) {
	var focus Focus
	focus.BeginFrame()
	focus.Preserve()
	focus.Preserve()
	if !focus.preserve {
		t.Fatal("Preserve did not retain focus for the current frame")
	}

	focus.BeginFrame()
	if focus.preserve {
		t.Fatal("focus preservation leaked into the next frame")
	}
}

func TestFocusPreservePreventsPointerFocusClear(t *testing.T) {
	tests := []struct {
		name     string
		preserve bool
		focused  bool
	}{
		{name: "preserved", preserve: true, focused: true},
		{name: "not preserved", focused: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tag := new(int)
			router := new(input.Router)
			registerFocusTestTarget(router, tag, nil)
			router.Source().Execute(key.FocusCmd{Tag: tag})
			if !router.Source().Focused(tag) {
				t.Fatal("failed to establish initial focus")
			}

			var focus Focus
			focus.BeginFrame()
			focus.pointerPress = true
			if test.preserve {
				focus.Preserve()
			}
			registerFocusTestTarget(router, tag, &focus)
			if got := router.Source().Focused(tag); got != test.focused {
				t.Fatalf("focused = %v, want %v", got, test.focused)
			}
		})
	}
}

func registerFocusTestTarget(router *input.Router, tag event.Tag, focus *Focus) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         &ops,
	}
	event.Op(gtx.Ops, tag)
	gtx.Event(key.FocusFilter{Target: tag})
	if focus != nil {
		focus.ApplyFrameCommands(gtx)
	}
	router.Frame(&ops)
}
