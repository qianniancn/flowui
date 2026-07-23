package popover

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
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
	frame.BeginFrame(ctx)
	Popover("help", true, Spacer(24, 12), text.New("Body")).Layout(ctx, testLayoutContext())

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) == nil {
		t.Fatal("open popover did not keep state")
	}
}

func TestPopoverClosedKeepsVisibleStateForExitAnimation(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.transition.Set(1, 1, start)

	gtx := testLayoutContext()
	gtx.Now = start
	frame.BeginFrame(ctx)
	Popover("help", false, Spacer(24, 12), text.New("Body")).Layout(ctx, gtx)

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) == nil {
		t.Fatal("closing popover state was removed before exit animation")
	}
	if state.transition.Target() != 0 || state.transition.Current() != 1 {
		t.Fatalf("closing popover animation state = target %v value %v, want closing from visible", state.transition.Target(), state.transition.Current())
	}
}

func TestPopoverClosedRemovesStateWhenExitAnimationFinishes(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.transition.Set(1, 0, start)

	gtx := testLayoutContext()
	gtx.Now = start.Add(popoverExitDuration)
	frame.BeginFrame(ctx)
	Popover("help", false, Spacer(24, 12), text.New("Body")).Layout(ctx, gtx)

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) != nil {
		t.Fatal("closed popover state was not removed after exit animation")
	}
}

func TestPopoverDismissAreaRequestsClose(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	state.dismiss[0].Click()

	closed := false
	popover := Popover("help", true, Spacer(24, 12), text.New("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutPopoverFrame(ctx, popover)

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestPopoverDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	state.dismiss[0].Click()

	closed := false
	popover := Popover("help", true, Spacer(24, 12), text.New("Body")).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutPopoverFrame(ctx, popover)

	if closed {
		t.Fatal("non-dismissable popover closed from dismiss area")
	}
}

func TestPopoverPanelPosition(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(40, 20), text.New("Body")).Offset(8).AvoidOverflow(false)

	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.Placement(overlay.PopoverBottom).resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverBottom).Position; got != image.Pt(80, 78) {
		t.Fatalf("bottom position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverTop).resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverTop).Position; got != image.Pt(80, 2) {
		t.Fatalf("top position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverRight).resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverRight).Position; got != image.Pt(148, 40) {
		t.Fatalf("right position = %v", got)
	}
	if got := popover.Placement(overlay.PopoverLeft).resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverLeft).Position; got != image.Pt(12, 40) {
		t.Fatalf("left position = %v", got)
	}
}

func TestPopoverPanelAlignment(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(40, 20), text.New("Body")).Offset(8).AvoidOverflow(false)

	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverBottomStart).Position; got != image.Pt(100, 78) {
		t.Fatalf("bottom start position = %v", got)
	}
	if got := popover.resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverBottomEnd).Position; got != image.Pt(60, 78) {
		t.Fatalf("bottom end position = %v", got)
	}
	if got := popover.resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverRightStart).Position; got != image.Pt(148, 50) {
		t.Fatalf("right start position = %v", got)
	}
	if got := popover.resolvedPosition(ctx, gtx, trigger, panel, overlaySize, overlay.PopoverRightEnd).Position; got != image.Pt(148, 30) {
		t.Fatalf("right end position = %v", got)
	}
}

func TestPopoverFlipsBottomAndRightWhenPositiveSpaceIsInsufficient(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), text.New("Body")).Offset(8)

	panel := image.Pt(80, 80)
	overlaySize := image.Pt(120, 160)
	if got := popover.resolvedPosition(ctx, gtx, image.Rect(20, 120, 100, 140), panel, overlaySize, overlay.PopoverBottom).Placement.PopoverPlacement(); got != overlay.PopoverTop {
		t.Fatalf("bottom resolved placement = %v, want top", got)
	}
	if got := popover.resolvedPosition(ctx, gtx, image.Rect(100, 40, 120, 120), panel, overlaySize, overlay.PopoverRightEnd).Placement.PopoverPlacement(); got != overlay.PopoverLeftEnd {
		t.Fatalf("right resolved placement = %v, want left end", got)
	}
}

func TestPopoverFlipCanBeDisabled(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := Popover("help", true, Spacer(80, 120), text.New("Body")).Offset(8).ShouldFlip(false)

	got := popover.resolvedPosition(ctx, gtx, image.Rect(20, 120, 100, 140), image.Pt(80, 80), image.Pt(120, 160), overlay.PopoverBottom).Placement.PopoverPlacement()
	if got != overlay.PopoverBottom {
		t.Fatalf("resolved placement = %v, want bottom", got)
	}
}

