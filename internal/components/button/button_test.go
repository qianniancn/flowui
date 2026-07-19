package button

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotClickable = "clickable"

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *frame.Context, key, slot string, value *T) {
	frame.UseStateWith(ctx, key, slot, func() *T { return value })
}

func TestButtonOnClick(t *testing.T) {
	clicked := false
	b := Button("add", text.New("Add").Size(20)).
		OnClick(func() {
			clicked = true
		})
	if b.key != "add" {
		t.Fatalf("key = %q, want add", b.key)
	}
	if _, ok := b.child.(text.Widget); !ok {
		t.Fatalf("child = %T, want TextWidget", b.child)
	}
	if b.onClick == nil {
		t.Fatal("missing click handler")
	}

	b.onClick()
	if !clicked {
		t.Fatal("click handler was not called")
	}
}

func TestButtonDisabled(t *testing.T) {
	b := Button("save", text.New("Save")).Disabled(true)

	if !b.disabled {
		t.Fatal("button was not disabled")
	}
}

func TestButtonOptions(t *testing.T) {
	b := Button("save", text.New("Save")).
		Label("Save changes").
		Variant(ButtonOutline).
		Size(ButtonLarge).
		Loading(true).
		FullWidth().
		IconOnly()

	if b.variant != ButtonOutline {
		t.Fatal("button variant was not set")
	}
	if b.label != "Save changes" {
		t.Fatal("button label was not set")
	}
	if b.size != ButtonLarge {
		t.Fatal("button size was not set")
	}
	if !b.loading {
		t.Fatal("button was not loading")
	}
	if !b.fullWidth {
		t.Fatal("button was not full width")
	}
	if !b.iconOnly {
		t.Fatal("button was not icon only")
	}
}

func TestButtonThemeAppliesOnlyToCurrentInstance(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	themedProbe := &themeProbeWidget{}
	baseProbe := &themeProbeWidget{}
	base := Button("default", baseProbe)
	accent := color.NRGBA{R: 0x17, G: 0x72, B: 0x45, A: 0xff}
	themed := Button("themed", themedProbe).Theme(func(theme *theme.Theme) {
		theme.Components.Button.Radius = 0
		theme.Components.Button.PressedScaleMedium = 0.9
		theme.Palette.Accent = accent
	})
	resolved := themed.activeTheme(ctx)

	themed.Layout(ctx, testLayoutContext())
	if resolved.Components.Button.Radius != 0 || resolved.Components.Button.PressedScaleMedium != 0.9 || resolved.Palette.Accent != accent {
		t.Fatalf("resolved button theme = %#v", resolved)
	}
	if themedProbe.radius != 0 {
		t.Fatalf("button child radius = %v, want inherited instance radius 0", themedProbe.radius)
	}
	if activeTheme.Components.Button.Radius != 24 || activeTheme.Components.Button.PressedScaleMedium != 0.97 || activeTheme.Palette.Accent == accent {
		t.Fatalf("button theme mutated active theme: %#v", activeTheme)
	}
	base.Layout(ctx, testLayoutContext())
	if baseProbe.radius != 24 {
		t.Fatalf("sibling button radius = %v, want 24", baseProbe.radius)
	}
}

func TestButtonSemanticsIncludeLabelAndDisabledState(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	Button("save", text.New("Save")).Label("Save changes").Disabled(true).Layout(newContext(nil), gtx)
	router.Frame(&ops)
	node, ok := buttonSemanticNode(router.AppendSemantics(nil))
	if !ok || node.Desc.Label != "Save changes" || !node.Desc.Disabled {
		t.Fatalf("button semantics = %#v", node.Desc)
	}
}

func TestButtonDefaultLayout(t *testing.T) {
	dims := Button("save", text.New("Save")).Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y != 40 {
		t.Fatalf("button height = %d, want 40", dims.Size.Y)
	}
}

func TestButtonFullWidth(t *testing.T) {
	dims := Button("save", text.New("Save")).FullWidth().Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("button width = %d, want 300", dims.Size.X)
	}
}

func TestButtonIconOnly(t *testing.T) {
	dims := Button("save", text.New("S")).IconOnly().Layout(newContext(nil), testLayoutContext())

	if dims.Size != image.Pt(40, 40) {
		t.Fatalf("button size = %v, want (40,40)", dims.Size)
	}
}

func TestButtonFullWidthIconOnly(t *testing.T) {
	dims := Button("save", text.New("S")).FullWidth().IconOnly().Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("button width = %d, want 300", dims.Size.X)
	}
	if dims.Size.Y != 40 {
		t.Fatalf("button height = %d, want 40", dims.Size.Y)
	}
}

