package overlay

import (
	"image"
	"testing"
)

func TestPopoverPlacementMapping(t *testing.T) {
	tests := []struct {
		name      string
		popover   PopoverPlacement
		placement Placement
	}{
		{name: "bottom", popover: PopoverBottom, placement: Placement{Side: SideBottom, Align: AlignCenter}},
		{name: "top", popover: PopoverTop, placement: Placement{Side: SideTop, Align: AlignCenter}},
		{name: "left", popover: PopoverLeft, placement: Placement{Side: SideLeft, Align: AlignCenter}},
		{name: "right", popover: PopoverRight, placement: Placement{Side: SideRight, Align: AlignCenter}},
		{name: "bottom start", popover: PopoverBottomStart, placement: Placement{Side: SideBottom, Align: AlignStart}},
		{name: "bottom end", popover: PopoverBottomEnd, placement: Placement{Side: SideBottom, Align: AlignEnd}},
		{name: "top start", popover: PopoverTopStart, placement: Placement{Side: SideTop, Align: AlignStart}},
		{name: "top end", popover: PopoverTopEnd, placement: Placement{Side: SideTop, Align: AlignEnd}},
		{name: "left start", popover: PopoverLeftStart, placement: Placement{Side: SideLeft, Align: AlignStart}},
		{name: "left end", popover: PopoverLeftEnd, placement: Placement{Side: SideLeft, Align: AlignEnd}},
		{name: "right start", popover: PopoverRightStart, placement: Placement{Side: SideRight, Align: AlignStart}},
		{name: "right end", popover: PopoverRightEnd, placement: Placement{Side: SideRight, Align: AlignEnd}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.popover.Placement(); got != test.placement {
				t.Fatalf("PopoverPlacement.Placement() = %#v, want %#v", got, test.placement)
			}
			if got := test.placement.PopoverPlacement(); got != test.popover {
				t.Fatalf("Placement.PopoverPlacement() = %v, want %v", got, test.popover)
			}
		})
	}
}

func TestPopoverPlacementMappingDefaultsToBottomCenter(t *testing.T) {
	want := Placement{Side: SideBottom, Align: AlignCenter}
	if got := PopoverPlacement(255).Placement(); got != want {
		t.Fatalf("invalid PopoverPlacement maps to %#v, want %#v", got, want)
	}
	if got := (Placement{Side: Side(255), Align: Align(255)}).PopoverPlacement(); got != PopoverBottom {
		t.Fatalf("invalid Placement maps to %v, want %v", got, PopoverBottom)
	}
}

func TestAvoidOverflowPreservesNegativeLocalPlacement(t *testing.T) {
	position := image.Pt(-88, -10)
	if got := AvoidOverflow(position, image.Pt(80, 40), image.Pt(300, 200)); got != position {
		t.Fatalf("AvoidOverflow() = %v, want local top-left position %v", got, position)
	}
}

func TestResolvePositionConstrainsViewportRelativePlacement(t *testing.T) {
	result := ResolvePosition(PositionConfig{
		Trigger:          image.Pt(20, 20),
		TriggerOrigin:    image.Pt(5, 4),
		HasTriggerOrigin: true,
		Panel:            image.Pt(80, 60),
		Bounds:           image.Pt(100, 80),
		Placement:        Placement{Side: SideTop, Align: AlignEnd},
		AvoidOverflow:    true,
	})
	if result.Position != (image.Point{}) {
		t.Fatalf("viewport position = %v, want origin", result.Position)
	}
}

func TestAvoidViewportOverflowPinsOversizedPanelToOrigin(t *testing.T) {
	got := AvoidViewportOverflow(image.Pt(10, 10), image.Pt(140, 90), image.Pt(100, 80))
	if got != (image.Point{}) {
		t.Fatalf("oversized panel position = %v, want origin", got)
	}
}
