package menu

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type fixedMenuWidget struct {
	size image.Point
}

func (w fixedMenuWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type measurableMenuWidget struct {
	size         image.Point
	layoutCalls  int
	measureCalls int
}

func (w *measurableMenuWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.layoutCalls++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func (w *measurableMenuWidget) Measure(_ *core.Context, gtx layout.Context) layout.Dimensions {
	w.measureCalls++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func menuTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
}

func menuTestLayoutContext(router *input.Router, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(400, 400)},
		Source:      source,
		Ops:         new(op.Ops),
		Now:         now,
	}
}

func TestMenuOptionsAreImmutable(t *testing.T) {
	base := New("actions", []Item{{Key: "copy", Label: "Copy"}})
	configured := base.
		Sections([]Section{{Title: "Edit", Items: []Item{{Key: "paste", Label: "Paste"}}}}).
		AutoSeparateSections(false).
		BeforeContent(fixedMenuWidget{size: image.Pt(20, 20)}).
		AfterContent(fixedMenuWidget{size: image.Pt(20, 20)}).
		EmptyText("Nothing here").
		SelectionMode(SelectionMultiple).
		SelectedKey("copy").
		SelectedKeys([]string{"paste"}).
		DisabledKeys([]string{"delete"}).
		OnActionEvent(func(ActionEvent) {}).
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnCheckedChange(func(string, bool) {}).
		OnRadioChange(func(string, string) {}).
		CloseOnSelect(false).
		Disabled(true).
		Compact(true).
		Width(260)
	if configured.emptyText != "Nothing here" || configured.onActionEvent == nil || configured.onChange == nil || configured.onSelectionChange == nil || configured.onCheckedChange == nil || configured.onRadioChange == nil || !configured.disabled || !configured.compact || configured.width != 260 || len(configured.sections) != 1 || configured.autoSeparateSections || !configured.hasCloseOnSelect || configured.closeOnSelect {
		t.Fatalf("configured menu = %#v", configured)
	}
	if base.hasEmptyText || base.resolvedEmptyText(menuTestContext()) != "No actions" || base.onActionEvent != nil || base.disabled || base.compact || base.width != 0 || len(base.sections) != 0 {
		t.Fatalf("base menu was mutated: %#v", base)
	}
	if Separator().Kind != ItemSeparator || GroupLabel("Edit").Kind != ItemGroupLabel {
		t.Fatal("menu structural item constructors returned wrong kinds")
	}
}

func TestMenuAutoWidthMeasuresContentAndHonorsBounds(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	short := New("short", []Item{{Key: "one", Label: "One"}}).AutoWidth().MinWidth(160)
	long := New("long", []Item{{Key: "one", Label: "A considerably longer menu action"}}).AutoWidth().MinWidth(160).MaxWidth(280)
	shortWidth := short.preferredWidthPxWithVisual(ctx, gtx, short.visual(ctx))
	longWidth := long.preferredWidthPxWithVisual(ctx, gtx, long.visual(ctx))
	if longWidth <= shortWidth {
		t.Fatalf("long auto width = %d, short width = %d", longWidth, shortWidth)
	}
	shortGtx := gtx
	short.applyConstraintsWithVisual(ctx, &shortGtx, nil, short.visual(ctx))
	if got := shortGtx.Constraints.Max.X; got < gtx.Dp(160) {
		t.Fatalf("short constrained width = %d, want at least 160dp", got)
	}
	long.applyConstraintsWithVisual(ctx, &gtx, nil, long.visual(ctx))
	if got := gtx.Constraints.Max.X; got > gtx.Dp(280) {
		t.Fatalf("max width = %d, want <= 280dp", got)
	}
	if got, want := gtx.Constraints.Max.X, min(longWidth, gtx.Dp(280)); got != want {
		t.Fatalf("auto width = %d, want content width %d without the default fraction cap", got, want)
	}
}

