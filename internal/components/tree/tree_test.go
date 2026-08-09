package tree

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

type treeProbe struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *treeProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestFilterTreeItemsKeepsAncestors(t *testing.T) {
	items := []Item{
		{Key: "src", Label: "src", Children: []Item{
			{Key: "main", Label: "main.go"},
			{Key: "util", Label: "util.go"},
		}},
		{Key: "docs", Label: "docs", Children: []Item{
			{Key: "readme", Label: "README"},
		}},
	}
	filtered := filterTreeItems(items, "main")
	if len(filtered) != 1 || filtered[0].Key != "src" || len(filtered[0].Children) != 1 || filtered[0].Children[0].Key != "main" {
		t.Fatalf("filtered = %#v", filtered)
	}
	if len(filterTreeItems(items, "missing")) != 0 {
		t.Fatal("expected empty filter result")
	}
}

func TestFilterTreeItemsByKeepsAncestors(t *testing.T) {
	items := []Item{
		{Key: "src", Label: "Source", Children: []Item{
			{Key: "main", Label: "main.go"},
			{Key: "util", Label: "util.go"},
		}},
		{Key: "docs", Label: "Docs"},
	}
	filtered := filterTreeItemsBy(items, func(item Item) bool { return item.Key == "util" })
	if len(filtered) != 1 || filtered[0].Key != "src" || len(filtered[0].Children) != 1 || filtered[0].Children[0].Key != "util" {
		t.Fatalf("filtered = %#v", filtered)
	}
}

func TestTreeCustomFilterVersionControlsCache(t *testing.T) {
	items := []Item{{Key: "root", Label: "First"}, {Key: "other", Label: "Other"}}
	want := "First"
	match := func(item Item) bool { return item.Label == want }
	state := new(treeState)
	first := state.resolveVisible(New("tree", "", items).DataVersion(1).FilterFunc(match).FilterVersion(1))
	want = "Other"
	stale := state.resolveVisible(New("tree", "", items).DataVersion(1).FilterFunc(match).FilterVersion(1))
	if len(first) != 1 || first[0].item.Key != "root" || len(stale) != 1 || stale[0].item.Key != "root" {
		t.Fatalf("same filter version did not reuse cache: first %#v stale %#v", first, stale)
	}
	updated := state.resolveVisible(New("tree", "", items).DataVersion(1).FilterFunc(match).FilterVersion(2))
	if len(updated) != 1 || updated[0].item.Key != "other" {
		t.Fatalf("new filter version was not resolved: %#v", updated)
	}
}

func TestItemsFromSimpleBuildsStableHierarchy(t *testing.T) {
	probe := &treeProbe{size: image.Pt(20, 12)}
	items := ItemsFromSimple([]SimpleItem{
		NewSimpleItem("readme", "docs", "README"),
		NewSimpleItem("workspace", "", "Workspace"),
		{Key: "docs", ParentKey: "workspace", Label: "Docs", Content: probe, DragDisabled: true},
		NewSimpleItem("orphan", "missing", "Orphan"),
	})
	if len(items) != 2 || items[0].Key != "workspace" || items[1].Key != "orphan" {
		t.Fatalf("roots = %#v", items)
	}
	if len(items[0].Children) != 1 || items[0].Children[0].Key != "docs" || len(items[0].Children[0].Children) != 1 || items[0].Children[0].Children[0].Key != "readme" {
		t.Fatalf("hierarchy = %#v", items[0])
	}
	if items[0].Children[0].Content != probe || !items[0].Children[0].DragDisabled {
		t.Fatalf("simple item properties were not preserved: %#v", items[0].Children[0])
	}
}

func TestItemsFromSimplePromotesCyclesToRoots(t *testing.T) {
	items := ItemsFromSimple([]SimpleItem{
		NewSimpleItem("a", "b", "A"),
		NewSimpleItem("b", "a", "B"),
	})
	if len(items) != 2 || items[0].Key != "a" || items[1].Key != "b" {
		t.Fatalf("cycle roots = %#v", items)
	}
}

func TestTreeFlattenIncludesOnlyExpandedDescendants(t *testing.T) {
	items := treeTestItems()
	collapsed := flattenVisibleItems(items, treeKeySet(nil))
	if got := treeFlatKeys(collapsed); !treeSameKeys(got, []string{"project", "archive"}) {
		t.Fatalf("collapsed keys = %#v", got)
	}
	expanded := flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	want := []string{"project", "src", "main", "ui", "readme", "archive"}
	if got := treeFlatKeys(expanded); !treeSameKeys(got, want) {
		t.Fatalf("expanded keys = %#v, want %#v", got, want)
	}
	if expanded[2].depth != 2 || expanded[2].parentKey != "src" {
		t.Fatalf("main metadata = depth %d parent %q, want 2/src", expanded[2].depth, expanded[2].parentKey)
	}
	if expanded[2].isLast || len(expanded[2].ancestorsLast) != 2 || expanded[2].ancestorsLast[0] || expanded[2].ancestorsLast[1] {
		t.Fatalf("main guide metadata = last %v ancestors %#v", expanded[2].isLast, expanded[2].ancestorsLast)
	}
}

func TestTreeDataVersionCachesFlattenedItems(t *testing.T) {
	state := new(treeState)
	firstItems := []Item{{Key: "root", Label: "First", Children: []Item{{Key: "child", Label: "Child"}}}}
	first := state.resolveVisible(New("tree", "", firstItems).DataVersion(1))
	changedItems := []Item{{Key: "root", Label: "Changed", Children: []Item{{Key: "child", Label: "Child"}}}}
	second := state.resolveVisible(New("tree", "", changedItems).DataVersion(1))
	if len(first) != 1 || second[0].item.Label != "First" || &first[0] != &second[0] {
		t.Fatalf("same-version Tree data was not reused: first %#v second %#v", first, second)
	}
	third := state.resolveVisible(New("tree", "", changedItems).DataVersion(2))
	if third[0].item.Label != "Changed" || &third[0] == &second[0] {
		t.Fatalf("new-version Tree data was not resolved: %#v", third)
	}
	expanded := state.resolveVisible(New("tree", "", changedItems).DataVersion(2).ExpandedKeys([]string{"root"}))
	if got := treeFlatKeys(expanded); !treeSameKeys(got, []string{"root", "child"}) {
		t.Fatalf("expanded Tree cache = %#v", got)
	}
}

func TestTreeRetainsInteractionStateOnlyForViewportRows(t *testing.T) {
	items := make([]Item, 500)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	widget := New("large", "", items).DataVersion(1).MaxHeight(80)
	layoutTreeFrame(ctx, router, widget, time.Unix(1, 0))
	layoutTreeFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)))
	state := treePeekState(ctx, "large")
	if len(state.items) == 0 || len(state.items) >= 30 {
		t.Fatalf("retained Tree row states = %d, want viewport rows only", len(state.items))
	}
	if len(state.keyFilters) == 0 || len(state.keyFilters) >= 300 {
		t.Fatalf("Tree key filters = %d, want viewport rows only", len(state.keyFilters))
	}
}

func BenchmarkTreeDataVersion(b *testing.B) {
	items := make([]Item, 10_000)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
	}
	widget := New("tree", "", items)
	for _, benchmark := range []struct {
		name   string
		widget Widget
	}{
		{name: "unversioned", widget: widget},
		{name: "versioned", widget: widget.DataVersion(1)},
		{name: "versioned-filter", widget: widget.DataVersion(1).Filter("Item")},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			state := new(treeState)
			b.ReportAllocs()
			for b.Loop() {
				visible := state.resolveVisible(benchmark.widget)
				runtime.KeepAlive(visible)
			}
		})
	}
}

func BenchmarkTreeLargeKeySets(b *testing.B) {
	keys := make([]string, 10_000)
	for index := range keys {
		keys[index] = fmt.Sprintf("item-%d", index)
	}
	target := keys[len(keys)-1]
	for _, benchmark := range []struct {
		name string
		tree Widget
		run  func(Widget) bool
	}{
		{name: "selected", tree: Widget{selectionMode: SelectionMultiple, selectedKeys: keys}, run: func(tree Widget) bool { return tree.itemSelected(target) }},
		{name: "disabled", tree: Widget{disabledKeys: keys}, run: func(tree Widget) bool { return tree.itemDisabled(Item{Key: target}) }},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			state := new(treeState)
			b.ReportAllocs()
			for b.Loop() {
				benchmark.tree.selectedKeySet = state.selectedKeys.Resolve(benchmark.tree.selectedKeys)
				benchmark.tree.disabledKeySet = state.disabledKeys.Resolve(benchmark.tree.disabledKeys)
				runtime.KeepAlive(benchmark.run(benchmark.tree))
			}
		})
	}
}

