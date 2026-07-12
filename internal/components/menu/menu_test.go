package menu

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

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
		EmptyText("Nothing here").
		OnAction(func(string) {}).
		OnCheckedChange(func(string, bool) {}).
		OnRadioChange(func(string, string) {}).
		Disabled(true).
		Width(260)
	if configured.emptyText != "Nothing here" || configured.onAction == nil || configured.onCheckedChange == nil || configured.onRadioChange == nil || !configured.disabled || configured.width != 260 || len(configured.sections) != 1 {
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
	if tokens.Width != 220 || tokens.MaxHeight != 360 || tokens.Padding != 6 || tokens.Radius != 24 || tokens.BorderWidth != 0 || tokens.ItemGap != 2 || tokens.ItemMinHeight != 36 || tokens.ItemRadius != 16 {
		t.Fatalf("menu geometry = %#v", tokens)
	}
	if tokens.ItemPaddingX != 10 || tokens.ItemPaddingY != 6 || tokens.ItemContentGap != 12 || tokens.ItemTextSize != 14 || tokens.ShortcutTextSize != 12 {
		t.Fatalf("menu content tokens = %#v", tokens)
	}
	if tokens.IndicatorSize != 16 || tokens.IndicatorContentGap != 2 || tokens.CheckmarkSize != 10 || tokens.RadioDotSize != 8 || tokens.SubmenuIndicatorSize != 12 {
		t.Fatalf("menu indicator tokens = %#v", tokens)
	}
	if tokens.FocusRingWidth != 2 || tokens.PressedScale != 0.98 || tokens.ContextMenuOffset != 2 || tokens.EnterScale != 0.9 || tokens.ExitScale != 0.95 || tokens.ShadowOpacity != 1 {
		t.Fatalf("menu state tokens = %#v", tokens)
	}
	dark := theme.DarkTheme().Components.Menu
	if dark.ShadowOpacity != 0 || dark.BorderWidth != 1 {
		t.Fatalf("dark menu elevation = shadow %v border %v", dark.ShadowOpacity, dark.BorderWidth)
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
	if dims.Size.X != 220 || dims.Size.Y != 86 {
		t.Fatalf("menu size = %v, want 220x86", dims.Size)
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
	if dims.Size != image.Pt(220, 360) {
		t.Fatalf("long menu size = %v, want 220x360", dims.Size)
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
	if !ok || anchor.Min.Y != 6 {
		t.Fatalf("first visible submenu anchor = %v, present %v", anchor, ok)
	}
	if _, ok := state.anchors["action-4"]; ok {
		t.Fatal("offscreen submenu item retained an anchor")
	}
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

func layoutMenuFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) {
	gtx := menuTestLayoutContext(router, now)
	frame.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
}
