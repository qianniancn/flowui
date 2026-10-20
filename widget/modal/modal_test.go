package modal

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
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotClickable = "clickable"

const modalEnterDuration = 250 * time.Millisecond

func (s *modalState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	duration := modalEnterDuration
	if len(motions) > 0 {
		motion := motions[0]
		if motion.ControlDuration > 0 {
			duration = motion.ControlDuration
		} else if motion.DefaultDuration > 0 {
			duration = motion.DefaultDuration
		}
	}
	return s.progressFor(gtx, open, duration, motions...)
}

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func newContextWithTheme(_ any, value *theme.Theme) *core.Context {
	if value != nil && value.VisualProvider == nil {
		clone := *value
		clone.VisualProvider = testtheme.BehaviorProvider()
		value = &clone
	}
	return core.New(nil, value, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.NeutralTheme()
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

func TestModalClosedReturnsZeroAndDoesNotClaimState(t *testing.T) {
	ctx := newContext(nil)
	dims := New("Settings", text.New("Body")).Key("settings").Layout(ctx, testLayoutContext())

	if dims.Size != (image.Point{}) {
		t.Fatalf("closed modal size = %v, want zero", dims.Size)
	}
	if core.StateLen(ctx) != 0 {
		t.Fatalf("closed modal claimed state, len = %d", core.StateLen(ctx))
	}
}

func TestModalOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	New("Settings", text.New("Body")).Key("settings").Open(true).Layout(ctx, testLayoutContext())

	if testComponentState[modalState](ctx, "settings", stateSlotModal) == nil {
		t.Fatal("open modal did not keep state")
	}
}

func TestModalRetainsCloseButtonAcrossFrames(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true)

	layoutModalFrameAt(ctx, router, modal, time.Unix(1, 0))
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	if state == nil || state.closeButton == nil {
		t.Fatal("open modal did not initialize its close button")
	}
	first := state.closeButton

	layoutModalFrameAt(ctx, router, modal, time.Unix(1, int64(time.Millisecond)))
	if state.closeButton != first {
		t.Fatal("modal recreated its close button between frames")
	}
}

func TestModalVisualResolversUseEffectiveOpenState(t *testing.T) {
	measure := func(open bool) image.Point {
		ctx, state := modalTestContextWithState("settings")
		gtx := testLayoutContext()
		gtx.Constraints = layout.Exact(image.Pt(260, 180))
		body := &constraintProbeWidget{size: image.Pt(80, 24)}
		modal := New("Settings", body).Key("settings").CloseButton(false)
		visual := modalTestVisual(ctx, modal)
		visual.Defaults = visual.Defaults.On(style.Open, style.Style{}.Padding(20))
		state.open = open
		modal.layoutDialogFrameWithVisual(ctx, gtx, state, visual)
		return image.Pt(body.constraints.Max.X, body.constraints.Max.Y)
	}

	closedSize := measure(false)
	openSize := measure(true)
	if openSize == closedSize {
		t.Fatalf("effective open state did not reach visual resolver: closed=%v open=%v", closedSize, openSize)
	}
}

func TestModalOpenFocusesTrap(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)

	layoutModalFrame(ctx, router, New("Settings", text.New("Body")).Key("settings").Open(true))
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	if state == nil {
		t.Fatal("open modal did not keep state")
	}
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("modal did not move focus into trap")
	}
	if router.Source().Focused(&state.close) {
		t.Fatal("modal focused close button before keyboard navigation")
	}
}

func TestModalOpenWithoutCloseButtonFocusesTrap(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)

	layoutModalFrame(ctx, router, New("Settings", text.New("Body")).Key("settings").Open(true).CloseButton(false))
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	if state == nil {
		t.Fatal("open modal did not keep state")
	}
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("modal without close button did not move focus into trap")
	}
}

func TestModalNaturallyDisabledDefersInitialFocusUntilEnabled(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true)

	layoutModalFrameEnabled(ctx, router, modal, false)
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	if router.Source().Focused(&state.focusTarget) {
		t.Fatal("naturally disabled modal took focus")
	}

	layoutModalFrameEnabled(ctx, router, modal, true)
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("enabled modal did not consume its deferred focus request")
	}
}