func BenchmarkTreeLargeLayout(b *testing.B) {
	items := make([]Item, 10_000)
	selected := make([]string, len(items))
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: "Item"}
		selected[index] = items[index].Key
	}
	tree := New("large", "", items).DataVersion(1).SelectionMode(SelectionMultiple).SelectedKeys(selected)
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	var router input.Router
	b.ReportAllocs()
	for b.Loop() {
		layoutTreeFrame(ctx, &router, tree, time.Time{})
	}
}

func TestTreeRejectsEmptyAndDuplicateKeysAcrossDepths(t *testing.T) {
	tests := []struct {
		name  string
		items []Item
	}{
		{"empty", []Item{{Label: "Missing"}}},
		{"duplicate", []Item{{Key: "same", Children: []Item{{Key: "same"}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected invalid Tree keys to panic")
				}
			}()
			state := new(treeState)
			state.checkItems(test.items)
		})
	}
}

func TestTreeToggleKeysDoesNotMutateAndDeduplicates(t *testing.T) {
	original := []string{"project", "project", "archive"}
	next := toggleTreeKey(original, "project")
	if !treeSameKeys(original, []string{"project", "project", "archive"}) {
		t.Fatalf("original keys mutated: %#v", original)
	}
	if !treeSameKeys(next, []string{"archive"}) {
		t.Fatalf("collapsed keys = %#v, want archive", next)
	}
	next = toggleTreeKey(original, "src")
	if !treeSameKeys(next, []string{"project", "archive", "src"}) {
		t.Fatalf("expanded keys = %#v", next)
	}
}

func TestTreeRowClickSelectsAndRunsAction(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	state := new(treeState)
	state.beginFrame()
	state.item("readme").clickable.Click()
	state.endFrame()
	treeSetState(ctx, "files", state)
	var changed, action string
	New("files", "", treeTestItems()).
		ExpandedKeys([]string{"project"}).
		OnChange(func(key string) { changed = key }).
		OnAction(func(key string) { action = key }).
		Layout(ctx, treeLayoutContext(nil, image.Pt(360, 240), time.Time{}))
	if changed != "readme" || action != "readme" {
		t.Fatalf("click callbacks = changed %q action %q, want readme/readme", changed, action)
	}
}

func TestTreeRowClickExpansionIsOptional(t *testing.T) {
	for _, test := range []struct {
		name string
		tree Widget
		want []string
	}{
		{name: "default", tree: New("files", "", treeTestItems())},
		{name: "enabled", tree: New("files", "", treeTestItems()).ExpandOnRowClick(true), want: []string{"project"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := treeTestContext(nil, locale.LanguageEnglish)
			state := new(treeState)
			state.beginFrame()
			state.item("project").clickable.Click()
			state.endFrame()
			treeSetState(ctx, "files", state)

			selected := ""
			var expanded []string
			test.tree.
				OnChange(func(key string) { selected = key }).
				OnExpandedChange(func(keys []string) { expanded = keys }).
				Layout(ctx, treeLayoutContext(nil, image.Pt(360, 240), time.Time{}))
			if selected != "project" || !treeSameKeys(expanded, test.want) {
				t.Fatalf("row click = selected %q expanded %#v, want project/%#v", selected, expanded, test.want)
			}
		})
	}
}

func TestTreeLeafRowClickDoesNotRequestExpansion(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	state := new(treeState)
	state.beginFrame()
	state.item("readme").clickable.Click()
	state.endFrame()
	treeSetState(ctx, "files", state)

	called := false
	New("files", "", treeTestItems()).
		ExpandedKeys([]string{"project"}).
		ExpandOnRowClick(true).
		OnExpandedChange(func([]string) { called = true }).
		Layout(ctx, treeLayoutContext(nil, image.Pt(360, 240), time.Time{}))
	if called {
		t.Fatal("leaf row click requested expansion")
	}
}

func TestTreeToggleClickOnlyChangesExpansion(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	state := new(treeState)
	state.beginFrame()
	state.item("project").toggle.Click()
	state.endFrame()
	treeSetState(ctx, "files", state)
	var selected string
	var expanded []string
	New("files", "", treeTestItems()).
		OnChange(func(key string) { selected = key }).
		OnExpandedChange(func(keys []string) { expanded = keys }).
		Layout(ctx, treeLayoutContext(nil, image.Pt(360, 240), time.Time{}))
	if selected != "" {
		t.Fatalf("toggle selected %q", selected)
	}
	if !treeSameKeys(expanded, []string{"project"}) {
		t.Fatalf("expanded = %#v, want project", expanded)
	}
}

func TestTreeAsyncBranchRequestsLoadAndExpansion(t *testing.T) {
	items := []Item{{Key: "remote", Label: "Remote", ChildrenState: ChildrenUnloaded}}
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	state := new(treeState)
	state.beginFrame()
	state.item("remote").toggle.Click()
	state.endFrame()
	treeSetState(ctx, "async", state)
	loaded := ""
	var expanded []string
	New("async", "", items).
		OnLoadChildren(func(key string) { loaded = key }).
		OnExpandedChange(func(keys []string) { expanded = keys }).
		Layout(ctx, treeLayoutContext(nil, image.Pt(360, 120), time.Time{}))
	if loaded != "remote" || !treeSameKeys(expanded, []string{"remote"}) {
		t.Fatalf("async toggle = load %q expanded %#v", loaded, expanded)
	}

	loaded = ""
	expanded = nil
	New("async-error", "", []Item{{Key: "remote", Label: "Remote", ChildrenState: ChildrenError, LoadError: "Offline"}}).
		ExpandedKeys([]string{"remote"}).
		OnLoadChildren(func(key string) { loaded = key }).
		OnExpandedChange(func(keys []string) { expanded = keys }).
		requestItemToggle(Item{Key: "remote", ChildrenState: ChildrenError})
	if loaded != "remote" || expanded != nil {
		t.Fatalf("async retry = load %q expanded %#v", loaded, expanded)
	}
}

func TestTreeLoadEventIncludesRetryAndExpansion(t *testing.T) {
	var events []LoadEvent
	var legacy []string
	initial := New("async", "", nil).
		OnLoadChildren(func(key string) { legacy = append(legacy, key) }).
		OnLoadChildrenEvent(func(event LoadEvent) { events = append(events, event) })
	initial.requestItemToggle(Item{Key: "remote", ChildrenState: ChildrenUnloaded})

	retry := New("async", "", nil).
		ExpandedKeys([]string{"remote"}).
		OnLoadChildrenEvent(func(event LoadEvent) { events = append(events, event) })
	retry.requestItemToggle(Item{Key: "remote", ChildrenState: ChildrenError})

	if !treeSameKeys(legacy, []string{"remote"}) {
		t.Fatalf("legacy load calls = %#v", legacy)
	}
	if len(events) != 2 || events[0] != (LoadEvent{Key: "remote"}) || events[1] != (LoadEvent{Key: "remote", Retry: true, Expanded: true}) {
		t.Fatalf("load events = %#v", events)
	}
}

func TestTreeAsyncStateControlsBranchAndErrorDescription(t *testing.T) {
	if !treeItemExpandable(Item{ChildrenState: ChildrenUnloaded}) || !treeItemExpandable(Item{ChildrenState: ChildrenLoading}) || !treeItemExpandable(Item{ChildrenState: ChildrenError}) {
		t.Fatal("an asynchronous Tree item was not expandable")
	}
	item := Item{Description: "Remote files", ChildrenState: ChildrenError, LoadError: "Connection failed"}
	if got := treeItemDescription(item); got != "Connection failed" {
		t.Fatalf("error description = %q", got)
	}
	item.ChildrenState = ChildrenLoaded
	if got := treeItemDescription(item); got != "Remote files" {
		t.Fatalf("loaded description = %q", got)
	}
}