func TestMenuAutoWidthIncludesLeadingContentGap(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := fixedMenuWidget{size: image.Pt(16, 16)}
	widget := New("leading-gap", nil)
	item := Item{Key: "roadmap", Label: "Roadmap", Leading: leading}
	visual := widget.visual(ctx)
	metrics := visual.Metrics
	itemVisual := visual.Item(theme.MenuItemRequest{Variant: uint8(item.Variant)})
	measurementStyle := widget.resolveItemMeasurementStyleWithVisual(ctx, item, visual)
	itemColors := itemVisual.Colors(measurementStyle)
	want := widget.measureWidgetWidth(ctx, gtx, leading) +
		gtx.Dp(metrics.ItemContentGap) +
		visual.TextWidth(gtx, theme.MenuTextRequest{Text: item.Label, Color: itemColors.Foreground, Role: theme.MenuTextItem, Style: measurementStyle}) +
		gtx.Dp(metrics.ItemPaddingX)*2
	if got := widget.preferredItemWidthPxWithVisual(ctx, gtx, item, visual); got != want {
		t.Fatalf("auto item width = %d, want %d including leading content gap", got, want)
	}
	trailing := fixedMenuWidget{size: image.Pt(16, 16)}
	trailingItem := Item{Key: "count", Label: "", Trailing: trailing}
	wantTrailing := gtx.Dp(metrics.ItemPaddingX)*2 + widget.measureWidgetWidth(ctx, gtx, trailing)
	if got := widget.preferredItemWidthPxWithVisual(ctx, gtx, trailingItem, visual); got != wantTrailing {
		t.Fatalf("auto item width = %d, want %d without a leading gap", got, wantTrailing)
	}
}

func TestMenuAutoWidthMeasuresEmptyText(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := New("empty", nil).EmptyText("No matching actions").AutoWidth()
	visual := widget.visual(ctx)
	width := widget.preferredWidthPxWithVisual(ctx, gtx, visual)
	if width <= gtx.Dp(visual.Metrics.Padding*2) {
		t.Fatalf("empty menu width = %d, expected the empty text to contribute", width)
	}
}

func TestMenuAutoWidthIncludesSectionTitlePadding(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := Sections("section-width", []Section{{Title: "A considerably long section title"}}).AutoWidth()
	visual := widget.visual(ctx)
	metrics := visual.Metrics
	title := visual.TextWidth(gtx, theme.MenuTextRequest{Text: "A considerably long section title", Color: visual.Colors.Description, Role: theme.MenuTextSection})
	width := widget.preferredWidthPxWithVisual(ctx, gtx, visual)
	minimum := title + 2*gtx.Dp(metrics.SectionPaddingX) + 2*gtx.Dp(metrics.Padding)
	if width < minimum {
		t.Fatalf("section title width = %d, want at least %d including padding", width, minimum)
	}
}

func TestMenuAutoWidthUsesSideEffectFreeWidgetMeasurement(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	indicator := &measurableMenuWidget{size: image.Pt(22, 16)}
	widget := New("measurable", []Item{{
		Key: "more", Label: "More", Leading: leading,
		Children: []Item{{Key: "copy", Label: "Copy"}}, SubmenuIndicator: indicator,
	}}).AutoWidth()
	width := widget.preferredWidthPxWithVisual(ctx, gtx, widget.visual(ctx))
	if width <= 0 || leading.measureCalls != 1 || indicator.measureCalls != 1 || leading.layoutCalls != 0 || indicator.layoutCalls != 0 {
		t.Fatalf("measurement width %d leading %#v indicator %#v", width, leading, indicator)
	}
}

func TestMenuAutoWidthCachesVersionedMeasurement(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	widget := New("cached-width", []Item{{Key: "more", Label: "More", Leading: leading}}).
		AutoWidth().DataVersion(1)
	state := &menuState{}
	first := widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	second := widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	if first != second || leading.measureCalls != 1 {
		t.Fatalf("cached width = %d/%d, measurement calls = %d", first, second, leading.measureCalls)
	}
	widget = widget.DataVersion(2)
	widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	if leading.measureCalls != 2 {
		t.Fatalf("version change reused width measurement, calls = %d", leading.measureCalls)
	}
}

