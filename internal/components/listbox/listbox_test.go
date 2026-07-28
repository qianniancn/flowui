package listbox

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/nav"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func newContextWithTheme(_ any, value *theme.Theme) *frame.Context {
	return frame.New(nil, value, locale.LanguageAuto)
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

func TestListBoxOptions(t *testing.T) {
	box := ListBox("files", "go", listBoxTestItems()).
		EmptyText("Nothing here").
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnAction(func(string) {}).
		SelectionMode(ListBoxSelectionNone).
		DisabledKeys([]string{"rust"}).
		Disabled(true).
		FullWidth().
		AllowEmptySelection().
		HideIndicator().
		MaxHeight(120)

	if box.key != "files" || box.selectedKey != "go" || len(box.items) != 3 {
		t.Fatal("listbox constructor did not set fields")
	}
	if box.emptyText != "Nothing here" || box.onChange == nil || box.onSelectionChange == nil || box.onAction == nil {
		t.Fatal("listbox callbacks/options were not set")
	}
	if box.selectionMode != ListBoxSelectionNone || !listBoxSameKeys(box.disabledKeys, []string{"rust"}) {
		t.Fatal("listbox selection options were not set")
	}
	if !box.disabled || !box.fullWidth || !box.allowEmpty || !box.hideIndicator || box.maxHeight != 120 {
		t.Fatal("listbox visual options were not set")
	}
}

func TestListBoxMultipleConstructor(t *testing.T) {
	box := ListBoxMultiple("files", []string{"go", "rust"}, listBoxTestItems())

	if box.key != "files" || box.selectionMode != ListBoxSelectionMultiple || len(box.selectedKeys) != 2 || len(box.items) != 3 {
		t.Fatal("multi listbox constructor did not set fields")
	}
	if box.emptyText != "No items" {
		t.Fatalf("empty text = %q, want No items", box.emptyText)
	}
}

func TestListBoxSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := ListBoxSections("files", "go", sections)

	if box.key != "files" || box.selectedKey != "go" || len(box.sections) != 2 {
		t.Fatal("section listbox constructor did not set fields")
	}
	if got := box.allItems(); len(got) != 3 {
		t.Fatalf("all items = %d, want 3", len(got))
	}
}

func TestListBoxMultipleSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := ListBoxMultipleSections("files", []string{"go"}, sections)

	if box.key != "files" || box.selectionMode != ListBoxSelectionMultiple || len(box.selectedKeys) != 1 || len(box.sections) != 2 {
		t.Fatal("multi section listbox constructor did not set fields")
	}
}

func TestListBoxRejectsDuplicateItemKeys(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]ListBoxItem{
			{Key: "go", Label: "Go"},
			{Key: "go", Label: "Go again"},
		})
	})
}

func TestListBoxRejectsEmptyItemKey(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]ListBoxItem{{Label: "Missing key"}})
	})
}

func TestListBoxClickActivatesItem(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	ListBox("languages", "go", listBoxTestItems()).
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "rust" {
		t.Fatalf("changed = %q, want rust", changed)
	}
	if action != "rust" {
		t.Fatalf("action = %q, want rust", action)
	}
}

func TestListBoxPointerClickDoesNotShowKeyboardFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "go"
	widget := func() ListBoxWidget {
		return ListBox("languages", selected, listBoxTestItems()).
			FullWidth().
			OnChange(func(key string) { selected = key })
	}
	layoutFrame := func(now time.Time) {
		viewport := image.Pt(300, 200)
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
		frame.BeginFrameWithViewport(ctx, viewport)
		widget().Layout(ctx, gtx)
		frame.ApplyFrameCommands(ctx, gtx)
		frame.EndFrame(ctx)
		router.Frame(&ops)
	}

	now := time.Unix(4, 0)
	layoutFrame(now)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(2 * time.Millisecond))
	layoutFrame(now.Add(3 * time.Millisecond))

	state, _ := frame.PeekState[listBoxState](ctx, "languages", stateSlotListBox)
	item := state.items["rust"]
	if selected != "rust" {
		t.Fatalf("selected = %q, want rust", selected)
	}
	if !router.Source().Focused(&item.Clickable) {
		t.Fatal("pointer click did not focus the selected item")
	}
	if frame.FocusVisible(ctx, &item.Clickable, true) {
		t.Fatal("pointer click displayed a keyboard focus ring")
	}
}

