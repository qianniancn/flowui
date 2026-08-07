package input

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/components/button"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/render"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

const (
	stateSlotClickable = "clickable"
	stateSlotEditor    = "editor"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func resolveInputTestStyle(activeTheme *theme.Theme, declaration flowstyle.Style, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	ctx := frame.New(nil, activeTheme, locale.LanguageAuto)
	return styleruntime.ResolveStatic(ctx, state, declaration, flowstyle.Style{}, flowstyle.Style{}, flowstyle.Style{})
}

func resolveInputTestPart(activeTheme *theme.Theme, declaration flowstyle.Style, part flowstyle.Part, state flowstyle.StyleState) flowstyle.ResolvedStyle {
	ctx := frame.New(nil, activeTheme, locale.LanguageAuto)
	return styleruntime.ResolvePartStatic(ctx, part, state, declaration, flowstyle.Style{}, flowstyle.Style{}, flowstyle.Style{})
}

func resolvedBackground(value flowstyle.ResolvedStyle) color.NRGBA {
	if value.Paint == nil {
		return color.NRGBA{}
	}
	brush, _ := styleruntime.Brush(value.Paint.Background)
	return brush.ColorAt(.5)
}

func Button(key string, child frame.Widget) button.ButtonWidget {
	return button.Button(key, child)
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
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

func TestInputSyncsValue(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Input("name", "Ada").Hint("Name").Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Ops:         &ops,
	})

	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	if got := editor.Text(); got != "Ada" {
		t.Fatalf("editor text = %q, want Ada", got)
	}
	if editor.Submit {
		t.Fatal("submit enabled without OnSubmit")
	}
}

func TestInputEnablesSubmit(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Input("name", "Ada").OnSubmit(func(string) {}).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Ops:         &ops,
	})

	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	if !editor.Submit {
		t.Fatal("submit was not enabled")
	}
}

func TestInputDisabled(t *testing.T) {
	i := Input("name", "Ada").Disabled(true)

	if !i.disabled {
		t.Fatal("input was not disabled")
	}
}

func TestInputDefaultLayout(t *testing.T) {
	dims := Input("name", "").Hint("Name").Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y != 36 {
		t.Fatalf("input height = %d, want 36", dims.Size.Y)
	}
}

func TestInputFullWidth(t *testing.T) {
	dims := Input("name", "").FullWidth().Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("input width = %d, want 300", dims.Size.X)
	}
}

func TestInputFrameKeepsInnerWidth(t *testing.T) {
	var got layout.Constraints
	child := func(gtx layout.Context) layout.Dimensions {
		got = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(1, 20))}
	}

	activeTheme := theme.DefaultTheme()
	resolved := resolveInputTestStyle(&activeTheme, inputDefaultDeclaration(&activeTheme, InputPrimary, true), flowstyle.StyleState{})
	dims := Input("name", "").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(inputState), resolved, true, child)

	if got.Min != image.Pt(276, 0) {
		t.Fatalf("inner minimum = %v, want (276,0)", got.Min)
	}
	if dims.Size != image.Pt(300, 36) {
		t.Fatalf("input size = %v, want (300,36)", dims.Size)
	}
}

func TestInputTypeConfig(t *testing.T) {
	for _, test := range []struct {
		name      string
		inputType InputType
		mask      rune
		hint      key.InputHint
		filter    string
	}{
		{name: "text", inputType: InputText, hint: key.HintText},
		{name: "email", inputType: InputEmail, hint: key.HintEmail},
		{name: "number", inputType: InputNumber, hint: key.HintNumeric, filter: "0123456789+-.eE"},
		{name: "password", inputType: InputPassword, mask: '\u2022', hint: key.HintPassword},
	} {
		t.Run(test.name, func(t *testing.T) {
			mask, hint, filter := inputTypeConfig(test.inputType)
			if mask != test.mask || hint != test.hint || filter != test.filter {
				t.Fatalf("config = mask %q hint %v filter %q, want %q, %v, %q", mask, hint, filter, test.mask, test.hint, test.filter)
			}
		})
	}
}