func TestButtonLoadingKeepsIntrinsicWidth(t *testing.T) {
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		for _, size := range []ButtonSize{ButtonSmall, ButtonMedium, ButtonLarge} {
			normalGtx := testLayoutContext()
			normalGtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			normal := Button("load", text.New("Load saved value")).
				Size(size).
				Layout(newContext(nil), normalGtx)
			loadingGtx := testLayoutContext()
			loadingGtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			loading := Button("load", text.New("Load saved value")).
				Size(size).
				Loading(true).
				Layout(newContext(nil), loadingGtx)
			if loading.Size != normal.Size {
				t.Fatalf("scale %.2f size %d loading dimensions = %v, want %v", scale, size, loading.Size, normal.Size)
			}
		}
	}
}

func TestButtonSizeStyle(t *testing.T) {
	theme := DefaultTheme()
	small := buttonSizeStyle(&theme, ButtonSmall, false)
	large := buttonSizeStyle(&theme, ButtonLarge, false)
	icon := buttonSizeStyle(&theme, ButtonMedium, true)

	if small.height != 36 {
		t.Fatalf("small height = %v, want 36", small.height)
	}
	if large.height != 44 || large.textSize != 16 {
		t.Fatalf("large style = %+v, want height 44 text 16", large)
	}
	if icon.inset.Left != 0 || icon.inset.Right != 0 {
		t.Fatalf("icon inset = %+v, want horizontal zero", icon.inset)
	}
	styled := Button("shape", text.New("Shape")).style(&theme, new(widget.Clickable))
	if styled.radius != 24 || styled.borderWidth != 1 || theme.Components.Button.SpinnerStrokeWidth != 2 {
		t.Fatalf("button shape = radius %v border %v spinner stroke %v, want 24/1/2", styled.radius, styled.borderWidth, theme.Components.Button.SpinnerStrokeWidth)
	}
}

func TestButtonSpinnerSize(t *testing.T) {
	theme := DefaultTheme()
	if size := buttonSpinnerSize(&theme, ButtonSmall); size != 14 {
		t.Fatalf("small spinner = %v, want 14", size)
	}
	if size := buttonSpinnerSize(&theme, ButtonMedium); size != 16 {
		t.Fatalf("medium spinner = %v, want 16", size)
	}
	if size := buttonSpinnerSize(&theme, ButtonLarge); size != 18 {
		t.Fatalf("large spinner = %v, want 18", size)
	}
}

func TestButtonSpinnerPhase(t *testing.T) {
	start := time.Unix(1, 0)
	if phase := buttonSpinnerPhase(time.Time{}, buttonSpinnerPeriod); phase != 0 {
		t.Fatalf("zero phase = %v, want 0", phase)
	}
	if phase := buttonSpinnerPhase(start, buttonSpinnerPeriod); phase != buttonSpinnerPhase(start.Add(buttonSpinnerPeriod), buttonSpinnerPeriod) {
		t.Fatalf("phase did not repeat after one period")
	}
	if phase := buttonSpinnerPhase(start, 0); phase != 0 {
		t.Fatalf("disabled motion phase = %v, want 0", phase)
	}
}

func TestButtonVariantColors(t *testing.T) {
	theme := DefaultTheme()
	outline := buttonColors(&theme, ButtonOutline)
	ghost := buttonColors(&theme, ButtonGhost)
	danger := buttonColors(&theme, ButtonDanger)

	if !outline.hasBorder {
		t.Fatal("outline button is missing border")
	}
	wantOutlineHover := theme.Palette.DefaultColor()
	wantOutlineHover.A = byte(uint16(wantOutlineHover.A) * 0x99 / 0xff)
	if outline.hover != wantOutlineHover {
		t.Fatalf("outline hover = %v, want 60%% default color %v", outline.hover, wantOutlineHover)
	}
	if ghost.pressed != ghost.hover {
		t.Fatalf("ghost pressed = %v, want hover color %v", ghost.pressed, ghost.hover)
	}
	if outline.pressed != ghost.hover {
		t.Fatalf("outline pressed = %v, want default color %v", outline.pressed, ghost.hover)
	}
	if danger.bg.A == 0 {
		t.Fatal("danger button background is transparent")
	}
}