func TestStackedModalRestoresUnderlyingFocusScopeAfterExit(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	start := time.Unix(1, 0)
	outer := New("Outer", text.New("Outer body")).Key("outer").Open(true)
	innerOpen := true
	inner := func() Widget {
		return New("Inner", text.New("Inner body")).Key("inner").Open(innerOpen)
	}

	layoutModalPairFrameAt(ctx, router, outer, inner(), start)
	layoutModalPairFrameAt(ctx, router, outer, inner(), start.Add(modalEnterDuration))
	outerState := testComponentState[modalState](ctx, "outer", stateSlotModal)
	innerState := testComponentState[modalState](ctx, "inner", stateSlotModal)
	if !router.Source().Focused(&innerState.focusTarget) {
		t.Fatal("top stacked modal did not own focus")
	}

	innerOpen = false
	closingAt := start.Add(modalEnterDuration + time.Millisecond)
	layoutModalPairFrameAt(ctx, router, outer, inner(), closingAt)
	if !router.Source().Focused(&innerState.focusTarget) {
		t.Fatal("exiting modal released focus before its animation completed")
	}

	layoutModalPairFrameAt(ctx, router, outer, inner(), closingAt.Add(modalEnterDuration))
	if !router.Source().Focused(&outerState.focusTarget) {
		t.Fatal("underlying modal did not regain focus after the top modal exited")
	}
}

func TestModalExitMovesInternalFocusToTrap(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	start := time.Unix(1, 0)
	open := true
	build := func() Widget {
		return New("Settings", text.New("Body")).Key("settings").Open(open)
	}

	layoutModalFrameAt(ctx, router, build(), start)
	layoutModalFrameAt(ctx, router, build(), start.Add(modalEnterDuration))
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	router.Source().Execute(key.FocusCmd{Tag: &state.close})
	if !router.Source().Focused(&state.close) {
		t.Fatal("modal close button did not receive setup focus")
	}

	open = false
	layoutModalFrameAt(ctx, router, build(), start.Add(modalEnterDuration+time.Millisecond))
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("modal exit did not move disabled content focus back into the trap")
	}
}

func TestModalTabFromTrapReachesFooterButton(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		Footer(button.New("Action").Key("footer-action"))

	layoutModalFrame(ctx, router, modal)
	footer := testComponentState[widget.Clickable](ctx, "footer-action", stateSlotClickable)
	if footer == nil {
		t.Fatal("footer button state was not kept")
	}

	router.MoveFocus(key.FocusForward)

	if !router.Source().Focused(footer) {
		t.Fatal("tab from modal trap did not reach footer button")
	}
	if router.Source().Focused(&testComponentState[modalState](ctx, "settings", stateSlotModal).dialog) {
		t.Fatal("modal dialog blocker received keyboard focus")
	}
}

func TestModalFocusWrapsAtBoundaries(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true)

	layoutModalFrame(ctx, router, modal)
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	router.MoveFocus(key.FocusBackward)
	layoutModalFrame(ctx, router, modal)

	if !router.Source().Focused(&state.close) {
		t.Fatal("shift-tab from modal trap did not wrap to close button")
	}

	router.MoveFocus(key.FocusForward)
	layoutModalFrame(ctx, router, modal)

	if !router.Source().Focused(&state.focusTarget) {
		t.Fatalf(
			"tab past modal close button did not wrap to trap: focusTarget=%v focusStart=%v focusEnd=%v close=%v",
			router.Source().Focused(&state.focusTarget),
			router.Source().Focused(&state.focusStart),
			router.Source().Focused(&state.focusEnd),
			router.Source().Focused(&state.close),
		)
	}
}

