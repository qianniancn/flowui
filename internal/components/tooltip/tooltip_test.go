package tooltip

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestTooltipHeroUIDefaultTokens(t *testing.T) {
	value := theme.DefaultTheme().Components.Tooltip
	if value.Delay != 1500*time.Millisecond || value.CloseDelay != 500*time.Millisecond {
		t.Fatalf("delays = %v/%v, want 1500ms/500ms", value.Delay, value.CloseDelay)
	}
	if value.Offset != 3 || value.ArrowOffset != 7 || value.Padding != 8 || value.Radius != 12 || value.BorderWidth != 1 || value.MaxWidth != 320 {
		t.Fatalf("layout tokens = %+v", value)
	}
	if value.TextSize != 12 || value.ArrowSize != 12 || value.AnimationDistance != 4 || value.AnimationScale != .90 || value.ExitScale != .95 {
		t.Fatalf("visual tokens = %+v", value)
	}
}

func TestTooltipBorderUsesThemeColor(t *testing.T) {
	themeValue := theme.DefaultTheme()
	style := tooltipStyleFor(&themeValue)
	if style.border != themeValue.Palette.Border {
		t.Fatalf("border = %v, want %v", style.border, themeValue.Palette.Border)
	}
}

func TestTooltipShowAndCloseDelays(t *testing.T) {
	start := time.Unix(1, 0)
	state := new(tooltipState)
	state.updateActive(testContextAt(start), true, false, 1500*time.Millisecond, 500*time.Millisecond)
	if state.open || state.showAt != start.Add(1500*time.Millisecond) {
		t.Fatalf("initial show state = open %v at %v", state.open, state.showAt)
	}

	state.updateActive(testContextAt(start.Add(1499*time.Millisecond)), true, false, 1500*time.Millisecond, 500*time.Millisecond)
	if state.open {
		t.Fatal("tooltip opened before the show delay")
	}
	state.updateActive(testContextAt(start.Add(1500*time.Millisecond)), true, false, 1500*time.Millisecond, 500*time.Millisecond)
	if !state.open {
		t.Fatal("tooltip did not open when the show delay elapsed")
	}

	leave := start.Add(2 * time.Second)
	state.updateActive(testContextAt(leave), false, false, 1500*time.Millisecond, 500*time.Millisecond)
	if !state.open || state.hideAt != leave.Add(500*time.Millisecond) {
		t.Fatalf("close state = open %v at %v", state.open, state.hideAt)
	}
	state.updateActive(testContextAt(leave.Add(500*time.Millisecond)), false, false, 1500*time.Millisecond, 500*time.Millisecond)
	if state.open {
		t.Fatal("tooltip stayed open after the close delay")
	}
}

func TestTooltipLeaveCancelsPendingShow(t *testing.T) {
	start := time.Unix(1, 0)
	state := new(tooltipState)
	state.updateActive(testContextAt(start), true, false, time.Second, time.Second)
	state.updateActive(testContextAt(start.Add(100*time.Millisecond)), false, false, time.Second, time.Second)
	if !state.showAt.IsZero() || state.open {
		t.Fatalf("pending show was not cancelled: open %v at %v", state.open, state.showAt)
	}
}

func TestTooltipReenterCancelsPendingClose(t *testing.T) {
	start := time.Unix(1, 0)
	state := &tooltipState{active: true, open: true}
	state.updateActive(testContextAt(start), false, false, 0, time.Second)
	state.updateActive(testContextAt(start.Add(100*time.Millisecond)), true, false, 0, time.Second)
	if !state.hideAt.IsZero() || !state.open {
		t.Fatalf("re-enter state = open %v hideAt %v", state.open, state.hideAt)
	}
}

func TestTooltipPeerCloseSkipsCloseDelay(t *testing.T) {
	state := &tooltipState{
		active: true,
		open:   true,
		hideAt: time.Unix(2, 0),
	}
	state.closeForPeer()
	if state.open || !state.showAt.IsZero() || !state.hideAt.IsZero() {
		t.Fatalf("peer close did not close immediately: %+v", state)
	}
}