func TestButtonFocusDoesNotChangeBackground(t *testing.T) {
	theme := DefaultTheme()
	button := Button("ghost", text.New("Ghost")).Variant(ButtonGhost)
	style := button.style(&theme, new(widget.Clickable))
	colors := buttonColors(&theme, ButtonGhost)

	if style.bg != colors.bg {
		t.Fatalf("background = %v, want %v", style.bg, colors.bg)
	}
}

func TestButtonFocusesOnPress(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutButtonFrame(ctx, router)
	button := testComponentState[widget.Clickable](ctx, "save", stateSlotClickable)
	if button == nil {
		t.Fatal("missing button state")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(20, 20),
	})
	layoutButtonFrame(ctx, router)

	if !router.Source().Focused(button) {
		t.Fatal("button did not gain focus after press")
	}
}

func TestFocusClearsOnOutsidePress(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutButtonFrame(ctx, router)
	button := testComponentState[widget.Clickable](ctx, "save", stateSlotClickable)
	if button == nil {
		t.Fatal("missing button state")
	}
	router.Source().Execute(key.FocusCmd{Tag: button})
	layoutButtonFrame(ctx, router)

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(200, 120),
	})
	layoutButtonFrame(ctx, router)

	if router.Source().Focused(button) {
		t.Fatal("button kept focus after outside press")
	}
}

func TestButtonLoadingBlocksClick(t *testing.T) {
	clicked := false
	clickable := new(widget.Clickable)
	clickable.Click()
	ctx := newContext(nil)
	testSetComponentState(ctx, "save", stateSlotClickable, clickable)

	Button("save", text.New("Save")).
		Loading(true).
		OnClick(func() {
			clicked = true
		}).
		Layout(ctx, testLayoutContext())

	if clicked {
		t.Fatal("loading button handled a click")
	}
}

func TestButtonPressedScale(t *testing.T) {
	theme := DefaultTheme()
	if scale := buttonPressedScale(&theme, ButtonSmall); scale != 0.98 {
		t.Fatalf("small scale = %v, want 0.98", scale)
	}
	if scale := buttonPressedScale(&theme, ButtonMedium); scale != 0.97 {
		t.Fatalf("medium scale = %v, want 0.97", scale)
	}
	if scale := buttonPressedScale(&theme, ButtonLarge); scale != 0.96 {
		t.Fatalf("large scale = %v, want 0.96", scale)
	}
}

func TestButtonAnimationScalePressIn(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start.Add(buttonPressInDuration / 2)
	theme := DefaultTheme()

	scale := buttonAnimationScale(gtx, []widget.Press{{Start: start}}, &theme, ButtonMedium, false)

	if scale <= buttonPressedScale(&theme, ButtonMedium) || scale >= 1 {
		t.Fatalf("press scale = %v, want between pressed and rest", scale)
	}
}

func TestButtonAnimationScaleRelease(t *testing.T) {
	end := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = end.Add(buttonPressOutDuration / 2)
	theme := DefaultTheme()

	scale := buttonAnimationScale(gtx, []widget.Press{{Start: end.Add(-time.Second), End: end}}, &theme, ButtonMedium, false)

	if scale <= buttonPressedScale(&theme, ButtonMedium) || scale >= 1 {
		t.Fatalf("release scale = %v, want between pressed and rest", scale)
	}
}

func TestButtonAnimationScaleDone(t *testing.T) {
	end := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = end.Add(buttonPressOutDuration)
	theme := DefaultTheme()

	scale := buttonAnimationScale(gtx, []widget.Press{{Start: end.Add(-time.Second), End: end}}, &theme, ButtonMedium, false)

	if scale != 1 {
		t.Fatalf("scale = %v, want 1", scale)
	}
}

func TestButtonAnimationScaleDisabled(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start.Add(buttonPressInDuration / 2)
	theme := DefaultTheme()

	scale := buttonAnimationScale(gtx, []widget.Press{{Start: start}}, &theme, ButtonMedium, true)

	if scale != 1 {
		t.Fatalf("disabled scale = %v, want 1", scale)
	}
}

func TestButtonBackgroundTransition(t *testing.T) {
	state := new(buttonState)
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	to := color.NRGBA{R: 110, G: 120, B: 130, A: 255}

	gtx := testLayoutContext()
	gtx.Now = start
	if got := state.background(gtx, from); got != from {
		t.Fatalf("initial background = %v, want %v", got, from)
	}

	gtx.Now = start
	if got := state.background(gtx, to); got != from {
		t.Fatalf("transition start = %v, want %v", got, from)
	}

	gtx.Now = start.Add(buttonColorDuration / 2)
	mid := state.background(gtx, to)
	if mid == from || mid == to {
		t.Fatalf("transition midpoint = %v, want between %v and %v", mid, from, to)
	}

	gtx.Now = start.Add(buttonColorDuration)
	if got := state.background(gtx, to); got != to {
		t.Fatalf("transition end = %v, want %v", got, to)
	}
}

