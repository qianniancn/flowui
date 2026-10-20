package listbox

import (
	"errors"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/navigation"
	"github.com/qianniancn/flowui/widget/text"
)

func listBoxItemColors(foreground, description, indicator color.NRGBA) theme.ListBoxItemColorsResolver {
	return func(style.ResolvedStyle) theme.ListBoxItemColors {
		return theme.ListBoxItemColors{Foreground: foreground, Description: description, Indicator: indicator}
	}
}

func listBoxTestItemDeclaration(foreground color.NRGBA) style.Style {
	return style.Style{}.
		MinHeight(40).
		PaddingX(4).
		PaddingY(4).
		Radius(4).
		Background(style.RGBA(0x00000000)).
		BorderWidth(1).
		BorderColor(style.RGBA(0x00000000)).
		Opacity(1).
		Scale(1).
		TextColor(style.Solid(foreground))
}

func listBoxTestItemDeclarationWithout(omission string) style.Style {
	item := style.Style{}.
		MinHeight(40).
		TextColor(style.Solid(color.NRGBA{A: 255}))
	if omission != "padding" {
		item = item.PaddingX(4).PaddingY(4)
	}
	if omission != "radius" {
		item = item.Radius(4)
	}
	if omission != "background" {
		item = item.Background(style.RGBA(0x00000000))
	}
	if omission != "border" {
		item = item.BorderWidth(1).BorderColor(style.RGBA(0x00000000))
	}
	if omission != "opacity" {
		item = item.Opacity(1)
	}
	if omission == "focus" {
		return item.On(style.FocusVisible, style.Style{}.BorderWidth(0))
	}
	return item.On(style.FocusVisible, style.Style{}.BorderWidth(1).BorderColor(style.Solid(color.NRGBA{B: 0xff, A: 0xff})))
}

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
}

func newContextWithTheme(_ any, value *theme.Theme) *core.Context {
	clone := *value
	if clone.VisualProvider == nil {
		clone.VisualProvider = testtheme.Behavior().VisualProvider
	}
	return core.New(nil, &clone, locale.LanguageEnglish)
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
	return value
}

func testSetComponentState[T any](ctx *core.Context, key, slot string, value *T) {
	core.UseStateWith(ctx, key, slot, func() *T { return value })
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
	box := New(listBoxTestItems()).Key("files").SelectedKey("go").
		EmptyText("Nothing here").
		OnChange(func(string) {}).
		OnSelectionChange(func([]string) {}).
		OnAction(func(string) {}).
		SelectionMode(SelectionNone).
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
	if box.selectionMode != SelectionNone || !listBoxSameKeys(box.disabledKeys, []string{"rust"}) {
		t.Fatal("listbox selection options were not set")
	}
	if !box.disabled || !box.fullWidth || !box.allowEmpty || !box.hideIndicator || box.maxHeight != 120 {
		t.Fatal("listbox visual options were not set")
	}
}

type testListBoxVisualProvider struct {
	theme.ComponentVisualProvider
	visual theme.ListBoxVisual
}

func (provider testListBoxVisualProvider) ListBox(_ *theme.Theme, request theme.ListBoxRequest) theme.ListBoxVisual {
	visual := provider.visual
	visual.Override = theme.ListBoxOverrideStyle(request)
	return visual
}

func TestListBoxUsesActiveVisualProvider(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	configured := base.ListBox(active, theme.ListBoxRequest{})
	configured.Metrics.ItemMinHeight = 91
	active.VisualProvider = testListBoxVisualProvider{ComponentVisualProvider: base, visual: configured}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	visual := state.listBoxVisualFor(ctx)
	if visual.Metrics.ItemMinHeight != 91 || visual.Item == nil {
		t.Fatalf("provider visual = %#v", visual)
	}
	if cached := state.listBoxVisualFor(ctx); cached.Metrics.ItemMinHeight != 91 {
		t.Fatalf("cached provider visual = %#v", cached)
	}
}

func TestListBoxVisualCacheTracksHitsAndThemeInvalidation(t *testing.T) {
	firstTheme := testtheme.Behavior()
	firstContext := core.New(nil, firstTheme, locale.LanguageEnglish)
	state := new(listBoxState)
	state.listBoxVisualFor(firstContext)
	state.listBoxVisualFor(firstContext)
	stats := state.visualCache.stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("visual cache stats = %#v, want one hit and one miss", stats)
	}

	secondTheme := testtheme.Behavior()
	secondContext := core.New(nil, secondTheme, locale.LanguageEnglish)
	state.listBoxVisualFor(secondContext)
	stats = state.visualCache.stats()
	if stats.Misses != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("visual cache theme stats = %#v, want second miss and one invalidation", stats)
	}
}

