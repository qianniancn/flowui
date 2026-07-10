package flowui

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
)

func TestDeferOverlayRecordsInputOps(t *testing.T) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(100, 100)},
		Ops:         &ops,
	}
	tag := new(int)

	newContext(nil).deferOverlay(gtx, func(gtx layout.Context) layout.Dimensions {
		event.Op(gtx.Ops, tag)
		return layout.Dimensions{Size: image.Pt(10, 10)}
	})

	var router input.Router
	router.Frame(&ops)
}

func TestOverlayDismissRectsExcludeTarget(t *testing.T) {
	bounds := image.Rect(-20, -10, 100, 80)
	excluded := image.Rect(20, 10, 80, 60)
	rects := overlayDismissRects(bounds, excluded)

	if rects[0] != image.Rect(-20, -10, 100, 10) {
		t.Fatalf("top dismiss rect = %v", rects[0])
	}
	for _, rect := range rects {
		if !rect.Intersect(excluded).Empty() {
			t.Fatalf("dismiss rect %v overlaps excluded target %v", rect, excluded)
		}
	}
}

func TestOverlayDismissRectsCoverBoundsWhenTargetIsOutside(t *testing.T) {
	bounds := image.Rect(0, 0, 100, 80)
	rects := overlayDismissRects(bounds, image.Rect(120, 90, 140, 110))

	if rects[0] != bounds {
		t.Fatalf("full dismiss rect = %v, want %v", rects[0], bounds)
	}
}

func TestOverlayRawPositionAlignments(t *testing.T) {
	trigger := image.Pt(40, 20)
	panel := image.Pt(80, 40)

	if got := overlayRawPosition(trigger, panel, 8, overlayPlacement{side: overlaySideBottom, align: overlayAlignCenter}); got != image.Pt(-20, 28) {
		t.Fatalf("bottom center position = %v, want (-20,28)", got)
	}
	if got := overlayRawPosition(trigger, panel, 8, overlayPlacement{side: overlaySideBottom, align: overlayAlignStart}); got != image.Pt(0, 28) {
		t.Fatalf("bottom start position = %v, want (0,28)", got)
	}
	if got := overlayRawPosition(trigger, panel, 8, overlayPlacement{side: overlaySideRight, align: overlayAlignEnd}); got != image.Pt(48, -20) {
		t.Fatalf("right end position = %v, want (48,-20)", got)
	}
}

func TestOverlayResolvePositionFlipsAndAvoidsOverflow(t *testing.T) {
	result := overlayResolvePosition(overlayPositionConfig{
		Trigger:       image.Pt(80, 120),
		Panel:         image.Pt(80, 80),
		Bounds:        image.Pt(120, 160),
		Offset:        8,
		Placement:     overlayPlacement{side: overlaySideBottom, align: overlayAlignEnd},
		Flip:          true,
		AvoidOverflow: true,
	})

	if result.Placement != (overlayPlacement{side: overlaySideTop, align: overlayAlignEnd}) {
		t.Fatalf("resolved placement = %#v, want top end", result.Placement)
	}
	if result.Position != image.Pt(0, -88) {
		t.Fatalf("resolved position = %v, want (0,-88)", result.Position)
	}
	if result.Rect != image.Rect(0, -88, 80, -8) {
		t.Fatalf("resolved rect = %v, want (0,-88)-(80,-8)", result.Rect)
	}
}

func TestOverlayResolvePlacementFlipsTopAndLeft(t *testing.T) {
	top := overlayResolvePlacement(
		image.Pt(80, 20),
		image.Pt(80, 80),
		image.Pt(160, 180),
		8,
		overlayPlacement{side: overlaySideTop, align: overlayAlignStart},
	)
	if top != (overlayPlacement{side: overlaySideBottom, align: overlayAlignStart}) {
		t.Fatalf("top resolved placement = %#v, want bottom start", top)
	}

	left := overlayResolvePlacement(
		image.Pt(20, 80),
		image.Pt(80, 80),
		image.Pt(180, 160),
		8,
		overlayPlacement{side: overlaySideLeft, align: overlayAlignEnd},
	)
	if left != (overlayPlacement{side: overlaySideRight, align: overlayAlignEnd}) {
		t.Fatalf("left resolved placement = %#v, want right end", left)
	}
}

func TestOverlayPanelTransformOriginUsesPanelLocalTriggerCenter(t *testing.T) {
	origin := overlayPanelTransformOrigin(
		image.Pt(41, 20),
		image.Pt(-30, 28),
		image.Pt(120, 60),
		overlayPlacement{side: overlaySideBottom, align: overlayAlignStart},
	)

	if origin.X != 50.5 || origin.Y != 0 {
		t.Fatalf("origin = %v, want (50.5,0)", origin)
	}
}

func TestOverlayPositionSupportsOffsetTriggerRect(t *testing.T) {
	result := overlayResolvePosition(overlayPositionConfig{
		Trigger:          image.Pt(120, 36),
		TriggerOrigin:    image.Pt(10, 24),
		HasTriggerOrigin: true,
		Panel:            image.Pt(120, 80),
		Bounds:           image.Pt(300, 240),
		Offset:           6,
		Placement:        overlayPlacement{side: overlaySideBottom, align: overlayAlignStart},
	})

	if result.Position != image.Pt(10, 66) {
		t.Fatalf("offset trigger position = %v, want (10,66)", result.Position)
	}
	origin := overlayPanelTransformOriginAt(image.Rect(10, 24, 130, 60), result.Position, result.Rect.Size(), result.Placement)
	if origin != (f32.Point{X: 60, Y: 0}) {
		t.Fatalf("offset trigger origin = %v, want (60,0)", origin)
	}
}