func TestButtonTransparentBackgroundTransitionKeepsColor(t *testing.T) {
	to := color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff}

	fadeIn := render.LerpColor(color.NRGBA{}, to, 0.5)
	if fadeIn.R != to.R || fadeIn.G != to.G || fadeIn.B != to.B {
		t.Fatalf("fade in color = %v, want rgb from %v", fadeIn, to)
	}
	if fadeIn.A == 0 || fadeIn.A == to.A {
		t.Fatalf("fade in alpha = %d, want between 0 and %d", fadeIn.A, to.A)
	}

	fadeOut := render.LerpColor(to, color.NRGBA{}, 0.5)
	if fadeOut.R != to.R || fadeOut.G != to.G || fadeOut.B != to.B {
		t.Fatalf("fade out color = %v, want rgb from %v", fadeOut, to)
	}
	if fadeOut.A == 0 || fadeOut.A == to.A {
		t.Fatalf("fade out alpha = %d, want between 0 and %d", fadeOut.A, to.A)
	}
}

func TestButtonFocusOpacityTransition(t *testing.T) {
	state := new(buttonState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.focusOpacity(gtx, false); got != 0 {
		t.Fatalf("initial focus = %v, want 0", got)
	}

	gtx.Now = start
	if got := state.focusOpacity(gtx, true); got != 0 {
		t.Fatalf("focus start = %v, want 0", got)
	}

	gtx.Now = start.Add(buttonFocusDuration / 2)
	mid := state.focusOpacity(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("focus midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(buttonFocusDuration)
	if got := state.focusOpacity(gtx, true); got != 1 {
		t.Fatalf("focus end = %v, want 1", got)
	}
}

func TestDisabledButtonAnimationInvalidates(t *testing.T) {
	var router input.Router
	var ops op.Ops
	now := time.Unix(1, 0)
	theme := DefaultTheme()
	normal := buttonColors(&theme, ButtonPrimary).bg
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	ctx := newContext(nil)
	state := new(buttonState)
	state.background(gtx, normal)
	testSetComponentState(ctx, "save", stateSlotButton, state)

	Button("save", text.New("Save")).Disabled(true).Layout(ctx, gtx)
	router.Frame(&ops)

	if _, wake := router.WakeupTime(); !wake {
		t.Fatal("disabled button animation did not request a redraw")
	}
}

func TestLoadingButtonInvalidates(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	}

	Button("save", text.New("Save")).Loading(true).Layout(newContext(nil), gtx)
	router.Frame(&ops)

	if _, wake := router.WakeupTime(); !wake {
		t.Fatal("loading button did not request a redraw")
	}
}

func TestLoadingButtonRespectsDisabledMotion(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	}
	themeValue := theme.DefaultTheme()
	themeValue.Motion.Enabled = false
	ctx := frame.New(nil, &themeValue, locale.LanguageAuto)

	Button("save", text.New("Save")).Loading(true).Layout(ctx, gtx)
	router.Frame(&ops)

	if _, wake := router.WakeupTime(); wake {
		t.Fatal("loading button requested redraw with motion disabled")
	}
}

func TestButtonPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	Button("save", probe).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("button child was laid out with enabled context")
	}
}

func TestButtonPassesForegroundColorToCustomContent(t *testing.T) {
	probe := &foregroundProbeWidget{}
	ctx := newContext(nil)

	Button("icon", probe).Layout(ctx, testLayoutContext())

	want := buttonColors(frame.ActiveTheme(ctx), ButtonPrimary).fg
	if probe.foreground != want {
		t.Fatalf("child foreground = %#v, want %#v", probe.foreground, want)
	}
}

type enabledProbeWidget struct {
	enabled bool
}

func (w *enabledProbeWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

type foregroundProbeWidget struct {
	foreground color.NRGBA
}

func (w *foregroundProbeWidget) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	w.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

type themeProbeWidget struct {
	radius unit.Dp
}

func (w *themeProbeWidget) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	w.radius = frame.ActiveTheme(ctx).Components.Button.Radius
	return layout.Dimensions{Size: image.Pt(16, 16)}
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

func layoutButtonFrame(ctx *frame.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	Button("save", text.New("Save")).Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func buttonSemanticNode(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			return node, true
		}
		if child, ok := buttonSemanticNode(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}
