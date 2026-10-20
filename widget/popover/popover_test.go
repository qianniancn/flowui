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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

const (
	popoverEnterDuration = 150 * time.Millisecond
	popoverExitDuration  = 100 * time.Millisecond
)

func (s *popoverState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	return s.progressFor(gtx, open, popoverEnterDuration, popoverExitDuration, motions...)
}

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.NeutralTheme()
}

func Spacer(width, height int) layoutui.SpacerWidget {
	return layoutui.Spacer(width, height)
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *core.Context, key, slot string, value *T) {
	core.UseStateWith(ctx, key, slot, func() *T { return value })
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

func popoverTestVisual(ctx *core.Context, value Widget) theme.PopoverVisual {
	return value.visualWithCache(ctx, nil, nil)
}

func TestPopoverClosedLaysOutTriggerAndDoesNotClaimState(t *testing.T) {
	ctx := newContext(nil)
	dims := New(Spacer(24, 12), text.New("Body")).Key("help").Layout(ctx, testLayoutContext())

	if dims.Size != image.Pt(24, 12) {
		t.Fatalf("closed popover trigger size = %v, want (24,12)", dims.Size)
	}
	if core.StateLen(ctx) != 0 {
		t.Fatalf("closed popover claimed state, len = %d", core.StateLen(ctx))
	}
}

func TestPopoverOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	core.BeginFrame(ctx)
	New(Spacer(24, 12), text.New("Body")).Key("help").Open(true).Layout(ctx, testLayoutContext())

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
	core.BeginFrame(ctx)
	New(Spacer(24, 12), text.New("Body")).Key("help").Layout(ctx, gtx)

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
	core.BeginFrame(ctx)
	New(Spacer(24, 12), text.New("Body")).Key("help").Layout(ctx, gtx)

	if testComponentState[popoverState](ctx, "help", stateSlotPopover) != nil {
		t.Fatal("closed popover state was not removed after exit animation")
	}
}

func TestPopoverDismissAreaRequestsClose(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	closed := false
	popover := New(Spacer(24, 12), text.New("Body")).Key("help").Open(true).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutPopoverFrame(ctx, popover)
	state.dismiss[0].Click()
	layoutPopoverFrame(ctx, popover)

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestPopoverDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := popoverTestContextWithState("help")
	closed := false
	popover := New(Spacer(24, 12), text.New("Body")).Key("help").Open(true).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutPopoverFrame(ctx, popover)
	state.dismiss[0].Click()
	layoutPopoverFrame(ctx, popover)

	if closed {
		t.Fatal("non-dismissable popover closed from dismiss area")
	}
}

func TestPopoverPanelPosition(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := New(Spacer(40, 20), text.New("Body")).Key("help").Open(true).Offset(8).AvoidOverflow(false)
	visual := popoverTestVisual(ctx, popover)

	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.Placement(core.PopoverBottom).resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverBottom, visual).Position; got != image.Pt(80, 78) {
		t.Fatalf("bottom position = %v", got)
	}
	if got := popover.Placement(core.PopoverTop).resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverTop, visual).Position; got != image.Pt(80, 2) {
		t.Fatalf("top position = %v", got)
	}
	if got := popover.Placement(core.PopoverRight).resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverRight, visual).Position; got != image.Pt(148, 40) {
		t.Fatalf("right position = %v", got)
	}
	if got := popover.Placement(core.PopoverLeft).resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverLeft, visual).Position; got != image.Pt(12, 40) {
		t.Fatalf("left position = %v", got)
	}
}

func TestPopoverPanelAlignment(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := New(Spacer(40, 20), text.New("Body")).Key("help").Open(true).Offset(8).AvoidOverflow(false)
	visual := popoverTestVisual(ctx, popover)

	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(80, 40)
	overlaySize := image.Pt(300, 200)
	if got := popover.resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverBottomStart, visual).Position; got != image.Pt(100, 78) {
		t.Fatalf("bottom start position = %v", got)
	}
	if got := popover.resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverBottomEnd, visual).Position; got != image.Pt(60, 78) {
		t.Fatalf("bottom end position = %v", got)
	}
	if got := popover.resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverRightStart, visual).Position; got != image.Pt(148, 50) {
		t.Fatalf("right start position = %v", got)
	}
	if got := popover.resolvedPositionWithVisual(ctx, gtx, trigger, panel, overlaySize, core.PopoverRightEnd, visual).Position; got != image.Pt(148, 30) {
		t.Fatalf("right end position = %v", got)
	}
}

func TestPopoverFlipsBottomAndRightWhenPositiveSpaceIsInsufficient(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := New(Spacer(80, 120), text.New("Body")).Key("help").Open(true).Offset(8)
	visual := popoverTestVisual(ctx, popover)

	panel := image.Pt(80, 80)
	overlaySize := image.Pt(120, 160)
	if got := popover.resolvedPositionWithVisual(ctx, gtx, image.Rect(20, 120, 100, 140), panel, overlaySize, core.PopoverBottom, visual).Placement.PopoverPlacement(); got != core.PopoverTop {
		t.Fatalf("bottom resolved placement = %v, want top", got)
	}
	if got := popover.resolvedPositionWithVisual(ctx, gtx, image.Rect(100, 40, 120, 120), panel, overlaySize, core.PopoverRightEnd, visual).Placement.PopoverPlacement(); got != core.PopoverLeftEnd {
		t.Fatalf("right resolved placement = %v, want left end", got)
	}
}