func TestModalNestedOverlayFocusCannotEscapeToBackground(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	background := new(widget.Clickable)
	nested := new(nestedFocusableOverlay)
	modal := New("Settings", nested).Key("settings").Open(true).CloseButton(false)

	layoutModalWithBackgroundFrame(ctx, router, modal, background)
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("modal with an initially open nested overlay did not take focus")
	}
	router.Source().Execute(key.FocusCmd{Tag: &nested.button})
	if !router.Source().Focused(&nested.button) {
		t.Fatal("nested overlay button did not receive focus")
	}

	router.MoveFocus(key.FocusForward)
	layoutModalWithBackgroundFrame(ctx, router, modal, background)
	if router.Source().Focused(background) {
		t.Fatal("Tab from nested overlay escaped to the background")
	}
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("Tab from nested overlay did not wrap to the modal focus trap")
	}
}

func TestModalClosedKeepsVisibleStateForExitAnimation(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	start := time.Unix(1, 0)
	state.transition.Set(1, 1, start)

	gtx := testLayoutContext()
	gtx.Now = start
	New("Settings", text.New("Body")).Key("settings").Layout(ctx, gtx)

	if testComponentState[modalState](ctx, "settings", stateSlotModal) == nil {
		t.Fatal("closing modal state was removed before exit animation")
	}
	if state.transition.Target() != 0 || state.transition.Current() != 1 {
		t.Fatalf("closing modal animation state = target %v value %v, want closing from visible", state.transition.Target(), state.transition.Current())
	}
}

func TestModalClosedRemovesStateWhenExitAnimationFinishes(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	start := time.Unix(1, 0)
	state.transition.Set(1, 0, start)

	gtx := testLayoutContext()
	gtx.Now = start.Add(modalEnterDuration)
	New("Settings", text.New("Body")).Key("settings").Layout(ctx, gtx)

	if testComponentState[modalState](ctx, "settings", stateSlotModal) != nil {
		t.Fatal("closed modal state was not removed after exit animation")
	}
}

func TestModalBackdropBlocksBackgroundClicksWhenNotDismissable(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	backgroundClicked := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Dismissable(false)

	layoutModalOverButtonFrame(ctx, router, modal, func() {
		backgroundClicked = true
	})
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(20, 20),
	})
	layoutModalOverButtonFrame(ctx, router, modal, func() {
		backgroundClicked = true
	})
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 20),
	})
	layoutModalOverButtonFrame(ctx, router, modal, func() {
		backgroundClicked = true
	})

	if backgroundClicked {
		t.Fatal("non-dismissable modal allowed backdrop click to reach background")
	}
}

func TestModalDialogBlocksBackgroundClicks(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var background widget.Clickable
	backgroundClicked := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true)
	pos := f32.Pt(150, 100)

	layoutModalOverClickableFrame(ctx, router, modal, &background, &backgroundClicked)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  pos,
	})
	layoutModalOverClickableFrame(ctx, router, modal, &background, &backgroundClicked)
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  pos,
	})
	layoutModalOverClickableFrame(ctx, router, modal, &background, &backgroundClicked)

	if backgroundClicked {
		t.Fatal("modal dialog allowed click to reach background")
	}
}

func TestModalExitDialogStillBlocksBackgroundClicks(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.transition.Set(1, 1, time.Time{})
	router := new(input.Router)
	background := new(widget.Clickable)
	backgroundClicked := false
	modal := New("Settings", text.New("Body")).Key("settings")
	position := f32.Pt(150, 100)

	layoutModalOverClickableFrame(ctx, router, modal, background, &backgroundClicked)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutModalOverClickableFrame(ctx, router, modal, background, &backgroundClicked)
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutModalOverClickableFrame(ctx, router, modal, background, &backgroundClicked)

	if backgroundClicked {
		t.Fatal("exiting modal dialog allowed a click to reach the background")
	}
}

func TestModalCloseButtonRequestsClose(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	closed := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	// Open first so the paint pass does not drain the programmatic click before
	// policy input runs; the following frame owns overlay policy input.
	layoutModalFrame(ctx, new(input.Router), modal)
	state.close.Click()
	layoutModalFrame(ctx, new(input.Router), modal)

	if !closed {
		t.Fatal("close button did not request close")
	}
}

