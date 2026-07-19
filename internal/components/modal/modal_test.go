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
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotClickable = "clickable"

func TestModalCloseGeometryMatchesHeroUI(t *testing.T) {
	component := theme.DefaultTheme().Components.Modal
	if component.CloseInset != 16 {
		t.Fatalf("modal close inset = %v, want 16", component.CloseInset)
	}
}

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func newContextWithTheme(_ any, value *theme.Theme) *frame.Context {
	return frame.New(nil, value, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func Button(key string, child frame.Widget) button.ButtonWidget {
	return button.Button(key, child)
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

func TestModalOptions(t *testing.T) {
	var open bool
	modal := Modal("settings", true, "Settings", text.New("Body")).
		Header(text.New("Custom header")).
		Body(text.New("Custom body")).
		Footer(text.New("Footer")).
		Icon(text.New("!")).
		OnOpenChange(func(next bool) {
			open = next
		}).
		Size(ModalLarge).
		Placement(ModalBottom).
		Backdrop(ModalBackdropBlur).
		Scroll(ModalScrollOutside).
		Animation(ModalAnimationBounceScale).
		Dismissable(false).
		KeyboardDismissDisabled(true).
		CloseButton(false)

	if modal.key != "settings" || !modal.open || modal.title != "Settings" || modal.body == nil {
		t.Fatal("modal constructor/options did not set base fields")
	}
	if modal.header == nil || modal.footer == nil || modal.icon == nil || modal.onOpenChange == nil {
		t.Fatal("modal content/callback options were not set")
	}
	if modal.size != ModalLarge || modal.placement != ModalBottom || modal.backdrop != ModalBackdropBlur || modal.scroll != ModalScrollOutside {
		t.Fatal("modal visual options were not set")
	}
	if modal.animation != ModalAnimationBounceScale {
		t.Fatal("modal animation option was not set")
	}
	if modal.isDismissable() || !modal.keyboardDismissDisabled || modal.showCloseButton() {
		t.Fatal("modal behavior options were not set")
	}
	modal.onOpenChange(true)
	if !open {
		t.Fatal("modal onOpenChange did not receive true")
	}
}

func TestModalClosedReturnsZeroAndDoesNotClaimState(t *testing.T) {
	ctx := newContext(nil)
	dims := Modal("settings", false, "Settings", text.New("Body")).Layout(ctx, testLayoutContext())

	if dims.Size != (image.Point{}) {
		t.Fatalf("closed modal size = %v, want zero", dims.Size)
	}
	if frame.StateLen(ctx) != 0 {
		t.Fatalf("closed modal claimed state, len = %d", frame.StateLen(ctx))
	}
}

func TestModalOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	Modal("settings", true, "Settings", text.New("Body")).Layout(ctx, testLayoutContext())

	if testComponentState[modalState](ctx, "settings", stateSlotModal) == nil {
		t.Fatal("open modal did not keep state")
	}
}

func TestModalOpenFocusesTrap(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)

	layoutModalFrame(ctx, router, Modal("settings", true, "Settings", text.New("Body")))
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

	layoutModalFrame(ctx, router, Modal("settings", true, "Settings", text.New("Body")).CloseButton(false))
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
	modal := Modal("settings", true, "Settings", text.New("Body"))

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
	outer := Modal("outer", true, "Outer", text.New("Outer body"))
	innerOpen := true
	inner := func() ModalWidget {
		return Modal("inner", innerOpen, "Inner", text.New("Inner body"))
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
	build := func() ModalWidget {
		return Modal("settings", open, "Settings", text.New("Body"))
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
	modal := Modal("settings", true, "Settings", text.New("Body")).
		Footer(Button("footer-action", text.New("Action")))

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
	modal := Modal("settings", true, "Settings", text.New("Body"))

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
	modal := Modal("settings", true, "Settings", nested).CloseButton(false)

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
	Modal("settings", false, "Settings", text.New("Body")).Layout(ctx, gtx)

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
	Modal("settings", false, "Settings", text.New("Body")).Layout(ctx, gtx)

	if testComponentState[modalState](ctx, "settings", stateSlotModal) != nil {
		t.Fatal("closed modal state was not removed after exit animation")
	}
}

func TestModalBackdropBlocksBackgroundClicksWhenNotDismissable(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	backgroundClicked := false
	modal := Modal("settings", true, "Settings", text.New("Body")).Dismissable(false)

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
	modal := Modal("settings", true, "Settings", text.New("Body"))
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
	modal := Modal("settings", false, "Settings", text.New("Body"))
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
	state.close.Click()

	closed := false
	modal := Modal("settings", true, "Settings", text.New("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutModalFrame(ctx, new(input.Router), modal)

	if !closed {
		t.Fatal("close button did not request close")
	}
}

func TestModalDismissAreaRequestsClose(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.dismiss[0].Click()

	closed := false
	modal := Modal("settings", true, "Settings", text.New("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutModalFrame(ctx, new(input.Router), modal)

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestModalDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.dismiss[0].Click()

	closed := false
	modal := Modal("settings", true, "Settings", text.New("Body")).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		})
	layoutModalFrame(ctx, new(input.Router), modal)

	if closed {
		t.Fatal("non-dismissable modal closed from dismiss area")
	}
}

func TestModalEscapeRequestsClose(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	closed := false
	modal := Modal("settings", true, "Settings", text.New("Body")).
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
	modal := Modal("settings", true, "Settings", text.New("Body")).
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

	constraints := Modal("settings", true, "Settings", text.New("Body")).
		dialogConstraints(ctx, gtx, image.Pt(500, 300))

	if constraints.Min.X != 360 || constraints.Max.X != 360 {
		t.Fatalf("modal width constraints = %+v, want width 360", constraints)
	}
	if constraints.Max.Y != 280 {
		t.Fatalf("modal max height = %d, want 280", constraints.Max.Y)
	}
}

func TestModalScrollOutsideKeepsScrollingWithoutScrollbarGutter(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(260, 160)}
	body := &constraintProbeWidget{size: image.Pt(120, 500)}
	modal := Modal("settings", true, "Settings", body).
		Scroll(ModalScrollOutside)

	dims := modal.layoutDialogFrame(ctx, gtx, state)

	if dims.Size.Y > gtx.Constraints.Max.Y {
		t.Fatalf("outside scroll dialog height = %d, want <= %d", dims.Size.Y, gtx.Constraints.Max.Y)
	}
	if state.outsideList.Position.Length <= dims.Size.Y {
		t.Fatalf("outside scroll content length = %d, want greater than viewport %d", state.outsideList.Position.Length, dims.Size.Y)
	}
	wantWidth := gtx.Constraints.Max.X - gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.Padding)*2
	if body.constraints.Max.X != wantWidth {
		t.Fatalf("outside scroll body width = %d, want %d without scrollbar gutter", body.constraints.Max.X, wantWidth)
	}
}

func TestModalScrollOutsideKeepsFullSurfaceHeight(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(260, 160))
	modal := Modal("settings", true, "Settings", fixedModalWidget{size: image.Pt(120, 20)}).
		Size(ModalFull).
		Scroll(ModalScrollOutside)

	dims := modal.layoutDialogFrame(ctx, gtx, state)

	if dims.Size != image.Pt(260, 160) {
		t.Fatalf("full outside scroll dimensions = %v, want (260,160)", dims.Size)
	}
	if state.outsideList.Position.Length <= 0 {
		t.Fatalf("full outside scroll content length = %d, want positive", state.outsideList.Position.Length)
	}
}

func TestModalScrollOutsideKeepsCoverSurfaceHeight(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(260, 160))
	modal := Modal("settings", true, "Settings", fixedModalWidget{size: image.Pt(120, 20)}).
		Size(ModalCover).
		Scroll(ModalScrollOutside)

	dims := modal.layoutDialogFrame(ctx, gtx, state)

	if dims.Size != image.Pt(260, 160) {
		t.Fatalf("cover outside scroll dimensions = %v, want (260,160)", dims.Size)
	}
	if state.outsideList.Position.Length <= 0 {
		t.Fatalf("cover outside scroll content length = %d, want positive", state.outsideList.Position.Length)
	}
}

func TestModalLargeSizesRelaxFooterSectionMinimumConstraints(t *testing.T) {
	for name, size := range map[string]ModalSize{
		"cover": ModalCover,
		"full":  ModalFull,
	} {
		t.Run(name, func(t *testing.T) {
			ctx, state := modalTestContextWithState("settings")
			gtx := testLayoutContext()
			gtx.Constraints = layout.Exact(image.Pt(260, 160))
			footer := &constraintProbeWidget{size: image.Pt(80, 32)}
			modal := Modal("settings", true, "Settings", fixedModalWidget{size: image.Pt(120, 20)}).
				Size(size).
				Footer(footer)

			dims := modal.layoutDialogFrame(ctx, gtx, state)

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

func TestModalLargeSizesAnchorFooterToBottom(t *testing.T) {
	for name, size := range map[string]ModalSize{
		"cover": ModalCover,
		"full":  ModalFull,
	} {
		t.Run(name, func(t *testing.T) {
			pos := Modal("settings", true, "Settings", text.New("Body")).
				Size(size).
				dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

			if pos != image.Pt(140, 88) {
				t.Fatalf("%s footer position = %v, want (140,88)", name, pos)
			}
		})
	}
}

func TestModalMediumKeepsFooterAfterContent(t *testing.T) {
	pos := Modal("settings", true, "Settings", text.New("Body")).
		dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

	if pos != image.Pt(140, 48) {
		t.Fatalf("medium footer position = %v, want (140,48)", pos)
	}
}

func TestModalPlacement(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()

	top := Modal("top", true, "Top", text.New("Body")).Placement(ModalTop).
		dialogPosition(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80))
	bottom := Modal("bottom", true, "Bottom", text.New("Body")).Placement(ModalBottom).
		dialogPosition(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80))

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
	if got := modalDialogScale(0, 0); got != 0.95 {
		t.Fatalf("dialog fallback scale = %v, want 0.95", got)
	}
}

func TestModalResolvedAnimationDefaultsToFadeScale(t *testing.T) {
	if got := Modal("settings", true, "Settings", text.New("Body")).resolvedAnimation(); got != ModalAnimationFadeScale {
		t.Fatalf("default animation = %v, want fade scale", got)
	}
}

func TestModalFadeAnimationDoesNotTransformDialog(t *testing.T) {
	ctx := newContext(nil)
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationFade).
		dialogMotion(ctx, testLayoutContext(), image.Rect(10, 20, 110, 120), 0.5, true)

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
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationSlideDown).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true)

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := -float32(gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide down offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalMotionBoundsUsesAnimatedRect(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationSlideDown).
		dialogMotion(ctx, gtx, rect, 0, true)

	got := overlay.AffineRectBounds(rect, motion.transform)
	want := rect.Add(image.Pt(0, -gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.AnimationDistance)))
	if got != want {
		t.Fatalf("motion bounds = %v, want %v", got, want)
	}
}

