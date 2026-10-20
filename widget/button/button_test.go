package button

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotClickable = "clickable"

func newContext(_ any) *core.Context {
	activeTheme := *testtheme.Behavior()
	return core.New(nil, &activeTheme, locale.LanguageAuto)
}

func resolvedBackground(value buttonStyle) color.NRGBA {
	if value.root.Paint == nil {
		return color.NRGBA{}
	}
	background, ok := value.root.Paint.Background.(style.SolidColor)
	if ok {
		return background.Color
	}
	if pointer, pointerOK := value.root.Paint.Background.(*style.SolidColor); pointerOK && pointer != nil {
		return pointer.Color
	}
	return color.NRGBA{}
}

func resolvedScaleX(value buttonStyle) float32 {
	if value.root.Trans == nil || value.root.Trans.ScaleX == nil {
		return 1
	}
	return *value.root.Trans.ScaleX
}

// TestButtonStyledWrapperAddsChrome pins the provider-boundary styling path:
// a caller adds instance chrome through Styled while Button remains behavior-only.
func TestButtonStyledWrapperAddsChrome(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)

	defaultStyle := New("").Content(nil).staticStyle(ctx, style.State{})
	fluentStyle := (New("").Content(nil)).Style(style.Style{}.
		BorderColor(style.Solid(activeTheme.Palette.Border)).
		BorderWidth(1)).staticStyle(ctx, style.State{})
	if defaultStyle.root.Paint == nil || defaultStyle.root.Paint.Border != nil {
		t.Fatal("default button unexpectedly has external style border")
	}
	if fluentStyle.root.Paint == nil || fluentStyle.root.Paint.Border == nil {
		t.Fatal("fluent button style did not add border chrome")
	}
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *core.Context, key, slot string, value *T) {
	core.UseStateWith(ctx, key, slot, func() *T { return value })
}

func TestButtonOnClick(t *testing.T) {
	clicked := false
	b := New("").Content(text.New("Add").Size(20)).Key("add").
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
	b := New("Save").Disabled(true)

	if !b.disabled {
		t.Fatal("button was not disabled")
	}
}

func TestWidgetCascadesScopeBeforeInstance(t *testing.T) {
	ctx := newContext(nil)
	scope := style.Style{}.
		PaddingX(12).
		Background(style.RGB(0xaaaaaa)).
		TextColor(style.RGB(0x123456))

	restore := core.PushStyle(ctx, scope)
	defer restore()
	instanceBackground := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	button := (New("Styled")).Style(style.Style{}.
		PaddingX(20).
		Background(solid(instanceBackground)).
		On(style.Loading, style.Style{}.Opacity(0.25)))

	resolved := button.staticStyle(ctx, style.State{Loading: true})
	if inset := buttonInset(resolved.root); inset.Left != 20 || inset.Right != 20 {
		t.Fatalf("instance padding = %+v, want horizontal 20", inset)
	}
	if got := resolvedBackground(resolved); got != instanceBackground {
		t.Fatalf("instance background = %#v, want %#v", got, instanceBackground)
	}
	foreground, _ := resolved.visual.ForegroundColor(resolved.content)
	if foreground != (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}) {
		t.Fatalf("scoped foreground = %#v", foreground)
	}
	if opacity := buttonOpacity(resolved.root); opacity != 0.25 {
		t.Fatalf("loading opacity = %v, want 0.25", opacity)
	}
}

func TestWidgetPartsOverrideContentAndIndicator(t *testing.T) {
	ctx := newContext(nil)
	labelColor := color.NRGBA{R: 1, A: 0xff}
	iconColor := color.NRGBA{G: 2, A: 0xff}
	indicatorColor := color.NRGBA{B: 3, A: 0xff}
	declaration := style.Style{}.
		Part(style.PartLabel, style.Style{}.TextColor(solid(labelColor)).FontSize(18)).
		Part(style.PartIcon, style.Style{}.TextColor(solid(iconColor))).
		Part(style.PartIndicator, style.Style{}.TextColor(solid(indicatorColor)))

	label := (New("Label")).Style(declaration).staticStyle(ctx, style.State{})
	labelForeground, _ := label.visual.ForegroundColor(label.content)
	labelIndicator, _ := label.visual.ForegroundColor(label.indicatorPart)
	if labelForeground != labelColor || label.content.Text == nil || label.content.Text.FontSize == nil || *label.content.Text.FontSize != 18 || labelIndicator != indicatorColor {
		t.Fatalf("label button parts = fg %#v size %v indicator %#v", labelForeground, label.content.Text, labelIndicator)
	}
	icon := (New("Icon").IconOnly()).Style(declaration).staticStyle(ctx, style.State{})
	iconForeground, _ := icon.visual.ForegroundColor(icon.content)
	if iconForeground != iconColor {
		t.Fatalf("icon part = %#v, want %#v", iconForeground, iconColor)
	}
}

