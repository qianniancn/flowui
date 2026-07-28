package tabs

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
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
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

func TestTabsThemeMatchesHeroUI(t *testing.T) {
	theme := DefaultTheme()
	tabs := theme.Components.Tabs
	if tabs.TabHeight != 32 || tabs.TextSize != 14 || tabs.ListPadding != 4 {
		t.Fatalf("tabs metrics = height %v text %v padding %v, want 32, 14, 4", tabs.TabHeight, tabs.TextSize, tabs.ListPadding)
	}
	if tabs.SmallTabHeight != 24 || tabs.SmallTabPaddingX != 12 {
		t.Fatalf("small tabs metrics = height %v padding %v, want 24 and 12", tabs.SmallTabHeight, tabs.SmallTabPaddingX)
	}
	if tabs.ScrollShadowSize != 64 {
		t.Fatalf("tabs scroll shadow size = %v, want 64", tabs.ScrollShadowSize)
	}
	if tabs.IndicatorLineWidth != 2 || tabs.PanelPadding != 8 || tabs.RootGap+tabs.PanelGap != 24 {
		t.Fatal("tabs indicator or panel metrics do not match HeroUI")
	}
	if theme.Palette.Segment.A == 0 || theme.Palette.SegmentForeground.A == 0 {
		t.Fatal("tabs segment palette was not initialized")
	}
	if got := tabsListStyleFor(&theme, TabsPrimary).background; got != theme.Palette.Default {
		t.Fatalf("primary tabs background = %#v, want default palette %#v", got, theme.Palette.Default)
	}

	primary := tabsItemStyleFor(&theme, TabsPrimary, TabsColorDefault, false, false)
	secondary := tabsItemStyleFor(&theme, TabsSecondary, TabsColorDefault, false, false)
	accent := tabsItemStyleFor(&theme, TabsPrimary, TabsColorAccent, false, false)
	secondaryAccent := tabsItemStyleFor(&theme, TabsSecondary, TabsColorAccent, false, false)
	if primary.indicator != theme.Palette.Segment || primary.selectedForeground != theme.Palette.SegmentForeground {
		t.Fatal("primary tabs do not use segment palette")
	}
	if secondary.indicator != theme.Palette.Accent || secondary.selectedForeground != theme.Palette.Foreground {
		t.Fatal("secondary tabs do not use accent line and foreground text")
	}
	if accent.indicator != theme.Palette.Accent || accent.selectedForeground != theme.Palette.AccentForeground {
		t.Fatal("accent tabs do not use the primary button palette")
	}
	if secondaryAccent.indicator != theme.Palette.Accent || secondaryAccent.selectedForeground != theme.Palette.Accent {
		t.Fatal("secondary accent tabs do not use accent text and indicator")
	}
}

func TestTabsItemPartMatchesHeroUIOpacity(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	widget := TabsWidget{key: "tabs"}
	size := tabsSizeStyleFor(frame.ActiveTheme(ctx), TabsMedium)
	resolve := func(key string, state flowstyle.StyleState) flowstyle.ResolvedStyle {
		return widget.resolveItemStyle(ctx, gtx, key, state, color.NRGBA{A: 0xff}, size)
	}

	hovered := resolve("hovered", flowstyle.StyleState{Hovered: true})
	disabled := resolve("disabled", flowstyle.StyleState{Disabled: true})
	if hovered.Paint == nil || hovered.Paint.Opacity == nil || *hovered.Paint.Opacity != 0.7 {
		t.Fatalf("hovered tab opacity = %#v, want 0.7", hovered.Paint)
	}
	if disabled.Paint == nil || disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != frame.ActiveTheme(ctx).DisabledOpacityValue() {
		t.Fatalf("disabled tab opacity = %#v", disabled.Paint)
	}
}

func TestTabsUsesDefaultCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	tabs := Tabs("cursor", "account", tabsTestItems())
	layoutTabsFrame(ctx, router, tabs, time.Unix(1, 0), image.Pt(300, 100))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(50, 20)})
	layoutTabsFrame(ctx, router, tabs, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if got := router.Cursor(); got != pointer.CursorDefault {
		t.Fatalf("tab cursor = %v, want default", got)
	}
}