func TestTooltipCoordinatorWarmupAndCooldown(t *testing.T) {
	start := time.Unix(1, 0)
	coordinator := new(tooltipCoordinator)
	coordinator.open("first", 100*time.Millisecond)
	if !coordinator.warmed || coordinator.activeKey != "first" {
		t.Fatalf("open coordinator = %+v", coordinator)
	}

	coordinator.beginCooldown(testContextAt(start), "first", 100*time.Millisecond)
	if coordinator.cooldownAt != start.Add(tooltipCooldown) {
		t.Fatalf("cooldown = %v, want %v", coordinator.cooldownAt, start.Add(tooltipCooldown))
	}
	coordinator.update(testContextAt(start.Add(tooltipCooldown - time.Millisecond)))
	if !coordinator.warmed {
		t.Fatal("coordinator cooled before 500ms")
	}
	coordinator.update(testContextAt(start.Add(tooltipCooldown)))
	if coordinator.warmed || coordinator.activeKey != "" || !coordinator.cooldownAt.IsZero() {
		t.Fatalf("coordinator did not cool down: %+v", coordinator)
	}
}

func TestTooltipCoordinatorCoolsWhenActiveTooltipUnmounts(t *testing.T) {
	start := time.Unix(1, 0)
	coordinator := new(tooltipCoordinator)
	coordinator.open("first", 300*time.Millisecond)
	coordinator.BeginFrame()
	coordinator.seen = map[string]struct{}{"second": {}}
	coordinator.finishFrame(testContextAt(start))
	if coordinator.cooldownAt != start.Add(tooltipCooldown) {
		t.Fatalf("unmounted cooldown = %v, want %v", coordinator.cooldownAt, start.Add(tooltipCooldown))
	}
	coordinator.BeginFrame()
	coordinator.seen = map[string]struct{}{"second": {}}
	coordinator.finishFrame(testContextAt(start.Add(100 * time.Millisecond)))
	if coordinator.cooldownAt != start.Add(tooltipCooldown) {
		t.Fatal("unmounted cooldown was extended by another frame")
	}
}

func TestTooltipCoordinatorUsesLongerCloseDelay(t *testing.T) {
	start := time.Unix(1, 0)
	coordinator := &tooltipCoordinator{warmed: true, activeKey: "help"}
	coordinator.beginCooldown(testContextAt(start), "help", 900*time.Millisecond)
	if coordinator.cooldownAt != start.Add(900*time.Millisecond) {
		t.Fatalf("cooldown = %v, want close delay", coordinator.cooldownAt)
	}
}

func TestTooltipDisabledClosesImmediately(t *testing.T) {
	state := &tooltipState{hovered: true, active: true, open: true}
	state.updateActive(testContextAt(time.Unix(1, 0)), true, true, time.Second, time.Second)
	if state.open || state.active || state.hovered || !state.showAt.IsZero() || !state.hideAt.IsZero() {
		t.Fatalf("disabled state was not cleared: %+v", state)
	}
}

func TestTooltipOffsetUsesArrowDefault(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContextAt(time.Time{})
	if got := Tooltip("plain", nil, nil).popup(1, false).offsetPx(ctx, gtx); got != 3 {
		t.Fatalf("plain offset = %d, want 3", got)
	}
	if got := Tooltip("arrow", nil, nil).Arrow(true).popup(1, false).offsetPx(ctx, gtx); got != 7 {
		t.Fatalf("arrow offset = %d, want 7", got)
	}
	if got := Tooltip("custom", nil, nil).Arrow(true).Offset(11).popup(1, false).offsetPx(ctx, gtx); got != 11 {
		t.Fatalf("custom offset = %d, want 11", got)
	}
}

func TestTooltipPositionAndFlip(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContextAt(time.Time{})
	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(80, 40)
	bounds := image.Pt(300, 200)
	value := Tooltip("help", nil, nil).Delay(0).Offset(3).AvoidOverflow(false)
	popup := value.popup(1, false)
	if got := popup.resolvedPosition(ctx, gtx, trigger, panel, bounds).Position; got != image.Pt(80, 7) {
		t.Fatalf("top position = %v, want (80,7)", got)
	}

	edgeTrigger := image.Rect(20, 2, 60, 22)
	if got := popup.resolvedPosition(ctx, gtx, edgeTrigger, panel, bounds).Placement.PopoverPlacement(); got != overlay.PopoverBottom {
		t.Fatalf("resolved placement = %v, want bottom", got)
	}
}

