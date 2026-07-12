package state

import (
	"image"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
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

func TestPointerFocusOriginSurvivesDelayedFocus(t *testing.T) {
	var focus Focus
	tag := new(int)
	focus.BeginFrame()
	focus.Request(tag, FocusOriginPointer)
	for range 4 {
		focus.BeginFrame()
	}
	if focus.Visible(tag, true) {
		t.Fatal("delayed pointer focus was reported as keyboard-visible")
	}
}

func TestFocusOnPressRecordsPointerOrigin(t *testing.T) {
	var focus Focus
	tag := new(int)
	focus.BeginFrame()
	focus.OnPress(tag, []widget.Press{{}}, 0)
	for range 4 {
		focus.BeginFrame()
	}
	if focus.Visible(tag, true) {
		t.Fatal("pointer press focus was reported as keyboard-visible")
	}
}

func TestKeyboardFocusOriginRemainsVisible(t *testing.T) {
	var focus Focus
	tag := new(int)
	focus.BeginFrame()
	focus.Request(tag, FocusOriginKeyboard)
	for range 4 {
		focus.BeginFrame()
	}
	if !focus.Visible(tag, true) {
		t.Fatal("delayed keyboard focus was hidden")
	}
}

func TestDifferentFocusedTargetClearsStaleOrigin(t *testing.T) {
	var focus Focus
	pointerTarget := new(int)
	keyboardTarget := new(int)
	focus.Request(pointerTarget, FocusOriginPointer)
	focus.ApplyFrameCommands(layout.Context{Ops: new(op.Ops)})
	focus.BeginFrame()
	if !focus.Visible(keyboardTarget, true) {
		t.Fatal("unrequested focus did not default to keyboard-visible")
	}
	focus.Visible(keyboardTarget, false)
	if !focus.Visible(pointerTarget, true) {
		t.Fatal("stale pointer origin hid a later native keyboard focus")
	}
}

func TestCurrentlyFocusedTargetDoesNotClearNewRequest(t *testing.T) {
	var focus Focus
	current := new(int)
	next := new(int)
	focus.Visible(current, true)
	focus.Request(next, FocusOriginPointer)
	if !focus.Visible(current, true) {
		t.Fatal("current keyboard focus changed before the pending command applied")
	}
	if focus.Visible(next, true) {
		t.Fatal("current focus cleared the pending pointer origin")
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
