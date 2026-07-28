package combobox

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
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

func TestComboBoxDataVersionCachesVisibleItems(t *testing.T) {
	widget := ComboBox("cached", "one", []ComboBoxItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}}).DataVersion(1)
	state := new(comboBoxState)
	state.checkItems(widget.items, true, widget.dataVersion)
	first := state.visibleItems(widget, "", "One")
	second := state.visibleItems(widget, "", "One")
	if &first[0] != &second[0] {
		t.Fatal("unchanged ComboBox data version did not reuse visible items")
	}
	state.checkItems(widget.items, true, 2)
	third := state.visibleItems(widget.DataVersion(2), "", "One")
	if &first[0] == &third[0] {
		t.Fatal("changed ComboBox data version reused stale visible items")
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
	state.popoverTransition.Set(1, 1, time.Time{})
	layoutComboBoxFrame(ctx, router, combo)

	item := state.items["dog"]
	if item == nil {
		t.Fatal("missing dog item state")
	}
	position := f32.Pt(20, 106)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutComboBoxFrame(ctx, router, combo)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutComboBoxFrame(ctx, router, combo)

	if got != "dog" {
		t.Fatalf("selected key = %q, want dog", got)
	}
	if state.open {
		t.Fatal("combobox stayed open after selection")
	}
	if frame.FocusVisible(ctx, &state.editor, true) {
		t.Fatal("pointer-selected item exposed keyboard-visible editor focus")
	}
	if text := state.editor.Text(); text != "Dog" {
		t.Fatalf("editor text = %q, want Dog", text)
	}
	start, end := state.editor.Selection()
	if want := len([]rune(state.editor.Text())); start != want || end != want {
		t.Fatalf("editor selection = (%d, %d), want caret at %d", start, end, want)
	}
}

func TestComboBoxIgnoresEscapeWhileAnotherOverlayIsTopmost(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	combo := ComboBox("animal", "", comboBoxTestItems())
	layoutComboBoxFrame(ctx, router, combo)
	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	router.Source().Execute(key.FocusCmd{Tag: &state.editor})
	state.popoverTransition.Set(1, 1, time.Time{})
	layoutComboBoxBelowOverlayFrame(ctx, router, combo)
	layoutComboBoxBelowOverlayFrame(ctx, router, combo)
	if !state.open {
		t.Fatal("combobox closed after another overlay became topmost")
	}
	if !router.Source().Focused(&state.editor) {
		t.Fatal("combobox editor lost focus after another overlay became topmost")
	}

	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutComboBoxBelowOverlayFrame(ctx, router, combo)
	if !state.open {
		t.Fatal("background combobox consumed Escape while another overlay was topmost")
	}
	if !router.Source().Focused(&state.editor) {
		t.Fatal("background combobox lost editor focus while another overlay was topmost")
	}
}

func TestComboBoxPanelPaddingPressKeepsFocusAndBlocksBackground(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	combo := ComboBox("animal", "", comboBoxTestItems()).FullWidth()
	var background widget.Clickable
	backgroundClicked := false
	start := time.Unix(1, 0)

	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start)
	comboState := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	router.Source().Execute(key.FocusCmd{Tag: &comboState.editor})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(time.Millisecond))
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(time.Millisecond+comboBoxAnimationDuration))
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(2*time.Millisecond+comboBoxAnimationDuration))

	position := f32.Pt(200, 49)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(3*time.Millisecond+comboBoxAnimationDuration))

	if !comboState.open {
		t.Fatal("panel padding press closed the combobox")
	}
	if !router.Source().Focused(&comboState.editor) {
		t.Fatal("panel padding press cleared editor focus")
	}
	if state.ActivePresses(background.History()) != 0 || backgroundClicked {
		t.Fatal("panel padding press reached the background control")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(4*time.Millisecond+comboBoxAnimationDuration))

	if !comboState.open {
		t.Fatal("panel padding release closed the combobox")
	}
	if !router.Source().Focused(&comboState.editor) {
		t.Fatal("panel padding release cleared editor focus")
	}
	if frame.FocusVisible(ctx, &comboState.editor, true) {
		t.Fatal("panel padding click exposed keyboard-visible editor focus")
	}
	if backgroundClicked {
		t.Fatal("panel padding click reached the background control")
	}
}