func TestLoadingUsesResolvedIndicatorColor(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	indicatorColor := color.NRGBA{R: 0xee, G: 0x11, B: 0x77, A: 0xff}
	behavior := New("Loading").Loading(true)
	b := (behavior).Style(style.Style{}.Part(style.PartIndicator, style.Style{}.TextColor(solid(indicatorColor))))

	resolved := b.staticStyle(ctx, style.State{Loading: true})
	gotIndicator, _ := resolved.visual.ForegroundColor(resolved.indicatorPart)
	if gotIndicator != indicatorColor {
		t.Fatalf("resolved indicator = %#v, want %#v", gotIndicator, indicatorColor)
	}

	window, err := headless.NewWindow(64, 64)
	if err != nil {
		t.Skipf("headless rendering unavailable: %v", err)
	}
	defer window.Release()

	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(64, 64)},
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	}
	behavior.layoutContent(ctx, gtx, resolved)
	if err := window.Frame(&ops); err != nil {
		t.Fatalf("render loading indicator: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	if err := window.Screenshot(img); err != nil {
		t.Fatalf("capture loading indicator: %v", err)
	}
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			got := img.RGBAAt(x, y)
			if got.R > 0xb0 && got.G < 0x70 && got.B > 0x40 {
				return
			}
		}
	}
	t.Fatalf("loading indicator did not render resolved color %#v", indicatorColor)
}

func TestRippleUsesResolvedBackgroundColor(t *testing.T) {
	ctx := newContext(nil)
	waveColorValue := color.NRGBA{R: 0x12, G: 0x9a, B: 0x70, A: 0xff}
	b := (New("Ripple")).Style(style.Style{}.Background(solid(waveColorValue)))

	resolved := b.staticStyle(ctx, style.State{})
	if resolved.visual.RippleColor == nil {
		t.Fatal("button provider returned no ripple color resolver")
	}
	got, ok := resolved.visual.RippleColor(resolved.root)
	if !ok || got != waveColorValue {
		t.Fatalf("ripple color = %#v, enabled %v; want %#v", got, ok, waveColorValue)
	}
}

func TestWidgetCanOverrideHeight(t *testing.T) {
	dims := (New("Tall")).Style(style.Style{}.Height(52)).
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
	New("Save").Label("Save changes").Disabled(true).Layout(newContext(nil), gtx)
	router.Frame(&ops)
	node, ok := buttonSemanticNode(router.AppendSemantics(nil))
	if !ok || node.Desc.Label != "Save changes" || !node.Desc.Disabled {
		t.Fatalf("button semantics = %#v", node.Desc)
	}
}

func TestButtonTextAndCustomContentKeepAccessibleLabel(t *testing.T) {
	plain := New("Save all pending changes")
	plainWidth := plain.Layout(newContext(nil), testLayoutContext()).Size.X
	emptyWidth := New("").Content(nil).Layout(newContext(nil), testLayoutContext()).Size.X
	if plainWidth <= emptyWidth {
		t.Fatal("button text did not contribute to layout width")
	}
	custom := plain.Content(core.WidgetFunc(func(_ *core.Context, _ layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(16, 16)}
	}))
	for _, test := range []struct {
		name   string
		value  Widget
		label  string
		custom bool
	}{
		{name: "text", value: plain, label: "Save all pending changes"},
		{name: "custom", value: custom, label: "Save all pending changes", custom: true},
		{name: "override", value: custom.Label("Save document"), label: "Save document", custom: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := new(input.Router)
			gtx := testLayoutContext()
			gtx.Source = router.Source()
			dims := test.value.Layout(newContext(nil), gtx)
			router.Frame(gtx.Ops)
			node, ok := buttonSemanticNode(router.AppendSemantics(nil))
			if !ok || node.Desc.Label != test.label {
				t.Fatalf("button semantics = %#v, want label %q", node.Desc, test.label)
			}
			if test.custom && dims.Size.X >= plainWidth {
				t.Fatal("custom content did not replace the button text")
			}
		})
	}
}

