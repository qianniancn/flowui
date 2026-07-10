package popover

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func newContextWithTheme(_ any, value *theme.Theme) *frame.Context {
	return frame.New(nil, value, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func Spacer(width, height int) layoutui.SpacerWidget {
	return layoutui.Spacer(width, height)
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *frame.Context, key, slot string, value *T) {
	frame.UseStateWith(ctx, key, slot, func() *T { return value })
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func TestPopoverOptions(t *testing.T) {
	var open bool
	popover := Popover("help", true, Spacer(10, 10), text.New("Body")).
		Heading("Help").
		OnOpenChange(func(next bool) {
			open = next
		}).
		Placement(overlay.PopoverTop).
		Offset(12).
		ShouldFlip(false).
		AvoidOverflow(false).
		Arrow(true).
		Dismissable(false).
		KeyboardDismissDisabled(true)

	if popover.key != "help" || !popover.open || popover.trigger == nil || popover.content == nil {
		t.Fatal("popover constructor did not set base fields")
	}
	if popover.heading != "Help" || popover.placement != overlay.PopoverTop || !popover.hasOffset || popover.offset != 12 {
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
	dims := Popover("help", false, Spacer(24, 12), text.New("Body")).Layout(ctx, testLayoutContext())

	if dims.Size != image.Pt(24, 12) {
		t.Fatalf("closed popover trigger size = %v, want (24,12)", dims.Size)
	}
	if frame.StateLen(ctx) != 0 {
		t.Fatalf("closed popover claimed state, len = %d", frame.StateLen(ctx))
	}
}

func TestPopoverOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	Popover("help", true, Spacer(24, 12), text.New("Body")).Layout(ctx, testLayoutContext())

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) == nil {
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
	Popover("help", false, Spacer(24, 12), text.New("Body")).Layout(ctx, gtx)

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) == nil {
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
	Popover("help", false, Spacer(24, 12), text.New("Body")).Layout(ctx, gtx)

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) != nil {
		t.Fatal("closed popover state was not removed after exit animation")
	}
}

func TestPopoverDismissAreaRequestsClose(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	state.dismiss[0].Click()

	closed := false
	Popover("help", true, Spacer(24, 12), text.New("Body")).
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
	Popover("help", true, Spacer(24, 12), text.New("Body")).
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
	popover := Popover("help", true, Spacer(40, 20), text.New("Body")).Offset(8).AvoidOverflow(false)

	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.Placement(overlay.PopoverBottom).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverBottom); got != image.Pt(-20, 28) {
		t.Fatalf("bottom position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverTop).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverTop); got != image.Pt(-20, -48) {
		t.Fatalf("top position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverRight).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverRight); got != image.Pt(48, -10) {
		t.Fatalf("right position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverLeft).panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverLeft); got != image.Pt(-88, -10) {
		t.Fatalf("left position = %v", got)
	}
}

func TestPopoverPanelAlignment(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(40, 20), text.New("Body")).Offset(8).AvoidOverflow(false)

	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverBottomStart); got != image.Pt(0, 28) {
		t.Fatalf("bottom start position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverBottomEnd); got != image.Pt(-40, 28) {
		t.Fatalf("bottom end position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverRightStart); got != image.Pt(48, 0) {
		t.Fatalf("right start position = %v", got)
	}
	if got := popover.panelPosition(ctx, gtx, image.Pt(40, 20), panel, overlaySize, overlay.PopoverRightEnd); got != image.Pt(48, -20) {
		t.Fatalf("right end position = %v", got)
	}
}

func TestPopoverFlipsBottomAndRightWhenPositiveSpaceIsInsufficient(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), text.New("Body")).Offset(8)

	panel := image.Pt(80, 80)
	overlaySize := image.Pt(120, 160)
	if got := popover.resolvedPlacement(ctx, gtx, image.Pt(80, 120), panel, overlaySize, overlay.PopoverBottom); got != overlay.PopoverTop {
		t.Fatalf("bottom resolved placement = %v, want top", got)
	}
	if got := popover.resolvedPlacement(ctx, gtx, image.Pt(100, 80), panel, overlaySize, overlay.PopoverRightEnd); got != overlay.PopoverLeftEnd {
		t.Fatalf("right resolved placement = %v, want left end", got)
	}
}

func TestPopoverFlipCanBeDisabled(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), text.New("Body")).Offset(8).ShouldFlip(false)

	got := popover.resolvedPlacement(ctx, gtx, image.Pt(80, 120), image.Pt(80, 80), image.Pt(120, 160), overlay.PopoverBottom)
	if got != overlay.PopoverBottom {
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
	body, ok := Popover("help", true, Spacer(24, 12), text.New("Body")).
		styleContent(ctx, text.New("Body"), popoverStyleFor(frame.ActiveTheme(ctx))).(text.Widget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	if body.ConfiguredSize() != frame.ActiveTheme(ctx).Components.Popover.BodyTextSize {
		t.Fatalf("body text size = %v, want %v", body.ConfiguredSize(), frame.ActiveTheme(ctx).Components.Popover.BodyTextSize)
	}
	if col, _ := body.ConfiguredColor(); col != frame.ActiveTheme(ctx).Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", col)
	}
}

func popoverTestContextWithState(key string) (*frame.Context, *popoverState) {
	state := new(popoverState)
	ctx := newContext(nil)
	testSetComponentState(ctx, key, stateSlotPopover, state)
	return ctx, state
}