func TestListBoxVisualOverrideBypassesBaseCache(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingListBoxVisualProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	state.listBoxVisualFor(ctx)
	if len(provider.requests) != 1 {
		t.Fatalf("base visual calls = %d, want 1", len(provider.requests))
	}
	state.listBoxVisualForRequest(ctx, theme.ListBoxRequest{
		Override: &theme.ListBoxOverride{Declaration: style.Style{}.MinHeight(72)},
	})
	stats := state.visualCache.stats()
	if stats.Hits != 0 || stats.Misses != 1 {
		t.Fatalf("override changed base cache stats = %#v, want one base miss", stats)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("override visual calls = %d, want 2 total", len(provider.requests))
	}
	state.listBoxVisualFor(ctx)
	stats = state.visualCache.stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("base cache after override = %#v, want one hit and one miss", stats)
	}
}

func TestListBoxItemVisualCacheTracksThemeAndState(t *testing.T) {
	state := new(listBoxState)
	active := testtheme.Behavior()
	calls := 0
	visual := theme.ListBoxVisual{Item: func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		calls++
		return theme.ListBoxItemVisual{
			Defaults: style.Style{}.Part(style.PartItem, listBoxTestItemDeclaration(color.NRGBA{A: 255})),
			Colors:   listBoxItemColors(color.NRGBA{R: byte(calls), A: 255}, color.NRGBA{R: byte(calls), A: 255}, color.NRGBA{R: byte(calls), A: 255}),
		}
	}}
	request := theme.ListBoxItemRequest{}
	first := state.itemVisualFor(active, visual, request)
	second := state.itemVisualFor(active, visual, request)
	firstColors := first.Colors(style.ResolvedStyle{})
	secondColors := second.Colors(style.ResolvedStyle{})
	if calls != 1 || firstColors.Foreground != secondColors.Foreground {
		t.Fatalf("same request calls=%d, first=%#v second=%#v", calls, first, second)
	}
	stats := state.itemVisualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("cache stats after hit = %#v, want one hit and one miss", stats)
	}
	request.State.Hovered = true
	state.itemVisualFor(active, visual, request)
	if calls != 2 {
		t.Fatalf("state change calls=%d, want 2", calls)
	}
	stats = state.itemVisualCache.Stats()
	if stats.Misses != 2 {
		t.Fatalf("cache misses after state change = %d, want 2", stats.Misses)
	}
	nextTheme := testtheme.Behavior()
	state.itemVisualFor(nextTheme, visual, theme.ListBoxItemRequest{})
	if calls != 3 {
		t.Fatalf("theme change calls=%d, want 3", calls)
	}
	stats = state.itemVisualCache.Stats()
	if stats.ThemeInvalidations != 1 || stats.Size != 1 {
		t.Fatalf("theme invalidation stats = %#v, cache size = %d", stats, stats.Size)
	}
}

func TestListBoxItemVisualCacheIsBoundedAndEvictsLeastRecentlyUsed(t *testing.T) {
	state := new(listBoxState)
	active := testtheme.Behavior()
	calls := 0
	visual := theme.ListBoxVisual{Item: func(theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		calls++
		return theme.ListBoxItemVisual{
			Defaults: style.Style{}.Part(style.PartItem, listBoxTestItemDeclaration(color.NRGBA{A: 255})),
			Colors:   listBoxItemColors(color.NRGBA{R: byte(calls), A: 255}, color.NRGBA{R: byte(calls), A: 255}, color.NRGBA{R: byte(calls), A: 255}),
		}
	}}
	first := theme.ListBoxItemRequest{Variant: 1}
	state.itemVisualFor(active, visual, first)
	for i := range listBoxItemVisualCacheCapacity - 1 {
		state.itemVisualFor(active, visual, theme.ListBoxItemRequest{Variant: uint8(i + 2)})
	}
	state.itemVisualFor(active, visual, first)
	for i := range listBoxItemVisualCacheCapacity {
		state.itemVisualFor(active, visual, theme.ListBoxItemRequest{Variant: uint8(i + listBoxItemVisualCacheCapacity + 2)})
	}
	stats := state.itemVisualCache.Stats()
	if got := stats.Size; got != listBoxItemVisualCacheCapacity {
		t.Fatalf("cache size = %d, want %d", got, listBoxItemVisualCacheCapacity)
	}
	if stats.Evictions == 0 {
		t.Fatal("cache did not record an eviction")
	}
	state.itemVisualFor(active, visual, first)
	if calls <= listBoxItemVisualCacheCapacity {
		t.Fatalf("first request was unexpectedly retained after pressure, calls=%d", calls)
	}
}

