package flowui

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
)

func TestModalOptions(t *testing.T) {
	var open bool
	modal := Modal("settings", true, "Settings", Text("Body")).
		Header(Text("Custom header")).
		Body(Text("Custom body")).
		Footer(Text("Footer")).
		Icon(Text("!")).
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
	dims := Modal("settings", false, "Settings", Text("Body")).Layout(ctx, testLayoutContext())

	if dims.Size != (image.Point{}) {
		t.Fatalf("closed modal size = %v, want zero", dims.Size)
	}
	if len(ctx.modals) != 0 {
		t.Fatalf("closed modal claimed state, len = %d", len(ctx.modals))
	}
}

func TestModalOpenKeepsState(t *testing.T) {
	ctx := newContext(nil)
	Modal("settings", true, "Settings", Text("Body")).Layout(ctx, testLayoutContext())

	if ctx.modals["settings"] == nil {
		t.Fatal("open modal did not keep state")
	}
}

func TestModalOpenFocusesTrap(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)

	layoutModalFrame(ctx, router, Modal("settings", true, "Settings", Text("Body")))
	state := ctx.modals["settings"]
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

	layoutModalFrame(ctx, router, Modal("settings", true, "Settings", Text("Body")).CloseButton(false))
	state := ctx.modals["settings"]
	if state == nil {
		t.Fatal("open modal did not keep state")
	}
	if !router.Source().Focused(&state.focusTarget) {
		t.Fatal("modal without close button did not move focus into trap")
	}
}

func TestModalTabFromTrapReachesFooterButton(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := Modal("settings", true, "Settings", Text("Body")).
		Footer(Button("footer-action", Text("Action")))

	layoutModalFrame(ctx, router, modal)
	footer := ctx.clickables["footer-action"]
	if footer == nil {
		t.Fatal("footer button state was not kept")
	}

	router.MoveFocus(key.FocusForward)

	if !router.Source().Focused(footer) {
		t.Fatal("tab from modal trap did not reach footer button")
	}
	if router.Source().Focused(&ctx.modals["settings"].dialog) {
		t.Fatal("modal dialog blocker received keyboard focus")
	}
}