func TestTreeToggleHoverHighlightsRow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := treeTestContext(&activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	tree := New("files", "", treeTestItems())
	start := time.Unix(1, 0)
	layoutTreeFrame(ctx, router, tree, start)

	router.Queue(pointer.Event{
		Kind:     pointer.Move,
		Source:   pointer.Mouse,
		Position: f32.Pt(20, 20),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(time.Millisecond))
	item := treePeekState(ctx, "files").items["project"]
	if !item.toggle.Hovered() || item.clickable.Hovered() {
		t.Fatalf("hover state = toggle %v row %v", item.toggle.Hovered(), item.clickable.Hovered())
	}
	style := treeItemStyleFor(&activeTheme, false, item.toggle.Hovered() || item.clickable.Hovered(), false)
	if style.background != activeTheme.Palette.SurfaceTertiary {
		t.Fatalf("toggle hover background = %#v, want %#v", style.background, activeTheme.Palette.SurfaceTertiary)
	}
}

func TestTreeDisabledNodeIgnoresClickAndToggle(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	state := new(treeState)
	state.beginFrame()
	itemState := state.item("project")
	itemState.clickable.Click()
	itemState.toggle.Click()
	state.endFrame()
	treeSetState(ctx, "files", state)
	called := false
	New("files", "", treeTestItems()).
		DisabledKeys([]string{"project"}).
		OnChange(func(string) { called = true }).
		OnExpandedChange(func([]string) { called = true }).
		Layout(ctx, treeLayoutContext(nil, image.Pt(360, 240), time.Time{}))
	if called {
		t.Fatal("disabled node handled an interaction")
	}
}

func TestTreeSelectionNoneOnlyRunsAction(t *testing.T) {
	var changed, action string
	tree := New("files", "", treeTestItems()).
		SelectionMode(SelectionNone).
		OnChange(func(key string) { changed = key }).
		OnAction(func(key string) { action = key })
	tree.activateWithModifiers(nil, nil, "project", 0)
	if changed != "" || action != "project" {
		t.Fatalf("callbacks = changed %q action %q", changed, action)
	}
}

func TestTreeCascadeCheckedKeysDeriveParentsAndHalfChecked(t *testing.T) {
	checked, half := deriveCascadeCheckedKeys(treeTestItems(), treeKeySet([]string{"main"}))
	if !treeSameKeys(checked, []string{"main"}) {
		t.Fatalf("checked = %#v, want only main", checked)
	}
	if !treeSameKeys(half, []string{"project", "src"}) {
		t.Fatalf("half = %#v, want project and src half-checked", half)
	}

	checked, half = deriveCascadeCheckedKeys(treeTestItems(), treeKeySet([]string{"main", "ui", "readme"}))
	if !treeSameKeys(checked, []string{"main", "project", "readme", "src", "ui"}) {
		t.Fatalf("checked = %#v, want project and src promoted to checked", checked)
	}
	if len(half) != 0 {
		t.Fatalf("half = %#v, want none", half)
	}
}

func TestTreeCascadeCheckboxesTreatDisabledNodesAsBoundaries(t *testing.T) {
	items := []Item{{
		Key: "root", Label: "Root", Children: []Item{
			{Key: "enabled", Label: "Enabled"},
			{Key: "locked", Label: "Locked", DisableCheckbox: true, Children: []Item{{Key: "nested", Label: "Nested"}}},
			{Key: "disabled", Label: "Disabled", Disabled: true, Children: []Item{{Key: "deep", Label: "Deep"}}},
		},
	}}
	isDisabled := func(item Item) bool { return item.Disabled || item.DisableCheckbox }
	checked, half := deriveCascadeCheckedKeysWith(items, treeKeySet([]string{"root"}), isDisabled)
	if !treeSameKeys(checked, []string{"enabled", "root"}) || len(half) != 0 {
		t.Fatalf("parent cascade = checked %#v half %#v", checked, half)
	}
	checked, half = deriveCascadeCheckedKeysWith(items, treeKeySet([]string{"enabled"}), isDisabled)
	if !treeSameKeys(checked, []string{"enabled", "root"}) || len(half) != 0 {
		t.Fatalf("enabled child cascade = checked %#v half %#v", checked, half)
	}
	checked, half = deriveCascadeCheckedKeysWith(items, treeKeySet([]string{"nested", "deep"}), isDisabled)
	if !treeSameKeys(checked, []string{"deep", "nested"}) || len(half) != 0 {
		t.Fatalf("isolated checks = checked %#v half %#v", checked, half)
	}
}

func TestTreeCheckRequestDoesNotModifyDisabledBranches(t *testing.T) {
	items := []Item{{
		Key: "root", Label: "Root", Children: []Item{
			{Key: "enabled", Label: "Enabled"},
			{Key: "locked", Label: "Locked", DisableCheckbox: true, Children: []Item{{Key: "nested", Label: "Nested"}}},
		},
	}}
	var nextChecked, nextHalf []string
	tree := New("checks", "", items).
		Checkable(true).
		OnCheckedChange(func(checked, half []string) {
			nextChecked, nextHalf = checked, half
		})
	state := new(treeState)
	tree.checkedKeySet = state.checkedKeys.Resolve(nil)
	tree.resolveCheckState(state)
	tree.requestCheck("root", true)
	if !treeSameKeys(nextChecked, []string{"enabled", "root"}) || len(nextHalf) != 0 {
		t.Fatalf("check request = checked %#v half %#v", nextChecked, nextHalf)
	}
}

func TestTreeDisabledPreservesCascadeCheckState(t *testing.T) {
	tree := New("checks", "", treeTestItems()).
		Checkable(true).
		CheckedKeys([]string{"main"}).
		Disabled(true)
	state := new(treeState)
	tree.checkedKeySet = state.checkedKeys.Resolve(tree.checkedKeys)
	tree.resolveCheckState(state)
	if !treeSameKeys(tree.resolvedChecked, []string{"main"}) || !treeSameKeys(tree.resolvedHalf, []string{"project", "src"}) {
		t.Fatalf("disabled Tree cascade = checked %#v half %#v", tree.resolvedChecked, tree.resolvedHalf)
	}
}

func TestTreeDisabledKeysStopCascadeCheckState(t *testing.T) {
	items := []Item{{Key: "root", Label: "Root", Children: []Item{
		{Key: "enabled", Label: "Enabled"},
		{Key: "locked", Label: "Locked"},
	}}}
	tree := New("checks", "", items).
		Checkable(true).
		CheckedKeys([]string{"root"}).
		DisabledKeys([]string{"locked"})
	state := new(treeState)
	tree.checkedKeySet = state.checkedKeys.Resolve(tree.checkedKeys)
	tree.disabledKeySet = state.disabledKeys.Resolve(tree.disabledKeys)
	tree.resolveCheckState(state)
	if !treeSameKeys(tree.resolvedChecked, []string{"enabled", "root"}) || len(tree.resolvedHalf) != 0 {
		t.Fatalf("disabled keys cascade = checked %#v half %#v", tree.resolvedChecked, tree.resolvedHalf)
	}
}

func TestTreeRowClickExpandsHonorsExpandAction(t *testing.T) {
	clickTree := New("files", "", treeTestItems()).ExpandOnRowClick(true)
	if !clickTree.rowClickExpands(1) {
		t.Fatal("single click should expand by default")
	}
	doubleTree := New("files", "", treeTestItems()).ExpandAction(ExpandActionDoubleClick)
	if doubleTree.rowClickExpands(1) {
		t.Fatal("single click should not expand in double-click mode")
	}
	if !doubleTree.rowClickExpands(2) {
		t.Fatal("double click should expand in double-click mode")
	}
}

func TestTreeAllowEmptySelection(t *testing.T) {
	changed := "not-called"
	tree := New("files", "project", treeTestItems()).
		AllowEmptySelection().
		OnChange(func(key string) { changed = key })
	tree.activateWithModifiers(nil, nil, "project", 0)
	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
}

func TestTreeMultipleSelectionUsesDesktopModifiers(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project"}))
	state := new(treeState)
	state.selectionAnchor = "project"
	selected := []string{"project"}
	tree := func() Widget {
		return New("files", "", treeTestItems()).
			SelectionMode(SelectionMultiple).
			SelectedKeys(selected).
			OnSelectionChange(func(keys []string) { selected = keys })
	}

	tree().activateWithModifiers(state, visible, "readme", key.ModShift)
	if !treeSameKeys(selected, []string{"project", "src", "readme"}) {
		t.Fatalf("shift range = %#v", selected)
	}
	tree().activateWithModifiers(state, visible, "src", key.ModCtrl)
	if !treeSameKeys(selected, []string{"project", "readme"}) {
		t.Fatalf("ctrl toggle = %#v", selected)
	}
	tree().activateWithModifiers(state, visible, "archive", 0)
	if !treeSameKeys(selected, []string{"archive"}) {
		t.Fatalf("plain selection = %#v", selected)
	}
}

func TestTreeSelectedDragKeysPreserveOrderAndExcludeDescendants(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project", "src"}))
	tree := New("files", "", treeTestItems()).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"ui", "project", "main", "archive"})
	if got := tree.selectedDragKeys(visible); !treeSameKeys(got, []string{"project", "archive"}) {
		t.Fatalf("drag keys = %#v, want project/archive", got)
	}
	items := treeTestItems()
	items[0].DragDisabled = true
	visible = flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	tree = New("files", "", items).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"project", "main", "archive"})
	if got := tree.selectedDragKeys(visible); !treeSameKeys(got, []string{"main", "archive"}) {
		t.Fatalf("drag keys with disabled parent = %#v, want main/archive", got)
	}
}