func TestInputAppliesEditorConfiguration(t *testing.T) {
	ctx := newContext(nil)
	Input("password", "secret").
		Type(InputPassword).
		ReadOnly(true).
		MaxLength(12).
		Layout(ctx, testLayoutContext())
	editor := testComponentState[widget.Editor](ctx, "password", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	if editor.Mask != '\u2022' || editor.InputHint != key.HintPassword || editor.Filter != "" || !editor.ReadOnly || editor.MaxLen != 12 {
		t.Fatalf("editor configuration = mask %q hint %v filter %q readonly %v max %d", editor.Mask, editor.InputHint, editor.Filter, editor.ReadOnly, editor.MaxLen)
	}
}

func TestInputNumberFiltersTextAndTypeSwitchResetsEditor(t *testing.T) {
	ctx := newContext(nil)
	frame.BeginFrame(ctx)
	Input("value", "").Type(InputNumber).Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	editor := testComponentState[widget.Editor](ctx, "value", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	editor.Insert("1a2")
	if got := editor.Text(); got != "12" {
		t.Fatalf("number input text = %q, want 12", got)
	}

	frame.BeginFrame(ctx)
	Input("value", "12").Type(InputText).Layout(ctx, testLayoutContext())
	frame.EndFrame(ctx)
	if editor.Filter != "" || editor.Mask != 0 || editor.InputHint != key.HintText {
		t.Fatalf("text configuration retained previous type: mask %q hint %v filter %q", editor.Mask, editor.InputHint, editor.Filter)
	}
}

func TestInputStylesMatchHeroUIStates(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	primaryDeclaration := inputDefaultDeclaration(&activeTheme, InputPrimary, false)
	secondaryDeclaration := inputDefaultDeclaration(&activeTheme, InputSecondary, false)
	primary := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{})
	secondary := resolveInputTestStyle(&activeTheme, secondaryDeclaration, flowstyle.StyleState{})
	secondaryHovered := resolveInputTestStyle(&activeTheme, secondaryDeclaration, flowstyle.StyleState{Hovered: true})
	hovered := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Hovered: true})
	focused := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Focused: true})
	invalid := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Invalid: true})
	focusedInvalid := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Focused: true, Invalid: true})
	disabled := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Disabled: true})

	if resolvedBackground(primary) != activeTheme.Palette.FieldBackgroundColor() || primary.Paint == nil || len(primary.Paint.Shadows) != 3 {
		t.Fatalf("primary style = %#v", primary)
	}
	if resolvedBackground(secondary) != activeTheme.Palette.DefaultColor() || secondary.Paint == nil || len(secondary.Paint.Shadows) != 0 {
		t.Fatalf("secondary style = %#v", secondary)
	}
	if resolvedBackground(secondaryHovered) != activeTheme.Palette.DefaultHoverColor() {
		t.Fatalf("secondary hover background = %#v", resolvedBackground(secondaryHovered))
	}
	wantPrimaryHover := color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff}
	if resolvedBackground(hovered) != wantPrimaryHover {
		t.Fatalf("hover background = %#v, want %#v", resolvedBackground(hovered), wantPrimaryHover)
	}
	if focused.Paint == nil || focused.Paint.Outline == nil || focused.Paint.Outline.Color != (flowstyle.SolidColor{Color: activeTheme.Palette.Focus}) || focused.Paint.Outline.Width != 2 || resolvedBackground(focused) != activeTheme.Palette.FieldFocusColor() {
		t.Fatalf("focused style = %#v", focused)
	}
	if invalid.Paint == nil || invalid.Paint.Outline == nil || invalid.Paint.Outline.Color != (flowstyle.SolidColor{Color: activeTheme.Palette.Danger}) || invalid.Paint.Outline.Width != 1 {
		t.Fatalf("invalid style = %#v", invalid)
	}
	if focusedInvalid.Paint == nil || focusedInvalid.Paint.Outline == nil || focusedInvalid.Paint.Outline.Color != (flowstyle.SolidColor{Color: activeTheme.Palette.Danger}) || focusedInvalid.Paint.Outline.Width != 2 {
		t.Fatalf("focused invalid style = %#v", focusedInvalid)
	}
	if disabled.Paint == nil || disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled opacity = %#v", disabled.Paint)
	}
}