func TestMenuAutoWidthCacheInvalidatesLayoutOptions(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	widget := New("cached-options", []Item{{Key: "more", Label: "More", Leading: leading}}).
		AutoWidth().DataVersion(1)
	state := &menuState{}
	widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))

	widget = widget.Compact(true)
	widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	if leading.measureCalls != 2 {
		t.Fatalf("compact option reused width measurement, calls = %d", leading.measureCalls)
	}
	widget = widget.MinWidth(240)
	widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	if leading.measureCalls != 3 {
		t.Fatalf("minimum width option reused width measurement, calls = %d", leading.measureCalls)
	}
	widget = widget.MaxWidth(320)
	widget.preferredWidthPxCachedWithVisual(ctx, gtx, state, widget.visual(ctx))
	if leading.measureCalls != 4 {
		t.Fatalf("maximum width option reused width measurement, calls = %d", leading.measureCalls)
	}
}

func TestMenuActionEventIncludesSubmenuPath(t *testing.T) {
	var got ActionEvent
	root := New("actions", []Item{{Key: "more", Children: []Item{{Key: "copy", Label: "Copy"}}}}).
		OnActionEvent(func(event ActionEvent) { got = event })
	state := &menuState{key: "actions"}
	child := root.submenu(state, root.items[0])
	child.activate(entry{item: child.items[0]})
	if got.Key != "copy" || len(got.Path) != 2 || got.Path[0] != "more" || got.Path[1] != "copy" || got.Item.Label != "Copy" {
		t.Fatalf("action event = %#v", got)
	}
}

func TestMenuDataVersionCachesEntries(t *testing.T) {
	widget := Sections("cached", []Section{{Title: "Edit", Items: []Item{{Key: "copy", Label: "Copy"}}}}).DataVersion(1)
	state := new(menuState)
	entries := state.resolveEntries(widget)
	cachedEntries := state.resolveEntries(widget)
	if &entries[0] != &cachedEntries[0] {
		t.Fatal("unchanged Menu data version did not reuse entries")
	}
	updatedEntries := state.resolveEntries(widget.DataVersion(2))
	if &entries[0] == &updatedEntries[0] {
		t.Fatal("changed Menu data version reused stale entries")
	}
}

func TestMenuDataCacheInvalidatesAutomaticSectionSeparators(t *testing.T) {
	widget := Sections("sections", []Section{
		{Title: "Edit", Items: []Item{{Key: "copy", Label: "Copy"}}},
		{Title: "View", Items: []Item{{Key: "zoom", Label: "Zoom"}}},
	}).DataVersion(1)
	state := new(menuState)
	withSeparators := state.resolveEntries(widget)
	withoutSeparators := state.resolveEntries(widget.AutoSeparateSections(false))
	if len(withSeparators) != len(withoutSeparators)+1 {
		t.Fatalf("section separator cache = %d/%d entries, want one separator difference", len(withSeparators), len(withoutSeparators))
	}
	for _, entry := range withoutSeparators {
		if entry.separator {
			t.Fatal("automatic section separator remained after disabling it")
		}
	}
}

func TestMenuCompactTokensAndSubmenu(t *testing.T) {
	ctx := menuTestContext()
	base := New("actions", nil)
	metrics := base.Compact(true).visual(ctx).Metrics
	if metrics.Padding != 4 || metrics.Radius != 8 || metrics.ItemGap != 0 || metrics.ItemMinHeight != 28 || metrics.ItemRadius != 4 || metrics.ItemPaddingX != 8 || metrics.ItemPaddingY != 3 {
		t.Fatalf("compact menu metrics = %#v", metrics)
	}
	root := base.Compact(true).visual(ctx).Root.Resolve(style.State{})
	if root.Paint == nil || root.Paint.Radius == nil || *root.Paint.Radius != 8 {
		t.Fatalf("compact menu root radius = %#v, want 8", root.Paint)
	}
	state := &menuState{key: "actions"}
	child := base.Compact(true).submenu(state, Item{Key: "more"})
	if !child.compact {
		t.Fatal("submenu did not inherit compact mode")
	}
}