func TestTooltipArrowTracksTriggerForAlignedAndShiftedPanels(t *testing.T) {
	trigger := image.Rect(100, 50, 140, 70)
	panel := image.Pt(100, 30)
	if got := overlay.ArrowAnchor(trigger, image.Pt(100, 10), panel, overlay.PopoverTopStart, 0, 12); got != 20 {
		t.Fatalf("top start arrow = %v, want 20", got)
	}
	if got := overlay.ArrowAnchor(trigger, image.Pt(40, 10), panel, overlay.PopoverTopEnd, 0, 12); got != 80 {
		t.Fatalf("top end arrow = %v, want 80", got)
	}
	if got := overlay.ArrowAnchor(trigger, image.Pt(0, 10), panel, overlay.PopoverTop, 12, 12); got != 82 {
		t.Fatalf("overflow-shifted arrow = %v, want clamped 82", got)
	}
}

func TestTooltipTransformOriginUsesTriggerAnchor(t *testing.T) {
	trigger := image.Rect(100, 50, 140, 70)
	origin := overlay.PanelTransformOriginAt(trigger, image.Pt(100, 10), image.Pt(100, 30), overlay.PopoverTopStart.Placement())
	if origin != f32.Pt(20, 30) {
		t.Fatalf("transform origin = %v, want (20,30)", origin)
	}
}

func TestTooltipPanelHonorsThemePadding(t *testing.T) {
	themeValue := theme.DefaultTheme()
	themeValue.Components.Tooltip.Padding = 10
	themeValue.Components.Tooltip.Radius = 6
	ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
	gtx := testContextAt(time.Time{})
	dims := Tooltip("help", nil, fixedWidget{size: image.Pt(30, 10)}).popup(1, false).layoutPanel(ctx, gtx)
	if dims.Size != image.Pt(50, 30) {
		t.Fatalf("panel size = %v, want (50,30)", dims.Size)
	}
}

func TestPopupPanelUsesSharedTooltipTheme(t *testing.T) {
	themeValue := theme.DefaultTheme()
	themeValue.Components.Tooltip.Padding = 10
	themeValue.Components.Tooltip.MaxWidth = 72
	ctx := frame.New(nil, &themeValue, locale.LanguageAuto)
	gtx := testContextAt(time.Time{})
	content := new(constraintProbe)
	popup := NewPopup(content)
	panelGtx := gtx
	panelGtx.Constraints = popup.panelConstraints(ctx, gtx, gtx.Constraints.Max)

	dims := popup.layoutPanel(ctx, panelGtx)
	if content.constraints.Max != image.Pt(52, 180) {
		t.Fatalf("popup content constraints = %v, want max (52,180)", content.constraints)
	}
	if dims.Size != image.Pt(72, 32) {
		t.Fatalf("popup panel size = %v, want (72,32)", dims.Size)
	}
}

func TestPopupTransitionEnterExitAndReset(t *testing.T) {
	start := time.Unix(1, 0)
	transition := new(PopupTransition)
	if got := transition.Progress(testContextAt(start), true); got != 0 {
		t.Fatalf("initial enter progress = %v, want 0", got)
	}
	if got := transition.Progress(testContextAt(start.Add(tooltipEnterDuration/2)), true); got != .5 {
		t.Fatalf("half enter progress = %v, want 0.5", got)
	}
	if got := transition.Progress(testContextAt(start.Add(tooltipEnterDuration)), true); got != 1 {
		t.Fatalf("completed enter progress = %v, want 1", got)
	}

	exitStart := start.Add(tooltipEnterDuration + time.Millisecond)
	if got := transition.Progress(testContextAt(exitStart), false); got != 1 || !transition.Exiting() {
		t.Fatalf("initial exit = progress %v exiting %v, want 1/true", got, transition.Exiting())
	}
	if got := transition.Progress(testContextAt(exitStart.Add(tooltipExitDuration/2)), false); got != .5 {
		t.Fatalf("half exit progress = %v, want 0.5", got)
	}
	if got := transition.Progress(testContextAt(exitStart.Add(tooltipExitDuration)), false); got != 0 {
		t.Fatalf("completed exit progress = %v, want 0", got)
	}

	transition.Reset()
	if transition.Value() != 0 || transition.Exiting() {
		t.Fatalf("reset transition = value %v exiting %v", transition.Value(), transition.Exiting())
	}
}