func TestListBoxClickSelectedItemOnlyRunsAction(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	ListBox("languages", "go", listBoxTestItems()).
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
	if action != "go" {
		t.Fatalf("action = %q, want go", action)
	}
}

func TestListBoxClickSelectedItemCanClearSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := "not-called"
	ListBox("languages", "go", listBoxTestItems()).
		AllowEmptySelection().
		OnChange(func(key string) {
			changed = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
}

func TestListBoxSelectionModeNoneOnlyRunsAction(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed, action string
	ListBox("languages", "go", listBoxTestItems()).
		SelectionMode(ListBoxSelectionNone).
		OnChange(func(key string) {
			changed = key
		}).
		OnAction(func(key string) {
			action = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
	if action != "rust" {
		t.Fatalf("action = %q, want rust", action)
	}
}

func TestListBoxMultipleClickAddsSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed []string
	ListBoxMultiple("languages", []string{"go"}, listBoxTestItems()).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"go", "rust"}) {
		t.Fatalf("changed = %#v, want [go rust]", changed)
	}
}

func TestListBoxMultipleClickRemovesSelection(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"go": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	var changed []string
	ListBoxMultiple("languages", []string{"go", "rust"}, listBoxTestItems()).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"rust"}) {
		t.Fatalf("changed = %#v, want [rust]", changed)
	}
}

func TestListBoxMultipleDoesNotMutateSelectedKeys(t *testing.T) {
	selected := []string{"go", "go", "rust"}
	next := listBoxToggleSelectedKeys(selected, "zig")

	if !listBoxSameKeys(selected, []string{"go", "go", "rust"}) {
		t.Fatalf("selected = %#v, want original slice unchanged", selected)
	}
	if !listBoxSameKeys(next, []string{"go", "rust", "zig"}) {
		t.Fatalf("next = %#v, want deduped selection with zig", next)
	}
}

func TestListBoxMultipleDoesNotCallSingleChange(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	called := false
	ListBoxMultiple("languages", []string{"go"}, listBoxTestItems()).
		OnChange(func(string) {
			called = true
		}).
		OnSelectionChange(func([]string) {}).
		Layout(ctx, testLayoutContext())

	if called {
		t.Fatal("multi listbox called single-selection OnChange")
	}
}

func TestListBoxDisabledItemIgnoresClick(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"swift": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := false
	ListBox("languages", "go", listBoxTestItems()).
		OnChange(func(string) {
			changed = true
		}).
		Layout(ctx, testLayoutContext())

	if changed {
		t.Fatal("disabled item handled click")
	}
}

func TestListBoxDisabledKeysIgnoreClick(t *testing.T) {
	itemState := new(listBoxItemState)
	itemState.Clickable.Click()
	state := &listBoxState{
		items: map[string]*listBoxItemState{
			"rust": itemState,
		},
	}
	ctx := newContext(nil)
	testSetComponentState(ctx, "languages", stateSlotListBox, state)

	changed := false
	ListBox("languages", "go", listBoxTestItems()).
		DisabledKeys([]string{"rust"}).
		OnChange(func(string) {
			changed = true
		}).
		Layout(ctx, testLayoutContext())

	if changed {
		t.Fatal("disabled key handled click")
	}
}

// These exercise the item-model → nav wiring (Disabled field + disabledKeys
// feeding nav.List); the navigation algorithms themselves are unit-tested in
// package nav.

