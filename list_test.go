package flowui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestListKeepsState(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	List("items", 2, func(i int) Widget {
		return Text("item")
	}).Gap(8).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	state := ctx.lists["items"]
	if state == nil {
		t.Fatal("missing list state")
	}
	if state.Axis != layout.Vertical {
		t.Fatalf("axis = %v, want vertical", state.Axis)
	}
	if state.Gap != 8 {
		t.Fatalf("gap = %d, want 8", state.Gap)
	}
}

func TestListDisabled(t *testing.T) {
	l := List("items", 1, func(int) Widget {
		return Text("item")
	}).Disabled(true)

	if !l.disabled {
		t.Fatal("list was not disabled")
	}
}

func TestListPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	List("items", 1, func(int) Widget {
		return probe
	}).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("list item was laid out with enabled context")
	}
}

func TestListScrollOptions(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	List("items", 1, func(int) Widget {
		return Spacer(20, 20)
	}).AlignEnd().StickToEnd().ScrollAnyAxis().Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	state := ctx.lists["items"]
	if state.Alignment != layout.End {
		t.Fatalf("alignment = %v, want end", state.Alignment)
	}
	if !state.ScrollToEnd {
		t.Fatal("scroll-to-end was not enabled")
	}
	if !state.ScrollAnyAxis {
		t.Fatal("scroll-any-axis was not enabled")
	}
}