func TestComboBoxExitingPanelPaddingBlocksBackground(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := ""
	combo := ComboBox("animal", "", comboBoxTestItems()).
		FullWidth().
		OnChange(func(key string) {
			selected = key
		})
	var background widget.Clickable
	backgroundClicked := false
	start := time.Unix(1, 0)

	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start)
	comboState := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	router.Source().Execute(key.FocusCmd{Tag: &comboState.editor})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(time.Millisecond))
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(time.Millisecond+comboBoxAnimationDuration))
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, start.Add(2*time.Millisecond+comboBoxAnimationDuration))

	comboState.open = false
	closeStart := start.Add(3*time.Millisecond + comboBoxAnimationDuration)
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, closeStart)
	midExit := closeStart.Add(comboBoxAnimationDuration / 2)
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, midExit)
	if progress := comboState.popoverTransition.Current(); progress <= 0 || progress >= 1 {
		t.Fatalf("exit progress = %v, want between 0 and 1", progress)
	}

	position := f32.Pt(200, 49)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, midExit.Add(time.Millisecond))

	if comboState.open {
		t.Fatal("exiting panel padding press reopened the combobox")
	}
	if comboState.popoverTransition.Current() <= 0 {
		t.Fatal("panel finished exiting before the padding press was processed")
	}
	if state.ActivePresses(background.History()) != 0 || backgroundClicked {
		t.Fatal("exiting panel padding press reached the background control")
	}
	if selected != "" {
		t.Fatalf("exiting panel content selected %q", selected)
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutComboBoxOverBackgroundFrame(ctx, router, combo, &background, &backgroundClicked, midExit.Add(2*time.Millisecond))

	if comboState.open {
		t.Fatal("exiting panel padding release reopened the combobox")
	}
	if comboState.popoverTransition.Current() <= 0 {
		t.Fatal("panel finished exiting before the padding release was processed")
	}
	if backgroundClicked {
		t.Fatal("exiting panel padding click reached the background control")
	}
	if selected != "" {
		t.Fatalf("exiting panel content selected %q", selected)
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
	state.popoverTransition.Set(1, 1, time.Time{})
	layoutComboBoxFrame(ctx, router, combo(""))

	state.items["dog"].Clickable.Click()
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

func TestComboBoxTriggerUsesPointerCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	combo := ComboBox("animal", "", comboBoxTestItems()).FullWidth()
	layoutComboBoxFrame(ctx, router, combo)

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(20, 20)})
	layoutComboBoxFrame(ctx, router, combo)
	if got := router.Cursor(); got != pointer.CursorText {
		t.Fatalf("input cursor = %v, want %v", got, pointer.CursorText)
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(280, 20)})
	layoutComboBoxFrame(ctx, router, combo)
	if got := router.Cursor(); got != pointer.CursorPointer {
		t.Fatalf("trigger cursor = %v, want %v", got, pointer.CursorPointer)
	}
}

func TestComboBoxFirstTriggerClickStaysOpen(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	combo := ComboBox("animal", "dog", comboBoxTestItems()).FullWidth()
	layoutComboBoxFrame(ctx, router, combo)

	position := f32.Pt(280, 20)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutComboBoxFrame(ctx, router, combo)
	layoutComboBoxFrame(ctx, router, combo)

	state := testComponentState[comboBoxState](ctx, "animal", stateSlotComboBox)
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutComboBoxFrame(ctx, router, combo)
	if !state.open {
		t.Fatal("first trigger click closed combobox on release")
	}
	if !router.Source().Focused(&state.editor) {
		t.Fatal("trigger click did not focus the combobox editor")
	}
	if frame.FocusVisible(ctx, &state.editor, true) {
		t.Fatal("pointer trigger click exposed keyboard-visible editor focus")
	}
	if state.highlight != -1 {
		t.Fatalf("highlight = %d, want no keyboard highlight", state.highlight)
	}
}

func TestComboBoxItemHeightFollowsContent(t *testing.T) {
	ctx := newContext(nil)
	state := new(comboBoxState)
	state.beginFrame()
	var editor widget.Editor
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(400, 300)}
	combo := ComboBox("animal", "dog", comboBoxTestItems())
	want := gtx.Dp(frame.ActiveTheme(ctx).Components.ComboBox.ItemHeight)

	singleLine := combo.layoutItem(ctx, gtx, state, &editor, ComboBoxItem{Key: "dog", Label: "Dog"}, true, false)
	if singleLine.Size.Y != want {
		t.Fatalf("single-line item height = %d, want %d", singleLine.Size.Y, want)
	}
	described := combo.layoutItem(ctx, gtx, state, &editor, ComboBoxItem{Key: "panda", Label: "Panda", Description: "Black and white"}, false, false)
	if described.Size.Y <= want {
		t.Fatalf("described item height = %d, want greater than %d", described.Size.Y, want)
	}
}

func TestComboBoxSingleLineItemContentIsVerticallyCentered(t *testing.T) {
	size, offset := comboBoxItemFrame(
		layout.Constraints{Max: image.Pt(400, 300)},
		36,
		image.Pt(120, 20),
	)
	if size.Y != 36 || offset.Y != 8 {
		t.Fatalf("single-line frame = size %v, offset %v, want height 36 and Y offset 8", size, offset)
	}
}

func TestComboBoxSelectedItemHasNoBackground(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	style := comboBoxItemStyleFor(&activeTheme, false, false, false)
	if style.bg.A != 0 {
		t.Fatalf("selected item background = %#v, want transparent", style.bg)
	}
}

func TestComboBoxSelectionAnimation(t *testing.T) {
	state := new(comboBoxItemState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.Selection(gtx, false, comboBoxItemSelectDuration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.Selection(gtx, true, comboBoxItemSelectDuration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}
	gtx.Now = start.Add(comboBoxItemSelectDuration / 2)
	if got := state.Selection(gtx, true, comboBoxItemSelectDuration); got <= 0 || got >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", got)
	}
	gtx.Now = start.Add(comboBoxItemSelectDuration)
	if got := state.Selection(gtx, true, comboBoxItemSelectDuration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
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
		false,
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
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutComboBoxBelowOverlayFrame(ctx *frame.Context, router *input.Router, combo ComboBoxWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 260)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	combo.Layout(ctx, gtx)
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key: "test-top-overlay",
		Layout: func(layout.Context, image.Rectangle, bool) layout.Dimensions {
			return layout.Dimensions{}
		},
	})
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutComboBoxOverBackgroundFrame(ctx *frame.Context, router *input.Router, combo ComboBoxWidget, background *widget.Clickable, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(400, 300)
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	presses := state.ActivePresses(background.History())
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	frame.FocusOnPress(ctx, background, background.History(), presses)
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	comboGtx := gtx
	comboGtx.Constraints = layout.Constraints{Max: viewport}
	combo.Layout(ctx, comboGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