func TestListBoxMoveIndexSkipsDisabledAndDoesNotWrap(t *testing.T) {
	items := listBoxTestItems()
	if got, ok := nav.Move(sliceList(items, nil), 0, 1, false); !ok || got != 1 {
		t.Fatalf("move from first = %d %v, want 1 true", got, ok)
	}
	if got, ok := nav.Move(sliceList(items, nil), 1, 1, false); ok || got != 1 {
		t.Fatalf("move past end = %d %v, want current false", got, ok)
	}
	if got, ok := nav.Move(sliceList(items, nil), 1, -1, false); !ok || got != 0 {
		t.Fatalf("move up = %d %v, want 0 true", got, ok)
	}
	if got, ok := nav.Move(sliceList(items, []string{"rust"}), 0, 1, false); ok || got != 0 {
		t.Fatalf("move into disabled key = %d %v, want current false", got, ok)
	}
	if got, ok := nav.Move(sliceList(items, []string{"rust"}), 1, 1, false); !ok || got != 0 {
		t.Fatalf("move away from disabled current = %d %v, want 0 true", got, ok)
	}
}

func TestListBoxFirstLastEnabled(t *testing.T) {
	items := []ListBoxItem{
		{Key: "a", Label: "A", Disabled: true},
		{Key: "b", Label: "B"},
		{Key: "c", Label: "C", Disabled: true},
	}
	if got, ok := FirstEnabled(items, nil); !ok || got != 1 {
		t.Fatalf("first enabled = %d %v, want 1 true", got, ok)
	}
	if got, ok := LastEnabled(items, nil); !ok || got != 1 {
		t.Fatalf("last enabled = %d %v, want 1 true", got, ok)
	}
	if _, ok := FirstEnabled(items, []string{"b"}); ok {
		t.Fatal("disabled key was returned as first enabled")
	}
}

func TestListBoxTypeaheadSkipsDisabledAndWraps(t *testing.T) {
	items := []ListBoxItem{
		{Key: "alpha", Label: "Alpha"},
		{Key: "beta", Label: "Beta", Disabled: true},
		{Key: "bravo", Label: "Bravo"},
		{Key: "charlie", Label: "Charlie"},
	}

	if index, ok := nav.Match(sliceList(items, nil), 0, "b"); !ok || index != 2 {
		t.Fatalf("typeahead b = %d %v, want enabled Bravo", index, ok)
	}
	if index, ok := nav.Match(sliceList(items, nil), 3, "a"); !ok || index != 0 {
		t.Fatalf("wrapped typeahead a = %d %v, want Alpha", index, ok)
	}
	if index, ok := nav.Match(sliceList(items, []string{"bravo"}), 0, "b"); ok || index != 0 {
		t.Fatalf("disabled typeahead b = %d %v, want no match", index, ok)
	}
}

func TestListBoxTypeaheadBufferExpires(t *testing.T) {
	state := new(listBoxState)
	start := time.Unix(1, 0)
	if got := state.appendTypeahead(start, "n"); got != "n" {
		t.Fatalf("first typeahead = %q, want n", got)
	}
	if got := state.appendTypeahead(start.Add(100*time.Millisecond), "e"); got != "ne" {
		t.Fatalf("continued typeahead = %q, want ne", got)
	}
	if got := state.appendTypeahead(start.Add(100*time.Millisecond+listBoxTypeaheadTimeout+time.Millisecond), "g"); got != "g" {
		t.Fatalf("expired typeahead = %q, want g", got)
	}
}

func TestListBoxStyleUsesDangerPalette(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.Danger = color.NRGBA{R: 7, G: 8, B: 9, A: 255}

	style := listBoxItemStyleFor(&theme, ListBoxItemDanger, false)

	if style.fg != theme.Palette.Danger {
		t.Fatalf("danger foreground = %#v, want %#v", style.fg, theme.Palette.Danger)
	}
	if style.indicator != theme.Palette.Danger {
		t.Fatalf("danger indicator = %#v, want %#v", style.indicator, theme.Palette.Danger)
	}
}