func TestInputParentDisabledClearsHover(t *testing.T) {
	ctx := newContext(nil)
	state := &inputState{}
	state.Hovered = true
	frame.UseStateWith(ctx, "name", stateSlotInput, func() *inputState { return state })
	Input("name", "").Placeholder("Name").Layout(ctx, testLayoutContext().Disabled())
	if state.Hovered {
		t.Fatal("parent-disabled input kept its hover state")
	}
}

func TestInputUsesEnhancedThreeLayerShadow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	resolved := resolveInputTestStyle(&activeTheme, inputDefaultDeclaration(&activeTheme, InputPrimary, false), flowstyle.StyleState{})
	layers := resolved.Paint.Shadows
	if len(layers) != 3 {
		t.Fatalf("shadow layer count = %d, want 3", len(layers))
	}
	first, _ := styleruntime.Color(layers[0].Color)
	second, _ := styleruntime.Color(layers[1].Color)
	third, _ := styleruntime.Color(layers[2].Color)
	if layers[0].OffsetY != 0 || layers[0].Blur != 1 || first != (color.NRGBA{A: 0x38}) {
		t.Fatalf("first shadow layer = %#v", layers[0])
	}
	if layers[1].OffsetY != 1 || layers[1].Blur != 2 || second != (color.NRGBA{A: 0x38}) {
		t.Fatalf("second shadow layer = %#v", layers[1])
	}
	if layers[2].OffsetY != 2 || layers[2].Blur != 4 || third != (color.NRGBA{A: 0x26}) {
		t.Fatalf("third shadow layer = %#v", layers[2])
	}
	if got := render.ThemeShadow(activeTheme.Shadows.Control, color.NRGBA{A: 0xff}, .5).EffectiveLayers()[0].Color.A; got != 18 {
		t.Fatalf("half opacity shadow alpha = %d, want 18", got)
	}
}

func TestInputCustomBackgroundIsRenderedByCommonStyle(t *testing.T) {
	window, err := headless.NewWindow(80, 36)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	want := color.NRGBA{R: 0xd1, G: 0x23, B: 0x45, A: 0xff}
	var router input.Router
	var ops op.Ops
	Input("paint", "").FullWidth().Style(flowstyle.Style{}.Background(flowstyle.SolidColor{Color: want})).Layout(
		newContext(nil),
		layout.Context{Constraints: layout.Exact(image.Pt(80, 36)), Source: router.Source(), Ops: &ops},
	)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 80, 36))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := color.NRGBAModel.Convert(pixels.At(40, 18)).(color.NRGBA); got != want {
		t.Fatalf("input center pixel = %#v, want %#v", got, want)
	}
}

func TestDarkInputDisablesFieldShadow(t *testing.T) {
	activeTheme := theme.DarkTheme()
	resolved := resolveInputTestStyle(&activeTheme, inputDefaultDeclaration(&activeTheme, InputPrimary, false), flowstyle.StyleState{})
	layers := resolved.Paint.Shadows
	if len(layers) != 0 {
		t.Fatalf("dark shadow layer count = %d, want 0", len(layers))
	}
}

func TestInputSemanticsExposeLabelAndDisabledState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	Input("email", "").Placeholder("Email").Label("Email address").Disabled(true).Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	tree := router.AppendSemantics(nil)
	node, ok := inputSemanticNode(tree)
	if !ok {
		t.Fatalf("semantic tree does not contain an editor: %#v", tree)
	}
	if node.Desc.Label != "Email address" || !node.Desc.Disabled {
		t.Fatalf("semantics = label %q disabled %v", node.Desc.Label, node.Desc.Disabled)
	}
}

func inputSemanticNode(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Editor {
			return node, true
		}
		if child, ok := inputSemanticNode(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func TestInputClearsFocusOnOutsidePress(t *testing.T) {
	ctx, router, editor := focusedInput(t)

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(250, 80),
	})
	layoutInputFrame(ctx, router)

	if router.Source().Focused(editor) {
		t.Fatal("input kept focus after outside press")
	}
}