func TestMenuSubmenuInheritsDataAndSectionConfiguration(t *testing.T) {
	root := New("actions", []Item{{Key: "more", Children: []Item{{Key: "copy", Label: "Copy"}}}}).
		DataVersion(7).
		AutoSeparateSections(false).
		Compact(true)
	state := &menuState{key: "actions"}
	child := root.submenu(state, root.items[0])
	if child.dataVersion != 7 || !child.hasDataVersion || child.autoSeparateSections || !child.compact {
		t.Fatalf("submenu configuration = %#v", child)
	}
}

func TestMenuRuntimeHoverInspectionDoesNotClaimSubmenuState(t *testing.T) {
	widget := New("runtime-menu", []Item{{
		Key: "more", Label: "More", Children: []Item{{Key: "copy", Label: "Copy"}},
	}})
	ctx := menuTestContext()

	core.BeginFrame(ctx)
	runtime := widget.Runtime(ctx, "owner", "menu", nil)
	child := runtime.widget.submenu(runtime.state, runtime.widget.items[0])
	child.stateFor(ctx)
	core.EndFrame(ctx)

	core.BeginFrame(ctx)
	runtime = widget.Runtime(ctx, "owner", "menu", nil)
	runtime.state.openSubmenu = "more"
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("hover inspection claimed submenu key twice: %v", recovered)
		}
		core.EndFrame(ctx)
	}()
	runtime.HoveredWithSubmenus(ctx)
	child = runtime.widget.submenu(runtime.state, runtime.widget.items[0])
	child.stateFor(ctx)
}

func TestMenuCloseReportsPointerAndKeyboardModality(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	modalities := make([]bool, 0, 2)
	widget := New("actions", []Item{{Key: "copy", Label: "Copy"}}).withClose(func(visible bool) {
		modalities = append(modalities, visible)
	})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := core.PeekState[menuState](ctx, "actions", stateSlotMenu)
	state.item("copy").clickable.Click()
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if len(modalities) != 1 || modalities[0] {
		t.Fatalf("pointer close modalities = %v, want hidden focus", modalities)
	}

	router.Source().Execute(key.FocusCmd{Tag: &state.item("copy").clickable})
	router.Queue(key.Event{Name: key.NameEnter, State: key.Press})
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameEnter, State: key.Release})
	layoutMenuFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	if len(modalities) != 2 || !modalities[1] {
		t.Fatalf("keyboard close modalities = %v, want visible focus", modalities)
	}
}