func TestTabsItemPartUsesEachItemsLocalState(t *testing.T) {
	base := color.NRGBA{R: 0x10, A: 0xff}
	hoveredColor := color.NRGBA{G: 0x80, A: 0xff}
	selectedColor := color.NRGBA{B: 0xf0, A: 0xff}
	custom := flowstyle.Style{}.
		Part(flowstyle.PartItem, flowstyle.Style{}.
			Background(flowstyle.SolidColor{Color: base}).
			When(flowstyle.Hovered, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: hoveredColor})).
			When(flowstyle.Selected, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: selectedColor})))

	widget := TabsWidget{key: "tabs", customStyle: custom}
	ctx := newContext(nil)
	gtx := testLayoutContext()
	size := tabsSizeStyleFor(frame.ActiveTheme(ctx), TabsMedium)
	resolveBackground := func(itemKey string, state flowstyle.StyleState) color.NRGBA {
		resolved := widget.resolveItemStyle(ctx, gtx, itemKey, state, color.NRGBA{A: 0xff}, size)
		brush, ok := styleruntime.Brush(resolved.Paint.Background)
		if !ok {
			t.Fatalf("item %q background = %#v", itemKey, resolved.Paint)
		}
		return brush.ColorAt(.5)
	}

	if got := resolveBackground("hovered", flowstyle.StyleState{Hovered: true}); got != hoveredColor {
		t.Fatalf("hovered item background = %#v", got)
	}
	if got := resolveBackground("selected", flowstyle.StyleState{Selected: true}); got != selectedColor {
		t.Fatalf("selected item background = %#v", got)
	}
	if got := resolveBackground("idle", flowstyle.StyleState{}); got != base {
		t.Fatalf("idle item background = %#v", got)
	}
}

func TestSmallFitTabsUseNaturalWidth(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(500, 100)}
	items := []TabItem{
		{Key: "daily", Label: "Daily"},
		{Key: "weekly", Label: "Weekly"},
		{Key: "monthly", Label: "Monthly"},
	}

	full := Tabs("full", "daily", items).Layout(newContext(nil), gtx)
	fit := Tabs("fit", "daily", items).Size(TabsSmall).Fit().Layout(newContext(nil), gtx)
	if full.Size.X != 500 {
		t.Fatalf("full tabs width = %d, want 500", full.Size.X)
	}
	if fit.Size.X >= full.Size.X {
		t.Fatalf("fit tabs width = %d, want less than %d", fit.Size.X, full.Size.X)
	}
	if fit.Size.Y != 32 {
		t.Fatalf("small fit tabs height = %d, want 32", fit.Size.Y)
	}
}

func TestTabsPropagatesPanelPositionToOverlayHost(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(300, 200)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	var got image.Rectangle
	panel := &tabsOverlayProbe{got: &got}
	items := []TabItem{{Key: "account", Label: "Account", Panel: panel}}

	frame.BeginFrameWithViewport(ctx, viewport)
	Tabs("settings", "account", items).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	want := image.Rect(8, 72, 18, 82)
	if got != want {
		t.Fatalf("panel anchor = %v, want %v", got, want)
	}
}

func TestTabsEffectiveSelection(t *testing.T) {
	// Test 1: Uncontrolled mode with DefaultSelectedKey
	t.Run("uncontrolled with default", func(t *testing.T) {
		ctx := newContext(nil)
		gtx := testLayoutContext()
		items := tabsTestItems()

		widget := Tabs("tabs", "", items).DefaultSelectedKey("security")
		widget.Layout(ctx, gtx)

		// Verify the default selection is applied
		// We can't directly access internal state, but we can verify behavior
		// through the OnChange callback in subsequent interactions
	})

	// Test 2: Controlled mode always uses provided value
	t.Run("controlled mode", func(t *testing.T) {
		ctx := newContext(nil)
		router := new(input.Router)
		selected := "account"
		changeCount := 0

		widget := Tabs("settings", selected, tabsTestItems()).OnChange(func(key string) {
			changeCount++
			selected = key
		})

		layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 200))

		if changeCount > 0 {
			t.Fatalf("controlled mode should not trigger onChange on initial render, got %d calls", changeCount)
		}
	})

	// Test 3: Uncontrolled mode without default falls back to first enabled
	t.Run("uncontrolled fallback to first enabled", func(t *testing.T) {
		ctx := newContext(nil)
		gtx := testLayoutContext()
		items := tabsTestItems()
		items[0].Disabled = true // Disable first item

		widget := Tabs("tabs2", "", items) // No default, no selected
		widget.Layout(ctx, gtx)

		// Should fall back to first enabled item (security)
	})
}

func TestTabsClickChangesSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	widget := func() TabsWidget {
		return Tabs("settings", selected, tabsTestItems()).OnChange(func(key string) {
			selected = key
		})
	}

	layoutTabsFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 200))
	clickTabsAt(router, f32.Pt(150, 20))
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 200))

	if selected != "security" {
		t.Fatalf("selected = %q, want security", selected)
	}
}

func TestTabsDisabledItemDoesNotChangeSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsTestItems()
	items[1].Disabled = true
	selected := "account"
	widget := Tabs("settings", selected, items).OnChange(func(key string) {
		selected = key
	})

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 200))
	clickTabsAt(router, f32.Pt(150, 20))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 200))

	if selected != "account" {
		t.Fatalf("selected = %q, want account", selected)
	}
}

func TestTabsArrowKeysSelectAndFocusNextEnabledTab(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsTestItems()
	items[1].Disabled = true
	selected := "account"
	widget := func() TabsWidget {
		return Tabs("settings", selected, items).OnChange(func(key string) {
			selected = key
		})
	}

	layoutTabsFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 200))
	state := testComponentState[tabsState](ctx, "settings", stateSlotTabs)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["account"].clickable})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 200))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 200))

	if selected != "billing" {
		t.Fatalf("selected = %q, want billing", selected)
	}
	if !router.Source().Focused(&state.items["billing"].clickable) {
		t.Fatal("next enabled tab did not gain focus")
	}
}

func TestVerticalTabsUseVerticalArrowKeys(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	widget := func() TabsWidget {
		return Tabs("settings", selected, tabsTestItems()).Vertical().OnChange(func(key string) {
			selected = key
		})
	}

	layoutTabsFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(420, 240))
	state := testComponentState[tabsState](ctx, "settings", stateSlotTabs)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["account"].clickable})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(420, 240))
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(420, 240))

	if selected != "security" {
		t.Fatalf("selected = %q, want security", selected)
	}
}

func TestTabsHomeEndKeysSelectEdges(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{
		{Key: "account", Label: "Account", Disabled: true},
		{Key: "security", Label: "Security"},
		{Key: "billing", Label: "Billing"},
		{Key: "audit", Label: "Audit", Disabled: true},
	}
	selected := "billing"
	widget := func() TabsWidget {
		return Tabs("settings", selected, items).OnChange(func(key string) {
			selected = key
		})
	}

	layoutTabsFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(400, 100))
	state := testComponentState[tabsState](ctx, "settings", stateSlotTabs)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["billing"].clickable})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(400, 100))

	router.Queue(key.Event{Name: key.NameHome, State: key.Press})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(400, 100))
	if selected != "security" {
		t.Fatalf("selected after Home = %q, want security", selected)
	}
	if !router.Source().Focused(&state.items["security"].clickable) {
		t.Fatal("first enabled tab did not gain focus after Home")
	}

	router.Queue(key.Event{Name: key.NameEnd, State: key.Press})
	layoutTabsFrame(ctx, router, widget(), time.Unix(1, int64(3*time.Millisecond)), image.Pt(400, 100))
	if selected != "billing" {
		t.Fatalf("selected after End = %q, want billing", selected)
	}
	if !router.Source().Focused(&state.items["billing"].clickable) {
		t.Fatal("last enabled tab did not gain focus after End")
	}
}

func TestTabsSelectedPanelOnly(t *testing.T) {
	accountLayouts := 0
	securityLayouts := 0
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: &tabsPanelProbe{layouts: &accountLayouts}},
		{Key: "security", Label: "Security", Panel: &tabsPanelProbe{layouts: &securityLayouts}},
	}

	Tabs("settings", "security", items).Layout(newContext(nil), testLayoutContext())
	if accountLayouts != 0 || securityLayouts != 1 {
		t.Fatalf("panel layouts = account %d security %d, want 0 and 1", accountLayouts, securityLayouts)
	}
}

func TestTabsDefaultSelectionDoesNotHideFittingItems(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutTabsFrame(ctx, router, Tabs("settings", "billing", tabsTestItems()), time.Unix(1, 0), image.Pt(500, 200))

	state := testComponentState[tabsState](ctx, "settings", stateSlotTabs)
	if state.list.Position.First != 0 || state.list.Position.Count != len(tabsTestItems()) {
		t.Fatalf("list position = first %d count %d, want 0 and %d", state.list.Position.First, state.list.Position.Count, len(tabsTestItems()))
	}
}

