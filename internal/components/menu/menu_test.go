package menu

import (
	"fmt"
	"image"
	"image/color"
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
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/render"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type fixedMenuWidget struct {
	size image.Point
}

func (w fixedMenuWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type measurableMenuWidget struct {
	size         image.Point
	layoutCalls  int
	measureCalls int
}

func (w *measurableMenuWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.layoutCalls++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func (w *measurableMenuWidget) Measure(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.measureCalls++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func menuTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
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
	base := Menu("actions", []Item{{Key: "copy", Label: "Copy"}})
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
	if base.emptyText != "No actions" || base.onActionEvent != nil || base.disabled || base.compact || base.width != 0 || len(base.sections) != 0 {
		t.Fatalf("base menu was mutated: %#v", base)
	}
	if MenuSeparator().Kind != ItemSeparator || MenuGroupLabel("Edit").Kind != ItemGroupLabel {
		t.Fatal("menu structural item constructors returned wrong kinds")
	}
}

func TestMenuAutoWidthMeasuresContentAndHonorsBounds(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	short := Menu("short", []Item{{Key: "one", Label: "One"}}).AutoWidth().MinWidth(160)
	long := Menu("long", []Item{{Key: "one", Label: "A considerably longer menu action"}}).AutoWidth().MinWidth(160).MaxWidth(280)
	shortWidth := short.preferredWidthPx(ctx, gtx)
	longWidth := long.preferredWidthPx(ctx, gtx)
	if longWidth <= shortWidth {
		t.Fatalf("long auto width = %d, short width = %d", longWidth, shortWidth)
	}
	shortGtx := gtx
	short.applyConstraints(ctx, &shortGtx, nil)
	if got := shortGtx.Constraints.Max.X; got < gtx.Dp(160) {
		t.Fatalf("short constrained width = %d, want at least 160dp", got)
	}
	long.applyConstraints(ctx, &gtx, nil)
	if got := gtx.Constraints.Max.X; got > gtx.Dp(280) {
		t.Fatalf("max width = %d, want <= 280dp", got)
	}
}

func TestMenuAutoWidthMeasuresEmptyText(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := Menu("empty", nil).EmptyText("No matching actions").AutoWidth()
	width := widget.preferredWidthPx(ctx, gtx)
	if width <= gtx.Dp(widget.themeTokens(ctx).Padding*2) {
		t.Fatalf("empty menu width = %d, expected the empty text to contribute", width)
	}
}

func TestMenuAutoWidthIncludesSectionTitlePadding(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := MenuSections("section-width", []Section{{Title: "A considerably long section title"}}).AutoWidth()
	tokens := widget.themeTokens(ctx)
	title := widget.measureTextWidth(ctx, gtx, "A considerably long section title", tokens.SectionTextSize, font.Medium)
	width := widget.preferredWidthPx(ctx, gtx)
	minimum := title + 2*gtx.Dp(tokens.SectionPaddingX) + 2*gtx.Dp(tokens.Padding)
	if width < minimum {
		t.Fatalf("section title width = %d, want at least %d including padding", width, minimum)
	}
}

func TestMenuAutoWidthUsesSideEffectFreeWidgetMeasurement(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	indicator := &measurableMenuWidget{size: image.Pt(22, 16)}
	widget := Menu("measurable", []Item{{
		Key: "more", Label: "More", Leading: leading,
		Children: []Item{{Key: "copy", Label: "Copy"}}, SubmenuIndicator: indicator,
	}}).AutoWidth()
	width := widget.preferredWidthPx(ctx, gtx)
	if width <= 0 || leading.measureCalls != 1 || indicator.measureCalls != 1 || leading.layoutCalls != 0 || indicator.layoutCalls != 0 {
		t.Fatalf("measurement width %d leading %#v indicator %#v", width, leading, indicator)
	}
}

func TestMenuAutoWidthCachesVersionedMeasurement(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	widget := Menu("cached-width", []Item{{Key: "more", Label: "More", Leading: leading}}).
		AutoWidth().DataVersion(1)
	state := &menuState{}
	first := widget.preferredWidthPxCached(ctx, gtx, state)
	second := widget.preferredWidthPxCached(ctx, gtx, state)
	if first != second || leading.measureCalls != 1 {
		t.Fatalf("cached width = %d/%d, measurement calls = %d", first, second, leading.measureCalls)
	}
	widget = widget.DataVersion(2)
	widget.preferredWidthPxCached(ctx, gtx, state)
	if leading.measureCalls != 2 {
		t.Fatalf("version change reused width measurement, calls = %d", leading.measureCalls)
	}
}

func TestMenuAutoWidthCacheInvalidatesLayoutOptions(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	leading := &measurableMenuWidget{size: image.Pt(48, 16)}
	widget := Menu("cached-options", []Item{{Key: "more", Label: "More", Leading: leading}}).
		AutoWidth().DataVersion(1)
	state := &menuState{}
	widget.preferredWidthPxCached(ctx, gtx, state)

	widget = widget.Compact(true)
	widget.preferredWidthPxCached(ctx, gtx, state)
	if leading.measureCalls != 2 {
		t.Fatalf("compact option reused width measurement, calls = %d", leading.measureCalls)
	}
	widget = widget.MinWidth(240)
	widget.preferredWidthPxCached(ctx, gtx, state)
	if leading.measureCalls != 3 {
		t.Fatalf("minimum width option reused width measurement, calls = %d", leading.measureCalls)
	}
	widget = widget.MaxWidth(320)
	widget.preferredWidthPxCached(ctx, gtx, state)
	if leading.measureCalls != 4 {
		t.Fatalf("maximum width option reused width measurement, calls = %d", leading.measureCalls)
	}
}

func TestMenuActionEventIncludesSubmenuPath(t *testing.T) {
	var got ActionEvent
	root := Menu("actions", []Item{{Key: "more", Children: []Item{{Key: "copy", Label: "Copy"}}}}).
		OnActionEvent(func(event ActionEvent) { got = event })
	state := &menuState{key: "actions"}
	child := root.submenu(state, root.items[0])
	child.activate(entry{item: child.items[0]})
	if got.Key != "copy" || len(got.Path) != 2 || got.Path[0] != "more" || got.Path[1] != "copy" || got.Item.Label != "Copy" {
		t.Fatalf("action event = %#v", got)
	}
}

func TestMenuDataVersionCachesEntries(t *testing.T) {
	widget := MenuSections("cached", []Section{{Title: "Edit", Items: []Item{{Key: "copy", Label: "Copy"}}}}).DataVersion(1)
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
	widget := MenuSections("sections", []Section{
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
	base := Menu("actions", nil)
	tokens := base.Compact(true).themeTokens(ctx)
	if tokens.Padding != 4 || tokens.Radius != 8 || tokens.ItemGap != 1 || tokens.ItemMinHeight != 28 || tokens.ItemRadius != 4 || tokens.ItemPaddingX != 8 || tokens.ItemPaddingY != 3 || tokens.ItemTextSize != 13 {
		t.Fatalf("compact menu tokens = %#v", tokens)
	}
	root := menuRootDeclaration(frame.ActiveTheme(ctx), tokens).Resolve(flowstyle.StyleState{})
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
	root := Menu("actions", []Item{{Key: "more", Children: []Item{{Key: "copy", Label: "Copy"}}}}).
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
	widget := Menu("runtime-menu", []Item{{
		Key: "more", Label: "More", Children: []Item{{Key: "copy", Label: "Copy"}},
	}})
	ctx := menuTestContext()

	frame.BeginFrame(ctx)
	runtime := widget.Runtime(ctx, "owner", "menu", nil)
	child := runtime.widget.submenu(runtime.state, runtime.widget.items[0])
	child.stateFor(ctx)
	frame.EndFrame(ctx)

	frame.BeginFrame(ctx)
	runtime = widget.Runtime(ctx, "owner", "menu", nil)
	runtime.state.openSubmenu = "more"
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("hover inspection claimed submenu key twice: %v", recovered)
		}
		frame.EndFrame(ctx)
	}()
	runtime.HoveredWithSubmenus(ctx)
	child = runtime.widget.submenu(runtime.state, runtime.widget.items[0])
	child.stateFor(ctx)
}

func TestMenuItemIndicatorUsesThemeColor(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	activeTheme.Components.Menu.IndicatorColor = want
	if got := menuItemStyle(&activeTheme, ItemDefault).indicator; got != want {
		t.Fatalf("menu indicator color = %#v, want %#v", got, want)
	}
}

func TestMenuCloseReportsPointerAndKeyboardModality(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	modalities := make([]bool, 0, 2)
	widget := Menu("actions", []Item{{Key: "copy", Label: "Copy"}}).withClose(func(visible bool) {
		modalities = append(modalities, visible)
	})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[menuState](ctx, "actions", stateSlotMenu)
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
	widget := Menu("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[menuState](ctx, "actions", stateSlotMenu)
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
	widget := Menu("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutMenuFrame(ctx, router, widget, start.Add(3*time.Millisecond))

	state, _ := frame.PeekState[menuState](ctx, "actions", stateSlotMenu)
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
	widget := Menu("actions", []Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[menuState](ctx, "actions", stateSlotMenu)
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
	frame.RequestFocusVisible(ctx, &item.clickable, true)
	if !menuItemFocusVisible(ctx, item, true) {
		t.Fatal("keyboard-focused submenu item lost its focus ring")
	}
	menuItemFocusVisible(ctx, item, false)

	frame.RequestFocusVisible(ctx, &item.clickable, false)
	for range 4 {
		frame.BeginFrame(ctx)
	}
	if menuItemFocusVisible(ctx, item, true) {
		t.Fatal("delayed pointer focus was exposed as keyboard focus")
	}

	nativeKeyboardItem := new(menuItemState)
	if !menuItemFocusVisible(ctx, nativeKeyboardItem, true) {
		t.Fatal("native keyboard focus without a pending request was hidden")
	}
}

func TestMenuUsesBalancedElevationShadow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	style := menuPanelStyle(&activeTheme)
	layers := render.ThemeShadow(activeTheme.Shadows.Menu, style.shadow, style.shadowOpacity).EffectiveLayers()
	if len(layers) != 3 {
		t.Fatalf("menu shadow layers = %d, want 3", len(layers))
	}
	if layers[0].OffsetY != 0 || layers[0].Blur != 6 || layers[0].Spread != 1 ||
		layers[1].OffsetY != 4 || layers[1].Blur != 12 ||
		layers[2].OffsetY != 12 || layers[2].Blur != 28 || layers[2].Spread != 2 {
		t.Fatalf("menu shadow layers = %#v", layers)
	}
}

func TestMenuShadowHonorsThemeColorAlpha(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	shadow := render.ThemeShadow(activeTheme.Shadows.Menu, color.NRGBA{A: 0x80}, 1)
	if len(shadow.Layers) != 3 || shadow.Layers[0].Color.A != 26 || shadow.Layers[1].Color.A != 13 || shadow.Layers[2].Color.A != 13 {
		t.Fatalf("alpha-aware Menu shadow = %#v", shadow)
	}
}

func TestMenuPartItemUsesPressedScale(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	resolved := flowstyle.CascadePart(
		flowstyle.StyleState{Pressed: true},
		flowstyle.PartItem,
		menuItemDefaultDeclaration(&activeTheme, activeTheme.Components.Menu),
	)
	if resolved.Trans == nil || resolved.Trans.ScaleX == nil || *resolved.Trans.ScaleX != activeTheme.Components.Menu.PressedScale {
		t.Fatalf("pressed menu item style = %#v", resolved.Trans)
	}
}

func TestMenuItemHoverUsesConfiguredThemeColor(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0x80}
	activeTheme.Components.Menu.HoverColor = want
	resolved := flowstyle.CascadePart(
		flowstyle.StyleState{Hovered: true},
		flowstyle.PartItem,
		menuItemDefaultDeclaration(&activeTheme, activeTheme.Components.Menu),
	)
	if resolved.Paint == nil {
		t.Fatal("hovered menu item has no paint")
	}
	background, ok := resolved.Paint.Background.(flowstyle.SolidColor)
	if !ok || background.Color != want {
		t.Fatalf("menu item hover background = %#v, want %#v", resolved.Paint.Background, want)
	}
}

func TestMenuDangerItemUsesDangerHover(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	resolved := flowstyle.CascadePart(
		flowstyle.StyleState{Hovered: true},
		flowstyle.PartItem,
		menuItemDefaultDeclaration(&activeTheme, activeTheme.Components.Menu),
		menuItemVariantDeclaration(&activeTheme, ItemDanger),
	)
	if resolved.Paint == nil || resolved.Paint.Background != flowstyle.TokenDangerSoftHover {
		t.Fatalf("danger menu hover background = %#v", resolved.Paint)
	}
}

func TestMenuLayoutUsesStableWidth(t *testing.T) {
	ctx := menuTestContext()
	dims := Menu("actions", []Item{{Key: "copy", Label: "Copy"}, {Key: "paste", Label: "Paste"}}).
		Layout(ctx, menuTestLayoutContext(nil, time.Unix(1, 0)))
	if dims.Size.X != 192 || dims.Size.Y != 86 {
		t.Fatalf("menu size = %v, want viewport-clamped 192x86", dims.Size)
	}
}

func TestMenuStyleWidthConstraintsArePreserved(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	widget := Menu("styled-width", []Item{{Key: "copy", Label: "Copy"}}).
		Style(flowstyle.Style{}.Width(unit.Dp(300)))
	dims := widget.Layout(ctx, gtx)
	if dims.Size.X != 300 {
		t.Fatalf("styled menu width = %d, want 300", dims.Size.X)
	}
}

func TestMenuWidthConstraintsPreserveParentMinimum(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	gtx.Constraints = layout.Constraints{Min: image.Pt(300, 0), Max: image.Pt(500, 400)}
	widget := Menu("constraint-conflict", []Item{{Key: "copy", Label: "Copy"}}).
		AutoWidth().MaxWidth(120)
	widget.applyConstraints(ctx, &gtx, nil)
	if got := gtx.Constraints.Max.X; got != 300 || gtx.Constraints.Min.X != 300 {
		t.Fatalf("conflicting width constraints = %v, want exact 300px parent minimum", gtx.Constraints)
	}
}

func TestMenuAutoWidthUsesItemPartFontWeight(t *testing.T) {
	ctx := menuTestContext()
	widget := Menu("weighted", []Item{{Key: "copy", Label: "Copy"}}).Style(
		flowstyle.Style{}.Part(flowstyle.PartItem, flowstyle.Style{}.FontWeight(int(font.Bold))),
	)
	if got := widget.preferredItemFontWeight(ctx, widget.items[0]); got != font.Bold {
		t.Fatalf("measured item font weight = %v, want %v", got, font.Bold)
	}
}

func TestMenuAutoWidthUsesItemPartPadding(t *testing.T) {
	ctx := menuTestContext()
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	base := Menu("padding", []Item{{Key: "copy", Label: "Copy"}})
	wide := base.Style(flowstyle.Style{}.Part(flowstyle.PartItem, flowstyle.Style{}.PaddingX(40)))
	baseWidth := base.preferredItemWidthPx(ctx, gtx, base.items[0])
	wideWidth := wide.preferredItemWidthPx(ctx, gtx, wide.items[0])
	wantDelta := 2 * (gtx.Dp(unit.Dp(40)) - gtx.Dp(base.themeTokens(ctx).ItemPaddingX))
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
	menu := Menu("long-actions", items)
	gtx := menuTestLayoutContext(nil, time.Unix(1, 0))
	frame.BeginFrame(ctx)
	dims := menu.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	if dims.Size != image.Pt(192, 400) {
		t.Fatalf("long menu size = %v, want viewport-constrained 192x400", dims.Size)
	}
	state, ok := frame.PeekState[menuState](ctx, "long-actions", stateSlotMenu)
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
	menu := Menu("scroll-anchors", items)
	start := time.Unix(1, 0)
	frame.BeginFrame(ctx)
	menu.Layout(ctx, menuTestLayoutContext(nil, start))
	frame.EndFrame(ctx)
	state, _ := frame.PeekState[menuState](ctx, "scroll-anchors", stateSlotMenu)
	state.list.ScrollTo(5)

	frame.BeginFrame(ctx)
	menu.Layout(ctx, menuTestLayoutContext(nil, start.Add(time.Millisecond)))
	frame.EndFrame(ctx)
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

func TestMenuSectionsSupportExplicitSeparatorsAndHeroSpacing(t *testing.T) {
	without := MenuSections("sections", []Section{
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
		widget := Menu("fruit", []Item{{Key: "apple", Label: "Apple"}}).
			SelectionMode(SelectionSingle).
			OnChange(func(key string) { selected = key }).
			OnSelectionChange(func(keys []string) { selection = keys })
		if !widget.activate(widget.actionableEntries()[0]) || selected != "apple" || len(selection) != 1 || selection[0] != "apple" {
			t.Fatalf("single selection = %q %v", selected, selection)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		selection := []string(nil)
		widget := Menu("styles", []Item{{Key: "bold", Label: "Bold"}}).
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
		widget := MenuSections("alignment", []Section{{
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
	widget := Menu("actions", []Item{
		{Key: "copy", Label: "Copy", OnAction: func() { itemActions["copy"]++ }},
		{Key: "favorite", Label: "Favorite", Kind: ItemCheckbox, Checked: false, KeepOpen: true, OnAction: func() { itemActions["favorite"]++ }},
		{Key: "compact", Label: "Compact", Kind: ItemRadio, RadioGroup: "density", Value: "compact", KeepOpen: true, OnAction: func() { itemActions["compact"]++ }},
	}).
		OnActionEvent(func(event ActionEvent) { action = event.Key }).
		OnCheckedChange(func(key string, value bool) { checkedKey, checked = key, value }).
		OnRadioChange(func(group, value string) { radioGroup, radioValue = group, value })

	frame.BeginFrame(ctx)
	state := widget.stateFor(ctx)
	state.item("copy").clickable.Click()
	state.item("favorite").clickable.Click()
	state.item("compact").clickable.Click()
	widget.layout(ctx, gtx, state, true)
	frame.EndFrame(ctx)
	if action != "compact" || checkedKey != "favorite" || !checked || radioGroup != "density" || radioValue != "compact" || itemActions["copy"] != 1 || itemActions["favorite"] != 1 || itemActions["compact"] != 1 {
		t.Fatalf("callbacks = action %q checkbox %q/%v radio %q/%q item actions %v", action, checkedKey, checked, radioGroup, radioValue, itemActions)
	}
}

func TestMenuKeyboardNavigationWrapsAndSkipsDisabled(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := Menu("actions", []Item{
		{Key: "copy", Label: "Copy"},
		{Key: "paste", Label: "Paste", Disabled: true},
		{Key: "delete", Label: "Delete"},
	})
	start := time.Unix(1, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[menuState](ctx, "actions", stateSlotMenu)
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
	Menu("actions", []Item{{Key: "copy", Label: "Copy"}, {Key: "copy", Label: "Copy again"}}).
		Layout(menuTestContext(), menuTestLayoutContext(nil, time.Unix(1, 0)))
}

func TestEmptyMenuClearsStaleSubmenuActivity(t *testing.T) {
	ctx := menuTestContext()
	widget := Menu("empty", nil)
	state := widget.stateFor(ctx)
	state.submenuActive = true
	widget.layout(ctx, menuTestLayoutContext(nil, time.Unix(1, 0)), state, true)
	if state.submenuActive {
		t.Fatal("empty Menu retained stale submenu activity")
	}
}

func TestMenuRootNavigationHandsOffOnlyAtLeafBoundary(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	next := 0
	widget := Menu("root-navigation", []Item{
		{Key: "leaf", Label: "Leaf"},
		{Key: "submenu", Label: "Submenu", Children: []Item{{Key: "child", Label: "Child"}}},
	})
	widget.onRootNext = func() { next++ }
	start := time.Unix(9, 0)
	layoutMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[menuState](ctx, "root-navigation", stateSlotMenu)
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

func layoutMenuFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) {
	gtx := menuTestLayoutContext(router, now)
	frame.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
}