func TestTreeKeyboardRightExpandsThenEntersChild(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	expanded := []string{}
	widget := func() Widget {
		return New("files", "project", treeTestItems()).
			ExpandedKeys(expanded).
			OnExpandedChange(func(keys []string) { expanded = keys })
	}
	start := time.Unix(1, 0)
	layoutTreeFrame(ctx, router, widget(), start)
	state := treePeekState(ctx, "files")
	router.Source().Execute(key.FocusCmd{Tag: &state.items["project"].clickable})
	layoutTreeFrame(ctx, router, widget(), start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutTreeFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	if !treeSameKeys(expanded, []string{"project"}) {
		t.Fatalf("Right on collapsed branch expanded %#v, want project", expanded)
	}

	layoutTreeFrame(ctx, router, widget(), start.Add(3*time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutTreeFrame(ctx, router, widget(), start.Add(4*time.Millisecond))
	if !router.Source().Focused(&state.items["src"].clickable) {
		t.Fatal("Right on expanded branch did not focus first child")
	}
}

func TestTreeKeyboardLeftReturnsToParentThenCollapses(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	expanded := []string{"project", "src"}
	widget := func() Widget {
		return New("files", "main", treeTestItems()).
			ExpandedKeys(expanded).
			OnExpandedChange(func(keys []string) { expanded = keys })
	}
	start := time.Unix(2, 0)
	layoutTreeFrame(ctx, router, widget(), start)
	state := treePeekState(ctx, "files")
	router.Source().Execute(key.FocusCmd{Tag: &state.items["main"].clickable})
	layoutTreeFrame(ctx, router, widget(), start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameLeftArrow, State: key.Press})
	layoutTreeFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	if !router.Source().Focused(&state.items["src"].clickable) {
		t.Fatal("Left on leaf did not focus its parent")
	}
	router.Queue(key.Event{Name: key.NameLeftArrow, State: key.Press})
	layoutTreeFrame(ctx, router, widget(), start.Add(3*time.Millisecond))
	if treeContainsKey(expanded, "src") {
		t.Fatalf("Left on expanded branch did not collapse src: %#v", expanded)
	}
}

func TestTreeKeyboardEnterSelectsFocusedNode(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	selected := ""
	widget := func() Widget {
		return New("files", selected, treeTestItems()).
			OnChange(func(key string) { selected = key })
	}
	start := time.Unix(3, 0)
	layoutTreeFrame(ctx, router, widget(), start)
	state := treePeekState(ctx, "files")
	router.Source().Execute(key.FocusCmd{Tag: &state.items["archive"].clickable})
	layoutTreeFrame(ctx, router, widget(), start.Add(time.Millisecond))
	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	layoutTreeFrame(ctx, router, widget(), start.Add(2*time.Millisecond))
	if selected != "archive" {
		t.Fatalf("selected = %q, want archive", selected)
	}
}

func TestTreePointerFocusThenF2StartsAndCommitsRename(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	items := treeTestItems()
	items[1].Renamable = true
	renamedKey, renamedLabel := "", ""
	widgetValue := New("files", "archive", items).OnRename(func(key, label string) {
		renamedKey, renamedLabel = key, label
	})
	start := time.Unix(8, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	state := treePeekState(ctx, "files")
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	if !router.Source().Focused(&state.items["archive"].clickable) {
		t.Fatal("pointer press did not focus the renamable row")
	}
	router.Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameF2, State: key.Press})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(3*time.Millisecond))
	if state.renameKey != "archive" || state.renameEditor.Text() != "Archive" {
		t.Fatalf("rename state = key %q text %q", state.renameKey, state.renameEditor.Text())
	}
	state.renameEditor.SetText("History")
	state.finishRename(widgetValue, true)
	if renamedKey != "archive" || renamedLabel != "History" || state.renameKey != "" {
		t.Fatalf("rename result = %q/%q active %q", renamedKey, renamedLabel, state.renameKey)
	}
}