func TestModalSlideUpAnimationOffsetsFromBottom(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	progress := float32(0.25)
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationSlideUp).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true)

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := float32(gtx.Dp(frame.ActiveTheme(ctx).Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide up offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalBounceScaleOvershootsOnlyWhenOpening(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)

	openMotion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationBounceScale).
		dialogMotion(ctx, gtx, rect, 0.65, true)
	openScale, _, _, _, _, _ := openMotion.transform.Elems()
	if openScale <= 1 {
		t.Fatalf("opening bounce scale = %v, want overshoot", openScale)
	}

	closeMotion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationBounceScale).
		dialogMotion(ctx, gtx, rect, 0.65, false)
	closeScale, _, _, _, _, _ := closeMotion.transform.Elems()
	if closeScale <= 0 || closeScale >= 1 {
		t.Fatalf("closing bounce scale = %v, want clean scale below 1", closeScale)
	}
}

func TestModalBounceScaleClampsToFinalScale(t *testing.T) {
	if got := modalDialogBounceScale(0.95, 1.035, 1); got != 1 {
		t.Fatalf("final bounce scale = %v, want 1", got)
	}
}

func TestModalZoomOutAnimationStartsLarge(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationZoomOut).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), 0, true)

	scale, _, _, _, _, _ := motion.transform.Elems()
	if scale <= 1 {
		t.Fatalf("zoom out start scale = %v, want greater than 1", scale)
	}
	if got := modalDialogZoomOutScale(1); got != 1 {
		t.Fatalf("final zoom out scale = %v, want 1", got)
	}
}

