package overlay

import (
	"image"
	"testing"
)

func TestArrowAnchorClampsToRoundedPanelCorners(t *testing.T) {
	trigger := image.Rect(0, 0, 20, 20)
	panel := image.Pt(100, 50)

	if got := ArrowAnchor(trigger, image.Pt(0, 30), panel, PopoverBottom, 12, 8); got != 16 {
		t.Fatalf("left-clamped arrow anchor = %v, want 16", got)
	}
	if got := ArrowAnchor(image.Rect(80, 0, 100, 20), image.Pt(0, 30), panel, PopoverBottom, 12, 8); got != 84 {
		t.Fatalf("right-clamped arrow anchor = %v, want 84", got)
	}
	if got := ArrowAnchor(image.Rect(40, 0, 60, 20), image.Pt(0, 30), panel, PopoverBottom, 12, 8); got != 50 {
		t.Fatalf("centered arrow anchor = %v, want 50", got)
	}
}

func TestArrowRectExtendsFromTheCorrectPanelEdge(t *testing.T) {
	panel := image.Pt(100, 50)
	want := map[Side]struct {
		placement PopoverPlacement
		rect      image.Rectangle
	}{
		SideTop:    {PopoverTop, image.Rect(34, 50, 47, 59)},
		SideBottom: {PopoverBottom, image.Rect(34, -8, 47, 1)},
		SideLeft:   {PopoverLeft, image.Rect(100, 34, 109, 47)},
		SideRight:  {PopoverRight, image.Rect(-8, 34, 1, 47)},
	}
	for side, expected := range want {
		got := ArrowRect(panel, expected.placement, 40, 12)
		if got != expected.rect {
			t.Errorf("ArrowRect(%v) = %v, want %v", side, got, expected.rect)
		}
	}
}
