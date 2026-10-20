package input

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
)

const (
	stateSlotClickable = "clickable"
	stateSlotEditor    = "editor"
)

func newContext(_ any) *core.Context {
	activeTheme := *testtheme.Behavior()
	return core.New(nil, &activeTheme, locale.LanguageAuto)
}

func resolveInputTestStyle(activeTheme *theme.Theme, declaration style.Style, state style.State) style.ResolvedStyle {
	ctx := core.New(nil, activeTheme, locale.LanguageAuto)
	return computed.ResolveStatic(ctx, state, declaration, style.Style{}, style.Style{}, style.Style{})
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
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

	New("Ada").Key("name").Hint("Name").Layout(ctx, layout.Context{
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
	if !editor.Submit {
		t.Fatal("single-line input did not consume carriage returns")
	}
}

func TestInputDisabled(t *testing.T) {
	i := New("Ada").Key("name").Disabled(true)

	if !i.disabled {
		t.Fatal("input was not disabled")
	}
}

func TestInputDefaultLayout(t *testing.T) {
	dims := New("").Key("name").Hint("Name").Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y != 36 {
		t.Fatalf("input height = %d, want 36", dims.Size.Y)
	}
}

func TestInputTypographyComesFromResolvedStyle(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	visual := inputVisualDeclarations(
		&activeTheme,
		Primary,
		false,
		theme.InputModeSingleLine,
		Medium,
		nil,
	)
	declaration := visual.Defaults
	custom := style.Style{}.
		FontSize(18).
		LineHeight(28).
		TextColor(style.RGB(0x123456))
	resolved := computed.ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, custom)
	providerStyle := visual.EditorStyle(resolved, style.ResolvedStyle{}, style.ResolvedStyle{})
	textSize, lineHeight := providerStyle.TextSize, providerStyle.LineHeight
	if textSize != 18 || lineHeight != 28 {
		t.Fatalf("resolved typography = %v/%v, want 18/28", textSize, lineHeight)
	}
	if providerStyle.Foreground != (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}) {
		t.Fatalf("provider foreground = %#v", providerStyle.Foreground)
	}
}

func TestInputEditorAdapterCarriesResolvedTypographyAndColors(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.
		FontSize(18).
		LineHeight(28).
		LineHeightScale(1.25).
		Typeface("monospace").
		FontStyle(font.Italic).
		FontWeight(int(font.Bold)).
		TextColor(style.RGB(0x123456)).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.RGB(0x654321))).
		Part(style.PartSelection, style.Style{}.Background(style.RGB(0xaabbcc)))
	root := computed.ResolveStatic(ctx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	placeholder := computed.ResolvePartStatic(ctx, style.PartPlaceholder, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	selection := computed.ResolvePartStatic(ctx, style.PartSelection, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	visual := inputVisualDeclarations(&activeTheme, Primary, false, theme.InputModeSingleLine, Medium, nil)
	resolved := visual.EditorStyle(root, placeholder, selection)

	if resolved.TextSize != 18 || resolved.LineHeight != 28 || !resolved.HasLineHeightScale || resolved.LineHeightScale != 1.25 {
		t.Fatalf("resolved editor metrics = %#v", resolved)
	}
	if !resolved.HasTypeface || resolved.Typeface != "monospace" || !resolved.HasFontStyle || resolved.FontStyle != font.Italic || !resolved.HasFontWeight || resolved.FontWeight != font.Bold {
		t.Fatalf("resolved editor font = %#v", resolved)
	}
	if resolved.Foreground != (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}) ||
		resolved.Placeholder != (color.NRGBA{R: 0x65, G: 0x43, B: 0x21, A: 0xff}) ||
		resolved.Selection != (color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff}) {
		t.Fatalf("resolved editor colors = %#v", resolved)
	}

	target := material.Editor(core.ActiveMaterial(ctx), &widget.Editor{}, "")
	applyInputEditorStyle(&target, resolved)
	if target.TextSize != 18 || target.LineHeight != 28 || target.LineHeightScale != 1.25 ||
		target.Font != (font.Font{Typeface: "monospace", Style: font.Italic, Weight: font.Bold}) {
		t.Fatalf("material editor typography = %#v", target)
	}
	if target.Color != resolved.Foreground || target.HintColor != resolved.Placeholder || target.SelectionColor != resolved.Selection {
		t.Fatalf("material editor colors = %#v", target)
	}
}

func TestInputFullWidth(t *testing.T) {
	dims := New("").Key("name").FullWidth().Layout(newContext(nil), testLayoutContext())

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

	activeTheme := *testtheme.Behavior()
	visual := inputVisualDeclarations(
		&activeTheme,
		Primary,
		true,
		theme.InputModeSingleLine,
		Medium,
		nil,
	)
	declaration := visual.Defaults
	resolved := resolveInputTestStyle(&activeTheme, declaration, style.State{})
	dims := New("").Key("name").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(inputState), resolved, true, child)

	if got.Min != image.Pt(276, 0) {
		t.Fatalf("inner minimum = %v, want (276,0)", got.Min)
	}
	if dims.Size != image.Pt(300, 36) {
		t.Fatalf("input size = %v, want (300,36)", dims.Size)
	}
}

