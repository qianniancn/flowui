package combobox

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
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

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestComboBoxOptions(t *testing.T) {
	c := ComboBox("animal", "cat", comboBoxTestItems()).
		Hint("Animal").
		InputValue("ca").
		EmptyText("Nothing").
		Variant(field.Secondary).
		Invalid(true).
		Disabled(true).
		FullWidth().
		AllowCustomValue()

	if c.key != "animal" {
		t.Fatalf("key = %q, want animal", c.key)
	}
	if c.hint != "Animal" {
		t.Fatal("hint was not set")
	}
	if c.inputValue != "ca" || !c.hasInputValue {
		t.Fatal("input value was not set")
	}
	if c.emptyText != "Nothing" {
		t.Fatal("empty text was not set")
	}
	if c.variant != field.Secondary {
		t.Fatal("variant was not set")
	}
	if !c.invalid || !c.disabled || !c.fullWidth || !c.allowCustomValue {
		t.Fatal("boolean option was not set")
	}
}

func TestComboBoxSyncsSelectedLabel(t *testing.T) {
	ctx := newContext(nil)
	layoutComboBoxFrame(ctx, new(input.Router), ComboBox("animal", "cat", comboBoxTestItems()))

	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	if state == nil {
		t.Fatal("missing combobox state")
	}
	if got := state.editor.Text(); got != "Cat" {
		t.Fatalf("editor text = %q, want Cat", got)
	}
}

func TestComboBoxSyncDoesNotDispatchInputChange(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var inputChanges int
	combo := ComboBox("animal", "cat", comboBoxTestItems()).
		OnInputChange(func(string) {
			inputChanges++
		})

	layoutComboBoxFrame(ctx, router, combo)

	if inputChanges != 0 {
		t.Fatalf("input changes = %d, want 0", inputChanges)
	}
	if testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox).open {
		t.Fatal("combobox opened while syncing selected key")
	}
}

func TestComboBoxInputValueOverridesSelectedLabel(t *testing.T) {
	ctx := newContext(nil)
	layoutComboBoxFrame(ctx, new(input.Router),
		ComboBox("animal", "cat", comboBoxTestItems()).InputValue("do"),
	)

	if got := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox).editor.Text(); got != "do" {
		t.Fatalf("editor text = %q, want do", got)
	}
}

func TestComboBoxUpdatesChangedSelectedLabel(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", []ComboBoxItem{
		{Key: "cat", Label: "Cat"},
	}))

	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", []ComboBoxItem{
		{Key: "cat", Label: "Kitten"},
	}))

	if got := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox).editor.Text(); got != "Kitten" {
		t.Fatalf("editor text = %q, want Kitten", got)
	}
}

func TestComboBoxClearsRemovedSelectedItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", []ComboBoxItem{
		{Key: "cat", Label: "Cat"},
	}))

	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", []ComboBoxItem{
		{Key: "dog", Label: "Dog"},
	}))

	if got := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox).editor.Text(); got != "" {
		t.Fatalf("editor text = %q, want empty", got)
	}
}

func TestComboBoxKeepsUserFilterWhenSelectedLabelUnchanged(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", comboBoxTestItems()))
	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	state.editor.SetText("do")

	layoutComboBoxFrame(ctx, router, ComboBox("animal", "cat", comboBoxTestItems()))

	if got := state.editor.Text(); got != "do" {
		t.Fatalf("editor text = %q, want user filter", got)
	}
}

func TestComboBoxFiltering(t *testing.T) {
	items := comboBoxTestItems()
	visible := comboBoxVisibleItems(items, "do", "")

	if len(visible) != 1 || items[visible[0]].Key != "dog" {
		t.Fatalf("visible = %v, want only dog", visible)
	}

	visible = comboBoxVisibleItems(items, "Cat", "Cat")
	if len(visible) != len(items) {
		t.Fatalf("visible count = %d, want all items", len(visible))
	}
}

func TestComboBoxClickSelectsItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var got string
	combo := ComboBox("animal", "", comboBoxTestItems()).
		OnChange(func(key string) {
			got = key
		})

	layoutComboBoxFrame(ctx, router, combo)
	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	if state == nil {
		t.Fatal("missing combobox state")
	}
	router.Source().Execute(key.FocusCmd{Tag: &state.editor})
	state.popover = 1
	state.popoverFrom = 1
	state.popoverTo = 1
	state.popoverReady = true
	layoutComboBoxFrame(ctx, router, combo)

	item := state.items["dog"]
	if item == nil {
		t.Fatal("missing dog item state")
	}
	item.clickable.Click()
	layoutComboBoxFrame(ctx, router, combo)

	if got != "dog" {
		t.Fatalf("selected key = %q, want dog", got)
	}
	if state.open {
		t.Fatal("combobox stayed open after selection")
	}
	if text := state.editor.Text(); text != "Dog" {
		t.Fatalf("editor text = %q, want Dog", text)
	}
}