func TestModalCloseButtonPointerFocusIsNotKeyboardVisible(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Size(Full)
	start := time.Unix(1, 0)
	layoutModalFrameAt(ctx, router, modal, start)
	layoutModalFrameAt(ctx, router, modal, start.Add(modalEnterDuration))
	state := testComponentState[modalState](ctx, "settings", stateSlotModal)
	position := f32.Pt(272, 28)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutModalFrameAt(ctx, router, modal, start.Add(modalEnterDuration+time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutModalFrameAt(ctx, router, modal, start.Add(modalEnterDuration+2*time.Millisecond))
	layoutModalFrameAt(ctx, router, modal, start.Add(modalEnterDuration+3*time.Millisecond))

	if !router.Source().Focused(&state.close) {
		t.Fatal("pointer click did not focus the close button")
	}
	if core.FocusVisible(ctx, &state.close, true) {
		t.Fatal("close button click exposed keyboard-visible focus")
	}
}

func TestModalDismissAreaRequestsClose(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	closed := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutModalFrame(ctx, new(input.Router), modal)
	state.dismiss[0].Click()
	layoutModalFrame(ctx, new(input.Router), modal)

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestModalDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	closed := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutModalFrame(ctx, new(input.Router), modal)
	state.dismiss[0].Click()
	layoutModalFrame(ctx, new(input.Router), modal)

	if closed {
		t.Fatal("non-dismissable modal closed from dismiss area")
	}
}

func TestModalEscapeRequestsClose(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	closed := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		CloseButton(false)

	layoutModalFrame(ctx, router, modal)
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutModalFrame(ctx, router, modal)

	if !closed {
		t.Fatal("Escape did not request close")
	}
}

func TestModalKeyboardDismissDisabledIgnoresEscape(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	closed := false
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).
		KeyboardDismissDisabled(true).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		CloseButton(false)

	layoutModalFrame(ctx, router, modal)
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutModalFrame(ctx, router, modal)

	if closed {
		t.Fatal("keyboard-dismiss-disabled modal closed from Escape")
	}
}

func TestModalDialogConstraintsUsesTheme(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Modal.MediumWidth = 360
	theme.Components.Modal.Margin = 10
	ctx := newContextWithTheme(nil, &theme)
	gtx := testLayoutContext()

	modal := New("Settings", text.New("Body")).Key("settings").Open(true)
	constraints := modal.dialogConstraintsWithVisual(ctx, gtx, image.Pt(500, 300), modalTestVisual(ctx, modal))

	if constraints.Min.X != 360 || constraints.Max.X != 360 {
		t.Fatalf("modal width constraints = %+v, want width 360", constraints)
	}
	if constraints.Max.Y != 280 {
		t.Fatalf("modal max height = %d, want 280", constraints.Max.Y)
	}
}

func TestScrollOutsideKeepsScrollingWithoutScrollbarGutter(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(260, 160)}
	body := &constraintProbeWidget{size: image.Pt(120, 500)}
	modal := New("Settings", body).Key("settings").Open(true).
		Scroll(ScrollOutside)

	dims := modal.layoutDialogFrameWithVisual(ctx, gtx, state, modalTestVisual(ctx, modal))

	if dims.Size.Y > gtx.Constraints.Max.Y {
		t.Fatalf("outside scroll dialog height = %d, want <= %d", dims.Size.Y, gtx.Constraints.Max.Y)
	}
	if state.outsideList.Position.Length <= dims.Size.Y {
		t.Fatalf("outside scroll content length = %d, want greater than viewport %d", state.outsideList.Position.Length, dims.Size.Y)
	}
	wantWidth := gtx.Constraints.Max.X - gtx.Dp(core.ActiveTheme(ctx).Components.Modal.Padding)*2
	if body.constraints.Max.X != wantWidth {
		t.Fatalf("outside scroll body width = %d, want %d without scrollbar gutter", body.constraints.Max.X, wantWidth)
	}
}