func TestModalPopAnimationOvershoots(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	motion := Modal("settings", true, "Settings", text.New("Body")).
		Animation(ModalAnimationPop).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), 0.58, true)

	scale, _, _, _, _, _ := motion.transform.Elems()
	if scale <= 1 {
		t.Fatalf("pop scale = %v, want overshoot", scale)
	}
	if got := modalDialogPopScale(1); got != 1 {
		t.Fatalf("final pop scale = %v, want 1", got)
	}
}

func TestModalBackdropStyle(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.Modal.BlurBackdrop.A = 200

	if got := modalStyleFor(&theme, ModalBackdropBlur, ModalMedium).backdrop.A; got != 200 {
		t.Fatalf("blur backdrop alpha = %d, want 200", got)
	}
	if got := modalStyleFor(&theme, ModalBackdropTransparent, ModalMedium).backdrop.A; got != 0 {
		t.Fatalf("transparent backdrop alpha = %d, want 0", got)
	}
}

func TestModalBodyTextStyle(t *testing.T) {
	ctx := newContext(nil)
	body, ok := Modal("settings", true, "Settings", text.New("Body")).
		styleBody(ctx, text.New("Body")).(text.Widget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	if body.ConfiguredSize() != frame.ActiveTheme(ctx).Components.Modal.BodyTextSize {
		t.Fatalf("body text size = %v, want %v", body.ConfiguredSize(), frame.ActiveTheme(ctx).Components.Modal.BodyTextSize)
	}
	if col, _ := body.ConfiguredColor(); col != frame.ActiveTheme(ctx).Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", col)
	}
}