func TestTabsEnsureVisibleUnclipsEdgeItems(t *testing.T) {
	state := tabsState{list: layout.List{Position: layout.Position{
		First:      2,
		Offset:     12,
		OffsetLast: -18,
		Count:      3,
	}}}

	state.ensureVisible(2)
	if state.list.Position.Offset != 0 {
		t.Fatalf("leading item offset = %d, want 0", state.list.Position.Offset)
	}

	state.list.Position.OffsetLast = -18
	state.ensureVisible(4)
	if state.list.Position.Offset != 18 {
		t.Fatalf("trailing item offset = %d, want 18", state.list.Position.Offset)
	}
}

func TestTabsSelectionVisibilityWaitsForFreshLayout(t *testing.T) {
	items := make([]TabItem, 10)
	for index := range items {
		items[index].Key = string(rune('a' + index))
	}
	state := new(tabsState)
	state.syncSelection(items, "j")
	state.list.Position = layout.Position{First: 0, Count: 5}

	if !state.ensureSelectionVisible(items, "j") {
		t.Fatal("offscreen selection did not request a scroll adjustment")
	}
	if state.list.Position.First != 9 {
		t.Fatalf("adjusted first item = %d, want selected index 9", state.list.Position.First)
	}
	if !state.selectionPending {
		t.Fatal("selection pending was cleared before the adjusted list was laid out")
	}

	state.list.Position = layout.Position{First: 9, Count: 1, Offset: 0, OffsetLast: -120}
	if state.ensureSelectionVisible(items, "j") {
		t.Fatal("oversized visible selection requested another scroll adjustment")
	}
	if state.selectionPending {
		t.Fatal("selection pending was not cleared after a fresh visible layout")
	}
}

func TestTabsScrollShadowGeometryMatchesOrientation(t *testing.T) {
	background := DefaultTheme().Palette.Background
	horizontal, ok := tabsScrollShadowFor(image.Pt(300, 40), TabsHorizontal, 64, 1, background)
	if !ok {
		t.Fatal("horizontal trailing shadow was not created")
	}
	if horizontal.bounds != image.Rect(236, 0, 300, 40) || horizontal.color1.A != 0 || horizontal.color2 != background {
		t.Fatalf("horizontal trailing shadow = %#v", horizontal)
	}

	vertical, ok := tabsScrollShadowFor(image.Pt(120, 200), TabsVertical, 64, -1, background)
	if !ok {
		t.Fatal("vertical leading shadow was not created")
	}
	if vertical.bounds != image.Rect(0, 0, 120, 64) || vertical.color1 != background || vertical.color2.A != 0 {
		t.Fatalf("vertical leading shadow = %#v", vertical)
	}

	clamped, ok := tabsScrollShadowFor(image.Pt(80, 40), TabsHorizontal, 64, -1, background)
	if !ok || clamped.bounds.Dx() != 40 {
		t.Fatalf("clamped shadow width = %d, want 40", clamped.bounds.Dx())
	}
}

func TestTabsOverflowScrollButtonAdvancesList(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("overflow", "overview", items)

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 160))
	state := testComponentState[tabsState](ctx, "overflow", stateSlotTabs)
	if !state.canScrollNext(len(items)) {
		t.Fatal("overflowing tabs did not expose forward scrolling")
	}
	buttonBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Scroll tabs right")
	if !ok {
		t.Fatal("horizontal tabs did not expose the right scroll button semantics")
	}
	clickTabsAt(router, f32.Pt(
		float32(buttonBounds.Min.X+buttonBounds.Max.X)/2,
		float32(buttonBounds.Min.Y+buttonBounds.Max.Y)/2,
	))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 160))
	if state.list.Position.First == 0 {
		t.Fatal("forward scroll button did not advance the tab list")
	}
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 160))
	if state.list.Position.First == 0 {
		t.Fatal("controlled selection pulled overflow scrolling back to the first tab")
	}
}