func TestListBoxLayoutKeepsState(t *testing.T) {
	ctx := newContext(nil)
	dims := ListBox("languages", "go", listBoxTestItems()).Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("listbox returned empty dimensions")
	}
}

func TestListBoxSectionsLayout(t *testing.T) {
	ctx := newContext(nil)
	dims := ListBoxSections("languages", "go", listBoxTestSections()).Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing section listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("section listbox returned empty dimensions")
	}
}

func TestListBoxDataVersionCachesEntries(t *testing.T) {
	widget := ListBoxSections("cached", "one", []ListBoxSection{{Title: "Group", Items: []ListBoxItem{{Key: "one", Label: "One"}}}}).DataVersion(1)
	state := new(listBoxState)
	entries, items := state.resolveEntries(widget)
	cachedEntries, cachedItems := state.resolveEntries(widget)
	if &entries[0] != &cachedEntries[0] || &items[0] != &cachedItems[0] {
		t.Fatal("unchanged ListBox data version did not reuse entries")
	}
	updatedEntries, _ := state.resolveEntries(widget.DataVersion(2))
	if &entries[0] == &updatedEntries[0] {
		t.Fatal("changed ListBox data version reused stale entries")
	}
}

func BenchmarkListBoxLargeData(b *testing.B) {
	items := make([]ListBoxItem, 10_000)
	for index := range items {
		items[index] = ListBoxItem{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := ListBox("large", "", items)
	for _, benchmark := range []struct {
		name   string
		widget ListBoxWidget
	}{
		{name: "unversioned", widget: widget},
		{name: "versioned", widget: widget.DataVersion(1)},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			state := new(listBoxState)
			state.resolveEntries(benchmark.widget)
			b.ReportAllocs()
			for b.Loop() {
				entries, resolved := state.resolveEntries(benchmark.widget)
				runtime.KeepAlive(entries)
				runtime.KeepAlive(resolved)
			}
		})
	}
}

func BenchmarkListBoxLargeSelection(b *testing.B) {
	keys := make([]string, 10_000)
	for index := range keys {
		keys[index] = fmt.Sprintf("item-%d", index)
	}
	widget := ListBoxMultiple("large", keys, nil)
	target := keys[len(keys)-1]
	state := new(listBoxState)
	b.ReportAllocs()
	for b.Loop() {
		widget.selectedKeySet = state.selectedKeys.Resolve(keys)
		runtime.KeepAlive(widget.isSelected(target))
	}
}