func TestPointerActivatedSubmenuItemDoesNotShowKeyboardFocusRing(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := New("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := core.PeekState[menuState](ctx, "actions", stateSlotMenu)
	state.item("share").clickable.Click()
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	item := state.item("share")
	if menuItemFocusVisible(ctx, item, router.Source().Focused(&item.clickable)) {
		t.Fatal("pointer-activated submenu item exposed a keyboard focus ring")
	}
}

func TestPointerClickOnSubmenuItemDoesNotShowKeyboardFocusRing(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := New("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutMenuFrame(ctx, router, widget, start.Add(3*time.Millisecond))

	state, _ := core.PeekState[menuState](ctx, "actions", stateSlotMenu)
	item := state.item("share")
	if !router.Source().Focused(&item.clickable) {
		t.Fatal("pointer-clicked submenu item did not receive focus")
	}
	if menuItemFocusVisible(ctx, item, true) {
		t.Fatal("pointer-clicked submenu item exposed a keyboard focus ring")
	}
}

func TestPointerPressImmediatelyClearsSubmenuItemKeyboardFocusRing(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := New("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := core.PeekState[menuState](ctx, "actions", stateSlotMenu)
	item := state.item("share")
	router.Source().Execute(key.FocusCmd{Tag: &item.clickable})
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutMenuFrame(ctx, router, widget, start.Add(200*time.Millisecond))
	keyboardGtx := menuTestLayoutContext(nil, start.Add(200*time.Millisecond))
	if got := item.focus.Opacity(keyboardGtx, true); got != 1 {
		t.Fatalf("keyboard focus opacity = %v, want 1", got)
	}

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutMenuFrame(ctx, router, widget, start.Add(201*time.Millisecond))
	if menuItemFocusVisible(ctx, item, router.Source().Focused(&item.clickable)) {
		t.Fatal("pointer press retained keyboard-visible focus")
	}
	gtx := menuTestLayoutContext(nil, start.Add(201*time.Millisecond))
	if got := item.focus.Opacity(gtx, false); got != 0 {
		t.Fatalf("pointer press focus opacity = %v, want 0", got)
	}
}

func TestSubmenuItemFocusRingUsesRequestedOrigin(t *testing.T) {
	ctx := menuTestContext()
	item := new(menuItemState)
	core.RequestFocusVisible(ctx, &item.clickable, true)
	if !menuItemFocusVisible(ctx, item, true) {
		t.Fatal("keyboard-focused submenu item lost its focus ring")
	}
	menuItemFocusVisible(ctx, item, false)

	core.RequestFocusVisible(ctx, &item.clickable, false)
	for range 4 {
		core.BeginFrame(ctx)
	}
	if menuItemFocusVisible(ctx, item, true) {
		t.Fatal("delayed pointer focus was exposed as keyboard focus")
	}

	nativeKeyboardItem := new(menuItemState)
	if !menuItemFocusVisible(ctx, nativeKeyboardItem, true) {
		t.Fatal("native keyboard focus without a pending request was hidden")
	}
}

func TestMenuLayoutUsesStableWidth(t *testing.T) {
	ctx := menuTestContext()
	dims := New("actions", []Item{{Key: "copy", Label: "Copy"}, {Key: "paste", Label: "Paste"}}).
		Layout(ctx, menuTestLayoutContext(nil, time.Unix(1, 0)))
	if dims.Size.X != 192 || dims.Size.Y != 80 {
		t.Fatalf("menu size = %v, want viewport-clamped 192x80", dims.Size)
	}
}

func TestMenuStyleWidthConstraintsArePreserved(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := New("styled-width", []Item{{Key: "copy", Label: "Copy"}})
	styled := (widget).Style(style.Style{}.Width(unit.Dp(300)))
	dims := styled.Layout(ctx, gtx)
	if dims.Size.X != 300 {
		t.Fatalf("styled menu width = %d, want 300", dims.Size.X)
	}
}

func TestMenuWidthConstraintsPreserveParentMinimum(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	gtx.Constraints = layout.Constraints{Min: image.Pt(300, 0), Max: image.Pt(500, 400)}
	widget := New("constraint-conflict", []Item{{Key: "copy", Label: "Copy"}}).
		AutoWidth().MaxWidth(120)
	widget.applyConstraintsWithVisual(ctx, &gtx, nil, widget.visual(ctx))
	if got := gtx.Constraints.Max.X; got != 300 || gtx.Constraints.Min.X != 300 {
		t.Fatalf("conflicting width constraints = %v, want exact 300px parent minimum", gtx.Constraints)
	}
}

func TestMenuAutoWidthUsesItemPartTypography(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := New("weighted", []Item{{Key: "copy", Label: "Copy"}})
	visual := widget.visualWithOverride(ctx, &theme.MenuOverride{Declaration: style.Style{}.Part(style.PartItem, style.Style{}.FontWeight(int(font.Bold)))})
	part := widget.resolveItemMeasurementStyleWithVisual(ctx, widget.items[0], visual)
	itemVisual := visual.Item(theme.MenuItemRequest{})
	if got := visual.TextWidth(gtx, theme.MenuTextRequest{Text: widget.items[0].Label, Color: itemVisual.Colors(part).Foreground, Role: theme.MenuTextItem, Style: part}); got <= 0 {
		t.Fatalf("measured item width = %d, want positive width", got)
	}
}

func TestMenuAutoWidthResolvesItemVisualOnce(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := New("visual-once", []Item{{Key: "copy", Label: "Copy", Shortcut: "Ctrl+C", Leading: fixedMenuWidget{size: image.Pt(16, 16)}}})
	visual := widget.visual(ctx)
	original := visual.Item
	itemCalls := 0
	visual.Item = func(request theme.MenuItemRequest) theme.MenuItemVisual {
		itemCalls++
		return original(request)
	}
	if got := widget.preferredItemWidthPxWithVisual(ctx, gtx, widget.items[0], visual); got <= 0 {
		t.Fatalf("auto item width = %d, want positive width", got)
	}
	if itemCalls != 1 {
		t.Fatalf("item visual calls = %d, want one call per measurement", itemCalls)
	}
}

func TestMenuAutoWidthUsesItemPartPadding(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	base := New("padding", []Item{{Key: "copy", Label: "Copy"}})
	wideVisual := base.visualWithOverride(ctx, &theme.MenuOverride{Declaration: style.Style{}.Part(style.PartItem, style.Style{}.PaddingX(40))})
	baseVisual := base.visual(ctx)
	baseWidth := base.preferredItemWidthPxWithVisual(ctx, gtx, base.items[0], baseVisual)
	wideWidth := base.preferredItemWidthPxWithVisual(ctx, gtx, base.items[0], wideVisual)
	wantDelta := 2 * (gtx.Dp(unit.Dp(40)) - gtx.Dp(baseVisual.Metrics.ItemPaddingX))
	if wideWidth-baseWidth != wantDelta {
		t.Fatalf("item padding width delta = %d, want %d", wideWidth-baseWidth, wantDelta)
	}
}

func TestLongMenuIsScrollableWithinMaximumHeight(t *testing.T) {
	ctx := menuTestContext()
	items := make([]Item, 20)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("action-%d", index), Label: fmt.Sprintf("Action %d", index)}
	}
	menu := New("long-actions", items)
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	core.BeginFrame(ctx)
	dims := menu.Layout(ctx, gtx)
	core.EndFrame(ctx)
	if dims.Size != image.Pt(192, 400) {
		t.Fatalf("long menu size = %v, want viewport-constrained 192x400", dims.Size)
	}
	state, ok := core.PeekState[menuState](ctx, "long-actions", stateSlotMenu)
	if !ok || state.list.Position.Count >= len(items) || state.list.Position.Length <= dims.Size.Y {
		t.Fatalf("long menu list position = %#v", state.list.Position)
	}
}

func TestScrolledMenuAnchorsOnlyVisibleSubmenuItems(t *testing.T) {
	ctx := menuTestContext()
	items := make([]Item, 20)
	for index := range items {
		items[index] = Item{Key: fmt.Sprintf("action-%d", index), Label: fmt.Sprintf("Action %d", index), Kind: ItemSubmenu, Children: []Item{{Key: fmt.Sprintf("child-%d", index), Label: "Child"}}}
	}
	menu := New("scroll-anchors", items)
	start := time.Unix(1, 0)
	core.BeginFrame(ctx)
	menu.Layout(ctx, menuTestLayoutContext(nil, start))
	core.EndFrame(ctx)
	state, _ := core.PeekState[menuState](ctx, "scroll-anchors", stateSlotMenu)
	state.list.ScrollTo(5)

	core.BeginFrame(ctx)
	menu.Layout(ctx, menuTestLayoutContext(nil, start.Add(time.Millisecond)))
	core.EndFrame(ctx)
	if state.list.Position.First != 5 {
		t.Fatalf("first visible entry = %d, want 5", state.list.Position.First)
	}
	anchor, ok := state.anchors["action-5"]
	if !ok || anchor.Min.Y != 8 {
		t.Fatalf("first visible submenu anchor = %v, present %v", anchor, ok)
	}
	if _, ok := state.anchors["action-4"]; ok {
		t.Fatal("offscreen submenu item retained an anchor")
	}
}

func TestSectionsSupportExplicitSeparatorsAndHeroSpacing(t *testing.T) {
	without := Sections("sections", []Section{
		{Title: "One", Items: []Item{{Key: "one", Label: "One"}}},
		{Title: "Two", Items: []Item{{Key: "two", Label: "Two"}}},
	}).AutoSeparateSections(false)
	with := without.Sections([]Section{
		{Title: "One", Items: []Item{{Key: "one", Label: "One"}}},
		{Title: "Two", SeparatorBefore: true, Items: []Item{{Key: "two", Label: "Two"}}},
	})
	if len(without.entries()) != 4 || len(with.entries()) != 5 || !with.entries()[2].separator {
		t.Fatalf("Menu section entries = without %d with %#v", len(without.entries()), with.entries())
	}
	entries := with.entries()
	if with.entryNeedsGap(entries, 1) || !with.entryNeedsGap(entries, 2) || !with.entryNeedsGap(entries, 3) {
		t.Fatalf("Menu section gaps = %#v", entries)
	}
}

func TestMenuSelectionCallbacks(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		selected := ""
		selection := []string(nil)
		widget := New("fruit", []Item{{Key: "apple", Label: "Apple"}}).
			SelectionMode(SelectionSingle).
			OnChange(func(key string) { selected = key }).
			OnSelectionChange(func(keys []string) { selection = keys })
		if !widget.activate(widget.actionableEntries()[0]) || selected != "apple" || len(selection) != 1 || selection[0] != "apple" {
			t.Fatalf("single selection = %q %v", selected, selection)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		selection := []string(nil)
		widget := New("styles", []Item{{Key: "bold", Label: "Bold"}}).
			SelectionMode(SelectionMultiple).
			SelectedKeys([]string{"italic"}).
			OnSelectionChange(func(keys []string) { selection = keys })
		widget.activate(widget.actionableEntries()[0])
		if len(selection) != 2 || selection[0] != "italic" || selection[1] != "bold" {
			t.Fatalf("multiple selection = %v", selection)
		}
	})

	t.Run("section override", func(t *testing.T) {
		selection := []string(nil)
		widget := Sections("alignment", []Section{{
			Title: "Alignment", SelectionMode: SelectionSingle, SelectedKeys: []string{"left"},
			OnSelectionChange: func(keys []string) { selection = keys },
			Items:             []Item{{Key: "right", Label: "Right"}},
		}})
		widget.activate(widget.actionableEntries()[0])
		if len(selection) != 1 || selection[0] != "right" {
			t.Fatalf("section selection = %v", selection)
		}
	})
}