func TestModalFocusWrapsAtBoundaries(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	modal := Modal("settings", true, "Settings", Text("Body"))

	layoutModalFrame(ctx, router, modal)
	state := ctx.modals["settings"]
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

func TestModalCloseFocusRingHiddenForPointerFocus(t *testing.T) {
	state := new(modalState)

	if !state.closeFocus.focusVisible(true, nil) {
		t.Fatal("keyboard focus should show modal close focus ring")
	}
	state.closeFocus.focusVisible(false, nil)
	if state.closeFocus.focusVisible(true, []widget.Press{{Start: time.Unix(1, 0)}}) {
		t.Fatal("pointer focus should hide modal close focus ring")
	}
}

func TestModalClosedKeepsVisibleStateForExitAnimation(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	start := time.Unix(1, 0)
	state.ready = true
	state.value = 1
	state.from = 1
	state.to = 1
	state.at = start

	gtx := testLayoutContext()
	gtx.Now = start
	Modal("settings", false, "Settings", Text("Body")).Layout(ctx, gtx)

	if ctx.modals["settings"] == nil {
		t.Fatal("closing modal state was removed before exit animation")
	}
	if state.to != 0 || state.value != 1 {
		t.Fatalf("closing modal animation state = from %v to %v value %v, want closing from visible", state.from, state.to, state.value)
	}
}

func TestModalClosedRemovesStateWhenExitAnimationFinishes(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	start := time.Unix(1, 0)
	state.ready = true
	state.value = 1
	state.from = 1
	state.to = 0
	state.at = start

	gtx := testLayoutContext()
	gtx.Now = start.Add(modalEnterDuration)
	Modal("settings", false, "Settings", Text("Body")).Layout(ctx, gtx)

	if ctx.modals["settings"] != nil {
		t.Fatal("closed modal state was not removed after exit animation")
	}
}

func TestModalBackdropBlocksBackgroundClicksWhenNotDismissable(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	backgroundClicked := false
	modal := Modal("settings", true, "Settings", Text("Body")).Dismissable(false)

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
	modal := Modal("settings", true, "Settings", Text("Body"))
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

func TestModalCloseButtonRequestsClose(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.close.Click()

	closed := false
	Modal("settings", true, "Settings", Text("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		Layout(ctx, testLayoutContext())

	if !closed {
		t.Fatal("close button did not request close")
	}
}

func TestModalDismissAreaRequestsClose(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.dismiss[0].Click()

	closed := false
	Modal("settings", true, "Settings", Text("Body")).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		Layout(ctx, testLayoutContext())

	if !closed {
		t.Fatal("dismiss area did not request close")
	}
}

func TestModalDismissableFalseIgnoresDismissArea(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	state.dismiss[0].Click()

	closed := false
	Modal("settings", true, "Settings", Text("Body")).
		Dismissable(false).
		OnOpenChange(func(open bool) {
			closed = !open
		}).
		Layout(ctx, testLayoutContext())

	if closed {
		t.Fatal("non-dismissable modal closed from dismiss area")
	}
}

func TestModalEscapeRequestsClose(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	closed := false
	modal := Modal("settings", true, "Settings", Text("Body")).
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
	modal := Modal("settings", true, "Settings", Text("Body")).
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

	constraints := Modal("settings", true, "Settings", Text("Body")).
		dialogConstraints(ctx, gtx, image.Pt(500, 300))

	if constraints.Min.X != 360 || constraints.Max.X != 360 {
		t.Fatalf("modal width constraints = %+v, want width 360", constraints)
	}
	if constraints.Max.Y != 280 {
		t.Fatalf("modal max height = %d, want 280", constraints.Max.Y)
	}
}

func TestModalScrollOutsideUsesDialogScrollContainer(t *testing.T) {
	ctx, state := modalTestContextWithState("settings")
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(260, 160)}
	modal := Modal("settings", true, "Settings", fixedModalWidget{size: image.Pt(120, 500)}).
		Scroll(ModalScrollOutside)

	dims := modal.layoutDialogFrame(ctx, gtx, state)

	if dims.Size.Y > gtx.Constraints.Max.Y {
		t.Fatalf("outside scroll dialog height = %d, want <= %d", dims.Size.Y, gtx.Constraints.Max.Y)
	}
	if state.outsideList.Position.Length <= dims.Size.Y {
		t.Fatalf("outside scroll content length = %d, want greater than viewport %d", state.outsideList.Position.Length, dims.Size.Y)
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
			pos := Modal("settings", true, "Settings", Text("Body")).
				Size(size).
				dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

			if pos != image.Pt(140, 88) {
				t.Fatalf("%s footer position = %v, want (140,88)", name, pos)
			}
		})
	}
}

func TestModalMediumKeepsFooterAfterContent(t *testing.T) {
	pos := Modal("settings", true, "Settings", Text("Body")).
		dialogFooterPosition(image.Pt(220, 120), image.Pt(80, 32), 48)

	if pos != image.Pt(140, 48) {
		t.Fatalf("medium footer position = %v, want (140,48)", pos)
	}
}

func TestModalPlacement(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()

	top := Modal("top", true, "Top", Text("Body")).Placement(ModalTop).
		dialogPosition(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80))
	bottom := Modal("bottom", true, "Bottom", Text("Body")).Placement(ModalBottom).
		dialogPosition(ctx, gtx, image.Pt(400, 300), image.Pt(200, 80))

	if top.Y >= bottom.Y {
		t.Fatalf("top placement y = %d, bottom y = %d, want top before bottom", top.Y, bottom.Y)
	}
}

