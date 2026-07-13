package menu

import (
	"fmt"
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
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type fixedMenuWidget struct {
	size image.Point
}

func (w fixedMenuWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
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
		OnAction(func(string) {}).
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnCheckedChange(func(string, bool) {}).
		OnRadioChange(func(string, string) {}).
		CloseOnSelect(false).
		Disabled(true).
		Width(260)
	if configured.emptyText != "Nothing here" || configured.onAction == nil || configured.onChange == nil || configured.onSelectionChange == nil || configured.onCheckedChange == nil || configured.onRadioChange == nil || !configured.disabled || configured.width != 260 || len(configured.sections) != 1 || configured.autoSeparateSections || !configured.hasCloseOnSelect || configured.closeOnSelect {
		t.Fatalf("configured menu = %#v", configured)
	}
	if base.emptyText != "No actions" || base.onAction != nil || base.disabled || base.width != 0 || len(base.sections) != 0 {
		t.Fatalf("base menu was mutated: %#v", base)
	}
	if MenuSeparator().Kind != ItemSeparator || MenuGroupLabel("Edit").Kind != ItemGroupLabel {
		t.Fatal("menu structural item constructors returned wrong kinds")
	}
}

func TestMenuHeroUIDefaultTheme(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Menu
	if tokens.Width != 220 || tokens.MaxHeight != 0 || tokens.Padding != 6 || tokens.Radius != 24 || tokens.BorderWidth != 0 || tokens.ItemGap != 2 || tokens.ItemMinHeight != 36 || tokens.ItemRadius != 16 {
		t.Fatalf("menu geometry = %#v", tokens)
	}
	if tokens.ItemPaddingX != 10 || tokens.ItemPaddingY != 6 || tokens.ItemContentGap != 12 || tokens.ItemTextSize != 14 || tokens.ShortcutTextSize != 14 || tokens.ShortcutHeight != 24 || tokens.ShortcutPaddingX != 8 {
		t.Fatalf("menu content tokens = %#v", tokens)
	}
	if tokens.IndicatorSize != 16 || tokens.IndicatorContentGap != 2 || tokens.CheckmarkSize != 10 || tokens.RadioDotSize != 8 || tokens.SubmenuIndicatorSize != 14 {
		t.Fatalf("menu indicator tokens = %#v", tokens)
	}
	if tokens.FocusRingWidth != 2 || tokens.FocusRingOffset != 2 || tokens.PressedScale != 0.98 || tokens.SubmenuGap != 8 || tokens.ContextMenuOffset != 2 || tokens.EnterScale != 0.9 || tokens.ExitScale != 0.95 || tokens.ShadowOpacity != 1 {
		t.Fatalf("menu state tokens = %#v", tokens)
	}
	dark := theme.DarkTheme().Components.Menu
	if dark.ShadowOpacity != 0 || dark.BorderWidth != 1 || dark.BorderColor != (color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d}) {
		t.Fatalf("dark menu elevation = shadow %v border %v", dark.ShadowOpacity, dark.BorderWidth)
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

func TestMenuFocusRingStaysInsideItemBounds(t *testing.T) {
	itemRect := image.Rect(0, 0, 200, 36)
	focusRect, focusRadius := menuFocusRingGeometry(itemRect, 12, 2, 2)
	if focusRect != image.Rect(3, 3, 197, 33) {
		t.Fatalf("focus rect = %v, want an inset ring", focusRect)
	}
	if focusRadius != 9 {
		t.Fatalf("focus radius = %d, want 9", focusRadius)
	}
	strokeBounds := focusRect.Inset(-1)
	if !strokeBounds.In(itemRect) {
		t.Fatalf("focus ring stroke %v escapes item bounds %v", strokeBounds, itemRect)
	}
}

func TestMenuHeroUIOverlayShadow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	style := menuPanelStyle(&activeTheme)
	layers := heroMenuShadow(style.shadow, style.shadowOpacity).EffectiveLayers()
	if len(layers) != 3 {
		t.Fatalf("menu shadow layers = %d, want 3", len(layers))
	}
	if layers[0].OffsetY != 2 || layers[0].Blur != 8 || layers[1].OffsetY != -6 || layers[1].Blur != 12 || layers[2].OffsetY != 14 || layers[2].Blur != 28 {
		t.Fatalf("menu shadow layers = %#v", layers)
	}
}

func TestMenuShadowHonorsThemeColorAlpha(t *testing.T) {
	shadow := heroMenuShadow(color.NRGBA{A: 0x80}, 1)
	if len(shadow.Layers) != 3 || shadow.Layers[0].Color.A != 8 || shadow.Layers[1].Color.A != 4 || shadow.Layers[2].Color.A != 10 {
		t.Fatalf("alpha-aware Menu shadow = %#v", shadow)
	}
}

func TestMenuHeroUIPressedScale(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	start := time.Unix(1, 0)
	gtx := menuTestLayoutContext(nil, start.Add(menuItemPressDuration/2))
	scale := menuItemScale(gtx, []widget.Press{{Start: start}}, &activeTheme, false)
	if scale <= activeTheme.Components.Menu.PressedScale || scale >= 1 {
		t.Fatalf("pressed menu scale = %v", scale)
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
	widget := Menu("actions", []Item{
		{Key: "copy", Label: "Copy"},
		{Key: "favorite", Label: "Favorite", Kind: ItemCheckbox, Checked: false, KeepOpen: true},
		{Key: "compact", Label: "Compact", Kind: ItemRadio, RadioGroup: "density", Value: "compact", KeepOpen: true},
	}).
		OnAction(func(key string) { action = key }).
		OnCheckedChange(func(key string, value bool) { checkedKey, checked = key, value }).
		OnRadioChange(func(group, value string) { radioGroup, radioValue = group, value })

	frame.BeginFrame(ctx)
	state := widget.stateFor(ctx)
	state.item("copy").clickable.Click()
	state.item("favorite").clickable.Click()
	state.item("compact").clickable.Click()
	widget.layout(ctx, gtx, state, true)
	frame.EndFrame(ctx)
	if action != "copy" || checkedKey != "favorite" || !checked || radioGroup != "density" || radioValue != "compact" {
		t.Fatalf("callbacks = action %q checkbox %q/%v radio %q/%q", action, checkedKey, checked, radioGroup, radioValue)
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