func TestScrollOutsideKeepsFullSurfaceHeight(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(260, 160))
	modal := New("Settings", fixedModalWidget{size: image.Pt(120, 20)}).Key("settings").Open(true).
		Size(Full).
		Scroll(ScrollOutside)

	dims := modal.layoutDialogFrameWithVisual(ctx, gtx, state, modalTestVisual(ctx, modal))

	if dims.Size != image.Pt(260, 160) {
		t.Fatalf("full outside scroll dimensions = %v, want (260,160)", dims.Size)
	}
	if state.outsideList.Position.Length <= 0 {
		t.Fatalf("full outside scroll content length = %d, want positive", state.outsideList.Position.Length)
	}
}

func TestScrollOutsideKeepsCoverSurfaceHeight(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(260, 160))
	modal := New("Settings", fixedModalWidget{size: image.Pt(120, 20)}).Key("settings").Open(true).
		Size(Cover).
		Scroll(ScrollOutside)

	dims := modal.layoutDialogFrameWithVisual(ctx, gtx, state, modalTestVisual(ctx, modal))

	if dims.Size != image.Pt(260, 160) {
		t.Fatalf("cover outside scroll dimensions = %v, want (260,160)", dims.Size)
	}
	if state.outsideList.Position.Length <= 0 {
		t.Fatalf("cover outside scroll content length = %d, want positive", state.outsideList.Position.Length)
	}
}

func TestLargeSizesRelaxFooterSectionMinimumConstraints(t *testing.T) {
	for name, size := range map[string]Size{
		"cover": Cover,
		"full":  Full,
	} {
		t.Run(name, func(t *testing.T) {
			ctx, state := modalTestContextWithState("settings")
			gtx := testLayoutContext()
			gtx.Constraints = layout.Exact(image.Pt(260, 160))
			footer := &constraintProbeWidget{size: image.Pt(80, 32)}
			modal := New("Settings", fixedModalWidget{size: image.Pt(120, 20)}).Key("settings").Open(true).
				Size(size).
				Footer(footer)

			dims := modal.layoutDialogFrameWithVisual(ctx, gtx, state, modalTestVisual(ctx, modal))

			if dims.Size != image.Pt(260, 160) {
				t.Fatalf("%s dimensions = %v, want (260,160)", name, dims.Size)
			}
			if !footer.laidOut {
				t.Fatal("footer was not laid out")
			}
			if footer.constraints.Min != (image.Point{}) {
				t.Fatalf("footer minimum constraints = %v, want zero", footer.constraints.Min)
			}
			if footer.constraints.Max.X <= 0 || footer.constraints.Max.Y <= 0 {
				t.Fatalf("footer maximum constraints = %v, want positive", footer.constraints.Max)
			}
		})
	}
}

func TestLargeSizesAnchorFooterToBottom(t *testing.T) {
	for name, size := range map[string]Size{
		"cover": Cover,
		"full":  Full,
	} {
		t.Run(name, func(t *testing.T) {
			pos := New("Settings", text.New("Body")).Key("settings").Open(true).
				Size(size).
				dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

			if pos != image.Pt(140, 88) {
				t.Fatalf("%s footer position = %v, want (140,88)", name, pos)
			}
		})
	}
}

func TestMediumKeepsFooterAfterContent(t *testing.T) {
	pos := New("Settings", text.New("Body")).Key("settings").Open(true).
		dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

	if pos != image.Pt(140, 48) {
		t.Fatalf("medium footer position = %v, want (140,48)", pos)
	}
}

func TestPlacement(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()

	topModal := New("Top", text.New("Body")).Key("top").Open(true).Placement(Top)
	top := topModal.dialogPositionWithVisual(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80), modalTestVisual(ctx, topModal))
	bottomModal := New("Bottom", text.New("Body")).Key("bottom").Open(true).Placement(Bottom)
	bottom := bottomModal.dialogPositionWithVisual(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80), modalTestVisual(ctx, bottomModal))

	if top.Y >= bottom.Y {
		t.Fatalf("top placement y = %d, bottom y = %d, want top before bottom", top.Y, bottom.Y)
	}
}

