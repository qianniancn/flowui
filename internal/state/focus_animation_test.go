package state

import (
	"testing"
	"time"

	"gioui.org/widget"
)

func TestFocusAnimationDistinguishesPointerFocus(t *testing.T) {
	var state FocusAnimation
	if !state.Visible(true, nil) {
		t.Fatal("keyboard focus was not visible")
	}
	state.Visible(false, nil)
	if state.Visible(true, []widget.Press{{}}) {
		t.Fatal("pointer focus was visible")
	}
}

func TestFocusAnimationDoesNotTreatStaleHistoryAsPointerFocus(t *testing.T) {
	var animation FocusAnimation
	start := time.Unix(1, 0)
	history := []widget.Press{{Start: start}}

	animation.Visible(false, nil)
	animation.Visible(false, history)
	if animation.Visible(true, history) {
		t.Fatal("pointer focus was visible")
	}
	animation.Visible(false, history)
	history[0].End = start.Add(time.Millisecond)
	if !animation.Visible(true, history) {
		t.Fatal("stale pointer history hid later keyboard focus")
	}
}

func TestFocusAnimationPointerPressHidesExistingKeyboardFocus(t *testing.T) {
	var animation FocusAnimation
	if !animation.Visible(true, nil) {
		t.Fatal("keyboard focus was not visible")
	}
	if animation.Visible(true, []widget.Press{{Start: time.Unix(1, 0)}}) {
		t.Fatal("pointer interaction did not hide keyboard focus")
	}
}

func TestFocusAnimationDropsUnappliedPointerFocus(t *testing.T) {
	var animation FocusAnimation
	history := []widget.Press{{Start: time.Unix(1, 0)}}
	animation.Visible(false, nil)
	animation.Visible(false, history)
	animation.Visible(false, history)
	animation.Visible(false, history)
	if !animation.Visible(true, history) {
		t.Fatal("unapplied pointer focus poisoned later keyboard focus")
	}
}