func TestMenuActivatesActionCheckboxAndRadioItems(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	action := ""
	checkedKey := ""
	checked := false
	radioGroup := ""
	radioValue := ""
	itemActions := map[string]int{}
	widget := New("actions", []Item{
		{Key: "copy", Label: "Copy", OnAction: func() { itemActions["copy"]++ }},
		{Key: "favorite", Label: "Favorite", Kind: ItemCheckbox, Checked: false, KeepOpen: true, OnAction: func() { itemActions["favorite"]++ }},
		{Key: "compact", Label: "Compact", Kind: ItemRadio, RadioGroup: "density", Value: "compact", KeepOpen: true, OnAction: func() { itemActions["compact"]++ }},
	}).
		OnActionEvent(func(event ActionEvent) { action = event.Key }).
		OnCheckedChange(func(key string, value bool) { checkedKey, checked = key, value }).
		OnRadioChange(func(group, value string) { radioGroup, radioValue = group, value })

	core.BeginFrame(ctx)
	state := widget.stateFor(ctx)
	state.item("copy").clickable.Click()
	state.item("favorite").clickable.Click()
	state.item("compact").clickable.Click()
	widget.layoutWithVisual(ctx, gtx, state, true, widget.visual(ctx))
	core.EndFrame(ctx)
	if action != "compact" || checkedKey != "favorite" || !checked || radioGroup != "density" || radioValue != "compact" || itemActions["copy"] != 1 || itemActions["favorite"] != 1 || itemActions["compact"] != 1 {
		t.Fatalf("callbacks = action %q checkbox %q/%v radio %q/%q item actions %v", action, checkedKey, checked, radioGroup, radioValue, itemActions)
	}
}