func TestModalProgressAnimation(t *testing.T) {
	state := new(modalState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.progress(gtx, true); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	gtx.Now = start.Add(modalEnterDuration / 2)
	mid := state.progress(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("progress midpoint = %v, want between 0 and 1", mid)
	}
	gtx.Now = start.Add(modalEnterDuration)
	if got := state.progress(gtx, true); got != 1 {
		t.Fatalf("progress end = %v, want 1", got)
	}
}

func TestModalProgressUsesProviderMotion(t *testing.T) {
	state := new(modalState)
	start := time.Unix(1, 0)
	motion := theme.MotionTheme{
		Enabled:         true,
		DurationScale:   1,
		ControlDuration: 100 * time.Millisecond,
		Ease:            theme.EaseLinear,
	}
	gtx := testLayoutContext()
	gtx.Now = start
	if got := state.progress(gtx, true, motion); got != 0 {
		t.Fatalf("initial provider progress = %v, want 0", got)
	}
	gtx.Now = start.Add(50 * time.Millisecond)
	if got := state.progress(gtx, true, motion); got < .49 || got > .51 {
		t.Fatalf("provider progress at half duration = %v, want about .5", got)
	}
	gtx.Now = start.Add(100 * time.Millisecond)
	if got := state.progress(gtx, true, motion); got != 1 {
		t.Fatalf("provider progress at duration = %v, want 1", got)
	}
}

func TestModalProgressRespectsDisabledProviderMotion(t *testing.T) {
	state := new(modalState)
	start := time.Unix(1, 0)
	motion := theme.MotionTheme{
		Enabled:         false,
		DurationScale:   1,
		ControlDuration: 100 * time.Millisecond,
		Ease:            theme.EaseLinear,
	}
	gtx := testLayoutContext()
	gtx.Now = start
	if got := state.progress(gtx, true, motion); got != 1 {
		t.Fatalf("disabled-motion progress = %v, want 1", got)
	}
}

func TestModalDialogScaleUsesThemeStartScale(t *testing.T) {
	if got := modalDialogScale(0.8, 0); got != 0.8 {
		t.Fatalf("dialog scale at start = %v, want 0.8", got)
	}
	if got := modalDialogScale(0.8, 1); got != 1 {
		t.Fatalf("dialog scale at end = %v, want 1", got)
	}
	got := modalDialogScale(0.8, 0.5)
	if got < 0.899 || got > 0.901 {
		t.Fatalf("dialog scale midpoint = %v, want about 0.9", got)
	}
	if got := modalDialogScale(0, 0); got != 1 {
		t.Fatalf("dialog fallback scale = %v, want neutral 1", got)
	}
}

func TestModalResolvedAnimationDefaultsToFadeScale(t *testing.T) {
	if got := New("Settings", text.New("Body")).Key("settings").Open(true).resolvedAnimation(); got != AnimationFadeScale {
		t.Fatalf("default animation = %v, want fade scale", got)
	}
}

func TestModalFadeAnimationDoesNotTransformDialog(t *testing.T) {
	ctx := newContext(nil)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationFade)
	motion := modal.dialogMotionWithVisual(ctx, testLayoutContext(), image.Rect(10, 20, 110, 120), 0.5, true, modalTestVisual(ctx, modal))

	sx, hx, ox, hy, sy, oy := motion.transform.Elems()
	if sx != 1 || hx != 0 || ox != 0 || hy != 0 || sy != 1 || oy != 0 {
		t.Fatalf("fade transform = %v %v %v %v %v %v, want identity", sx, hx, ox, hy, sy, oy)
	}
	if motion.opacity != 0.5 {
		t.Fatalf("fade opacity = %v, want 0.5", motion.opacity)
	}
}