func TestComboBoxSelectionDoesNotDispatchInputChange(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var selected string
	var inputChanges int
	combo := func(selectedKey string) ComboBoxWidget {
		return ComboBox("animal", selectedKey, comboBoxTestItems()).
			OnChange(func(key string) {
				selected = key
			}).
			OnInputChange(func(string) {
				inputChanges++
			})
	}

	layoutComboBoxFrame(ctx, router, combo(""))
	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	router.Source().Execute(key.FocusCmd{Tag: &state.editor})
	state.popover = 1
	state.popoverFrom = 1
	state.popoverTo = 1
	state.popoverReady = true
	layoutComboBoxFrame(ctx, router, combo(""))

	state.items["dog"].clickable.Click()
	layoutComboBoxFrame(ctx, router, combo(""))
	layoutComboBoxFrame(ctx, router, combo(selected))

	if selected != "dog" {
		t.Fatalf("selected key = %q, want dog", selected)
	}
	if inputChanges != 0 {
		t.Fatalf("input changes = %d, want 0", inputChanges)
	}
	if state.open {
		t.Fatal("combobox reopened after selection")
	}
}

func TestComboBoxFullWidth(t *testing.T) {
	dims := ComboBox("animal", "", comboBoxTestItems()).
		FullWidth().
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("combobox width = %d, want 300", dims.Size.X)
	}
}

func TestComboBoxOpenLayoutDoesNotTakeSpace(t *testing.T) {
	ctx := newContext(nil)
	state := new(comboBoxState)
	state.beginFrame()
	inputDims := layout.Dimensions{Size: image.Pt(300, 40)}
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Ops:         &ops,
	}
	dims := ComboBox("animal", "", comboBoxTestItems()).layoutOpen(
		ctx,
		gtx,
		state,
		&state.editor,
		inputDims,
		[]int{0, 1, 2},
		1,
	)

	if dims.Size != inputDims.Size {
		t.Fatalf("open combobox size = %v, want %v", dims.Size, inputDims.Size)
	}
}

func TestComboBoxHighlightSkipsDisabledItems(t *testing.T) {
	items := []ComboBoxItem{
		{Key: "mobile", Label: "Mobile", Disabled: true},
		{Key: "native", Label: "Native"},
		{Key: "server", Label: "Server", Disabled: true},
	}
	visible := []int{0, 1, 2}
	state := comboBoxState{highlight: 0}

	state.clampHighlight(items, visible)

	if state.highlight != 1 {
		t.Fatalf("highlight = %d, want 1", state.highlight)
	}
	if got := comboBoxMoveHighlight(items, visible, 0, 1); got != 1 {
		t.Fatalf("next highlight = %d, want 1", got)
	}
	if got := comboBoxFirstEnabled(items[:1], []int{0}); got != -1 {
		t.Fatalf("first enabled = %d, want -1", got)
	}
}

func TestComboBoxItemKeysMustBeUnique(t *testing.T) {
	mustPanic(t, func() {
		ComboBox("animal", "", []ComboBoxItem{
			{Key: "dog", Label: "Dog"},
			{Key: "dog", Label: "Dog again"},
		}).Layout(newContext(nil), testLayoutContext())
	})
}

func TestComboBoxItemKeysMustNotBeEmpty(t *testing.T) {
	mustPanic(t, func() {
		ComboBox("animal", "", []ComboBoxItem{
			{Label: "Missing key"},
		}).Layout(newContext(nil), testLayoutContext())
	})
}

func TestComboBoxItemStateSweep(t *testing.T) {
	state := new(comboBoxState)
	state.beginFrame()
	state.item("cat")
	state.endFrame()

	state.beginFrame()
	state.item("dog")
	state.endFrame()

	if state.items["cat"] != nil {
		t.Fatal("old item state was not removed")
	}
	if state.items["dog"] == nil {
		t.Fatal("current item state was not kept")
	}
}

func comboBoxTestItems() []ComboBoxItem {
	return []ComboBoxItem{
		{Key: "cat", Label: "Cat"},
		{Key: "dog", Label: "Dog"},
		{Key: "panda", Label: "Panda", Description: "Black and white"},
	}
}

func layoutComboBoxFrame(ctx *frame.Context, router *input.Router, combo ComboBoxWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 260)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	combo.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