func TestMenuKeyboardNavigationWrapsAndSkipsDisabled(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := New("actions", []Item{
		{Key: "copy", Label: "Copy"},
		{Key: "paste", Label: "Paste", Disabled: true},
		{Key: "delete", Label: "Delete"},
	})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := core.PeekState[menuState](ctx, "actions", stateSlotMenu)
	router.Source().Execute(key.FocusCmd{Tag: &state.item("copy").clickable})
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if !router.Source().Focused(&state.item("delete").clickable) {
		t.Fatal("down arrow did not skip disabled item")
	}
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if !router.Source().Focused(&state.item("copy").clickable) {
		t.Fatal("down arrow did not wrap to first item")
	}
}

func TestMenuRejectsDuplicateItemKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate menu item keys did not panic")
		}
	}()
	New("actions", []Item{{Key: "copy", Label: "Copy"}, {Key: "copy", Label: "Copy again"}}).
		Layout(menuTestContext(), menuTestLayoutContext(nil, time.Unix(1, 0)))
}

func TestEmptyMenuClearsStaleSubmenuActivity(t *testing.T) {
	ctx := menuTestContext()
	widget := New("empty", nil)
	state := widget.stateFor(ctx)
	state.submenuActive = true
	widget.layoutWithVisual(ctx, menuTestLayoutContext(nil, time.Unix(1, 0)), state, true, widget.visual(ctx))
	if state.submenuActive {
		t.Fatal("empty Menu retained stale submenu activity")
	}
}

