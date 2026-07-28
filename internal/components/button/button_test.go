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
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
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

func TestButtonWidgetCascadesScopeBeforeInstance(t *testing.T) {
	ctx := newContext(nil)
	scope := flowstyle.Style{}.
		PaddingX(12).
		Background(flowstyle.TokenSurface).
		TextColor(flowstyle.TokenForeground)

	restore := frame.PushStyle(ctx, scope)
	defer restore()
	instanceBackground := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	button := Button("styled", text.New("Styled")).Style(
		flowstyle.Style{}.
			PaddingX(20).
			Background(solid(instanceBackground)).
			When(flowstyle.Loading, flowstyle.Style{}.Opacity(0.25)),
	)

	resolved := button.staticStyle(ctx, flowstyle.StyleState{Loading: true})
	if resolved.inset.Left != 20 || resolved.inset.Right != 20 {
		t.Fatalf("instance padding = %+v, want horizontal 20", resolved.inset)
	}
	if resolved.bg != instanceBackground {
		t.Fatalf("instance background = %#v, want %#v", resolved.bg, instanceBackground)
	}
	if resolved.fg != frame.ActiveTheme(ctx).Palette.Foreground {
		t.Fatalf("scoped foreground = %#v", resolved.fg)
	}
	if resolved.opacity != 0.25 {
		t.Fatalf("loading opacity = %v, want 0.25", resolved.opacity)
	}
}

func TestButtonWidgetPartsOverrideContentAndIndicator(t *testing.T) {
	ctx := newContext(nil)
	labelColor := color.NRGBA{R: 1, A: 0xff}
	iconColor := color.NRGBA{G: 2, A: 0xff}
	indicatorColor := color.NRGBA{B: 3, A: 0xff}
	declaration := flowstyle.Style{}.
		Part(flowstyle.PartLabel, flowstyle.Style{}.TextColor(solid(labelColor)).FontSize(18)).
		Part(flowstyle.PartIcon, flowstyle.Style{}.TextColor(solid(iconColor))).
		Part(flowstyle.PartIndicator, flowstyle.Style{}.TextColor(solid(indicatorColor)))

	label := Button("label", text.New("Label")).Style(declaration).staticStyle(ctx, flowstyle.StyleState{})
	if label.fg != labelColor || label.textSize != 18 || label.indicator != indicatorColor {
		t.Fatalf("label button parts = fg %#v size %v indicator %#v", label.fg, label.textSize, label.indicator)
	}
	icon := Button("icon", text.New("Icon")).IconOnly().Style(declaration).staticStyle(ctx, flowstyle.StyleState{})
	if icon.fg != iconColor {
		t.Fatalf("icon part = %#v, want %#v", icon.fg, iconColor)
	}
}

func TestButtonWidgetCanOverrideHeight(t *testing.T) {
	dims := Button("tall", text.New("Tall")).
		Style(flowstyle.Style{}.Height(52)).
		Layout(newContext(nil), testLayoutContext())
	if dims.Size.Y != 52 {
		t.Fatalf("styled button height = %d, want 52", dims.Size.Y)
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
	ctx := frame.New(nil, &theme, locale.LanguageAuto)
	small := Button("small", text.New("Small")).Size(ButtonSmall).staticStyle(ctx, flowstyle.StyleState{})
	large := Button("large", text.New("Large")).Size(ButtonLarge).staticStyle(ctx, flowstyle.StyleState{})
	icon := Button("icon", text.New("Icon")).IconOnly().staticStyle(ctx, flowstyle.StyleState{})

	if small.height != 36 {
		t.Fatalf("small height = %v, want 36", small.height)
	}
	if large.height != 44 || large.textSize != 16 {
		t.Fatalf("large style = %+v, want height 44 text 16", large)
	}
	if icon.inset.Left != 0 || icon.inset.Right != 0 {
		t.Fatalf("icon inset = %+v, want horizontal zero", icon.inset)
	}
	styled := Button("shape", text.New("Shape")).staticStyle(ctx, flowstyle.StyleState{})
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
	ctx := frame.New(nil, &theme, locale.LanguageAuto)
	style := button.staticStyle(ctx, flowstyle.StyleState{})
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

func TestButtonScaleTransitionUsesPressAndReleaseDurations(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	theme := DefaultTheme()
	ctx := frame.New(nil, &theme, locale.LanguageAuto)
	button := Button("scale", text.New("Scale"))
	target := buttonPressedScale(&theme, ButtonMedium)

	gtx.Now = start
	button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{})
	button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{Pressed: true})
	gtx.Now = start.Add(buttonPressInDuration / 2)
	pressed := button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{Pressed: true}).scaleX
	if pressed <= target || pressed >= 1 {
		t.Fatalf("press scale = %v, want between %v and 1", pressed, target)
	}

	gtx.Now = start.Add(buttonPressInDuration)
	button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{Pressed: true})
	button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{})
	gtx.Now = start.Add(buttonPressInDuration + buttonPressOutDuration/2)
	released := button.resolveStyle(ctx, gtx, "scale", flowstyle.StyleState{}).scaleX
	if released <= target || released >= 1 {
		t.Fatalf("release scale = %v, want between %v and 1", released, target)
	}
}

func TestButtonBackgroundTransition(t *testing.T) {
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	to := color.NRGBA{R: 110, G: 120, B: 130, A: 255}
	button := Button("transition", text.New("Transition")).Style(
		flowstyle.Style{}.
			Background(solid(from)).
			When(flowstyle.Hovered, flowstyle.Style{}.Background(solid(to))),
	)
	ctx := newContext(nil)

	gtx := testLayoutContext()
	gtx.Now = start
	if got := button.resolveStyle(ctx, gtx, "transition", flowstyle.StyleState{}).bg; got != from {
		t.Fatalf("initial background = %v, want %v", got, from)
	}

	gtx.Now = start
	if got := button.resolveStyle(ctx, gtx, "transition", flowstyle.StyleState{Hovered: true}).bg; got != from {
		t.Fatalf("transition start = %v, want %v", got, from)
	}

	gtx.Now = start.Add(buttonColorDuration / 2)
	mid := button.resolveStyle(ctx, gtx, "transition", flowstyle.StyleState{Hovered: true}).bg
	if mid == from || mid == to {
		t.Fatalf("transition midpoint = %v, want between %v and %v", mid, from, to)
	}

	gtx.Now = start.Add(buttonColorDuration)
	if got := button.resolveStyle(ctx, gtx, "transition", flowstyle.StyleState{Hovered: true}).bg; got != to {
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

func TestDisabledButtonAnimationInvalidates(t *testing.T) {
	var router input.Router
	var ops op.Ops
	now := time.Unix(1, 0)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	ctx := newContext(nil)
	Button("save", text.New("Save")).resolveStyle(ctx, gtx, "save", flowstyle.StyleState{})

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