func TestTreeContextMenuOpensForUnselectedRow(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	selected, contextKey := "", ""
	widgetValue := New("files", selected, treeTestItems()).
		ContextMenu(menu.Menu("actions", []menu.Item{{Key: "open", Label: "Open"}})).
		OnChange(func(key string) { selected = key }).
		OnContextMenu(func(key string) { contextKey = key })
	start := time.Unix(10, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	state := treePeekState(ctx, "files")
	if selected != "" || contextKey != "archive" {
		t.Fatalf("context menu row = selected %q target %q, want empty/archive", selected, contextKey)
	}
	if !router.Source().Focused(&state.items["archive"].clickable) {
		t.Fatal("context menu did not focus its Tree row")
	}
}

func TestTreeContextMenuKeyDoesNotCollideWithUserKey(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	gtx := treeLayoutContext(nil, image.Pt(360, 240), time.Unix(10, 0))
	tree := New("files", "", treeTestItems()).
		ContextMenu(menu.Menu("actions", []menu.Item{{Key: "open", Label: "Open"}}))
	userMenu := menu.ContextMenu(
		"files-context-menu-archive",
		&treeProbe{size: image.Pt(20, 20)},
		menu.Menu("user-actions", []menu.Item{{Key: "open", Label: "Open"}}),
	)

	frame.BeginFrame(ctx)
	tree.Layout(ctx, gtx)
	userMenu.Layout(ctx, gtx)
	frame.EndFrame(ctx)
}

func TestTreeContextMenuPreservesExistingMultipleSelection(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	selected := []string{"project", "archive"}
	selectionChanged := false
	contextKey := ""
	widgetValue := New("files", "", treeTestItems()).
		SelectionMode(SelectionMultiple).
		SelectedKeys(selected).
		ContextMenu(menu.Menu("actions", []menu.Item{{Key: "open", Label: "Open"}})).
		OnSelectionChange(func(keys []string) {
			selectionChanged = true
			selected = keys
		}).
		OnContextMenu(func(key string) { contextKey = key })
	start := time.Unix(11, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonSecondary, Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	if selectionChanged || !treeSameKeys(selected, []string{"project", "archive"}) || contextKey != "archive" {
		t.Fatalf("selected context row = changed %v selected %#v target %q", selectionChanged, selected, contextKey)
	}
}

func TestTreeContextMenuOpensFromShiftF10(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	contextKey := ""
	widgetValue := New("files", "", treeTestItems()).
		ContextMenu(menu.Menu("actions", []menu.Item{{Key: "open", Label: "Open"}})).
		OnContextMenu(func(key string) { contextKey = key })
	start := time.Unix(12, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	state := treePeekState(ctx, "files")
	router.Source().Execute(key.FocusCmd{Tag: &state.items["archive"].clickable})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameF10, Modifiers: key.ModShift, State: key.Press})
	layoutTreeFrame(ctx, router, widgetValue, start.Add(2*time.Millisecond))
	if contextKey != "archive" {
		t.Fatalf("Shift+F10 context target = %q, want archive", contextKey)
	}
}

func TestTreeRenameRejectsEmptyAndHonorsCancel(t *testing.T) {
	called := false
	widgetValue := New("files", "", nil).OnRename(func(string, string) { called = true })
	state := new(treeState)
	state.beginRename(Item{Key: "file", Label: "File"})
	state.renameEditor.SetText("   ")
	state.finishRename(widgetValue, true)
	state.beginRename(Item{Key: "file", Label: "File"})
	state.renameEditor.SetText("Changed")
	state.finishRename(widgetValue, false)
	if called {
		t.Fatal("empty or cancelled rename invoked OnRename")
	}
}

func TestTreeRenameRequestStartsOncePerRevision(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	items := treeTestItems()
	items[1].Renamable = true
	widgetValue := New("files", "archive", items).
		OnRename(func(string, string) {}).
		RequestRename("archive", 1)
	start := time.Unix(13, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	state := treePeekState(ctx, "files")
	if state.renameKey != "archive" || state.renameEditor.Text() != "Archive" || !router.Source().Focused(&state.renameEditor) {
		t.Fatalf("rename request = key %q text %q focused %v", state.renameKey, state.renameEditor.Text(), router.Source().Focused(&state.renameEditor))
	}
	state.renameEditor.SetText("Editing")
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	if state.renameEditor.Text() != "Editing" {
		t.Fatalf("same revision restarted editor with %q", state.renameEditor.Text())
	}
	state.finishRename(widgetValue, false)
	layoutTreeFrame(ctx, router, widgetValue.RequestRename("archive", 2), start.Add(2*time.Millisecond))
	if state.renameKey != "archive" || state.renameEditor.Text() != "Archive" {
		t.Fatalf("new revision did not restart rename: key %q text %q", state.renameKey, state.renameEditor.Text())
	}
	state.finishRename(widgetValue, false)
	items[0].Renamable = true
	otherTarget := New("files", "project", items).
		OnRename(func(string, string) {}).
		RequestRename("project", 2)
	layoutTreeFrame(ctx, router, otherTarget, start.Add(3*time.Millisecond))
	if state.renameKey != "project" || state.renameEditor.Text() != "Project" {
		t.Fatalf("same revision with a new key did not start rename: key %q text %q", state.renameKey, state.renameEditor.Text())
	}
}

func TestTreeRenameRequestWaitsForHiddenTarget(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	items := treeTestItems()
	items[0].Children[0].Children[0].Renamable = true
	widgetValue := New("files", "", items).
		OnRename(func(string, string) {}).
		RequestRename("main", 1)
	start := time.Unix(14, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	state := treePeekState(ctx, "files")
	if state.renameKey != "" || state.renameRequestReady {
		t.Fatalf("hidden request was consumed: key %q ready %v", state.renameKey, state.renameRequestReady)
	}
	layoutTreeFrame(ctx, router, widgetValue.ExpandedKeys([]string{"project", "src"}), start.Add(time.Millisecond))
	if state.renameKey != "main" || state.renameEditor.Text() != "main.go" {
		t.Fatalf("visible request did not start rename: key %q text %q", state.renameKey, state.renameEditor.Text())
	}
}

func TestTreeKeyboardNavigationSkipsDisabledNodes(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project"}))
	tree := New("files", "", treeTestItems()).DisabledKeys([]string{"readme"})
	project := treeVisibleIndex(visible, "project")
	next, ok := treeMoveVisible(visible, tree, project, 1)
	if !ok || visible[next].item.Key != "src" {
		t.Fatalf("next enabled = %d/%v key %q, want src", next, ok, visible[next].item.Key)
	}
	readme := treeVisibleIndex(visible, "readme")
	next, ok = treeMoveVisible(visible, tree, readme-1, 1)
	if !ok || visible[next].item.Key != "archive" {
		t.Fatalf("disabled readme was not skipped: %d/%v key %q", next, ok, visible[next].item.Key)
	}
}

func TestTreeKeyboardActiveIndexSkipsHiddenAndDisabledSelections(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project"}))
	tree := New("files", "", treeTestItems()).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"project", "main", "readme"}).
		DisabledKeys([]string{"readme"})
	index := tree.keyboardActiveIndex(visible)
	if index < 0 || visible[index].item.Key != "project" {
		t.Fatalf("keyboard active index = %d, want visible project", index)
	}
}

func TestTreeTypeaheadSkipsDisabledAndWraps(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project"}))
	tree := New("files", "", treeTestItems()).DisabledKeys([]string{"archive"})
	index, ok := treeTypeaheadIndex(visible, tree, treeVisibleIndex(visible, "readme"), "p")
	if !ok || visible[index].item.Key != "project" {
		t.Fatalf("typeahead p = %d/%v, want project", index, ok)
	}
	if _, ok := treeTypeaheadIndex(visible, tree, 0, "a"); ok {
		t.Fatal("typeahead returned disabled archive")
	}
}

func TestTreeItemStylesFollowTheme(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	selected := treeItemStyleFor(&activeTheme, true, false, false)
	if selected.background != activeTheme.Palette.AccentSoft || selected.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("selected style = %#v/%#v", selected.background, selected.foreground)
	}
	hovered := treeItemStyleFor(&activeTheme, false, true, false)
	if hovered.background != activeTheme.Palette.SurfaceTertiary {
		t.Fatalf("hover background = %#v, want SurfaceTertiary", hovered.background)
	}
}

func TestTreeSmallSizeUsesThemeTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Tree.SmallRowRadius = 9
	activeTheme.Components.Tree.SmallItemTextSize = 15
	tokens := treeTokensFor(&activeTheme, SizeSmall)
	if tokens.RowRadius != 9 || tokens.ItemTextSize != 15 {
		t.Fatalf("small Tree tokens = radius %v text %v, want 9/15", tokens.RowRadius, tokens.ItemTextSize)
	}
}

func TestTreeSmallUsesCompactFileTreeGeometry(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := treeTokensFor(&activeTheme, SizeSmall)
	if tokens.Padding != 2 || tokens.Gap != 1 || tokens.RowHeight != 24 || tokens.DescriptionRowHeight != 40 || tokens.RowRadius != 4 {
		t.Fatalf("small Tree geometry = %+v", tokens)
	}
	if tokens.RowPaddingX != 4 || tokens.RowPaddingY != 2 || tokens.Indent != 12 || tokens.ChevronSlotSize != 16 || tokens.ChevronIconSize != 12 || tokens.ContentGap != 5 {
		t.Fatalf("small Tree hierarchy geometry = %+v", tokens)
	}
	if tokens.ItemTextSize != 13 || tokens.ItemDescriptionSize != 11 {
		t.Fatalf("small Tree typography = %+v", tokens)
	}

	ctx := treeTestContext(&activeTheme, locale.LanguageEnglish)
	gtx := treeLayoutContext(nil, image.Pt(320, 200), time.Time{})
	medium := New("medium", "", []Item{{Key: "one", Label: "One"}}).Layout(ctx, gtx)
	small := New("small", "", []Item{{Key: "one", Label: "One"}}).Size(SizeSmall).Layout(ctx, gtx)
	if small.Size.Y >= medium.Size.Y {
		t.Fatalf("small Tree height = %d, want less than medium %d", small.Size.Y, medium.Size.Y)
	}
}