func TestPopoverFlipCanBeDisabled(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	popover := New(Spacer(80, 120), text.New("Body")).Key("help").Open(true).Offset(8).ShouldFlip(false)
	visual := popoverTestVisual(ctx, popover)

	got := popover.resolvedPositionWithVisual(ctx, gtx, image.Rect(20, 120, 100, 140), image.Pt(80, 80), image.Pt(120, 160), core.PopoverBottom, visual).Placement.PopoverPlacement()
	if got != core.PopoverBottom {
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
	nested := New(Spacer(24, 16), text.New("Nested body")).Key("nested").Open(true)
	content := &countingPopoverContent{child: nested}
	outer := New(Spacer(24, 16), content).Key("outer").Open(true).
		Heading("Outer heading").
		Arrow(true)

	core.BeginFrameWithViewport(ctx, viewport)
	layoutui.Box(outer).Style(style.Style{}.PaddingLeft(50).PaddingTop(130)).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)

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
	popover := New(Spacer(24, 16), Spacer(40, 20)).Key("help").Open(true).
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
	build := func() Widget {
		return New(Spacer(24, 16), Spacer(40, 20)).Key("help").Open(open).
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
	popover := New(Spacer(24, 16), Spacer(40, 20)).Key("help")

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
	pos := core.AvoidOverflow(image.Pt(260, 180), image.Pt(80, 40), image.Pt(300, 200))

	if pos != image.Pt(220, 160) {
		t.Fatalf("adjusted position = %v, want (220,160)", pos)
	}
}

func TestPopoverSlideOffsetUsesThemeDistance(t *testing.T) {
	themeValue := *testtheme.Behavior()
	themeValue.Components.Popover.AnimationDistance = 10
	ctx := core.New(nil, &themeValue, locale.LanguageAuto)
	gtx := testLayoutContext()
	visual := popoverTestVisual(ctx, New(nil, nil).Key("help").Open(true))

	got := popoverSlideOffsetWithVisual(gtx, 0, core.PopoverBottom, visual)
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
	active := core.ActiveTheme(ctx)
	popover := New(Spacer(24, 12), text.New("Body")).Key("help").Open(true)
	visual := popoverTestVisual(ctx, popover)
	body, ok := popover.styleContentWithVisual(ctx, text.New("Body"), visual).(text.Widget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	resolved := text.ResolveStatic(ctx, body)
	if resolved.Text == nil || resolved.Text.FontSize == nil || *resolved.Text.FontSize != visual.Metrics.BodyTextSize {
		t.Fatalf("body text style = %#v", resolved.Text)
	}
	if col := resolved.Text.Color.(style.SolidColor).Color; col != active.Semantic().MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", col)
	}
}

func popoverTestContextWithState(key string) (*core.Context, *popoverState) {
	state := new(popoverState)
	ctx := newContext(nil)
	testSetComponentState(ctx, key, stateSlotPopover, state)
	return ctx, state
}

func layoutPopoverFrame(ctx *core.Context, popover Widget) {
	gtx := testLayoutContext()
	core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	popover.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
}

func layoutPopoverOverBackgroundFrame(ctx *core.Context, router *input.Router, popover Widget, background *widget.Clickable, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(300, 200)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	popover.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

type countingPopoverContent struct {
	child   core.Widget
	layouts int
}

func (c *countingPopoverContent) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	c.layouts++
	return c.child.Layout(ctx, gtx)
}

func TestPopoverUncontrolledMode(t *testing.T) {
	ctx := newContext(nil)
	core.BeginFrame(ctx)

	// Uncontrolled popover should start closed
	popover := New(Spacer(24, 12), text.New("Content")).Key("test")
	popover.Layout(ctx, testLayoutContext())

	state := testComponentState[popoverState](ctx, "test", stateSlotPopover)
	if state != nil {
		t.Fatal("uncontrolled closed popover should not claim state")
	}
}

func TestPopoverUncontrolledWithDefaultOpen(t *testing.T) {
	ctx := newContext(nil)
	core.BeginFrame(ctx)

	// Uncontrolled with DefaultOpen(true) should start open
	popover := New(Spacer(24, 12), text.New("Content")).Key("test").DefaultOpen(true)
	popover.Layout(ctx, testLayoutContext())

	state := testComponentState[popoverState](ctx, "test", stateSlotPopover)
	if state == nil {
		t.Fatal("uncontrolled popover with DefaultOpen(true) should claim state")
	}
	if !state.open {
		t.Fatal("uncontrolled popover with DefaultOpen(true) should be open")
	}
}

func TestPopoverDefaultOpenDoesNotReopenAfterClose(t *testing.T) {
	ctx := newContext(nil)
	value := New(Spacer(24, 12), text.New("Content")).Key("test").DefaultOpen(true)

	layoutPopoverFrame(ctx, value)
	state := testComponentState[popoverState](ctx, "test", stateSlotPopover)
	if state == nil || !state.open {
		t.Fatal("default-open popover did not start open")
	}
	state.requestOpen(value, false)
	state.transition.Set(0, 0, time.Time{})

	layoutPopoverFrame(ctx, value)
	if core.HasTopOverlay(ctx) {
		t.Fatal("default-open popover remained visible after close")
	}
	if testComponentState[popoverState](ctx, "test", stateSlotPopover) == nil {
		t.Fatal("default-open popover lost its consumed closed state")
	}

	layoutPopoverFrame(ctx, value)
	if core.HasTopOverlay(ctx) {
		t.Fatal("default-open popover reopened after its close")
	}
}

func TestPopoverUncontrolledOnOpenChange(t *testing.T) {
	ctx := newContext(nil)
	var called bool
	var calledValue bool

	onOpenChange := func(open bool) {
		called = true
		calledValue = open
	}

	// Start with closed popover
	popover := New(Spacer(24, 12), text.New("Content")).Key("test").
		OnOpenChange(onOpenChange)

	// First frame - closed
	core.BeginFrameWithViewport(ctx, image.Pt(300, 200))
	popover.Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)

	if called {
		t.Fatal("OnOpenChange should not be called on first frame when closed")
	}

	// Simulate dismiss click to request close - this should trigger requestOpen
	called = false
	core.BeginFrameWithViewport(ctx, image.Pt(300, 200))
	state := popoverStateFor(ctx, "test")

	// Manually call requestOpen to simulate user interaction
	state.requestOpen(popover, true)

	if !called {
		t.Fatal("OnOpenChange should be called when requesting open")
	}
	if !calledValue {
		t.Fatalf("OnOpenChange value = %v, want true", calledValue)
	}

	// Verify state is open
	if !state.open {
		t.Fatal("popover should be open after requestOpen(true)")
	}
}

