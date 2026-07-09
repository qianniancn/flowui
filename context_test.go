package flowui

import (
	"image/color"
	"testing"
)

func TestClickableKeys(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.clickable("save")

	mustPanic(t, func() {
		ctx.clickable("save")
	})

	ctx.beginFrame()
	ctx.clickable("save")
}

func TestDuplicateKeysAcrossWidgets(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.clickable("name")

	mustPanic(t, func() {
		ctx.editor("name")
	})
}

func TestDuplicateKeysIncludeBools(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.boolState("done")

	mustPanic(t, func() {
		ctx.clickable("done")
	})
}

func TestDuplicateKeysIncludeComboBoxes(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.comboBoxState("animal")

	mustPanic(t, func() {
		ctx.clickable("animal")
	})
}

func TestDuplicateKeysIncludeDatePickers(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.datePickerState("date")

	mustPanic(t, func() {
		ctx.clickable("date")
	})
}

func TestDuplicateKeysIncludeRadioGroups(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.radioGroupState("plan")

	mustPanic(t, func() {
		ctx.clickable("plan")
	})
}

func TestDuplicateKeysIncludeProgressBars(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.progressBarState("upload")

	mustPanic(t, func() {
		ctx.clickable("upload")
	})
}

func TestDuplicateKeysIncludeLists(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.listState("todos")

	mustPanic(t, func() {
		ctx.editor("todos")
	})
}

func TestDuplicateKeysIncludeScrolls(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.scrollState("body")

	mustPanic(t, func() {
		ctx.clickable("body")
	})
}

func TestContextUsesTheme(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.Accent = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	ctx := newContextWithTheme(nil, &theme)

	if ctx.Theme != &theme {
		t.Fatal("context did not use provided theme")
	}
}

func TestEndFrameKeepsUsedState(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	buttonKey, clickable := ctx.clickableWithKey("save")
	button := ctx.buttonState(buttonKey)
	editor := ctx.editor("name")
	inputKey, inputEditor := ctx.inputEditor("search")
	input := ctx.inputState(inputKey)
	combo := ctx.comboBoxState("animal")
	datePicker := ctx.datePickerState("date")
	checkboxKey, checkbox := ctx.boolStateWithKey("done")
	checkboxAnim := ctx.checkboxState(checkboxKey)
	radio := ctx.radioGroupState("plan")
	progress := ctx.progressBarState("upload")
	list := ctx.listState("todos")
	scroll := ctx.scrollState("body")

	ctx.endFrame()

	if ctx.clickables["save"] != clickable {
		t.Fatal("clickable state was not kept")
	}
	if ctx.buttons["save"] != button {
		t.Fatal("button state was not kept")
	}
	if ctx.editors["name"] != editor {
		t.Fatal("editor state was not kept")
	}
	if ctx.editors["search"] != inputEditor {
		t.Fatal("input editor state was not kept")
	}
	if ctx.inputs["search"] != input {
		t.Fatal("input state was not kept")
	}
	if ctx.combos["animal"] != combo {
		t.Fatal("combobox state was not kept")
	}
	if ctx.datePickers["date"] != datePicker {
		t.Fatal("date picker state was not kept")
	}
	if ctx.bools["done"] != checkbox {
		t.Fatal("checkbox state was not kept")
	}
	if ctx.checkboxes["done"] != checkboxAnim {
		t.Fatal("checkbox animation state was not kept")
	}
	if ctx.radioGroups["plan"] != radio {
		t.Fatal("radio group state was not kept")
	}
	if ctx.progressBars["upload"] != progress {
		t.Fatal("progress bar state was not kept")
	}
	if ctx.lists["todos"] != list {
		t.Fatal("list state was not kept")
	}
	if ctx.scrolls["body"] != scroll {
		t.Fatal("scroll state was not kept")
	}
}

func TestEndFrameRemovesUnusedState(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	buttonKey, _ := ctx.clickableWithKey("save")
	ctx.buttonState(buttonKey)
	ctx.editor("name")
	inputKey, _ := ctx.inputEditor("search")
	ctx.inputState(inputKey)
	ctx.comboBoxState("animal")
	ctx.datePickerState("date")
	checkboxKey, _ := ctx.boolStateWithKey("done")
	ctx.checkboxState(checkboxKey)
	ctx.radioGroupState("plan")
	ctx.progressBarState("upload")
	ctx.listState("todos")
	ctx.scrollState("body")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.endFrame()

	if len(ctx.clickables) != 0 {
		t.Fatalf("clickables = %d, want 0", len(ctx.clickables))
	}
	if len(ctx.buttons) != 0 {
		t.Fatalf("buttons = %d, want 0", len(ctx.buttons))
	}
	if len(ctx.editors) != 0 {
		t.Fatalf("editors = %d, want 0", len(ctx.editors))
	}
	if len(ctx.inputs) != 0 {
		t.Fatalf("inputs = %d, want 0", len(ctx.inputs))
	}
	if len(ctx.combos) != 0 {
		t.Fatalf("comboboxes = %d, want 0", len(ctx.combos))
	}
	if len(ctx.datePickers) != 0 {
		t.Fatalf("date pickers = %d, want 0", len(ctx.datePickers))
	}
	if len(ctx.bools) != 0 {
		t.Fatalf("checkboxes = %d, want 0", len(ctx.bools))
	}
	if len(ctx.checkboxes) != 0 {
		t.Fatalf("checkbox animation states = %d, want 0", len(ctx.checkboxes))
	}
	if len(ctx.radioGroups) != 0 {
		t.Fatalf("radio groups = %d, want 0", len(ctx.radioGroups))
	}
	if len(ctx.progressBars) != 0 {
		t.Fatalf("progress bars = %d, want 0", len(ctx.progressBars))
	}
	if len(ctx.lists) != 0 {
		t.Fatalf("lists = %d, want 0", len(ctx.lists))
	}
	if len(ctx.scrolls) != 0 {
		t.Fatalf("scrolls = %d, want 0", len(ctx.scrolls))
	}
}

func TestEndFrameRemovesOldStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	checkboxKey, _ := ctx.boolStateWithKey("action")
	ctx.checkboxState(checkboxKey)
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("action")
	ctx.endFrame()

	if ctx.bools["action"] != nil {
		t.Fatal("old checkbox state was not removed")
	}
	if ctx.checkboxes["action"] != nil {
		t.Fatal("old checkbox animation state was not removed")
	}
	if ctx.lists["action"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesOldInputStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	editorKey, _ := ctx.inputEditor("action")
	ctx.inputState(editorKey)
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("action")
	ctx.endFrame()

	if ctx.editors["action"] != nil {
		t.Fatal("old editor state was not removed")
	}
	if ctx.inputs["action"] != nil {
		t.Fatal("old input state was not removed")
	}
	if ctx.lists["action"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesOldComboBoxStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.comboBoxState("animal")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("animal")
	ctx.endFrame()

	if ctx.combos["animal"] != nil {
		t.Fatal("old combobox state was not removed")
	}
	if ctx.lists["animal"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesOldDatePickerStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.datePickerState("date")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("date")
	ctx.endFrame()

	if ctx.datePickers["date"] != nil {
		t.Fatal("old date picker state was not removed")
	}
	if ctx.lists["date"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesOldRadioGroupStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.radioGroupState("plan")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("plan")
	ctx.endFrame()

	if ctx.radioGroups["plan"] != nil {
		t.Fatal("old radio group state was not removed")
	}
	if ctx.lists["plan"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesOldProgressBarStateWhenKeyChangesKind(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.progressBarState("upload")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.listState("upload")
	ctx.endFrame()

	if ctx.progressBars["upload"] != nil {
		t.Fatal("old progress bar state was not removed")
	}
	if ctx.lists["upload"] == nil {
		t.Fatal("list state was not kept")
	}
}

func TestEndFrameRemovesInputStateWhenKeyChangesToEditor(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	editorKey, inputEditor := ctx.inputEditor("field")
	ctx.inputState(editorKey)
	ctx.endFrame()

	ctx.beginFrame()
	editor := ctx.editor("field")
	ctx.endFrame()

	if editor != inputEditor {
		t.Fatal("editor state was not reused")
	}
	if ctx.inputs["field"] != nil {
		t.Fatal("old input state was not removed")
	}
}

func TestEndFrameRemovesUnusedScopedCheckboxState(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	pop := ctx.pushKey("todo:1")
	checkboxKey, _ := ctx.boolStateWithKey("done")
	ctx.checkboxState(checkboxKey)
	pop()
	ctx.endFrame()

	ctx.beginFrame()
	pop = ctx.pushKey("todo:2")
	checkboxKey, _ = ctx.boolStateWithKey("done")
	ctx.checkboxState(checkboxKey)
	pop()
	ctx.endFrame()

	if ctx.bools["todo:1/done"] != nil {
		t.Fatal("old scoped checkbox state was not removed")
	}
	if ctx.checkboxes["todo:1/done"] != nil {
		t.Fatal("old scoped checkbox animation state was not removed")
	}
	if ctx.bools["todo:2/done"] == nil {
		t.Fatal("current scoped checkbox state was not kept")
	}
	if ctx.checkboxes["todo:2/done"] == nil {
		t.Fatal("current scoped checkbox animation state was not kept")
	}
}

func TestEndFrameRemovesOldStateWhenKeyChangesToScroll(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	ctx.listState("body")
	ctx.endFrame()

	ctx.beginFrame()
	ctx.scrollState("body")
	ctx.endFrame()

	if ctx.lists["body"] != nil {
		t.Fatal("old list state was not removed")
	}
	if ctx.scrolls["body"] == nil {
		t.Fatal("scroll state was not kept")
	}
}

func TestKeyScope(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()

	pop := ctx.pushKey("todo:1")
	ctx.clickable("delete")
	pop()
	ctx.clickable("delete")

	if ctx.clickables["todo:1/delete"] == nil {
		t.Fatal("missing scoped clickable")
	}
	if ctx.clickables["delete"] == nil {
		t.Fatal("missing root clickable")
	}
}

func TestEndFrameRemovesUnusedScopedState(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()
	pop := ctx.pushKey("todo:1")
	ctx.clickable("delete")
	pop()
	ctx.endFrame()

	ctx.beginFrame()
	pop = ctx.pushKey("todo:2")
	ctx.clickable("delete")
	pop()
	ctx.endFrame()

	if ctx.clickables["todo:1/delete"] != nil {
		t.Fatal("old scoped clickable was not removed")
	}
	if ctx.clickables["todo:2/delete"] == nil {
		t.Fatal("current scoped clickable was not kept")
	}
}

func TestDuplicateScopedKeys(t *testing.T) {
	ctx := newContext(nil)
	ctx.beginFrame()

	pop := ctx.pushKey("todo:1")
	defer pop()
	ctx.clickable("delete")
	mustPanic(t, func() {
		ctx.clickable("delete")
	})
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