func TestTreeGuideSegmentsFollowExpandedBranches(t *testing.T) {
	root := flatItem{item: Item{Children: []Item{{Key: "child"}}}}
	if got := treeGuideSegments(root, 4, 12, 16, 5, 24, true, false); len(got) != 1 || got[0] != (treeGuideSegment{from: image.Pt(12, 20), to: image.Pt(12, 24), extend: true}) {
		t.Fatalf("expanded root guides = %#v", got)
	}

	child := flatItem{depth: 1}
	if got := treeGuideSegments(child, 4, 12, 16, 5, 24, false, false); len(got) != 1 || got[0] != (treeGuideSegment{from: image.Pt(12, 0), to: image.Pt(12, 24), extend: true}) {
		t.Fatalf("vertical child guides = %#v", got)
	}
	if got := treeGuideSegments(child, 4, 12, 16, 5, 24, false, true); len(got) != 2 || got[1] != (treeGuideSegment{from: image.Pt(12, 12), to: image.Pt(37, 12)}) {
		t.Fatalf("continuing child guides = %#v", got)
	}
	child.isLast = true
	if got := treeGuideSegments(child, 4, 12, 16, 5, 24, false, false); len(got) != 1 || got[0] != (treeGuideSegment{from: image.Pt(12, 0), to: image.Pt(12, 24)}) {
		t.Fatalf("last vertical-only child guides = %#v", got)
	}
	if got := treeGuideSegments(child, 4, 12, 16, 5, 24, false, true); len(got) != 2 || got[0] != (treeGuideSegment{from: image.Pt(12, 0), to: image.Pt(12, 12)}) {
		t.Fatalf("last child guides = %#v", got)
	}

	nested := flatItem{depth: 2, ancestorsLast: []bool{false, false}}
	if got := treeGuideSegments(nested, 4, 12, 16, 5, 24, false, true); len(got) != 3 || got[0].from.X != 12 || got[1].from.X != 24 || got[2].to.X != 49 {
		t.Fatalf("nested guides = %#v", got)
	}
	nested.ancestorsLast = []bool{true, false}
	if got := treeGuideSegments(nested, 4, 12, 16, 5, 24, false, true); len(got) != 3 || got[0].from.X != 12 {
		t.Fatalf("continued ancestor guides = %#v", got)
	}
	nested.ancestorsLast = []bool{false, true}
	if got := treeGuideSegments(nested, 4, 12, 16, 5, 24, false, true); len(got) != 2 || got[0].from.X != 24 {
		t.Fatalf("completed ancestor guides = %#v", got)
	}

	branch := flatItem{item: Item{Children: []Item{{Key: "leaf"}}}, depth: 1}
	if got := treeGuideSegments(branch, 4, 12, 16, 5, 24, false, true); len(got) != 2 || got[1].to.X != 24 {
		t.Fatalf("branch connector guides = %#v", got)
	}
}

func TestTreeConnectorToggleBarsAreCentered(t *testing.T) {
	horizontal, vertical := treeConnectorToggleBars(14, 1)
	if horizontal != image.Rect(3, 7, 11, 8) {
		t.Fatalf("horizontal connector toggle bar = %v", horizontal)
	}
	if vertical != image.Rect(7, 3, 8, 11) {
		t.Fatalf("vertical connector toggle bar = %v", vertical)
	}
}

func TestTreeGuideDashesMatchReferencePattern(t *testing.T) {
	want := [][2]int{{0, 1}, {3, 4}, {6, 7}}
	got := treeGuideDashes(0, 8, 1, 3)
	if len(got) != len(want) {
		t.Fatalf("guide dashes = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("guide dash %d = %v, want %v", index, got[index], want[index])
		}
	}
}

func TestTreeDropTargetUsesRowQuarters(t *testing.T) {
	visible := []flatItem{{item: Item{Key: "one"}}, {item: Item{Key: "two", AcceptsChildren: true}}}
	heights := []int{24, 24}
	tests := []struct {
		y        float32
		position DropPosition
	}{
		{25, DropBefore},
		{32, DropInside},
		{44, DropAfter},
	}
	for _, test := range tests {
		target := treeDropTargetAt(visible, 0, test.y, heights, 1)
		if target.key != "two" || target.position != test.position {
			t.Fatalf("drop at %.0f = %#v, want two/%v", test.y, target, test.position)
		}
	}
	if target := treeDropTargetAt(visible, 0, 24, heights, 1); target.key != "" {
		t.Fatalf("drop in row gap = %#v, want no target", target)
	}
}

func TestTreeLeafDropTargetUsesRowHalves(t *testing.T) {
	visible := []flatItem{{item: Item{Key: "source"}}, {item: Item{Key: "file"}}}
	heights := []int{24, 24}
	for _, test := range []struct {
		y        float32
		position DropPosition
	}{
		{27, DropBefore},
		{38, DropAfter},
	} {
		target := treeDropTargetAt(visible, 0, test.y, heights, 1)
		if target.key != "file" || target.position != test.position {
			t.Fatalf("leaf drop at %.0f = %#v, want file/%v", test.y, target, test.position)
		}
	}
}

func TestTreeDragViewportAndAutoScrollEdges(t *testing.T) {
	heights := []int{24, 24, 24}
	position := layout.Position{First: 1, Offset: 4, Count: 2}
	if got := treeDragViewportY(heights, 1, position, 2, 12); got != 33 {
		t.Fatalf("drag viewport y = %v, want 33", got)
	}
	if got := treeAutoScrollDirection(10, 200, 24); got != -1 {
		t.Fatalf("top auto-scroll direction = %d", got)
	}
	if got := treeAutoScrollDirection(100, 200, 24); got != 0 {
		t.Fatalf("middle auto-scroll direction = %d", got)
	}
	if got := treeAutoScrollDirection(190, 200, 24); got != 1 {
		t.Fatalf("bottom auto-scroll direction = %d", got)
	}
}

func TestTreeDragAutoScrollAdvancesList(t *testing.T) {
	state := new(treeState)
	state.list.Axis = layout.Vertical
	start := time.Unix(9, 0)
	layoutList := func(now time.Time) {
		gtx := treeLayoutContext(nil, image.Pt(240, 72), now)
		state.list.Layout(gtx, 20, func(gtx layout.Context, _ int) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(240, 24)}
		})
	}
	layoutList(start)
	gtx := treeLayoutContext(nil, image.Pt(240, 72), start.Add(time.Millisecond))
	state.updateDragScroll(gtx, 70, 72, 24, 20)
	layoutList(start.Add(2 * time.Millisecond))
	if state.list.Position.First == 0 && state.list.Position.Offset == 0 {
		t.Fatal("drag auto-scroll did not advance the list")
	}
}

func TestTreeDragGeometryCachesVersionedRowsAndTracksOffscreenSources(t *testing.T) {
	items := make([]Item, 128)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: fmt.Sprintf("Item %d", index)}
	}
	tree := New("files", "", items).DataVersion(1)
	visible := flattenVisibleItems(items, nil)
	state := new(treeState)
	geometry := state.dragGeometryFor(tree, visible, 24, 40, 1)
	firstHeight := &geometry.heights[0]
	if again := state.dragGeometryFor(tree, visible, 24, 40, 1); &again.heights[0] != firstHeight {
		t.Fatal("versioned drag geometry was rebuilt without a data change")
	}

	sourceIndex := 2
	targetIndex := 96
	pointerY := float32(geometry.starts[targetIndex] - geometry.starts[sourceIndex] + 12)
	target := geometry.dropTargetAt(visible, sourceIndex, pointerY)
	if target.key != "item-96" || target.index != targetIndex {
		t.Fatalf("offscreen drop target = %#v, want item-96", target)
	}
	position := layout.Position{First: 80, Offset: 5, Count: 3}
	wantViewportY := float32(geometry.starts[sourceIndex] - geometry.starts[position.First] - position.Offset)
	if got := geometry.viewportY(position, sourceIndex, 0); got != wantViewportY {
		t.Fatalf("offscreen source viewport y = %v, want %v", got, wantViewportY)
	}
}

func TestTreeDragHoverExpandsOnceAfterDelay(t *testing.T) {
	state := new(treeState)
	start := time.Unix(7, 0)
	if state.updateDragHover(treeLayoutContext(nil, image.Pt(320, 200), start), "folder", true) {
		t.Fatal("drag hover expanded immediately")
	}
	if state.updateDragHover(treeLayoutContext(nil, image.Pt(320, 200), start.Add(treeDragExpandDelay-time.Millisecond)), "folder", true) {
		t.Fatal("drag hover expanded before its delay")
	}
	if !state.updateDragHover(treeLayoutContext(nil, image.Pt(320, 200), start.Add(treeDragExpandDelay)), "folder", true) {
		t.Fatal("drag hover did not expand at its delay")
	}
	if state.updateDragHover(treeLayoutContext(nil, image.Pt(320, 200), start.Add(time.Second)), "folder", true) {
		t.Fatal("drag hover expanded the same target twice")
	}
}