func TestModalSlideDownAnimationOffsetsFromTop(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	progress := float32(0.25)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationSlideDown)
	motion := modal.dialogMotionWithVisual(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true, modalTestVisual(ctx, modal))

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := -float32(gtx.Dp(core.ActiveTheme(ctx).Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide down offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalMotionBoundsUsesAnimatedRect(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationSlideDown)
	motion := modal.dialogMotionWithVisual(ctx, gtx, rect, 0, true, modalTestVisual(ctx, modal))

	got := core.AffineRectBounds(rect, motion.transform)
	want := rect.Add(image.Pt(0, -gtx.Dp(core.ActiveTheme(ctx).Components.Modal.AnimationDistance)))
	if got != want {
		t.Fatalf("motion bounds = %v, want %v", got, want)
	}
}

func TestModalSlideUpAnimationOffsetsFromBottom(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	progress := float32(0.25)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationSlideUp)
	motion := modal.dialogMotionWithVisual(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true, modalTestVisual(ctx, modal))

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := float32(gtx.Dp(core.ActiveTheme(ctx).Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide up offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalBounceScaleOvershootsOnlyWhenOpening(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)

	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationBounceScale)
	openMotion := modal.dialogMotionWithVisual(ctx, gtx, rect, 0.65, true, modalTestVisual(ctx, modal))
	openScale, _, _, _, _, _ := openMotion.transform.Elems()
	if openScale <= 1 {
		t.Fatalf("opening bounce scale = %v, want overshoot", openScale)
	}

	closeMotion := modal.dialogMotionWithVisual(ctx, gtx, rect, 0.65, false, modalTestVisual(ctx, modal))
	closeScale, _, _, _, _, _ := closeMotion.transform.Elems()
	if closeScale <= 0 || closeScale >= 1 {
		t.Fatalf("closing bounce scale = %v, want clean scale below 1", closeScale)
	}
}

func TestModalBounceScaleClampsToFinalScale(t *testing.T) {
	if got := modalDialogBounceScale(0.95, 1.035, 0.992, 1); got != 1 {
		t.Fatalf("final bounce scale = %v, want 1", got)
	}
}

func TestModalZoomOutAnimationStartsLarge(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationZoomOut)
	motion := modal.dialogMotionWithVisual(ctx, gtx, image.Rect(10, 20, 110, 120), 0, true, modalTestVisual(ctx, modal))

	scale, _, _, _, _, _ := motion.transform.Elems()
	if scale <= 1 {
		t.Fatalf("zoom out start scale = %v, want greater than 1", scale)
	}
	if got := modalDialogZoomOutScale(1.04, 1); got != 1 {
		t.Fatalf("final zoom out scale = %v, want 1", got)
	}
}

func TestModalPopAnimationOvershoots(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	modal := New("Settings", text.New("Body")).Key("settings").Open(true).Animation(AnimationPop)
	motion := modal.dialogMotionWithVisual(ctx, gtx, image.Rect(10, 20, 110, 120), 0.58, true, modalTestVisual(ctx, modal))

	scale, _, _, _, _, _ := motion.transform.Elems()
	if scale <= 1 {
		t.Fatalf("pop scale = %v, want overshoot", scale)
	}
	if got := modalDialogPopScale(0.92, 1.025, 0.997, 1); got != 1 {
		t.Fatalf("final pop scale = %v, want 1", got)
	}
}

func TestModalBackdropStyle(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Components.Modal.BlurBackdrop.A = 200
	provider := testtheme.BehaviorProvider()

	blur := provider.Modal(&activeTheme, theme.ModalRequest{Backdrop: theme.ModalBackdropBlur})
	blurred := style.CascadePart(style.State{}, style.PartBackdrop, blur.Defaults)
	if blurred.Paint == nil || blurred.Paint.Background == nil {
		t.Fatalf("blur backdrop declaration is incomplete: %#v", blurred.Paint)
	}
	if got := blurred.Paint.Background.(style.SolidColor).Color.A; got != 200 {
		t.Fatalf("blur backdrop alpha = %d, want 200", got)
	}
	transparent := provider.Modal(&activeTheme, theme.ModalRequest{Backdrop: theme.ModalBackdropTransparent})
	transparentBackdrop := style.CascadePart(style.State{}, style.PartBackdrop, transparent.Defaults)
	if transparentBackdrop.Paint == nil || transparentBackdrop.Paint.Background == nil {
		t.Fatalf("transparent backdrop declaration is incomplete: %#v", transparentBackdrop.Paint)
	}
	if got := transparentBackdrop.Paint.Background.(style.SolidColor).Color.A; got != 0 {
		t.Fatalf("transparent backdrop alpha = %d, want 0", got)
	}
}