func TestListBoxResolvedPartCacheTracksScopeAndTheme(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	visual := state.listBoxVisualFor(ctx)
	state.resolvedPartCache.prepare(ctx, active, visual.Override)
	request := theme.ListBoxItemRequest{State: style.State{Hovered: true}}
	itemVisual := state.itemVisualFor(active, visual, request)
	gtx := testLayoutContext()
	state.resolvedItemPart(ctx, gtx, active, "item-transition", request, itemVisual, visual.Override)
	state.resolvedItemPart(ctx, gtx, active, "item-transition", request, itemVisual, visual.Override)

	stats := state.resolvedPartCache.stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.Size != 1 {
		t.Fatalf("warm resolved cache stats = %#v, want one hit/miss and one entry", stats)
	}

	restore := core.PushStyle(ctx, style.Style{}.Padding(1))
	state.resolvedPartCache.prepare(ctx, active, visual.Override)
	state.resolvedItemPart(ctx, gtx, active, "item-transition", request, itemVisual, visual.Override)
	restore()
	stats = state.resolvedPartCache.stats()
	if stats.ScopeInvalidations != 1 || stats.Misses != 2 {
		t.Fatalf("scope invalidation stats = %#v, want one invalidation and two misses", stats)
	}
	state.resolvedPartCache.prepare(ctx, active, visual.Override)
	instanceStyle := style.Style{}.Padding(1)
	state.resolvedPartCache.prepare(ctx, active, instanceStyle)
	state.resolvedItemPart(ctx, gtx, active, "item-transition", request, itemVisual, instanceStyle)
	stats = state.resolvedPartCache.stats()
	if stats.OverrideInvalidations != 1 || stats.ScopeInvalidations != 2 || stats.Misses != 3 {
		t.Fatalf("override invalidation stats = %#v, want one override/two scope invalidations and three misses", stats)
	}

	replaced := *active
	core.ReplaceTheme(ctx, replaced)
	next := core.ActiveTheme(ctx)
	nextVisual := state.listBoxVisualFor(ctx)
	state.resolvedPartCache.prepare(ctx, next, nextVisual.Override)
	nextRequest := theme.ListBoxItemRequest{State: style.State{Hovered: true}}
	nextItemVisual := state.itemVisualFor(next, nextVisual, nextRequest)
	state.resolvedItemPart(ctx, gtx, next, "item-transition", nextRequest, nextItemVisual, nextVisual.Override)
	stats = state.resolvedPartCache.stats()
	if stats.ThemeInvalidations != 1 || stats.Misses != 4 {
		t.Fatalf("theme invalidation stats = %#v, want one invalidation and four misses", stats)
	}
}

func TestListBoxResolvedPartCacheBypassesUnsafeDeclarations(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	unsafeStyle := style.Style{}.Transition(style.PropOpacity, time.Second,
		style.TransitionEase(func(progress float32) float32 { return progress }))
	state.resolvedPartCache.prepare(ctx, active, unsafeStyle)
	if state.resolvedPartCache.enabled {
		t.Fatal("custom easing unexpectedly enabled the resolved part cache")
	}

	visual := theme.ListBoxVisual{Item: func(theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		return theme.ListBoxItemVisual{
			Defaults: style.Style{}.Part(style.PartItem,
				listBoxTestItemDeclaration(color.NRGBA{A: 255}).Transition(
					style.PropOpacity,
					time.Second,
					style.TransitionEase(func(progress float32) float32 { return progress }),
				)),
			Colors: listBoxItemColors(
				color.NRGBA{A: 255},
				color.NRGBA{A: 255},
				color.NRGBA{A: 255},
			),
		}
	}}
	request := theme.ListBoxItemRequest{}
	itemVisual := state.itemVisualFor(active, visual, request)
	if !itemVisual.CacheUnsafe {
		t.Fatal("unsafe item declaration was not normalized in the visual cache")
	}
	state.resolvedPartCache.prepare(ctx, active, style.Style{})
	state.resolvedItemPart(ctx, testLayoutContext(), active, "unsafe-item", request, itemVisual, style.Style{})
	if stats := state.resolvedPartCache.stats(); stats.Misses != 0 {
		t.Fatalf("unsafe item declaration populated resolved cache: %#v", stats)
	}
}

