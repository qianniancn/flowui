package input

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
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	stateSlotClickable = "clickable"
	stateSlotEditor    = "editor"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
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

func TestInputOptions(t *testing.T) {
	base := Input("name", "Ada")
	i := base.
		Placeholder("Name").
		Invalid(true).
		Variant(InputSecondary).
		FullWidth().
		Type(InputEmail).
		ReadOnly(true).
		MaxLength(24).
		Label("Name")

	if !i.invalid {
		t.Fatal("input was not invalid")
	}
	if i.variant != InputSecondary {
		t.Fatal("input variant was not set")
	}
	if !i.fullWidth {
		t.Fatal("input was not full width")
	}
	if i.hint != "Name" || i.inputType != InputEmail || !i.readOnly || i.maxLength != 24 || i.label != "Name" {
		t.Fatalf("configured input = %#v", i)
	}
	if base.hint != "" || base.inputType != InputText || base.readOnly || base.maxLength != 0 || base.label != "" {
		t.Fatalf("base input was mutated: %#v", base)
	}
	if got := base.MaxLength(-1).maxLength; got != 0 {
		t.Fatalf("negative max length = %d, want 0", got)
	}
}

func TestInputDefaultLayout(t *testing.T) {
	dims := Input("name", "").Hint("Name").Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y != 36 {
		t.Fatalf("input height = %d, want 36", dims.Size.Y)
	}
}

func TestInputHeroUIDefaultTheme(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Input
	if tokens.Height != 36 || tokens.Radius != 12 || tokens.PaddingX != 12 || tokens.TextSize != 14 || tokens.LineHeight != 20 {
		t.Fatalf("input geometry = %#v", tokens)
	}
	if tokens.FocusRingWidth != 2 || tokens.InvalidOutlineWidth != 1 || tokens.ShadowOpacity != 1 {
		t.Fatalf("input state tokens = %#v", tokens)
	}
	if dark := theme.DarkTheme().Components.Input.ShadowOpacity; dark != 0 {
		t.Fatalf("dark input shadow opacity = %v, want 0", dark)
	}
	if activeTheme.Palette.Background != (color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff}) {
		t.Fatalf("light background = %#v, want HeroUI neutral background", activeTheme.Palette.Background)
	}
	if activeTheme.Palette.Focus != activeTheme.Palette.Accent {
		t.Fatalf("focus = %#v, want accent %#v", activeTheme.Palette.Focus, activeTheme.Palette.Accent)
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
		return layout.Dimensions{Size: image.Pt(1, 1)}
	}

	Input("name", "").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(inputState), inputStyle{Opacity: 1}, true, child)

	if got.Min.X != 276 {
		t.Fatalf("inner min width = %d, want 276", got.Min.X)
	}
	if got.Min.Y != 0 {
		t.Fatalf("inner min height = %d, want 0", got.Min.Y)
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
	primary := inputStyleFor(&activeTheme, InputPrimary, false, false, false, false)
	secondary := inputStyleFor(&activeTheme, InputSecondary, false, false, false, false)
	secondaryHovered := inputStyleFor(&activeTheme, InputSecondary, true, false, false, false)
	hovered := inputStyleFor(&activeTheme, InputPrimary, true, false, false, false)
	focused := inputStyleFor(&activeTheme, InputPrimary, true, true, false, false)
	invalid := inputStyleFor(&activeTheme, InputPrimary, false, false, false, true)
	focusedInvalid := inputStyleFor(&activeTheme, InputPrimary, false, true, false, true)
	disabled := inputStyleFor(&activeTheme, InputPrimary, false, false, true, false)

	if primary.Background != activeTheme.Palette.Surface || primary.ShadowOpacity != 1 {
		t.Fatalf("primary style = %#v", primary)
	}
	if secondary.Background != activeTheme.Palette.SurfacePressed || secondary.ShadowOpacity != 0 {
		t.Fatalf("secondary style = %#v", secondary)
	}
	if secondaryHovered.Background != activeTheme.Palette.Border {
		t.Fatalf("secondary hover background = %#v", secondaryHovered.Background)
	}
	wantPrimaryHover := color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff}
	if hovered.Background != wantPrimaryHover {
		t.Fatalf("hover background = %#v, want %#v", hovered.Background, wantPrimaryHover)
	}
	if focused.Ring != activeTheme.Palette.Focus || focused.RingWidth != 2 || focused.Background != activeTheme.Palette.Surface {
		t.Fatalf("focused style = %#v", focused)
	}
	if invalid.Ring != activeTheme.Palette.Danger || invalid.RingWidth != 1 {
		t.Fatalf("invalid style = %#v", invalid)
	}
	if focusedInvalid.Ring != activeTheme.Palette.Danger || focusedInvalid.RingWidth != 2 {
		t.Fatalf("focused invalid style = %#v", focusedInvalid)
	}
	if disabled.Opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled opacity = %v", disabled.Opacity)
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

func TestInputRingWidthAnimation(t *testing.T) {
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start
	state := new(inputState)
	if got := state.RingWidth(gtx, 0); got != 0 {
		t.Fatalf("initial ring width = %v", got)
	}
	if got := state.RingWidth(gtx, 2); got != 0 {
		t.Fatalf("transition start width = %v", got)
	}
	gtx.Now = start.Add(inputTransitionDuration / 2)
	if got := state.RingWidth(gtx, 2); got <= 0 || got >= 2 {
		t.Fatalf("transition midpoint width = %v", got)
	}
	gtx.Now = start.Add(inputTransitionDuration)
	if got := state.RingWidth(gtx, 2); got != 2 {
		t.Fatalf("transition end width = %v", got)
	}
}

func TestInputUsesThreeLayerHeroUIShadow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	layers := render.ThemeShadow(activeTheme.Shadows.Control, activeTheme.Palette.Shadow, 1).EffectiveLayers()
	if len(layers) != 3 {
		t.Fatalf("shadow layer count = %d, want 3", len(layers))
	}
	if layers[0].OffsetY != 0 || layers[0].Blur != 1 || layers[0].Color.A != 0x0f {
		t.Fatalf("first shadow layer = %#v", layers[0])
	}
	if layers[1].OffsetY != 1 || layers[1].Blur != 2 || layers[1].Color.A != 0x0f {
		t.Fatalf("second shadow layer = %#v", layers[1])
	}
	if layers[2].OffsetY != 2 || layers[2].Blur != 4 || layers[2].Color.A != 0x0a {
		t.Fatalf("third shadow layer = %#v", layers[2])
	}
	if got := render.ThemeShadow(activeTheme.Shadows.Control, color.NRGBA{A: 0xff}, .5).EffectiveLayers()[0].Color.A; got != 18 {
		t.Fatalf("half opacity shadow alpha = %d, want 18", got)
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

	Input("name", "").
		OnChange(func(text string) {
			got = append(got, "change:"+text)
		}).
		OnSubmit(func(text string) {
			got = append(got, "submit:"+text)
		}).
		dispatchEvents(editor, inputEvents{
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
