package input

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
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

	Input("name", "").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(field.State), field.Style{}, child)

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