func TestListBoxMissingProviderPanics(t *testing.T) {
	active := theme.New(theme.DefaultSeed())
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defer func() {
		recovered := recover()
		if recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	New([]Item{{Key: "item", Label: "Item"}}).Key("neutral").
		Layout(ctx, testLayoutContext())
}

func TestListBoxIncompleteVisualProviderPanics(t *testing.T) {
	for _, test := range []struct {
		name   string
		remove func(*theme.ListBoxVisual)
	}{
		{name: "item", remove: func(visual *theme.ListBoxVisual) { visual.Item = nil }},
		{name: "indicator", remove: func(visual *theme.ListBoxVisual) { visual.Indicator = nil }},
		{name: "text", remove: func(visual *theme.ListBoxVisual) { visual.Text = nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			active := testtheme.Behavior()
			base := active.VisualProvider.(theme.ComponentVisualProvider)
			visual := base.ListBox(active, theme.ListBoxRequest{})
			test.remove(&visual)
			active.VisualProvider = testListBoxVisualProvider{
				ComponentVisualProvider: base,
				visual:                  visual,
			}
			ctx := core.New(nil, active, locale.LanguageEnglish)
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			New([]Item{{Key: "item", Label: "Item"}}).Key("incomplete").
				Layout(ctx, testLayoutContext())
		})
	}
}

func TestListBoxIncompleteItemVisualIsNotCached(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	visual := base.ListBox(active, theme.ListBoxRequest{})
	visual.Item = func(theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		return theme.ListBoxItemVisual{}
	}
	state := new(listBoxState)
	root := visual

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete item visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := state.itemVisualCache.Stats(); stats.Size != 0 {
			t.Fatalf("incomplete item visual entered cache: %#v", stats)
		}
	}()
	state.itemVisualFor(active, root, theme.ListBoxItemRequest{})
}

func TestListBoxIncompleteItemDeclarationIsNotCached(t *testing.T) {
	for _, test := range []struct {
		name    string
		request theme.ListBoxItemRequest
	}{
		{name: "padding"},
		{name: "background"},
		{name: "border"},
		{name: "radius"},
		{name: "opacity"},
		{name: "focus", request: theme.ListBoxItemRequest{State: style.State{FocusVisible: true}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			active := testtheme.Behavior()
			state := new(listBoxState)
			visual := theme.ListBoxVisual{Item: func(theme.ListBoxItemRequest) theme.ListBoxItemVisual {
				return theme.ListBoxItemVisual{
					Defaults: style.Style{}.Part(style.PartItem, listBoxTestItemDeclarationWithout(test.name)),
					Colors:   listBoxItemColors(color.NRGBA{A: 255}, color.NRGBA{A: 255}, color.NRGBA{A: 255}),
				}
			}}
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete item declaration panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
				if stats := state.itemVisualCache.Stats(); stats.Size != 0 {
					t.Fatalf("incomplete item declaration entered cache: %#v", stats)
				}
			}()
			state.itemVisualFor(active, visual, test.request)
		})
	}
}

func TestListBoxVisualValidationRejectsIncompleteStateDeclaration(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	visual := base.ListBox(active, theme.ListBoxRequest{})
	baseItem := visual.Item
	visual.Item = func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		if request.State.FocusVisible {
			return theme.ListBoxItemVisual{
				Defaults: style.Style{}.Part(style.PartItem, listBoxTestItemDeclarationWithout("focus")),
				Colors:   listBoxItemColors(color.NRGBA{A: 255}, color.NRGBA{A: 255}, color.NRGBA{A: 255}),
			}
		}
		return baseItem(request)
	}
	active.VisualProvider = testListBoxVisualProvider{ComponentVisualProvider: base, visual: visual}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	state := new(listBoxState)
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete state declaration panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := state.visualCache.stats(); stats.Size != 0 {
			t.Fatalf("incomplete visual entered root cache: %#v", stats)
		}
	}()
	state.listBoxVisualFor(ctx)
}

