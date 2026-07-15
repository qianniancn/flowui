package tree

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
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func TestTreeOptionsUseValueSemantics(t *testing.T) {
	items := treeTestItems()
	base := New("files", "readme", items)
	configured := base.
		ExpandedKeys([]string{"project"}).
		DisabledKeys([]string{"archive"}).
		EmptyText("Empty").
		OnChange(func(string) {}).
		OnExpandedChange(func([]string) {}).
		OnAction(func(string) {}).
		OnDrop(func(DropEvent) {}).
		Variant(VariantSurface).
		Size(SizeSmall).
		SelectionMode(SelectionNone).
		Disabled(true).
		AllowEmptySelection().
		Guides(true).
		GuideConnectors(true).
		GuideStyle(GuideDashed).
		MaxHeight(180)

	if len(base.expandedKeys) != 0 || len(base.disabledKeys) != 0 || base.variant != VariantDefault || base.size != SizeMedium || base.selectionMode != SelectionSingle {
		t.Fatal("configuring a Tree mutated the base value")
	}
	if configured.key != "files" || configured.selectedKey != "readme" || len(configured.items) != len(items) {
		t.Fatal("constructor fields were not retained")
	}
	if !treeSameKeys(configured.expandedKeys, []string{"project"}) || !treeSameKeys(configured.disabledKeys, []string{"archive"}) {
		t.Fatal("controlled key options were not retained")
	}
	if configured.emptyText != "Empty" || configured.onChange == nil || configured.onExpandedChange == nil || configured.onAction == nil || configured.onDrop == nil {
		t.Fatal("callbacks or empty text were not retained")
	}
	if configured.variant != VariantSurface || configured.size != SizeSmall || configured.selectionMode != SelectionNone || !configured.disabled || !configured.allowEmpty || !configured.guides || !configured.guideConnectors || configured.guideStyle != GuideDashed || configured.maxHeight != 180 {
		t.Fatal("behavior options were not retained")
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
	tree.activate("project")
	if changed != "" || action != "project" {
		t.Fatalf("callbacks = changed %q action %q", changed, action)
	}
}

func TestTreeAllowEmptySelection(t *testing.T) {
	changed := "not-called"
	New("files", "project", treeTestItems()).
		AllowEmptySelection().
		OnChange(func(key string) { changed = key }).
		activate("project")
	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
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

func TestTreeGeometryAndStyleMatchHeroUIPatterns(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Tree
	if tokens.Padding != 4 || tokens.Gap != 4 || tokens.RowHeight != 36 || tokens.RowRadius != 16 || tokens.RowPaddingX != 8 || tokens.RowPaddingY != 6 {
		t.Fatalf("Tree geometry = %+v", tokens)
	}
	if tokens.Indent != 20 || tokens.ChevronSlotSize != 20 || tokens.ChevronIconSize != 16 || tokens.ContentGap != 8 {
		t.Fatalf("Tree hierarchy geometry = %+v", tokens)
	}
	selected := treeItemStyleFor(&activeTheme, true, false, false)
	if selected.background != activeTheme.Palette.AccentSoft || selected.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("selected style = %#v/%#v", selected.background, selected.foreground)
	}
	hovered := treeItemStyleFor(&activeTheme, false, true, false)
	if hovered.background != activeTheme.Palette.SurfaceTertiary {
		t.Fatalf("hover background = %#v, want SurfaceTertiary", hovered.background)
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
	visible := []flatItem{{item: Item{Key: "one"}}, {item: Item{Key: "two"}}}
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

func TestTreeDropRejectsSelfDescendantsAndDisabledItems(t *testing.T) {
	items := treeTestItems()
	visible := flattenVisibleItems(items, treeKeySet([]string{"project", "src"}))
	tree := New("files", "", items)
	if treeDropAllowed(tree, visible, "project", "project") {
		t.Fatal("Tree allowed dropping an item onto itself")
	}
	if treeDropAllowed(tree, visible, "project", "main") {
		t.Fatal("Tree allowed dropping a parent into its descendant")
	}
	if !treeDropAllowed(tree, visible, "main", "archive") {
		t.Fatal("Tree rejected a valid sibling move")
	}
	if treeDropAllowed(tree.DisabledKeys([]string{"archive"}), visible, "main", "archive") {
		t.Fatal("Tree allowed dropping onto a disabled item")
	}
	target := treeDropIndicatorTarget(visible, treeDropTarget{key: "project", drawKey: "project", position: DropAfter})
	if target.drawKey != "readme" {
		t.Fatalf("expanded branch after-indicator = %q, want readme", target.drawKey)
	}
}

func TestTreeDragEmitsDropEvent(t *testing.T) {
	ctx := treeTestContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	var dropped DropEvent
	activated := ""
	tree := New("files", "", treeTestItems()).
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
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(4*time.Millisecond))
	state := treePeekState(ctx, "files")
	if state.dragSource != "project" || state.dropTarget.key != "archive" || state.dropTarget.drawKey != "archive" || state.dropTarget.position != DropInside {
		t.Fatalf("active drop target = source %q target %#v", state.dragSource, state.dropTarget)
	}
	router.Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: f32.Pt(100, 60),
	})
	layoutTreeFrame(ctx, router, tree, start.Add(5*time.Millisecond))
	layoutTreeFrame(ctx, router, tree, start.Add(6*time.Millisecond))
	if dropped != (DropEvent{SourceKey: "project", TargetKey: "archive", Position: DropInside}) {
		t.Fatalf("drop event = %#v", dropped)
	}
	if activated != "" {
		t.Fatalf("drag also activated %q", activated)
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

func TestTreeComposedContentInheritsSelectedColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	leading := &treeProbe{size: image.Pt(16, 16)}
	trailing := &treeProbe{size: image.Pt(32, 16)}
	item := Item{Key: "selected", Label: "Selected", Leading: leading, Trailing: trailing}
	New("tree", "selected", []Item{item}).Layout(
		treeTestContext(&activeTheme, locale.LanguageEnglish),
		treeLayoutContext(nil, image.Pt(320, 100), time.Time{}),
	)
	for name, probe := range map[string]*treeProbe{"leading": leading, "trailing": trailing} {
		if probe.layouts != 1 || probe.foreground != activeTheme.Palette.AccentSoftForeground || probe.background != activeTheme.Palette.AccentSoft {
			t.Errorf("%s content = layouts %d colors %#v/%#v", name, probe.layouts, probe.foreground, probe.background)
		}
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