func TestTreeDropRejectsSelfDescendantsAndDisabledItems(t *testing.T) {
	items := treeTestItems()
	items[1].AcceptsChildren = true
	visible := flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	tree := New("files", "", items)
	if treeDropAllowed(tree, visible, []string{"project"}, "project", DropBefore) {
		t.Fatal("Tree allowed dropping an item onto itself")
	}
	if treeDropAllowed(tree, visible, []string{"project"}, "main", DropBefore) {
		t.Fatal("Tree allowed dropping a parent into its descendant")
	}
	if !treeDropAllowed(tree, visible, []string{"main"}, "archive", DropInside) {
		t.Fatal("Tree rejected a valid sibling move")
	}
	if treeDropAllowed(tree.DisabledKeys([]string{"archive"}), visible, []string{"main"}, "archive", DropInside) {
		t.Fatal("Tree allowed dropping onto a disabled item")
	}
	items[1].DropDisabled = true
	visible = flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	if treeDropAllowed(New("files", "", items), visible, []string{"main"}, "archive", DropInside) {
		t.Fatal("Tree allowed dropping onto a drop-disabled item")
	}
	items = treeTestItems()
	items[1].AcceptsChildren = true
	items[0].Children[0].Children[0].DragDisabled = true
	visible = flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	if treeDropAllowed(New("files", "", items), visible, []string{"main"}, "archive", DropInside) {
		t.Fatal("Tree allowed a drag-disabled item as source")
	}
	if treeDropAllowed(tree, visible, []string{"main"}, "readme", DropInside) {
		t.Fatal("Tree allowed dropping inside a leaf")
	}
	if !treeDropAllowed(tree, visible, []string{"ui"}, "readme", DropBefore) {
		t.Fatal("Tree rejected dropping before a leaf")
	}
	target := treeDropIndicatorTarget(visible, treeDropTarget{key: "project", drawKey: "project", position: DropAfter})
	if target.drawKey != "readme" {
		t.Fatalf("expanded branch after-indicator = %q, want readme", target.drawKey)
	}
}

func TestTreeCanDropReceivesBatchAndControlsIndicator(t *testing.T) {
	visible := flattenVisibleItems(treeTestItems(), treeKeySet([]string{"project", "src"}))
	var proposed DropEvent
	tree := New("files", "", treeTestItems()).CanDrop(func(event DropEvent) bool {
		proposed = event
		return false
	})
	sources := []string{"main", "ui"}
	if tree.dropAllowed(visible, sources, "archive", DropBefore) {
		t.Fatal("CanDrop rejection was ignored")
	}
	if proposed.SourceKey != "main" || !treeSameKeys(proposed.SourceKeys, sources) || proposed.TargetKey != "archive" || proposed.Position != DropBefore {
		t.Fatalf("drop proposal = %#v", proposed)
	}
	proposed.SourceKeys[0] = "changed"
	if sources[0] != "main" {
		t.Fatal("drop proposal exposed the Tree source slice")
	}
}

func TestTreeDragLifecycleReportsValidTargetTransitions(t *testing.T) {
	var phases []string
	record := func(phase string) func(DragEvent) {
		return func(event DragEvent) {
			phases = append(phases, fmt.Sprintf("%s:%s:%d:%s", phase, event.TargetKey, event.Position, event.SourceKey))
			if phase == "start" {
				event.SourceKeys[0] = "changed"
			}
		}
	}
	tree := New("files", "", nil).
		OnDragStart(record("start")).
		OnDragEnter(record("enter")).
		OnDragLeave(record("leave")).
		OnDragOver(record("over")).
		OnDragEnd(record("end"))
	state := new(treeState)
	sources := []string{"one", "two"}
	first := treeDropTarget{key: "folder", position: DropInside}
	second := treeDropTarget{key: "archive", position: DropBefore}
	tree.updateDragLifecycle(state, sources, first)
	tree.updateDragLifecycle(state, sources, first)
	tree.updateDragLifecycle(state, sources, second)
	tree.endDragLifecycle(state)
	want := []string{
		"start::0:one",
		"enter:folder:1:one",
		"over:folder:1:one",
		"over:folder:1:one",
		"leave:folder:1:one",
		"enter:archive:0:one",
		"over:archive:0:one",
		"leave:archive:0:one",
		"end:archive:0:one",
	}
	if !treeSameKeys(phases, want) {
		t.Fatalf("drag phases = %#v, want %#v", phases, want)
	}
	if sources[0] != "one" || len(state.dragEventSources) != 0 || state.dragEventTarget.key != "" {
		t.Fatalf("drag lifecycle state = sources %#v state %#v", sources, state)
	}
}

func TestTreeDragEmitsDropEvent(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	var dropped DropEvent
	activated := ""
	items := []Item{
		{Key: "one", Label: "One"},
		{Key: "two", Label: "Two"},
		{Key: "folder", Label: "Folder", AcceptsChildren: true},
	}
	tree := New("files", "", items).
		SelectionMode(SelectionMultiple).
		SelectedKeys([]string{"one", "two"}).
		OnAction(func(key string) { activated = key }).
		OnDrop(func(event DropEvent) { dropped = event })
	start := time.Unix(4, 0)
	layoutTreeFrame(ctx, router, tree, start)
	layoutTreeFrame(ctx, router, tree, start.Add(time.Millisecond))

	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 20),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(2*time.Millisecond))
	router.Queue(pointer.Event{
		Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 26),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(3*time.Millisecond))
	router.Queue(pointer.Event{
		Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 100),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(4*time.Millisecond))
	state := treePeekState(ctx, "files")
	if state.dragSource != "one" || !treeSameKeys(state.dragSources, []string{"one", "two"}) || state.dropTarget.key != "folder" || state.dropTarget.drawKey != "folder" || state.dropTarget.position != DropInside {
		t.Fatalf("active drop target = source %q target %#v", state.dragSource, state.dropTarget)
	}
	router.Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: f32.Pt(100, 100),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(5*time.Millisecond))
	layoutTreeFrame(ctx, router, tree, start.Add(6*time.Millisecond))
	if dropped.SourceKey != "one" || !treeSameKeys(dropped.SourceKeys, []string{"one", "two"}) || dropped.TargetKey != "folder" || dropped.Position != DropInside {
		t.Fatalf("drop event = %#v", dropped)
	}
	if activated != "" {
		t.Fatalf("drag also activated %q", activated)
	}
}

func TestDraggableTreeReleasesOffscreenItemState(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	items := make([]Item, 80)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("item-%d", index), Label: fmt.Sprintf("Item %d", index)}
	}
	widgetValue := New("files", "", items).MaxHeight(72).OnDrop(func(DropEvent) {})
	start := time.Unix(15, 0)
	layoutTreeFrame(ctx, router, widgetValue, start)
	state := treePeekState(ctx, "files")
	initial := len(state.items)
	if initial == 0 {
		t.Fatal("draggable Tree did not create visible item state")
	}
	state.list.ScrollTo(60)
	layoutTreeFrame(ctx, router, widgetValue, start.Add(time.Millisecond))
	if _, retained := state.items["item-0"]; retained {
		t.Fatal("draggable Tree retained an offscreen item state")
	}
	if len(state.items) > initial+1 {
		t.Fatalf("draggable Tree state grew from %d to %d after scrolling", initial, len(state.items))
	}
}

func TestDraggableTreeStillHandlesRowClicks(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	selected := ""
	tree := New("files", "", treeTestItems()).
		OnChange(func(key string) { selected = key }).
		OnDrop(func(DropEvent) {})
	start := time.Unix(5, 0)
	layoutTreeFrame(ctx, router, tree, start)
	layoutTreeFrame(ctx, router, tree, start.Add(time.Millisecond))
	router.Queue(
		pointer.Event{
			Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
			Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 20),
		},
		pointer.Event{
			Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
			Position: f32.Pt(100, 20),
		},
	)
	layoutTreeFrame(ctx, router, tree, start.Add(2*time.Millisecond))
	if selected != "project" {
		t.Fatalf("selected = %q, want project", selected)
	}
}

func TestDraggableTreeStillHandlesToggleClicks(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	var expanded []string
	tree := New("files", "", treeTestItems()).
		OnExpandedChange(func(keys []string) { expanded = keys }).
		OnDrop(func(DropEvent) {})
	start := time.Unix(6, 0)
	layoutTreeFrame(ctx, router, tree, start)
	layoutTreeFrame(ctx, router, tree, start.Add(time.Millisecond))
	router.Queue(
		pointer.Event{
			Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
			Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20),
		},
		pointer.Event{
			Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
			Position: f32.Pt(20, 20),
		},
	)
	layoutTreeFrame(ctx, router, tree, start.Add(2*time.Millisecond))
	if !treeSameKeys(expanded, []string{"project"}) {
		t.Fatalf("expanded = %#v, want project", expanded)
	}
}