func TestPopupTransitionFollowsPointerAndSnapsAfterExit(t *testing.T) {
	start := time.Unix(2, 0)
	transition := new(PopupTransition)
	if got := transition.Position(testContextAt(start), f32.Pt(10, 20)); got != f32.Pt(10, 20) {
		t.Fatalf("initial popup position = %v", got)
	}
	if got := transition.Position(testContextAt(start), f32.Pt(110, 60)); got != f32.Pt(10, 20) {
		t.Fatalf("popup moved before transition started = %v", got)
	}
	if got := transition.Position(testContextAt(start.Add(tooltipMoveDuration/2)), f32.Pt(110, 60)); got != f32.Pt(97.5, 55) {
		t.Fatalf("half popup position = %v", got)
	}
	if got := transition.Position(testContextAt(start.Add(tooltipMoveDuration)), f32.Pt(110, 60)); got != f32.Pt(110, 60) {
		t.Fatalf("completed popup position = %v", got)
	}

	transition.Progress(testContextAt(start), true)
	transition.Progress(testContextAt(start.Add(tooltipEnterDuration)), true)
	exitStart := start.Add(tooltipEnterDuration + time.Millisecond)
	transition.Progress(testContextAt(exitStart), false)
	transition.Progress(testContextAt(exitStart.Add(tooltipExitDuration)), false)
	if got := transition.Position(testContextAt(exitStart.Add(tooltipExitDuration)), f32.Pt(200, 100)); got != f32.Pt(200, 100) {
		t.Fatalf("popup position after completed exit = %v", got)
	}
}

func TestPopupTransitionRespectsDisabledMotion(t *testing.T) {
	start := time.Unix(3, 0)
	transition := new(PopupTransition)
	motion := theme.MotionTheme{Enabled: false, DurationScale: 1}

	if got := transition.Progress(testContextAt(start), true, motion); got != 1 {
		t.Fatalf("disabled motion progress = %v, want 1", got)
	}
	transition.Position(testContextAt(start), f32.Pt(10, 20), motion)
	if got := transition.Position(testContextAt(start), f32.Pt(110, 60), motion); got != f32.Pt(110, 60) {
		t.Fatalf("disabled motion position = %v, want (110,60)", got)
	}
}

func TestPopupCanDisableTransformMotion(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContextAt(time.Unix(3, 0))
	popup := NewPopup(nil).TransformMotion(false).Progress(.5)
	if popup.transformMotionEnabled() {
		t.Fatal("popup transform motion remained enabled")
	}
	affine := popup.panelAffine(ctx, image.Rect(20, 20, 21, 21), image.Pt(30, 30), image.Pt(80, 40), overlay.PopoverRight)
	if got := affine.Transform(f32.Pt(12, 24)); got != f32.Pt(12, 24) {
		t.Fatalf("disabled popup transform = %v", got)
	}
	if got := popup.slideOffset(ctx, gtx, overlay.PopoverRight); got != (image.Point{}) {
		t.Fatalf("disabled popup slide offset = %v", got)
	}
}

func TestPopupWithoutAnchorOrContentHasNoLayoutSize(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx := testContextAt(time.Time{})
	if got := NewPopup(fixedWidget{size: image.Pt(20, 10)}).Layout(ctx, gtx, image.Rectangle{}).Size; got != (image.Point{}) {
		t.Fatalf("empty anchor size = %v, want zero", got)
	}
	if got := NewPopup(nil).Layout(ctx, gtx, image.Rect(10, 10, 20, 20)).Size; got != (image.Point{}) {
		t.Fatalf("nil content size = %v, want zero", got)
	}
}

func TestTooltipHoverAreaDoesNotBlockChildClick(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	child := new(clickProbe)
	value := Tooltip("help", child, text.New("Helpful text")).Delay(0)
	layoutTooltipFrame(ctx, router, value, time.Unix(1, 0))

	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 10)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(10, 10)},
	)
	layoutTooltipFrame(ctx, router, value, time.Unix(1, int64(time.Millisecond)))
	if child.clicks != 1 {
		t.Fatalf("child clicks = %d, want 1", child.clicks)
	}
}

func TestTooltipHoverEventStartsShowDelay(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	value := Tooltip("help", fixedWidget{size: image.Pt(40, 24)}, text.New("Helpful text"))
	start := time.Unix(1, 0)
	layoutTooltipFrame(ctx, router, value, start)

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(10, 10)})
	layoutTooltipFrame(ctx, router, value, start.Add(time.Millisecond))
	state, _ := frame.PeekState[tooltipState](ctx, "help", stateSlotTooltip)
	if state == nil || !state.hovered || state.showAt != start.Add(time.Millisecond+1500*time.Millisecond) {
		t.Fatalf("hover state = %+v", state)
	}
}