func TestPopoverFlipLaysOutNestedContentOnce(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(200, 160)
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	}
	nested := Popover("nested", true, Spacer(24, 16), text.New("Nested body"))
	content := &countingPopoverContent{child: nested}
	outer := Popover("outer", true, Spacer(24, 16), content).
		Heading("Outer heading").
		Arrow(true)

	frame.BeginFrameWithViewport(ctx, viewport)
	layoutui.Box(outer).PaddingLeft(50).PaddingTop(130).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)

	if content.layouts != 1 {
		t.Fatalf("outer content layouts = %d, want 1 after placement flip", content.layouts)
	}
	if testComponentState[popoverState](ctx, "nested", stateSlotPopover) == nil {
		t.Fatal("nested popover was not registered while laying out outer content")
	}
}

func TestPopoverArrowBlocksPointerWithoutDismissing(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.transition.Set(1, 1, start)
	router := new(input.Router)
	background := new(widget.Clickable)
	backgroundClicked := false
	closed := false
	popover := Popover("help", true, Spacer(24, 16), Spacer(40, 20)).
		Arrow(true).
		OnOpenChange(func(open bool) { closed = !open })

	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start)
	position := f32.Pt(36, 20)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start.Add(time.Millisecond))
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start.Add(2*time.Millisecond))

	if closed {
		t.Fatal("clicking the popover arrow dismissed it")
	}
	if backgroundClicked {
		t.Fatal("popover arrow click reached the background")
	}
}

func TestPopoverAnimatedPanelEdgeHasNoPointerHole(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.transition.Set(.5, 1, start)
	router := new(input.Router)
	background := new(widget.Clickable)
	backgroundClicked := false
	open := true
	build := func() PopoverWidget {
		return Popover("help", open, Spacer(24, 16), Spacer(40, 20)).
			OnOpenChange(func(next bool) { open = next })
	}

	layoutPopoverOverBackgroundFrame(ctx, router, build(), background, &backgroundClicked, start)
	position := f32.Pt(.5, 40)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, build(), background, &backgroundClicked, start.Add(time.Millisecond))
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, build(), background, &backgroundClicked, start.Add(2*time.Millisecond))

	if open {
		t.Fatal("click beside the animated popover did not dismiss it")
	}
	if backgroundClicked {
		t.Fatal("animated popover edge allowed the click to reach the background")
	}
}

func TestPopoverExitPanelStillBlocksBackground(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	start := time.Unix(1, 0)
	state.transition.Set(1, 1, start)
	router := new(input.Router)
	background := new(widget.Clickable)
	backgroundClicked := false
	popover := Popover("help", false, Spacer(24, 16), Spacer(40, 20))

	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start)
	position := f32.Pt(10, 30)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start.Add(time.Millisecond))
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutPopoverOverBackgroundFrame(ctx, router, popover, background, &backgroundClicked, start.Add(2*time.Millisecond))

	if backgroundClicked {
		t.Fatal("exiting popover panel allowed a click to reach the background")
	}
}

func TestPopoverAvoidsPositiveOverflow(t *testing.T) {
	pos := popoverAvoidOverflow(image.Pt(260, 180), image.Pt(80, 40), image.Pt(300, 200))

	if pos != image.Pt(220, 160) {
		t.Fatalf("adjusted position = %v, want (220,160)", pos)
	}
}

func TestPopoverSlideOffsetUsesThemeDistance(t *testing.T) {
	themeValue := theme.DefaultTheme()
	themeValue.Components.Popover.AnimationDistance = 10
	ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
	gtx := testLayoutContext()

	got := popoverSlideOffset(ctx, gtx, 0, overlay.PopoverBottom)
	if want := image.Pt(0, -gtx.Dp(10)); got != want {
		t.Fatalf("popover slide offset = %v, want %v", got, want)
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
	resolved := text.ResolveStyleStatic(ctx, body)
	if resolved.Text == nil || resolved.Text.FontSize == nil || *resolved.Text.FontSize != frame.ActiveTheme(ctx).Components.Popover.BodyTextSize {
		t.Fatalf("body text style = %#v", resolved.Text)
	}
	if col := resolved.Text.Color.(flowstyle.SolidColor).Color; col != frame.ActiveTheme(ctx).Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", col)
	}
}

func popoverTestContextWithState(key string) (*frame.Context, *popoverState) {
	state := new(popoverState)
	ctx := newContext(nil)
	testSetComponentState(ctx, key, stateSlotPopover, state)
	return ctx, state
}

func layoutPopoverFrame(ctx *frame.Context, popover PopoverWidget) {
	gtx := testLayoutContext()
	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	popover.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
}

func layoutPopoverOverBackgroundFrame(ctx *frame.Context, router *input.Router, popover PopoverWidget, background *widget.Clickable, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(300, 200)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	popover.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

type countingPopoverContent struct {
	child   frame.Widget
	layouts int
}

func (c *countingPopoverContent) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	c.layouts++
	return c.child.Layout(ctx, gtx)
}