func TestTypeConfig(t *testing.T) {
	for _, test := range []struct {
		name      string
		inputType Type
		mask      rune
		hint      key.InputHint
		filter    string
	}{
		{name: "text", inputType: Text, hint: key.HintText},
		{name: "email", inputType: Email, hint: key.HintEmail},
		{name: "number", inputType: Number, hint: key.HintNumeric, filter: "0123456789+-.eE"},
		{name: "password", inputType: Password, mask: '\u2022', hint: key.HintPassword},
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
	New("secret").Key("password").
		Type(Password).
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

func TestNumberFiltersTextAndTypeSwitchResetsEditor(t *testing.T) {
	ctx := newContext(nil)
	core.BeginFrame(ctx)
	New("").Key("value").Type(Number).Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)
	editor := testComponentState[widget.Editor](ctx, "value", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	editor.Insert("1a2")
	if got := editor.Text(); got != "12" {
		t.Fatalf("number input text = %q, want 12", got)
	}

	core.BeginFrame(ctx)
	New("12").Key("value").Type(Text).Layout(ctx, testLayoutContext())
	core.EndFrame(ctx)
	if editor.Filter != "" || editor.Mask != 0 || editor.InputHint != key.HintText {
		t.Fatalf("text configuration retained previous type: mask %q hint %v filter %q", editor.Mask, editor.InputHint, editor.Filter)
	}
}

func TestInputParentDisabledClearsHover(t *testing.T) {
	ctx := newContext(nil)
	state := &inputState{}
	state.Hovered = true
	core.UseStateWith(ctx, "name", stateSlotInput, func() *inputState { return state })
	New("").Key("name").Placeholder("Name").Layout(ctx, testLayoutContext().Disabled())
	if state.Hovered {
		t.Fatal("parent-disabled input kept its hover state")
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
	(New("").Key("paint").FullWidth()).Style(style.Style{}.Background(style.SolidColor{Color: want})).Layout(
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

func TestInputSemanticsExposeLabelAndDisabledState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	New("").Key("email").Placeholder("Email").Label("Email address").Disabled(true).Layout(ctx, gtx)
	core.EndFrame(ctx)
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

	widget := New("").Key("name").
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

func focusedInput(t *testing.T) (*core.Context, *input.Router, *widget.Editor) {
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

func focusedInputWithButton(t *testing.T) (*core.Context, *input.Router, *widget.Editor) {
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

func layoutInputFrame(ctx *core.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	New("").Key("name").Hint("Name").Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func TestBooleanAndStaticConfigMethods(t *testing.T) {
	field := New("Ada").Key("name").
		Disabled(true).
		Disabled(false).
		Invalid(true).
		Invalid(false).
		ReadOnly(true).
		ReadOnly(false).
		FullWidth()

	if field.disabled || field.invalid || field.readOnly {
		t.Fatal("boolean methods should honor the final value in the chain")
	}
	if !field.fullWidth {
		t.Fatal("FullWidth should enable the full-width flag")
	}
}

func TestUncontrolledPreservesInitialValueWithoutControl(t *testing.T) {
	field := New("").Key("input").Uncontrolled()
	if field.hasValue || !field.hasDefault || field.defaultValue != "" {
		t.Fatalf("Uncontrolled() = hasValue=%v hasDefault=%v defaultValue=%q", field.hasValue, field.hasDefault, field.defaultValue)
	}
}

func TestValueAfterDefaultValueWinsControl(t *testing.T) {
	field := New("").Key("input").DefaultValue("seed").Value("live")
	if !field.hasValue || field.value != "live" {
		t.Fatalf("Value after DefaultValue = hasValue=%v value=%q", field.hasValue, field.value)
	}
}

func TestDefaultValueAfterValueSwitchesToUncontrolled(t *testing.T) {
	field := New("old").Key("input").DefaultValue("seed")
	if field.hasValue || !field.hasDefault || field.defaultValue != "seed" {
		t.Fatalf("DefaultValue after constructor value = hasValue=%v hasDefault=%v defaultValue=%q", field.hasValue, field.hasDefault, field.defaultValue)
	}
}

func layoutInputAndButtonFrame(ctx *core.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)

	inputGtx := gtx
	inputGtx.Constraints = layout.Constraints{Max: image.Pt(200, 40)}
	New("").Key("name").Hint("Name").Layout(ctx, inputGtx)

	buttonGtx := gtx
	buttonGtx.Constraints = layout.Constraints{Max: image.Pt(80, 40)}
	stack := op.Offset(image.Pt(220, 0)).Push(gtx.Ops)
	button.New("Add").Key("add").Layout(ctx, buttonGtx)
	stack.Pop()

	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
