package flowui

import (
	"image"
	"testing"
	"time"
)

func TestPopoverOptions(t *testing.T) {
	var open bool
	popover := Popover("help", true, Spacer(10, 10), Text("Body")).
		Heading("Help").
		OnOpenChange(func(next bool) {
			open = next
		}).
		Placement(PopoverTop).
		Offset(12).
		ShouldFlip(false).
		AvoidOverflow(false).
		Arrow(true).
		Dismissable(false).
		KeyboardDismissDisabled(true)

	if popover.key != "help" || !popover.open || popover.trigger == nil || popover.content == nil {
		t.Fatal("popover constructor did not set base fields")
	}
	if popover.heading != "Help" || popover.placement != PopoverTop || !popover.hasOffset || popover.offset != 12 {
		t.Fatal("popover visual options were not set")
	}
	if popover.flipEnabled() || popover.overflowAvoidanceEnabled() {
		t.Fatal("popover positioning options were not set")
	}
	if !popover.showArrow() || popover.isDismissable() || !popover.keyboardDismissDisabled {
		t.Fatal("popover behavior options were not set")
	}
	popover.onOpenChange(true)
	if !open {
		t.Fatal("popover onOpenChange did not receive true")
	}
}

func TestPopoverClosedLaysOutTriggerAndDoesNotClaimState(t *testing.T) {
	ctx := newContext(nil)
	dims := Popover("help", false, Spacer(24, 12), Text("Body")).Layout(ctx, testLayoutContext())

	if dims.Size != image.Pt(24, 12) {
		t.Fatalf("closed popover trigger size = %v, want (24,12)", dims.Size)
	}
	if len(ctx.popovers) != 0 {
		t.Fatalf("closed popover claimed state, len = %d", len(ctx.popovers))
	}
}

func TestPopoverOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	Popover("help", true, Spacer(24, 12), Text("Body")).Layout(ctx, testLayoutContext())

	if ctx.popovers["help"] == nil {
		t.Fatal("open popover did not keep state")
	}
}

func TestPopoverClosedKeepsVisibleStateForExitAnimation(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.ready = true
	state.value = 1
	state.from = 1
	state.to = 1
	state.at = start

	gtx := testLayoutContext()
	gtx.Now = start
	Popover("help", false, Spacer(24, 12), Text("Body")).Layout(ctx, gtx)

	if ctx.popovers["help"] == nil {
		t.Fatal("closing popover state was removed before exit animation")
	}
	if state.to != 0 || state.value != 1 {
		t.Fatalf("closing popover animation state = from %v to %v value %v, want closing from visible", state.from, state.to, state.value)
	}
}

func TestPopoverClosedRemovesStateWhenExitAnimationFinishes(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.ready = true
	state.value = 1
	state.from = 1
	state.to = 0
	state.at = start

	gtx := testLayoutContext()
	gtx.Now = start.Add(popoverExitDuration)
	Popover("help", false, Spacer(24, 12), Text("Body")).Layout(ctx, gtx)

	if ctx.popovers["help"] != nil {
		t.Fatal("closed popover state was not removed after exit animation")
	}
}

func TestPopoverDismissAreaRequestsClose(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	state.dismiss[0].Click()

	closed := false
	Popover("help", true, Spacer(24, 12), Text("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		Layout(ctx, testLayoutContext())

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestPopoverDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	state.dismiss[0].Click()

	closed := false
	Popover("help", true, Spacer(24, 12), Text("Body")).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		Layout(ctx, testLayoutContext())

	if closed {
		t.Fatal("non-dismissable popover closed from dismiss area")
	}
}

func TestPopoverPanelPosition(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(40, 20), Text("Body")).Offset(8).AvoidOverflow(false)

	panel := image.Pt(80, 40)
	overlay := image.Pt(300, 200)
	if got := popover.Placement(PopoverBottom).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverBottom); got != image.Pt(-20, 28) {
		t.Fatalf("bottom position = %v", got)
	}
	if got := popover.Placement(PopoverTop).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverTop); got != image.Pt(-20, -48) {
		t.Fatalf("top position = %v", got)
	}
	if got := popover.Placement(PopoverRight).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverRight); got != image.Pt(48, -10) {
		t.Fatalf("right position = %v", got)
	}
	if got := popover.Placement(PopoverLeft).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverLeft); got != image.Pt(-88, -10) {
		t.Fatalf("left position = %v", got)
	}
}

func TestPopoverPanelAlignment(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(40, 20), Text("Body")).Offset(8).AvoidOverflow(false)

	panel := image.Pt(80, 40)
	overlay := image.Pt(300, 200)
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverBottomStart); got != image.Pt(0, 28) {
		t.Fatalf("bottom start position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverBottomEnd); got != image.Pt(-40, 28) {
		t.Fatalf("bottom end position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverRightStart); got != image.Pt(48, 0) {
		t.Fatalf("right start position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlay, PopoverRightEnd); got != image.Pt(48, -20) {
		t.Fatalf("right end position = %v", got)
	}
}

func TestPopoverFlipsBottomAndRightWhenPositiveSpaceIsInsufficient(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), Text("Body")).Offset(8)

	panel := image.Pt(80, 80)
	overlay := image.Pt(120, 160)
	if got := popover.resolvedPlacement(ctx, gtx, image.Pt(80, 120), panel, overlay, PopoverBottom); got != PopoverTop {
		t.Fatalf("bottom resolved placement = %v, want top", got)
	}
	if got := popover.resolvedPlacement(ctx, gtx, image.Pt(100, 80), panel, overlay, PopoverRightEnd); got != PopoverLeftEnd {
		t.Fatalf("right resolved placement = %v, want left end", got)
	}
}

func TestPopoverFlipCanBeDisabled(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), Text("Body")).Offset(8).ShouldFlip(false)

	got := popover.resolvedPlacement(ctx, gtx, image.Pt(80, 120), image.Pt(80, 80), image.Pt(120, 160), PopoverBottom)
	if got != PopoverBottom {
		t.Fatalf("resolved placement = %v, want bottom", got)
	}
}

func TestPopoverAvoidsPositiveOverflow(t *testing.T) {
	pos := popoverAvoidOverflow(image.Pt(260, 180), image.Pt(80, 40), image.Pt(300, 200))

	if pos != image.Pt(220, 160) {
		t.Fatalf("adjusted position = %v, want (220,160)", pos)
	}
}

func TestPopoverProgressAnimation(t *testing.T) {
	state := new(popoverState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.progress(gtx, true); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	gtx.Now = start.Add(popoverEnterDuration / 2)
	mid := state.progress(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("progress midpoint = %v, want between 0 and 1", mid)
	}
	gtx.Now = start.Add(popoverEnterDuration)
	if got := state.progress(gtx, true); got != 1 {
		t.Fatalf("progress end = %v, want 1", got)
	}
}

func TestPopoverBodyTextStyle(t *testing.T) {
	ctx := newContext(nil)
	body, ok := Popover("help", true, Spacer(24, 12), Text("Body")).
		styleContent(ctx, Text("Body"), popoverStyleFor(ctx.Theme)).(TextWidget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	if body.size != ctx.Theme.Components.Popover.BodyTextSize {
		t.Fatalf("body text size = %v, want %v", body.size, ctx.Theme.Components.Popover.BodyTextSize)
	}
	if body.color != ctx.Theme.Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", body.color)
	}
}

func popoverTestContextWithState(key string) (*Context, *popoverState) {
	state := new(popoverState)
	ctx := newContext(nil)
	ctx.popovers = map[string]*popoverState{key: state}
	return ctx, state
}