func TestModalDismissRectsExcludeDialog(t *testing.T) {
	rects := modalDismissRects(image.Pt(100, 80), image.Rect(20, 10, 80, 60))

	if rects[0] != image.Rect(0, 0, 100, 10) {
		t.Fatalf("top dismiss rect = %v", rects[0])
	}
	if rects[2].Intersect(image.Rect(20, 10, 80, 60)) != (image.Rectangle{}) {
		t.Fatal("dismiss rect overlaps dialog")
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
	if got := Modal("settings", true, "Settings", Text("Body")).resolvedAnimation(); got != ModalAnimationFadeScale {
		t.Fatalf("default animation = %v, want fade scale", got)
	}
}

func TestModalFadeAnimationDoesNotTransformDialog(t *testing.T) {
	ctx := newContext(nil)
	motion := Modal("settings", true, "Settings", Text("Body")).
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
	motion := Modal("settings", true, "Settings", Text("Body")).
		Animation(ModalAnimationSlideDown).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true)

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := -float32(gtx.Dp(ctx.Theme.Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide down offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalMotionBoundsUsesAnimatedRect(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)
	motion := Modal("settings", true, "Settings", Text("Body")).
		Animation(ModalAnimationSlideDown).
		dialogMotion(ctx, gtx, rect, 0, true)

	got := modalMotionBounds(rect, motion.transform)
	want := rect.Add(image.Pt(0, -gtx.Dp(ctx.Theme.Components.Modal.AnimationDistance)))
	if got != want {
		t.Fatalf("motion bounds = %v, want %v", got, want)
	}
}

func TestModalSlideUpAnimationOffsetsFromBottom(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	progress := float32(0.25)
	motion := Modal("settings", true, "Settings", Text("Body")).
		Animation(ModalAnimationSlideUp).
		dialogMotion(ctx, gtx, image.Rect(10, 20, 110, 120), progress, true)

	_, _, ox, _, _, oy := motion.transform.Elems()
	want := float32(gtx.Dp(ctx.Theme.Components.Modal.AnimationDistance)) * (1 - progress)
	if ox != 0 || oy != want {
		t.Fatalf("slide up offset = (%v,%v), want (0,%v)", ox, oy, want)
	}
}

func TestModalBounceScaleOvershootsOnlyWhenOpening(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	rect := image.Rect(10, 20, 110, 120)

	openMotion := Modal("settings", true, "Settings", Text("Body")).
		Animation(ModalAnimationBounceScale).
		dialogMotion(ctx, gtx, rect, 0.65, true)
	openScale, _, _, _, _, _ := openMotion.transform.Elems()
	if openScale <= 1 {
		t.Fatalf("opening bounce scale = %v, want overshoot", openScale)
	}

	closeMotion := Modal("settings", true, "Settings", Text("Body")).
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
	motion := Modal("settings", true, "Settings", Text("Body")).
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
	motion := Modal("settings", true, "Settings", Text("Body")).
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
	body, ok := Modal("settings", true, "Settings", Text("Body")).
		styleBody(ctx, Text("Body")).(TextWidget)
	if !ok {
		t.Fatal("styled body is not TextWidget")
	}
	if body.size != ctx.Theme.Components.Modal.BodyTextSize {
		t.Fatalf("body text size = %v, want %v", body.size, ctx.Theme.Components.Modal.BodyTextSize)
	}
	if body.color != ctx.Theme.Palette.MutedForeground {
		t.Fatalf("body text color = %#v, want muted foreground", body.color)
	}
}

func TestModalCloseLabelUsesContextLanguage(t *testing.T) {
	if got := modalCloseLabel(newContextWithThemeAndLanguage(nil, nil, LanguageEnglish)); got != "Close" {
		t.Fatalf("english close label = %q, want Close", got)
	}
	if got := modalCloseLabel(newContextWithThemeAndLanguage(nil, nil, LanguageChinese)); got != "关闭" {
		t.Fatalf("chinese close label = %q, want 关闭", got)
	}
}

func modalTestContextWithState(key string) (*Context, *modalState) {
	state := new(modalState)
	ctx := newContext(nil)
	ctx.modals = map[string]*modalState{key: state}
	return ctx, state
}

type fixedModalWidget struct {
	size image.Point
}

func (w fixedModalWidget) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type constraintProbeWidget struct {
	size        image.Point
	constraints layout.Constraints
	laidOut     bool
}

func (w *constraintProbeWidget) Layout(_ *Context, gtx layout.Context) layout.Dimensions {
	w.constraints = gtx.Constraints
	w.laidOut = true
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func layoutModalFrame(ctx *Context, router *input.Router, modal ModalWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx.beginFrame()
	modal.Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}

func layoutModalOverButtonFrame(ctx *Context, router *input.Router, modal ModalWidget, onClick func()) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx.beginFrame()
	Button("behind", Text("Behind")).OnClick(onClick).Layout(ctx, gtx)
	modal.Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}

func layoutModalOverClickableFrame(ctx *Context, router *input.Router, modal ModalWidget, background *widget.Clickable, clicked *bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx.beginFrame()
	for background.Clicked(gtx) {
		*clicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	modal.Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}