func TestTooltipFocusTriggerStartsShowDelay(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	value := Tooltip("help", fixedWidget{size: image.Pt(40, 24)}, text.New("Helpful text")).Trigger(TooltipFocus)
	start := time.Unix(1, 0)
	layoutTooltipFrame(ctx, router, value, start)
	state, _ := frame.PeekState[tooltipState](ctx, "help", stateSlotTooltip)
	if state == nil {
		t.Fatal("missing tooltip state")
	}

	router.Source().Execute(key.FocusCmd{Tag: state})
	layoutTooltipFrame(ctx, router, value, start.Add(time.Millisecond))
	if !state.focused || state.showAt != start.Add(time.Millisecond+1500*time.Millisecond) {
		t.Fatalf("focus state = %+v", state)
	}
}

func TestRapidTooltipSwitchClosesPeerAndUsesWarmup(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	start := time.Unix(1, 0)
	layoutTooltipPairFrame(ctx, router, start, false)

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(10, 10)})
	layoutTooltipPairFrame(ctx, router, start.Add(time.Millisecond), false)
	first, _ := frame.PeekState[tooltipState](ctx, "first", stateSlotTooltip)
	second, _ := frame.PeekState[tooltipState](ctx, "second", stateSlotTooltip)
	if first == nil || second == nil || !first.open || second.open {
		t.Fatalf("first hover states = first %+v second %+v", first, second)
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(70, 10)})
	layoutTooltipPairFrame(ctx, router, start.Add(2*time.Millisecond), false)
	if first.open || !first.hideAt.IsZero() {
		t.Fatalf("first tooltip retained close delay: %+v", first)
	}
	if !second.open || !second.showAt.IsZero() {
		t.Fatalf("second tooltip did not use warmup: %+v", second)
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(150, 50)})
	layoutTooltipPairFrame(ctx, router, start.Add(3*time.Millisecond), false)
	layoutTooltipPairFrame(ctx, router, start.Add(503*time.Millisecond), false)
	layoutTooltipPairFrame(ctx, router, start.Add(604*time.Millisecond), false)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(70, 10)})
	layoutTooltipPairFrame(ctx, router, start.Add(700*time.Millisecond), false)
	if second.open || second.showAt != start.Add(2200*time.Millisecond) {
		t.Fatalf("cold tooltip did not restore show delay: %+v", second)
	}
}

func TestRapidTooltipSwitchIsIndependentOfLayoutOrder(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	start := time.Unix(1, 0)
	layoutTooltipPairFrame(ctx, router, start, true)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(10, 10)})
	layoutTooltipPairFrame(ctx, router, start.Add(time.Millisecond), true)
	first, _ := frame.PeekState[tooltipState](ctx, "first", stateSlotTooltip)
	second, _ := frame.PeekState[tooltipState](ctx, "second", stateSlotTooltip)

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(70, 10)})
	layoutTooltipPairFrame(ctx, router, start.Add(2*time.Millisecond), true)
	if first.open || !second.open {
		t.Fatalf("reverse layout switch = first %+v second %+v", first, second)
	}
}

func testContextAt(now time.Time) layout.Context {
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      new(input.Router).Source(),
		Ops:         new(op.Ops),
		Now:         now,
	}
}

func layoutTooltipFrame(ctx *frame.Context, router *input.Router, value TooltipWidget, now time.Time) {
	ops := new(op.Ops)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(300, 200)),
		Source:      router.Source(),
		Ops:         ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	value.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(ops)
}

func layoutTooltipPairFrame(ctx *frame.Context, router *input.Router, now time.Time, reverse bool) {
	ops := new(op.Ops)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(160, 60)),
		Source:      router.Source(),
		Ops:         ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	first := func() {
		Tooltip("first", fixedWidget{size: image.Pt(40, 24)}, text.New("First")).Delay(0).Layout(ctx, gtx)
	}
	second := func() {
		offset := op.Offset(image.Pt(60, 0)).Push(gtx.Ops)
		Tooltip("second", fixedWidget{size: image.Pt(40, 24)}, text.New("Second")).Layout(ctx, gtx)
		offset.Pop()
	}
	if reverse {
		second()
		first()
	} else {
		first()
		second()
	}
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(ops)
}

type fixedWidget struct{ size image.Point }

func (w fixedWidget) Layout(*frame.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: w.size}
}

type constraintProbe struct {
	constraints layout.Constraints
}

func (p *constraintProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(200, 12))}
}

type clickProbe struct {
	clickable widget.Clickable
	clicks    int
}

func (p *clickProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	for p.clickable.Clicked(gtx) {
		p.clicks++
	}
	return p.clickable.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(40, 24)}
	})
}