func TestMenuRootNavigationHandsOffOnlyAtLeafBoundary(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	next := 0
	widget := New("root-navigation", []Item{
		{Key: "leaf", Label: "Leaf"},
		{Key: "submenu", Label: "Submenu", Children: []Item{{Key: "child", Label: "Child"}}},
	})
	widget.onRootNext = func() { next++ }
	start := time.Unix(9, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := core.PeekState[menuState](ctx, "root-navigation", stateSlotMenu)
	if state == nil {
		t.Fatal("Menu state is missing")
	}
	router.Source().Execute(key.FocusCmd{Tag: &state.item("leaf").clickable})
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if next != 1 || state.openSubmenu != "" {
		t.Fatalf("root leaf navigation = next %d submenu %q", next, state.openSubmenu)
	}

	router.Source().Execute(key.FocusCmd{Tag: &state.item("submenu").clickable})
	layoutMenuFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenuFrame(ctx, router, widget, start.Add(4*time.Millisecond))
	if next != 1 || state.openSubmenu != "submenu" {
		t.Fatalf("submenu navigation = next %d submenu %q", next, state.openSubmenu)
	}
	child := widget.submenu(state, widget.items[1])
	if child.onRootNext == nil {
		t.Fatal("submenu did not inherit Menubar root navigation")
	}
}

func layoutMenuFrame(ctx *core.Context, router *input.Router, widget Widget, now time.Time) {
	gtx := menuTestLayoutContext(router, now)
	core.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(gtx.Ops)
}