func TestListBoxStrictContextRejectsProviderMutation(t *testing.T) {
	active := testtheme.Behavior()
	ctx, err := core.NewWithCatalogStrict(nil, active, locale.LanguageEnglish, i18n.Catalog{})
	if err != nil {
		t.Fatalf("strict context construction failed: %v", err)
	}
	core.ActiveTheme(ctx).VisualProvider = nil
	mustPanic(t, func() {
		New([]Item{{Key: "item", Label: "Item"}}).Key("strict").
			Layout(ctx, testLayoutContext())
	})
}

func TestMultipleConstructor(t *testing.T) {
	box := New(listBoxTestItems()).Key("files").SelectedKeys([]string{"go", "rust"})

	if box.key != "files" || box.selectionMode != SelectionMultiple || len(box.selectedKeys) != 2 || len(box.items) != 3 {
		t.Fatal("multi listbox constructor did not set fields")
	}
	if box.hasEmptyText {
		t.Fatal("multi listbox constructor should leave the localized empty text unresolved")
	}
}

func TestSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := Sections("files", "go", sections)

	if box.key != "files" || box.selectedKey != "go" || len(box.sections) != 2 {
		t.Fatal("section listbox constructor did not set fields")
	}
	if got := box.allItems(); len(got) != 3 {
		t.Fatalf("all items = %d, want 3", len(got))
	}
}

func TestMultipleSectionsConstructor(t *testing.T) {
	sections := listBoxTestSections()
	box := MultipleSections("files", []string{"go"}, sections)

	if box.key != "files" || box.selectionMode != SelectionMultiple || len(box.selectedKeys) != 1 || len(box.sections) != 2 {
		t.Fatal("multi section listbox constructor did not set fields")
	}
}

func TestListBoxRejectsDuplicateItemKeys(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]Item{
			{Key: "go", Label: "Go"},
			{Key: "go", Label: "Go again"},
		})
	})
}