func BenchmarkListBoxLargeLayout(b *testing.B) {
	items := make([]ListBoxItem, 10_000)
	for index := range items {
		items[index] = ListBoxItem{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := ListBox("large", "", items).DataVersion(1)
	ctx := newContext(nil)
	b.ReportAllocs()
	for b.Loop() {
		gtx := testLayoutContext()
		frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		widget.Layout(ctx, gtx)
		frame.EndFrame(ctx)
	}
}

func BenchmarkListBoxLargeDisabledLayout(b *testing.B) {
	items := make([]ListBoxItem, 10_000)
	disabled := make([]string, len(items))
	for index := range items {
		items[index] = ListBoxItem{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
		disabled[index] = items[index].Key
	}
	widget := ListBox("large-disabled", "", items).DisabledKeys(disabled).DataVersion(1)
	ctx := newContext(nil)
	b.ReportAllocs()
	for b.Loop() {
		gtx := testLayoutContext()
		frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		widget.Layout(ctx, gtx)
		frame.EndFrame(ctx)
	}
}

func TestListBoxFullWidth(t *testing.T) {
	dims := ListBox("languages", "go", listBoxTestItems()).
		FullWidth().
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("width = %d, want 300", dims.Size.X)
	}
}

func TestListBoxMaxHeight(t *testing.T) {
	items := make([]ListBoxItem, 12)
	for i := range items {
		items[i] = ListBoxItem{Key: string(rune('a' + i)), Label: "Item"}
	}

	dims := ListBox("items", "", items).
		MaxHeight(80).
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y > 80 {
		t.Fatalf("height = %d, want <= 80", dims.Size.Y)
	}
}

func TestListBoxThemeControlsItemHeight(t *testing.T) {
	theme := DefaultTheme()
	theme.Components.ListBox.Padding = 0
	theme.Components.ListBox.ItemMinHeight = 72
	ctx := newContextWithTheme(nil, &theme)

	dims := ListBox("languages", "go", []ListBoxItem{{Key: "go", Label: "Go"}}).
		Layout(ctx, testLayoutContext())

	if dims.Size.Y < 72 {
		t.Fatalf("height = %d, want at least 72", dims.Size.Y)
	}
}

func TestListBoxPartItemResolvesPerItemState(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Components.ListBox.Padding = 0
	activeTheme.Components.ListBox.Gap = 0
	ctx := newContextWithTheme(nil, &activeTheme)
	declaration := flowstyle.Style{}.
		Part(flowstyle.PartItem, flowstyle.Style{}.
			Height(44).
			When(flowstyle.Selected, flowstyle.Style{}.Height(64)))

	dims := ListBox("languages", "go", []ListBoxItem{
		{Key: "go", Label: "Go"},
		{Key: "rust", Label: "Rust"},
	}).Style(declaration).Layout(ctx, testLayoutContext())

	if dims.Size.Y != 108 {
		t.Fatalf("listbox height = %d, want selected 64 + default 44", dims.Size.Y)
	}
}

func TestListBoxIndicatorVisibility(t *testing.T) {
	box := ListBox("languages", "go", listBoxTestItems())
	if !box.showIndicator(listBoxTestItems()[0]) {
		t.Fatal("single-selection listbox should show default indicator")
	}
	if box.SelectionMode(ListBoxSelectionNone).showIndicator(listBoxTestItems()[0]) {
		t.Fatal("none-selection listbox should hide default indicator")
	}
	if box.HideIndicator().showIndicator(ListBoxItem{
		Key:       "custom",
		Label:     "Custom",
		Indicator: func(bool) frame.Widget { return text.New("!") },
	}) {
		t.Fatal("hide indicator should override custom indicator")
	}
}

func TestListBoxCustomIndicatorReceivesSelectedState(t *testing.T) {
	ctx := newContext(nil)
	gotSelected := false
	item := ListBoxItem{
		Key:   "custom",
		Label: "Custom",
		Indicator: func(selected bool) frame.Widget {
			gotSelected = selected
			return nil
		},
	}

	ListBox("custom", "custom", []ListBoxItem{item}).
		layoutIndicator(ctx, testLayoutContext(), item, listBoxItemStyle{}, true)

	if !gotSelected {
		t.Fatal("custom indicator did not receive selected=true")
	}
}

func TestListBoxSelectionAnimation(t *testing.T) {
	state := new(listBoxItemState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.Selection(gtx, false, listBoxItemSelectDuration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.Selection(gtx, true, listBoxItemSelectDuration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(listBoxItemSelectDuration / 2)
	mid := state.Selection(gtx, true, listBoxItemSelectDuration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(listBoxItemSelectDuration)
	if got := state.Selection(gtx, true, listBoxItemSelectDuration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func listBoxTestItems() []ListBoxItem {
	return []ListBoxItem{
		{Key: "go", Label: "Go", Description: "Simple deployment"},
		{Key: "rust", Label: "Rust", Description: "Memory safety"},
		{Key: "swift", Label: "Swift", Description: "Unavailable", Disabled: true},
	}
}

func listBoxTestSections() []ListBoxSection {
	return []ListBoxSection{
		{
			Title: "Languages",
			Items: []ListBoxItem{
				{Key: "go", Label: "Go"},
				{Key: "rust", Label: "Rust"},
			},
		},
		{
			Title: "Unavailable",
			Items: []ListBoxItem{
				{Key: "swift", Label: "Swift", Disabled: true},
			},
		},
	}
}
