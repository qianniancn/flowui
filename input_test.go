package flowui

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

func TestInputSyncsValue(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Input("name", "Ada").Hint("Name").Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Ops:         &ops,
	})

	editor := ctx.editors["name"]
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

	editor := ctx.editors["name"]
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
	i := Input("name", "Ada").
		Invalid(true).
		Variant(InputSecondary).
		FullWidth()

	if !i.invalid {
		t.Fatal("input was not invalid")
	}
	if i.variant != InputSecondary {
		t.Fatal("input variant was not set")
	}
	if !i.fullWidth {
		t.Fatal("input was not full width")
	}
}

func TestInputDefaultLayout(t *testing.T) {
	dims := Input("name", "").Hint("Name").Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y != 40 {
		t.Fatalf("input height = %d, want 40", dims.Size.Y)
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

	Input("name", "").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(inputState), inputStyle{}, child)

	if got.Min.X != 276 {
		t.Fatalf("inner min width = %d, want 276", got.Min.X)
	}
	if got.Min.Y != 0 {
		t.Fatalf("inner min height = %d, want 0", got.Min.Y)
	}
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
	button := ctx.clickables["add"]
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

func TestInputStylePrimary(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, false, false, false, false)

	if style.bg != (color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}) {
		t.Fatalf("primary bg = %v, want white", style.bg)
	}
	if style.border.A != 0 || style.borderWidth != 0 {
		t.Fatalf("primary border = %v at %v, want none", style.border, style.borderWidth)
	}
	if style.shadowOpacity != 1 {
		t.Fatalf("primary shadow opacity = %v, want 1", style.shadowOpacity)
	}
}

func TestInputStyleSecondary(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputSecondary, false, false, false, false)

	if style.bg != (color.NRGBA{R: 0xf4, G: 0xf4, B: 0xf5, A: 0xff}) {
		t.Fatalf("secondary bg = %v, want default bg", style.bg)
	}
	if style.shadowOpacity != 0 {
		t.Fatalf("secondary shadow opacity = %v, want 0", style.shadowOpacity)
	}
}

func TestInputStyleHover(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, true, false, false, false)

	if style.bg != (color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff}) {
		t.Fatalf("hover bg = %v, want field hover", style.bg)
	}
	if style.border.A != 0 || style.borderWidth != 0 {
		t.Fatalf("hover border = %v at %v, want none", style.border, style.borderWidth)
	}
}

func TestInputStyleFocus(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, true, true, false, false)

	if style.border != (color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff}) {
		t.Fatalf("focus border = %v, want accent", style.border)
	}
	if style.borderWidth != unit.Dp(2) {
		t.Fatalf("focus border width = %v, want 2dp", style.borderWidth)
	}
}

func TestInputStyleInvalid(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, true, true, false, true)

	if style.border != (color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff}) {
		t.Fatalf("invalid border = %v, want danger", style.border)
	}
	if style.borderWidth != unit.Dp(2) {
		t.Fatalf("focused invalid border width = %v, want 2dp", style.borderWidth)
	}
}

func TestInputStyleInvalidUnfocused(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, false, false, false, true)

	if style.border != (color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff}) {
		t.Fatalf("invalid border = %v, want danger", style.border)
	}
	if style.borderWidth != unit.Dp(1) {
		t.Fatalf("invalid border width = %v, want 1dp", style.borderWidth)
	}
}

func TestInputStyleDisabled(t *testing.T) {
	theme := DefaultTheme()
	style := inputStyleFor(&theme, InputPrimary, true, true, true, true)

	if style.border.A != 0x7f {
		t.Fatalf("disabled border alpha = %d, want 127", style.border.A)
	}
	if style.fg.A != 0x7f {
		t.Fatalf("disabled fg alpha = %d, want 127", style.fg.A)
	}
	if style.shadowOpacity != 0.5 {
		t.Fatalf("disabled shadow opacity = %v, want 0.5", style.shadowOpacity)
	}
}

func TestInputDisabledClearsHover(t *testing.T) {
	state := &inputState{hovered: true}

	state.update(newContext(nil), testLayoutContext(), true, new(int))

	if state.hovered {
		t.Fatal("disabled input kept hover state")
	}
}

func TestInputBackgroundTransition(t *testing.T) {
	state := new(inputState)
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	to := color.NRGBA{R: 110, G: 120, B: 130, A: 255}
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.background(gtx, from); got != from {
		t.Fatalf("initial background = %v, want %v", got, from)
	}
	if got := state.background(gtx, to); got != from {
		t.Fatalf("transition start = %v, want %v", got, from)
	}

	gtx.Now = start.Add(inputColorDuration / 2)
	mid := state.background(gtx, to)
	if mid == from || mid == to {
		t.Fatalf("transition midpoint = %v, want between %v and %v", mid, from, to)
	}

	gtx.Now = start.Add(inputColorDuration)
	if got := state.background(gtx, to); got != to {
		t.Fatalf("transition end = %v, want %v", got, to)
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

func focusedInput(t *testing.T) (*Context, *input.Router, *widget.Editor) {
	t.Helper()
	ctx := newContext(nil)
	router := new(input.Router)

	layoutInputFrame(ctx, router)
	editor := ctx.editors["name"]
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

func focusedInputWithButton(t *testing.T) (*Context, *input.Router, *widget.Editor) {
	t.Helper()
	ctx := newContext(nil)
	router := new(input.Router)

	layoutInputAndButtonFrame(ctx, router)
	editor := ctx.editors["name"]
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

func layoutInputFrame(ctx *Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx.beginFrame()
	Input("name", "").Hint("Name").Layout(ctx, gtx)
	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}

func layoutInputAndButtonFrame(ctx *Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx.beginFrame()

	inputGtx := gtx
	inputGtx.Constraints = layout.Constraints{Max: image.Pt(200, 40)}
	Input("name", "").Hint("Name").Layout(ctx, inputGtx)

	buttonGtx := gtx
	buttonGtx.Constraints = layout.Constraints{Max: image.Pt(80, 40)}
	stack := op.Offset(image.Pt(220, 0)).Push(gtx.Ops)
	Button("add", Text("Add")).Layout(ctx, buttonGtx)
	stack.Pop()

	ctx.applyFrameCommands(gtx)
	ctx.endFrame()
	router.Frame(&ops)
}