func TestListBoxRejectsEmptyItemKey(t *testing.T) {
	state := new(listBoxState)
	mustPanic(t, func() {
		state.checkItems([]Item{{Label: "Missing key"}})
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
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
	widget := func() Widget {
		return New(listBoxTestItems()).Key("languages").SelectedKey(selected).
			FullWidth().
			OnChange(func(key string) { selected = key })
	}
	layoutFrame := func(now time.Time) {
		viewport := image.Pt(300, 200)
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
		core.BeginFrameWithViewport(ctx, viewport)
		widget().Layout(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
		router.Frame(&ops)
	}

	now := time.Unix(4, 0)
	layoutFrame(now)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(30, 70)})
	layoutFrame(now.Add(2 * time.Millisecond))
	layoutFrame(now.Add(3 * time.Millisecond))

	state, _ := core.PeekState[listBoxState](ctx, "languages", stateSlotListBox)
	item := state.items["rust"]
	if selected != "rust" {
		t.Fatalf("selected = %q, want rust", selected)
	}
	if !router.Source().Focused(&item.Clickable) {
		t.Fatal("pointer click did not focus the selected item")
	}
	if core.FocusVisible(ctx, &item.Clickable, true) {
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		AllowEmptySelection().
		OnChange(func(key string) {
			changed = key
		}).
		Layout(ctx, testLayoutContext())

	if changed != "" {
		t.Fatalf("changed = %q, want empty", changed)
	}
}

func TestSelectionModeNoneOnlyRunsAction(t *testing.T) {
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
		SelectionMode(SelectionNone).
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

func TestMultipleClickAddsSelection(t *testing.T) {
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
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go"}).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"go", "rust"}) {
		t.Fatalf("changed = %#v, want [go rust]", changed)
	}
}

func TestMultipleClickRemovesSelection(t *testing.T) {
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
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go", "rust"}).
		OnSelectionChange(func(keys []string) {
			changed = keys
		}).
		Layout(ctx, testLayoutContext())

	if !listBoxSameKeys(changed, []string{"rust"}) {
		t.Fatalf("changed = %#v, want [rust]", changed)
	}
}

func TestMultipleDoesNotMutateSelectedKeys(t *testing.T) {
	selected := []string{"go", "go", "rust"}
	next := listBoxToggleSelectedKeys(selected, "zig")

	if !listBoxSameKeys(selected, []string{"go", "go", "rust"}) {
		t.Fatalf("selected = %#v, want original slice unchanged", selected)
	}
	if !listBoxSameKeys(next, []string{"go", "rust", "zig"}) {
		t.Fatalf("next = %#v, want deduped selection with zig", next)
	}
}

func TestMultipleDoesNotCallSingleChange(t *testing.T) {
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
	New(listBoxTestItems()).Key("languages").SelectedKeys([]string{"go"}).
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
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
	New(listBoxTestItems()).Key("languages").SelectedKey("go").
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
// feeding navigation.List); the navigation algorithms themselves are unit-tested in
// package navigation.

func TestListBoxMoveIndexSkipsDisabledAndDoesNotWrap(t *testing.T) {
	items := listBoxTestItems()
	if got, ok := navigation.Move(sliceList(items, nil), 0, 1, false); !ok || got != 1 {
		t.Fatalf("move from first = %d %v, want 1 true", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, nil), 1, 1, false); ok || got != 1 {
		t.Fatalf("move past end = %d %v, want current false", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, nil), 1, -1, false); !ok || got != 0 {
		t.Fatalf("move up = %d %v, want 0 true", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, []string{"rust"}), 0, 1, false); ok || got != 0 {
		t.Fatalf("move into disabled key = %d %v, want current false", got, ok)
	}
	if got, ok := navigation.Move(sliceList(items, []string{"rust"}), 1, 1, false); !ok || got != 0 {
		t.Fatalf("move away from disabled current = %d %v, want 0 true", got, ok)
	}
}

func TestListBoxFirstLastEnabled(t *testing.T) {
	items := []Item{
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
	items := []Item{
		{Key: "alpha", Label: "Alpha"},
		{Key: "beta", Label: "Beta", Disabled: true},
		{Key: "bravo", Label: "Bravo"},
		{Key: "charlie", Label: "Charlie"},
	}

	if index, ok := navigation.Match(sliceList(items, nil), 0, "b"); !ok || index != 2 {
		t.Fatalf("typeahead b = %d %v, want enabled Bravo", index, ok)
	}
	if index, ok := navigation.Match(sliceList(items, nil), 3, "a"); !ok || index != 0 {
		t.Fatalf("wrapped typeahead a = %d %v, want Alpha", index, ok)
	}
	if index, ok := navigation.Match(sliceList(items, []string{"bravo"}), 0, "b"); ok || index != 0 {
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
	expired := start.Add(100*time.Millisecond + listBoxTypeaheadTimeout + time.Millisecond)
	if got := state.appendTypeahead(expired, "g"); got != "g" {
		t.Fatalf("expired typeahead = %q, want g", got)
	}
}

func TestListBoxLayoutKeepsState(t *testing.T) {
	ctx := newContext(nil)
	dims := New(listBoxTestItems()).Key("languages").SelectedKey("go").Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("listbox returned empty dimensions")
	}
}

func TestSectionsLayout(t *testing.T) {
	ctx := newContext(nil)
	dims := Sections("languages", "go", listBoxTestSections()).Layout(ctx, testLayoutContext())

	if testComponentState[listBoxState](ctx, "languages", stateSlotListBox) == nil {
		t.Fatal("missing section listbox state")
	}
	if dims.Size == (image.Point{}) {
		t.Fatal("section listbox returned empty dimensions")
	}
}

func TestListBoxDataVersionCachesEntries(t *testing.T) {
	widget := Sections("cached", "one", []Section{{Title: "Group", Items: []Item{{Key: "one", Label: "One"}}}}).DataVersion(1)
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

func TestListBoxFullWidth(t *testing.T) {
	dims := New(listBoxTestItems()).Key("languages").SelectedKey("go").
		FullWidth().
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.X != 300 {
		t.Fatalf("width = %d, want 300", dims.Size.X)
	}
}

func TestListBoxMaxHeight(t *testing.T) {
	items := make([]Item, 12)
	for i := range items {
		items[i] = Item{Key: string(rune('a' + i)), Label: "Item"}
	}

	dims := New(items).Key("items").
		MaxHeight(80).
		Layout(newContext(nil), testLayoutContext())

	if dims.Size.Y > 80 {
		t.Fatalf("height = %d, want <= 80", dims.Size.Y)
	}
}

func TestListBoxThemeControlsItemHeight(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	visual := base.ListBox(active, theme.ListBoxRequest{})
	visual.Metrics.Padding = 0
	visual.Metrics.ItemMinHeight = 72
	item := visual.Item
	visual.Item = func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		value := item(request)
		value.Defaults = style.Join(value.Defaults, style.Style{}.
			Part(style.PartItem, style.Style{}.MinHeight(72)))
		return value
	}
	active.VisualProvider = testListBoxVisualProvider{ComponentVisualProvider: base, visual: visual}
	ctx := newContextWithTheme(nil, active)

	dims := New([]Item{{Key: "go", Label: "Go"}}).Key("languages").SelectedKey("go").
		Layout(ctx, testLayoutContext())

	if dims.Size.Y < 72 {
		t.Fatalf("height = %d, want at least 72", dims.Size.Y)
	}
}

func TestListBoxPartItemResolvesPerItemState(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	visual := base.ListBox(active, theme.ListBoxRequest{})
	visual.Metrics.Padding = 0
	visual.Metrics.Gap = 0
	visual.Item = func(theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		return theme.ListBoxItemVisual{
			Defaults: style.Style{}.Part(style.PartItem, listBoxTestItemDeclaration(color.NRGBA{A: 255})),
			Colors:   listBoxItemColors(color.NRGBA{A: 255}, color.NRGBA{A: 255}, color.NRGBA{A: 255}),
		}
	}
	active.VisualProvider = testListBoxVisualProvider{ComponentVisualProvider: base, visual: visual}
	ctx := newContextWithTheme(nil, active)
	declaration := style.Style{}.
		Part(style.PartItem, style.Style{}.
			Height(44).
			On(style.Selected, style.Style{}.Height(64)))

	dims := (New([]Item{
		{Key: "go", Label: "Go"},
		{Key: "rust", Label: "Rust"},
	}).Key("languages").SelectedKey("go")).Style(declaration).Layout(ctx, testLayoutContext())

	if dims.Size.Y != 108 {
		t.Fatalf("listbox height = %d, want selected 64 + default 44", dims.Size.Y)
	}
}

func TestListBoxIndicatorVisibility(t *testing.T) {
	box := New(listBoxTestItems()).Key("languages").SelectedKey("go")
	if !box.showIndicator(listBoxTestItems()[0]) {
		t.Fatal("single-selection listbox should show default indicator")
	}
	if box.SelectionMode(SelectionNone).showIndicator(listBoxTestItems()[0]) {
		t.Fatal("none-selection listbox should hide default indicator")
	}
	if box.HideIndicator().showIndicator(Item{
		Key:       "custom",
		Label:     "Custom",
		Indicator: func(bool) core.Widget { return text.New("!") },
	}) {
		t.Fatal("hide indicator should override custom indicator")
	}
}

func TestListBoxCustomIndicatorReceivesSelectedState(t *testing.T) {
	ctx := newContext(nil)
	state := new(listBoxState)
	visual := state.listBoxVisualFor(ctx)
	gotSelected := false
	item := Item{
		Key:   "custom",
		Label: "Custom",
		Indicator: func(selected bool) core.Widget {
			gotSelected = selected
			return nil
		},
	}

	New([]Item{item}).Key("custom").SelectedKey("custom").
		layoutIndicator(ctx, testLayoutContext(), item, theme.ListBoxItemColors{}, 0, true, visual)

	if !gotSelected {
		t.Fatal("custom indicator did not receive selected=true")
	}
}

func TestListBoxSelectionAnimation(t *testing.T) {
	state := new(listBoxItemState)
	active := testtheme.Behavior()
	visual := active.VisualProvider.(theme.ListBoxProvider).ListBox(active, theme.ListBoxRequest{})
	duration := visual.Metrics.SelectionDuration
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.Selection(gtx, false, duration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.Selection(gtx, true, duration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.Selection(gtx, true, duration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.Selection(gtx, true, duration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func listBoxTestItems() []Item {
	return []Item{
		{Key: "go", Label: "Go", Description: "Simple deployment"},
		{Key: "rust", Label: "Rust", Description: "Memory safety"},
		{Key: "swift", Label: "Swift", Description: "Unavailable", Disabled: true},
	}
}

func listBoxTestSections() []Section {
	return []Section{
		{
			Title: "Languages",
			Items: []Item{
				{Key: "go", Label: "Go"},
				{Key: "rust", Label: "Rust"},
			},
		},
		{
			Title: "Unavailable",
			Items: []Item{
				{Key: "swift", Label: "Swift", Disabled: true},
			},
		},
	}
}