func TestButtonFullWidth(t *testing.T) {
	dims := New("Save").FullWidth().Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("button width = %d, want 300", dims.Size.X)
	}
}

func TestButtonFullWidthIconOnly(t *testing.T) {
	dims := New("S").FullWidth().IconOnly().Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("button width = %d, want 300", dims.Size.X)
	}
}

func TestLoadingKeepsIntrinsicWidth(t *testing.T) {
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		for _, size := range []Size{Small, Medium, Large} {
			normalGtx := testLayoutContext()
			normalGtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			normal := New("Load saved value").
				Size(size).
				Layout(newContext(nil), normalGtx)
			loadingGtx := testLayoutContext()
			loadingGtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			loading := New("Load saved value").
				Size(size).
				Loading(true).
				Layout(newContext(nil), loadingGtx)
			if loading.Size != normal.Size {
				t.Fatalf("scale %.2f size %d loading dimensions = %v, want %v", scale, size, loading.Size, normal.Size)
			}
		}
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

func TestLoadingBlocksClick(t *testing.T) {
	clicked := false
	clickable := new(widget.Clickable)
	clickable.Click()
	ctx := newContext(nil)
	testSetComponentState(ctx, "save", stateSlotClickable, clickable)

	New("Save").
		Loading(true).
		OnClick(func() {
			clicked = true
		}).
		Layout(ctx, testLayoutContext())

	if clicked {
		t.Fatal("loading button handled a click")
	}
}

func TestButtonScaleTransitionUsesPressAndReleaseDurations(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	theme := *testtheme.Behavior()
	ctx := core.New(nil, &theme, locale.LanguageAuto)
	button := New("Scale")
	visual := button.declarationsForOverride(&theme, nil, nil).visual
	pressedVisual := visual.Size.Resolve(style.State{Pressed: true})
	if pressedVisual.Trans == nil || pressedVisual.Trans.ScaleX == nil {
		t.Fatal("provider size layer has no pressed scale")
	}
	target := *pressedVisual.Trans.ScaleX
	pressDuration := ruleTweenDuration(t, visual.Size, style.PropTransform)
	releaseDuration := transitionDuration(t, visual.Defaults, style.PropTransform)

	gtx.Now = start
	button.resolveStyle(ctx, gtx, "scale", style.State{})
	button.resolveStyle(ctx, gtx, "scale", style.State{Pressed: true})
	gtx.Now = start.Add(pressDuration / 2)
	pressedStyle := button.resolveStyle(ctx, gtx, "scale", style.State{Pressed: true})
	pressed := resolvedScaleX(pressedStyle)
	if pressed < target || pressed > 1 {
		t.Fatalf("press scale = %v, want within %v..1", pressed, target)
	}

	gtx.Now = start.Add(pressDuration)
	button.resolveStyle(ctx, gtx, "scale", style.State{Pressed: true})
	button.resolveStyle(ctx, gtx, "scale", style.State{})
	gtx.Now = start.Add(pressDuration + releaseDuration/2)
	released := resolvedScaleX(button.resolveStyle(ctx, gtx, "scale", style.State{}))
	if released < target || released > 1 {
		t.Fatalf("release scale = %v, want within %v..1", released, target)
	}
}

func TestButtonBackgroundTransition(t *testing.T) {
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	to := color.NRGBA{R: 110, G: 120, B: 130, A: 255}
	button := (New("Transition")).Style(style.Style{}.
		Background(solid(from)).
		On(style.Hovered, style.Style{}.Background(solid(to))))
	ctx := newContext(nil)
	activeTheme := *testtheme.Behavior()
	span := transitionDuration(t, New("").Content(nil).declarationsForOverride(&activeTheme, nil, nil).visual.Defaults, style.PropBackgroundColor)

	gtx := testLayoutContext()
	gtx.Now = start
	if got := resolvedBackground(button.resolveStyle(ctx, gtx, "transition", style.State{})); got != from {
		t.Fatalf("initial background = %v, want %v", got, from)
	}

	gtx.Now = start
	if got := resolvedBackground(button.resolveStyle(ctx, gtx, "transition", style.State{Hovered: true})); got != from {
		t.Fatalf("transition start = %v, want %v", got, from)
	}

	gtx.Now = start.Add(span / 2)
	mid := resolvedBackground(button.resolveStyle(ctx, gtx, "transition", style.State{Hovered: true}))
	if mid == from || mid == to {
		t.Fatalf("transition midpoint = %v, want between %v and %v", mid, from, to)
	}

	gtx.Now = start.Add(span)
	if got := resolvedBackground(button.resolveStyle(ctx, gtx, "transition", style.State{Hovered: true})); got != to {
		t.Fatalf("transition end = %v, want %v", got, to)
	}
}

func TestButtonTransparentBackgroundTransitionKeepsColor(t *testing.T) {
	to := color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff}

	fadeIn := animation.LerpColor(color.NRGBA{}, to, 0.5)
	if fadeIn.R != to.R || fadeIn.G != to.G || fadeIn.B != to.B {
		t.Fatalf("fade in color = %v, want rgb from %v", fadeIn, to)
	}
	if fadeIn.A == 0 || fadeIn.A == to.A {
		t.Fatalf("fade in alpha = %d, want between 0 and %d", fadeIn.A, to.A)
	}

	fadeOut := animation.LerpColor(to, color.NRGBA{}, 0.5)
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
	New("Save").Key("save").resolveStyle(ctx, gtx, "save", style.State{})

	New("Save").Key("save").Disabled(true).Layout(ctx, gtx)
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

	New("Save").Loading(true).Layout(newContext(nil), gtx)
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
	themeValue := *testtheme.Behavior()
	themeValue.Motion.Enabled = false
	ctx := core.New(nil, &themeValue, locale.LanguageAuto)

	New("Save").Loading(true).Layout(ctx, gtx)
	router.Frame(&ops)

	if _, wake := router.WakeupTime(); wake {
		t.Fatal("loading button requested redraw with motion disabled")
	}
}

