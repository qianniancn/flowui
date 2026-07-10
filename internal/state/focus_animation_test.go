package state

import (
	"testing"

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
