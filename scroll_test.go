package flowui

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestScrollKeepsState(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll("body", Spacer(20, 200)).Horizontal().Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	})

	state := ctx.scrolls["body"]
	if state == nil {
		t.Fatal("missing scroll state")
	}
	if state.Axis != layout.Horizontal {
		t.Fatalf("axis = %v, want horizontal", state.Axis)
	}
}

func TestScrollPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	Scroll("body", probe).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("scroll child was laid out with enabled context")
	}
}

func TestScrollOptions(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Scroll("body", Spacer(20, 200)).
		AlignMiddle().
		StickToEnd().
		ScrollAnyAxis().
		Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(100, 100)},
			Ops:         &ops,
		})

	state := ctx.scrolls["body"]
	if state.Alignment != layout.Middle {
		t.Fatalf("alignment = %v, want middle", state.Alignment)
	}
	if !state.ScrollToEnd {
		t.Fatal("scroll-to-end was not enabled")
	}
	if !state.ScrollAnyAxis {
		t.Fatal("scroll-any-axis was not enabled")
	}
}