func TestTreeLayoutUsesThemeAndMaxHeight(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Tree.RowHeight = 48
	ctx := treeTestContext(&activeTheme, locale.LanguageEnglish)
	dims := New("files", "", []Item{{Key: "one", Label: "One"}}).
		Layout(ctx, treeLayoutContext(nil, image.Pt(320, 200), time.Time{}))
	if dims.Size.X != 320 || dims.Size.Y < 48 {
		t.Fatalf("theme dimensions = %v, want width 320 and height >= 48", dims.Size)
	}

	items := make([]Item, 20)
	for index := range items {
		items[index] = Item{Key: string(rune('a' + index)), Label: "Item"}
	}
	dims = New("many", "", items).MaxHeight(80).
		Layout(treeTestContext(nil, locale.LanguageEnglish), treeLayoutContext(nil, image.Pt(320, 300), time.Time{}))
	if dims.Size.Y > 80 {
		t.Fatalf("max-height dimensions = %v, want height <= 80", dims.Size)
	}
}

func TestTreeDragPreviewUsesCompactIndependentSize(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	dims := New("files", "", nil).Size(SizeSmall).layoutDragPreview(
		ctx,
		treeLayoutContext(nil, image.Pt(520, 24), time.Time{}),
		"A long item name that must not use the full Tree row width",
	)
	if dims.Size.X > 240 || dims.Size.Y <= 24 {
		t.Fatalf("drag preview dimensions = %v, want width <= 240 and independent row height", dims.Size)
	}
}

func TestTreeComposedContentInheritsSelectedColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	leading := &treeProbe{size: image.Pt(16, 16)}
	expandedLeading := &treeProbe{size: image.Pt(16, 16)}
	trailing := &treeProbe{size: image.Pt(32, 16)}
	item := Item{
		Key: "selected", Label: "Selected", Leading: leading, ExpandedLeading: expandedLeading, Trailing: trailing,
		Children: []Item{{Key: "child", Label: "Child"}},
	}
	New("tree", "selected", []Item{item}).ExpandedKeys([]string{"selected"}).Layout(
		treeTestContext(&activeTheme, locale.LanguageEnglish),
		treeLayoutContext(nil, image.Pt(320, 100), time.Time{}),
	)
	if leading.layouts != 0 {
		t.Fatalf("collapsed leading layouts = %d, want 0 while expanded", leading.layouts)
	}
	for name, probe := range map[string]*treeProbe{"expanded leading": expandedLeading, "trailing": trailing} {
		if probe.layouts != 1 || probe.foreground != activeTheme.Palette.AccentSoftForeground || probe.background != activeTheme.Palette.AccentSoft {
			t.Errorf("%s content = layouts %d colors %#v/%#v", name, probe.layouts, probe.foreground, probe.background)
		}
	}
}

func TestTreeCustomContentInheritsSelectedColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	content := &treeProbe{size: image.Pt(120, 18)}
	item := Item{Key: "selected", Label: "Selected", Content: content}
	New("tree", "selected", []Item{item}).Layout(
		treeTestContext(&activeTheme, locale.LanguageEnglish),
		treeLayoutContext(nil, image.Pt(320, 100), time.Time{}),
	)
	if content.layouts != 1 || content.foreground != activeTheme.Palette.AccentSoftForeground || content.background != activeTheme.Palette.AccentSoft {
		t.Fatalf("custom content = layouts %d colors %#v/%#v", content.layouts, content.foreground, content.background)
	}
}

func TestTreeCustomSwitcherTracksExpansion(t *testing.T) {
	collapsed := &treeProbe{size: image.Pt(12, 12)}
	expanded := &treeProbe{size: image.Pt(12, 12)}
	item := Item{
		Key: "folder", Label: "Folder",
		Children: []Item{{Key: "file", Label: "File"}},
		Switcher: collapsed, ExpandedSwitcher: expanded,
	}
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	New("tree", "", []Item{item}).ExpandedKeys([]string{"folder"}).Layout(
		ctx,
		treeLayoutContext(nil, image.Pt(320, 100), time.Time{}),
	)
	if collapsed.layouts != 0 || expanded.layouts != 1 {
		t.Fatalf("switcher layouts = %d/%d, want 0/1", collapsed.layouts, expanded.layouts)
	}
}

func TestTreeSemanticsExposeSelectionAndExpansion(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 160)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	New("files", "project", treeTestItems()).ExpandedKeys([]string{"project"}).Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	node, ok := treeSemanticNode(router.AppendSemantics(nil), "Project")
	if !ok || !node.Desc.Selected || node.Desc.Description != "Expanded" {
		t.Fatalf("Project semantics = found %v selected %v description %q", ok, node.Desc.Selected, node.Desc.Description)
	}
}

func TestTreeCheckboxSemanticsExposeCheckedAndPartialState(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 160)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	items := []Item{
		{Key: "checked", Label: "Checked"},
		{Key: "partial", Label: "Partial", Children: []Item{{Key: "child", Label: "Child"}}},
	}
	frame.BeginFrame(ctx)
	New("files", "", items).
		Checkable(true).
		CheckStrictly(true).
		CheckedKeys([]string{"checked"}).
		HalfCheckedKeys([]string{"partial"}).
		Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	nodes := router.AppendSemantics(nil)
	checked, ok := treeSemanticNodeByClass(nodes, "Checked", semantic.CheckBox)
	if !ok || !checked.Desc.Selected || checked.Desc.Disabled {
		t.Fatalf("checked checkbox semantics = found %v desc %#v", ok, checked.Desc)
	}
	partial, ok := treeSemanticNodeByClass(nodes, "Partial", semantic.CheckBox)
	if !ok || partial.Desc.Selected || partial.Desc.Description != "Partially selected" {
		t.Fatalf("partial checkbox semantics = found %v desc %#v", ok, partial.Desc)
	}
}

func TestTreeSemanticDescriptionUsesContextLanguage(t *testing.T) {
	item := Item{Key: "folder", Label: "Folder", Children: []Item{{Key: "file"}}}
	if got := treeSemanticDescription(treeTestContext(nil, locale.LanguageEnglish), item, false); got != "Collapsed" {
		t.Fatalf("english description = %q", got)
	}
	if got := treeSemanticDescription(treeTestContext(nil, locale.LanguageChinese), item, true); got != "已展开" {
		t.Fatalf("chinese description = %q", got)
	}
}

func treeTestItems() []Item {
	return []Item{
		{
			Key:   "project",
			Label: "Project",
			Children: []Item{
				{
					Key:   "src",
					Label: "Source",
					Children: []Item{
						{Key: "main", Label: "main.go"},
						{Key: "ui", Label: "ui.go"},
					},
				},
				{Key: "readme", Label: "README.md"},
			},
		},
		{Key: "archive", Label: "Archive"},
	}
}

func treeTestContext(activeTheme *theme.Theme, language locale.Language) *frame.Context {
	return frame.New(nil, activeTheme, language)
}

func treeLayoutContext(router *input.Router, max image.Point, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	var ops op.Ops
	return layout.Context{Constraints: layout.Constraints{Max: max}, Source: source, Ops: &ops, Now: now}
}

func layoutTreeFrame(ctx *frame.Context, router *input.Router, tree Widget, now time.Time) {
	gtx := treeLayoutContext(router, image.Pt(360, 240), now)
	frame.BeginFrame(ctx)
	tree.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func treeSetState(ctx *frame.Context, key string, value *treeState) {
	frame.UseStateWith(ctx, key, stateSlotTree, func() *treeState { return value })
}

func treePeekState(ctx *frame.Context, key string) *treeState {
	value, _ := frame.PeekState[treeState](ctx, key, stateSlotTree)
	return value
}

func treeSameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func treeFlatKeys(items []flatItem) []string {
	keys := make([]string, len(items))
	for index, item := range items {
		keys[index] = item.item.Key
	}
	return keys
}

func treeSemanticNode(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Label == label {
			return node, true
		}
		if child, ok := treeSemanticNode(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func treeSemanticNodeByClass(nodes []input.SemanticNode, label string, class semantic.ClassOp) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Label == label && node.Desc.Class == class {
			return node, true
		}
		if child, ok := treeSemanticNodeByClass(node.Children, label, class); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}