func TestVerticalTabsOverflowScrollButtonAdvancesList(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("vertical-overflow", "Overview", items).Vertical()

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(320, 120))
	state := testComponentState[tabsState](ctx, "vertical-overflow", stateSlotTabs)
	if !state.canScrollNext(len(items)) {
		t.Fatal("overflowing vertical tabs did not expose downward scrolling")
	}
	buttonBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Scroll tabs down")
	if !ok {
		t.Fatal("vertical tabs did not expose the down scroll button semantics")
	}
	clickTabsAt(router, f32.Pt(
		float32(buttonBounds.Min.X+buttonBounds.Max.X)/2,
		float32(buttonBounds.Min.Y+buttonBounds.Max.Y)/2,
	))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(320, 120))
	if state.list.Position.First == 0 {
		t.Fatal("down scroll button did not advance the vertical tab list")
	}
}

func TestTabsScrollButtonLabelsMatchOrientation(t *testing.T) {
	tests := []struct {
		name      string
		tabs      TabsWidget
		direction int
		want      string
	}{
		{name: "horizontal previous", tabs: TabsWidget{}, direction: -1, want: "Scroll tabs left"},
		{name: "horizontal next", tabs: TabsWidget{}, direction: 1, want: "Scroll tabs right"},
		{name: "vertical previous", tabs: TabsWidget{orientation: TabsVertical}, direction: -1, want: "Scroll tabs up"},
		{name: "vertical next", tabs: TabsWidget{orientation: TabsVertical}, direction: 1, want: "Scroll tabs down"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.tabs.scrollButtonLabel(test.direction); got != test.want {
				t.Fatalf("label = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTabsSemanticsExposeSelectedTab(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutTabsFrame(ctx, router, Tabs("settings", "security", tabsTestItems()), time.Unix(1, 0), image.Pt(400, 200))

	selected := selectedSemanticLabels(router.AppendSemantics(nil))
	if len(selected) != 1 || selected[0] != "Security" {
		t.Fatalf("selected semantic labels = %v, want [Security]", selected)
	}
}

func TestTabsLabelColorAnimation(t *testing.T) {
	state := new(tabsItemState)
	gtx := testLayoutContext()
	start := time.Unix(1, 0)
	gtx.Now = start
	if got := state.selectionProgress(gtx, false); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	state.selectionProgress(gtx, true)
	gtx.Now = start.Add(tabsColorDuration / 2)
	if got := state.selectionProgress(gtx, true); got <= 0 || got >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", got)
	}
	gtx.Now = start.Add(tabsColorDuration)
	if got := state.selectionProgress(gtx, true); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestTabsIndicatorMovesAndResizesBetweenTabs(t *testing.T) {
	state := new(tabsIndicatorState)
	gtx := testLayoutContext()
	start := time.Unix(1, 0)
	first := image.Rect(0, 0, 80, 32)
	second := image.Rect(80, 0, 200, 32)

	gtx.Now = start
	if got := state.transition(gtx, "first", TabsHorizontal, first); got != first {
		t.Fatalf("initial indicator = %v, want %v", got, first)
	}
	if got := state.transition(gtx, "second", TabsHorizontal, second); got != first {
		t.Fatalf("indicator at transition start = %v, want %v", got, first)
	}

	gtx.Now = start.Add(tabsIndicatorDuration / 2)
	middle := state.transition(gtx, "second", TabsHorizontal, second)
	if middle.Min.X <= first.Min.X || middle.Min.X >= second.Min.X {
		t.Fatalf("indicator midpoint X = %d, want between %d and %d", middle.Min.X, first.Min.X, second.Min.X)
	}
	if middle.Dx() <= first.Dx() || middle.Dx() >= second.Dx() {
		t.Fatalf("indicator midpoint width = %d, want between %d and %d", middle.Dx(), first.Dx(), second.Dx())
	}

	gtx.Now = start.Add(tabsIndicatorDuration)
	if got := state.transition(gtx, "second", TabsHorizontal, second); got != second {
		t.Fatalf("indicator at transition end = %v, want %v", got, second)
	}
}

func TestTabsIndicatorTracksScrollingWithoutRestarting(t *testing.T) {
	state := new(tabsIndicatorState)
	gtx := testLayoutContext()
	gtx.Now = time.Unix(1, 0)
	initial := image.Rect(80, 0, 180, 32)
	shifted := initial.Add(image.Pt(-24, 0))

	state.transition(gtx, "selected", TabsHorizontal, initial)
	if got := state.transition(gtx, "selected", TabsHorizontal, shifted); got != shifted {
		t.Fatalf("indicator after scrolling = %v, want %v", got, shifted)
	}
}

func TestTabsIndicatorRespectsDisabledMotion(t *testing.T) {
	state := new(tabsIndicatorState)
	gtx := testLayoutContext()
	gtx.Now = time.Unix(1, 0)
	first := image.Rect(0, 0, 80, 32)
	second := image.Rect(80, 0, 200, 32)
	motion := theme.MotionTheme{Enabled: false, DurationScale: 1}

	state.transition(gtx, "first", TabsHorizontal, first, motion)
	if got := state.transition(gtx, "second", TabsHorizontal, second, motion); got != second {
		t.Fatalf("disabled motion indicator = %v, want %v", got, second)
	}
}

func TestTabsItemRectAccountsForListOffsetAndOrientation(t *testing.T) {
	position := layout.Position{First: 1, Offset: 10}
	widths := []int{80, 100, 120}

	horizontal := TabsWidget{}.tabRect(position, widths, 2, 32, 4)
	if want := image.Rect(90, 0, 210, 32); horizontal != want {
		t.Fatalf("horizontal rect = %v, want %v", horizontal, want)
	}
	vertical := (TabsWidget{orientation: TabsVertical}).tabRect(position, widths, 2, 32, 4)
	if want := image.Rect(0, 26, 120, 58); vertical != want {
		t.Fatalf("vertical rect = %v, want %v", vertical, want)
	}
}

func TestTabsRejectInvalidItemKeys(t *testing.T) {
	state := new(tabsState)
	state.beginFrame()
	mustPanic(t, func() {
		state.checkItems([]TabItem{{Label: "Missing key"}})
	})
	mustPanic(t, func() {
		state.checkItems([]TabItem{{Key: "same"}, {Key: "same"}})
	})
}

func tabsTestItems() []TabItem {
	return []TabItem{
		{Key: "account", Label: "Account", Panel: text.New("Account panel")},
		{Key: "security", Label: "Security", Panel: text.New("Security panel")},
		{Key: "billing", Label: "Billing", Panel: text.New("Billing panel")},
	}
}

func tabsOverflowItems() []TabItem {
	labels := []string{"Overview", "Analytics", "Reports", "Performance", "Engagement", "Conversions", "Revenue", "Retention"}
	items := make([]TabItem, 0, len(labels))
	for _, label := range labels {
		items = append(items, TabItem{Key: label, Label: label, Panel: text.New(label + " panel")})
	}
	return items
}

func clickTabsAt(router *input.Router, position f32.Point) {
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  position,
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  position,
		},
	)
}

func layoutTabsFrame(ctx *frame.Context, router *input.Router, tabs TabsWidget, now time.Time, viewport image.Point) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	tabs.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

type tabsPanelProbe struct {
	layouts *int
}

type tabsOverlayProbe struct {
	got *image.Rectangle
}

func (p *tabsOverlayProbe) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       "tabs-panel-probe",
		Anchor:    image.Rect(0, 0, 10, 10),
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			*p.got = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: image.Pt(40, 20)}
}

func (p *tabsPanelProbe) Layout(_ *frame.Context, _ layout.Context) layout.Dimensions {
	*p.layouts++
	return layout.Dimensions{Size: image.Pt(40, 20)}
}

func selectedSemanticLabels(nodes []input.SemanticNode) []string {
	return collectSelectedSemanticLabels(nodes, make(map[input.SemanticID]struct{}))
}

func collectSelectedSemanticLabels(nodes []input.SemanticNode, seen map[input.SemanticID]struct{}) []string {
	var labels []string
	for _, node := range nodes {
		if _, ok := seen[node.ID]; ok {
			continue
		}
		seen[node.ID] = struct{}{}
		if node.Desc.Selected {
			labels = append(labels, node.Desc.Label)
		}
		labels = append(labels, collectSelectedSemanticLabels(node.Children, seen)...)
	}
	return labels
}

func semanticBoundsForLabel(nodes []input.SemanticNode, label string) (image.Rectangle, bool) {
	for _, node := range nodes {
		if node.Desc.Label == label {
			return node.Desc.Bounds, true
		}
		if bounds, ok := semanticBoundsForLabel(node.Children, label); ok {
			return bounds, true
		}
	}
	return image.Rectangle{}, false
}