func TestPopoverControlledMode(t *testing.T) {
	ctx := newContext(nil)

	// Controlled mode - explicitly call Open()
	popover := New(Spacer(24, 12), text.New("Content")).Key("test").
		Open(false)

	core.BeginFrame(ctx)
	popover.Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)

	// In controlled mode, internal state should follow Open() parameter
	state := testComponentState[popoverState](ctx, "test", stateSlotPopover)
	if state != nil {
		t.Fatal("controlled closed popover should not claim state")
	}

	// Open in controlled mode
	popover = popover.Open(true)

	core.BeginFrame(ctx)
	popover.Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)

	state = testComponentState[popoverState](ctx, "test", stateSlotPopover)
	if state == nil {
		t.Fatal("controlled open popover should claim state")
	}
	if !state.open {
		t.Fatal("controlled popover should be open when Open(true)")
	}
}

func TestPopoverOnlyCallsOnOpenChangeWhenValueChanges(t *testing.T) {
	ctx := newContext(nil)
	callCount := 0

	onOpenChange := func(open bool) {
		callCount++
	}

	popover := New(Spacer(24, 12), text.New("Content")).Key("test").
		OnOpenChange(onOpenChange)

	// Frame 1 - closed
	core.BeginFrame(ctx)
	popover.Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)

	if callCount != 0 {
		t.Fatalf("OnOpenChange call count = %d, want 0 on first frame", callCount)
	}

	// Request open via state
	core.BeginFrame(ctx)
	state := popoverStateFor(ctx, "test")
	state.requestOpen(popover, true)
	core.EndFrame(ctx)

	if callCount != 1 {
		t.Fatalf("OnOpenChange call count = %d, want 1 after requesting open", callCount)
	}

	// Frame with same value (no change) - just check current state
	core.BeginFrame(ctx)
	popover.Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)

	if callCount != 1 {
		t.Fatalf("OnOpenChange call count = %d, want 1 when value unchanged", callCount)
	}

	// Request same value again (should not trigger)
	core.BeginFrame(ctx)
	state.requestOpen(popover, true)
	core.EndFrame(ctx)

	if callCount != 1 {
		t.Fatalf("OnOpenChange call count = %d, want 1 when requesting same value", callCount)
	}

	// Request close (should trigger)
	core.BeginFrame(ctx)
	state.requestOpen(popover, false)
	core.EndFrame(ctx)

	if callCount != 2 {
		t.Fatalf("OnOpenChange call count = %d, want 2 after requesting close", callCount)
	}
}

func TestUncontrolledPreservesClosedInitialState(t *testing.T) {
	widget := New(nil, nil).Key("popover").Uncontrolled()
	if widget.hasOpen || !widget.hasDefaultOpen || widget.defaultOpen {
		t.Fatalf("Uncontrolled() = hasOpen=%v hasDefault=%v defaultOpen=%v", widget.hasOpen, widget.hasDefaultOpen, widget.defaultOpen)
	}
}