func TestInputClearsFocusWhenButtonPressed(t *testing.T) {
	ctx, router, editor := focusedInputWithButton(t)
	button := testComponentState[widget.Clickable](ctx, "add", stateSlotClickable)
	if button == nil {
		t.Fatal("missing button state")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(250, 20),
	})
	layoutInputAndButtonFrame(ctx, router)

	if router.Source().Focused(editor) {
		t.Fatal("input kept focus after button press")
	}
	if !router.Source().Focused(button) {
		t.Fatal("button did not gain focus after press")
	}
}

func TestInputKeepsFocusOnInsidePress(t *testing.T) {
	ctx, router, editor := focusedInput(t)

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(20, 20),
	})
	layoutInputFrame(ctx, router)

	if !router.Source().Focused(editor) {
		t.Fatal("input lost focus after inside press")
	}
}

func TestInputTracksPointerHover(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutInputFrame(ctx, router)
	router.Queue(pointer.Event{
		Kind:      pointer.Move,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(20, 18),
	})
	layoutInputFrame(ctx, router)
	state := testComponentState[inputState](ctx, "name", stateSlotInput)
	if state == nil || !state.Hovered {
		t.Fatalf("input hover state = %#v, want hovered", state)
	}
}

func TestInputDispatchesChangeBeforeSubmit(t *testing.T) {
	editor := new(widget.Editor)
	editor.SetText("Ada")
	var got []string

	widget := Input("name", "").
		OnChange(func(text string) {
			got = append(got, "change:"+text)
		}).
		OnSubmit(func(text string) {
			got = append(got, "submit:"+text)
		})

	state := &inputState{}
	state.bind(widget)

	widget.dispatchEvents(state, editor, inputEvents{
		changed:    true,
		submitted:  true,
		submitText: "Ada",
	})

	want := []string{"change:Ada", "submit:Ada"}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}

func TestInputEventsTrackSubmit(t *testing.T) {
	var events inputEvents
	events.add(widget.SubmitEvent{Text: "Ada"})

	if !events.submitted {
		t.Fatal("submit event was not tracked")
	}
	if events.submitText != "Ada" {
		t.Fatalf("submit text = %q, want Ada", events.submitText)
	}
}

func focusedInput(t *testing.T) (*frame.Context, *input.Router, *widget.Editor) {
	t.Helper()
	ctx := newContext(nil)
	router := new(input.Router)

	layoutInputFrame(ctx, router)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	router.Source().Execute(key.FocusCmd{Tag: editor})
	layoutInputFrame(ctx, router)

	if !router.Source().Focused(editor) {
		t.Fatal("input did not gain focus")
	}
	return ctx, router, editor
}

func focusedInputWithButton(t *testing.T) (*frame.Context, *input.Router, *widget.Editor) {
	t.Helper()
	ctx := newContext(nil)
	router := new(input.Router)

	layoutInputAndButtonFrame(ctx, router)
	editor := testComponentState[widget.Editor](ctx, "name", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	router.Source().Execute(key.FocusCmd{Tag: editor})
	layoutInputAndButtonFrame(ctx, router)

	if !router.Source().Focused(editor) {
		t.Fatal("input did not gain focus")
	}
	return ctx, router, editor
}

func layoutInputFrame(ctx *frame.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	Input("name", "").Hint("Name").Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutInputAndButtonFrame(ctx *frame.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)

	inputGtx := gtx
	inputGtx.Constraints = layout.Constraints{Max: image.Pt(200, 40)}
	Input("name", "").Hint("Name").Layout(ctx, inputGtx)

	buttonGtx := gtx
	buttonGtx.Constraints = layout.Constraints{Max: image.Pt(80, 40)}
	stack := op.Offset(image.Pt(220, 0)).Push(gtx.Ops)
	Button("add", text.New("Add")).Layout(ctx, buttonGtx)
	stack.Pop()

	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