func TestButtonPassesDisabledContext(t *testing.T) {
	probe := &enabledProbeWidget{}

	New("").Content(probe).Disabled(true).Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("button child was laid out with enabled context")
	}
}

func TestButtonPassesForegroundColorToCustomContent(t *testing.T) {
	probe := &foregroundProbeWidget{}
	ctx := newContext(nil)
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}

	(New("").Content(probe)).Style(style.Style{}.TextColor(solid(want))).
		Layout(ctx, testLayoutContext())

	if probe.foreground != want {
		t.Fatalf("child foreground = %#v, want %#v", probe.foreground, want)
	}
}

type enabledProbeWidget struct {
	enabled bool
}

func (w *enabledProbeWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

type foregroundProbeWidget struct {
	foreground color.NRGBA
}

func (w *foregroundProbeWidget) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
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

func solid(value color.NRGBA) style.SolidColor {
	return style.SolidColor{Color: value}
}

func transitionDuration(t *testing.T, declaration style.Style, property style.PropertyID) time.Duration {
	t.Helper()
	for _, transition := range declaration.Resolve(style.State{}).Transitions {
		if transition.Property == property {
			return transition.Duration
		}
	}
	t.Fatalf("missing %v transition", property)
	return 0
}

func ruleTweenDuration(t *testing.T, declaration style.Style, property style.PropertyID) time.Duration {
	t.Helper()
	for _, rule := range declaration.RuleTable() {
		if rule.Tween != nil && rule.Tween.Property == property {
			return rule.Tween.Duration
		}
	}
	t.Fatalf("missing %v rule tween", property)
	return 0
}

func layoutButtonFrame(ctx *core.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	New("Save").Key("save").Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
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

func TestBooleanAndStaticConfigMethods(t *testing.T) {
	button := New("Save").
		Disabled(true).
		Disabled(false).
		FullWidth().
		IconOnly()

	if button.disabled {
		t.Fatal("Disabled(false) should clear the disabled state")
	}
	if !button.fullWidth || !button.iconOnly {
		t.Fatal("FullWidth and IconOnly should enable their respective flags")
	}
}

func TestButtonStyleComposesAtProviderBoundary(t *testing.T) {
	value := New("Save").Style(style.New().Width(120))
	// The declaration lives on the widget, so later chain calls keep it.
	value = value.OnClick(func() {}).Disabled(false)
	if dims := value.Layout(newContext(nil), testLayoutContext()); dims.Size.X != 120 {
		t.Fatalf("styled button width = %d, want 120", dims.Size.X)
	}
}