func modalTestContextWithState(key string) (*frame.Context, *modalState) {
	state := new(modalState)
	ctx := newContext(nil)
	testSetComponentState(ctx, key, stateSlotModal, state)
	return ctx, state
}

type fixedModalWidget struct {
	size image.Point
}

type nestedFocusableOverlay struct {
	button widget.Clickable
}

func (w *nestedFocusableOverlay) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:   "nested-focusable",
		Layer: frame.OverlayLayerPopup,
		Layout: func(gtx layout.Context, _ image.Rectangle, _ bool) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(40, 24))
			return w.button.Layout(gtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(40, 24)}
			})
		},
	})
	return layout.Dimensions{Size: image.Pt(80, 24)}
}

func (w fixedModalWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type constraintProbeWidget struct {
	size        image.Point
	constraints layout.Constraints
	laidOut     bool
}

func (w *constraintProbeWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	w.laidOut = true
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func layoutModalFrame(ctx *frame.Context, router *input.Router, modal ModalWidget) {
	layoutModalFrameAt(ctx, router, modal, time.Time{})
}

func layoutModalFrameAt(ctx *frame.Context, router *input.Router, modal ModalWidget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	modal.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalFrameEnabled(ctx *frame.Context, router *input.Router, modal ModalWidget, enabled bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	modalGtx := gtx
	if !enabled {
		modalGtx = modalGtx.Disabled()
	}
	modal.Layout(ctx, modalGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalPairFrameAt(ctx *frame.Context, router *input.Router, first, second ModalWidget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	first.Layout(ctx, gtx)
	second.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalOverButtonFrame(ctx *frame.Context, router *input.Router, modal ModalWidget, onClick func()) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	Button("behind", text.New("Behind")).OnClick(onClick).Layout(ctx, gtx)
	modal.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalOverClickableFrame(ctx *frame.Context, router *input.Router, modal ModalWidget, background *widget.Clickable, clicked *bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*clicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	modal.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutModalWithBackgroundFrame(ctx *frame.Context, router *input.Router, modal ModalWidget, background *widget.Clickable) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	modal.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