func TestModalBackdropPartOverridesColorAndOpacity(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	want := style.RGB(0x123456)
	custom := style.Style{}.
		Part(style.PartBackdrop, style.Style{}.Background(want).Opacity(.35))
	visual := testtheme.BehaviorProvider().Modal(&activeTheme, theme.ModalRequest{})

	resolved := computed.ResolvePartStatic(
		ctx, style.PartBackdrop, style.State{Open: true},
		visual.Defaults,
		style.Style{}, style.Style{}, custom,
	)
	if resolved.Paint == nil || resolved.Paint.Background != want || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != .35 {
		t.Fatalf("backdrop part = %#v", resolved.Paint)
	}
}

func TestModalBodyTextStyle(t *testing.T) {
	ctx := newContext(nil)
	modal := New("Settings", text.New("Body")).Key("settings").Open(true)
	body, ok := modal.styleBodyWithVisual(ctx, text.New("Body"), modalTestVisual(ctx, modal)).(text.Widget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	resolved := text.ResolveStatic(ctx, body)
	if resolved.Text == nil || resolved.Text.FontSize == nil || *resolved.Text.FontSize != core.ActiveTheme(ctx).Components.Modal.BodyTextSize {
		t.Fatalf("body text style = %#v", resolved.Text)
	}
	if col := resolved.Text.Color.(style.SolidColor).Color; col != core.ActiveTheme(ctx).Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", col)
	}
}

func modalTestContextWithState(key string) (*core.Context, *modalState) {
	state := new(modalState)
	ctx := newContext(nil)
	testSetComponentState(ctx, key, stateSlotModal, state)
	return ctx, state
}

func modalTestVisual(ctx *core.Context, value Widget) theme.ModalVisual {
	return value.visualWithCache(ctx, nil, nil)
}

type fixedModalWidget struct {
	size image.Point
}

type nestedFocusableOverlay struct {
	button widget.Clickable
}

func (w *nestedFocusableOverlay) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:   "nested-focusable",
		Layer: core.OverlayLayerPopup,
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(40, 24))
			return w.button.Layout(gtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(40, 24)}
			})
		},
	})
	return layout.Dimensions{Size: image.Pt(80, 24)}
}

func (w fixedModalWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type constraintProbeWidget struct {
	size        image.Point
	constraints layout.Constraints
	laidOut     bool
}

func (w *constraintProbeWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	w.laidOut = true
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func layoutModalFrame(ctx *core.Context, router *input.Router, modal Widget) {
	layoutModalFrameAt(ctx, router, modal, time.Time{})
}

func layoutModalFrameAt(ctx *core.Context, router *input.Router, modal Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	modal.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func TestUncontrolledPreservesClosedInitialState(t *testing.T) {
	modal := New("", nil).Key("modal").Uncontrolled()
	if modal.hasOpen || !modal.hasDefaultOpen || modal.defaultOpen {
		t.Fatalf("Uncontrolled() = hasOpen=%v hasDefault=%v defaultOpen=%v", modal.hasOpen, modal.hasDefaultOpen, modal.defaultOpen)
	}
}

func layoutModalFrameEnabled(ctx *core.Context, router *input.Router, modal Widget, enabled bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	modalGtx := gtx
	if !enabled {
		modalGtx = modalGtx.Disabled()
	}
	modal.Layout(ctx, modalGtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalPairFrameAt(ctx *core.Context, router *input.Router, first, second Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	first.Layout(ctx, gtx)
	second.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalOverButtonFrame(ctx *core.Context, router *input.Router, modal Widget, onClick func()) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	button.New("Behind").OnClick(onClick).Layout(ctx, gtx)
	modal.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalOverClickableFrame(ctx *core.Context, router *input.Router, modal Widget, background *widget.Clickable, clicked *bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*clicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	modal.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalWithBackgroundFrame(ctx *core.Context, router *input.Router, modal Widget, background *widget.Clickable) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	modal.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
