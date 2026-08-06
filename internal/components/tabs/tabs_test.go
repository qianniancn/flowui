package tabs

import (
	"image"
	"image/color"
	"slices"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	buttonui "github.com/qianniancn/flowui/internal/components/button"
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
	if tabs.LargeTabHeight != 40 || tabs.LargeTextSize != 16 || tabs.ColorDuration <= 0 || tabs.IndicatorDuration <= 0 || tabs.PanelDuration <= 0 {
		t.Fatal("extended tabs theme metrics were not initialized")
	}
	if tabs.IndicatorWidth != 0 {
		t.Fatalf("default indicator width = %v, want automatic width", tabs.IndicatorWidth)
	}
	if tabs.ExtraContentGap <= 0 {
		t.Fatal("extra-content tabs tokens were not initialized")
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

func TestTabsHorizontalGapMatchesIndicatorGeometry(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{
		{Key: "daily", Label: "Daily"},
		{Key: "weekly", Label: "Weekly"},
		{Key: "monthly", Label: "Monthly"},
	}
	layoutTabsFrame(ctx, router, Tabs("horizontal-gap", "daily", items).Fit(), time.Unix(1, 0), image.Pt(400, 100))
	state := testComponentState[tabsState](ctx, "horizontal-gap", stateSlotTabs)
	wantGap := testLayoutContext().Dp(frame.ActiveTheme(ctx).Components.Tabs.TabGap)
	if state.list.Gap != wantGap {
		t.Fatalf("horizontal list gap = %d, want theme gap %d", state.list.Gap, wantGap)
	}
	wantLength := max(len(state.lastListWidths)-1, 0) * wantGap
	for _, width := range state.lastListWidths {
		wantLength += width
	}
	if state.list.Position.Length != wantLength {
		t.Fatalf("horizontal list length = %d, want widths plus gaps %d", state.list.Position.Length, wantLength)
	}
}

func TestTabsOverflowIndicatorIncludesGapBeforeLastVisibleTab(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectedKey := "Conversions"
	widget := Tabs("overflow-indicator-gap", selectedKey, tabsOverflowItems()).
		Overflow(TabsOverflowMenu).
		OverflowTrigger(text.New("More"))
	viewport := image.Pt(640, 160)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(viewport)
	overflow := widget.measureOverflow(ctx, gtx, selectedKey)
	if len(overflow.hidden) == 0 {
		t.Fatal("test setup did not produce hidden overflow items")
	}
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), viewport)
	state := testComponentState[tabsState](ctx, "overflow-indicator-gap", stateSlotTabs)
	if !state.indicator.set {
		t.Fatal("selected tab indicator was not initialized")
	}
	selectedIndex := tabsIndexByKey(overflow.visible, selectedKey)
	if selectedIndex < 0 {
		t.Fatal("selected tab was not promoted into the visible strip")
	}
	tabHeight := gtx.Dp(tabsSizeStyleFor(frame.ActiveTheme(ctx), widget.size).height)
	wantGap := gtx.Dp(frame.ActiveTheme(ctx).Components.Tabs.TabGap)
	start := -state.list.Position.Offset
	if selectedIndex >= state.list.Position.First {
		for index := state.list.Position.First; index < selectedIndex; index++ {
			start += state.lastListWidths[index] + wantGap
		}
	} else {
		for index := state.list.Position.First - 1; index >= selectedIndex; index-- {
			start -= state.lastListWidths[index] + wantGap
		}
	}
	want := image.Rect(start, 0, start+state.lastListWidths[selectedIndex], tabHeight)
	// The default primary variant uses the whole tab slot for its indicator.
	if got := state.indicator.to; got != want {
		t.Fatalf("selected overflow indicator = %v, want %v (horizontal gaps included)", got, want)
	}
}

func TestCenteredTabsCenterIntrinsicStrip(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	layoutTabsFrame(ctx, router, Tabs("centered", "a", items).Centered(true), time.Unix(1, 0), image.Pt(400, 100))

	bounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "A")
	if !ok {
		t.Fatal("centered tab was not exposed in semantics")
	}
	if bounds.Min.X <= 0 || bounds.Max.X >= 400 {
		t.Fatalf("centered tab bounds = %v, want an inset strip inside 400px", bounds)
	}
}

func TestCenteredVerticalTabsKeepPanelAtStart(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{{Key: "a", Label: "A", Panel: text.New("Panel")}}
	layoutTabsFrame(ctx, router, Tabs("centered-vertical", "a", items).Vertical().Centered(true), time.Unix(1, 0), image.Pt(300, 240))

	bounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "A")
	if !ok {
		t.Fatal("vertical tab was not exposed in semantics")
	}
	if bounds.Min.Y > 8 {
		t.Fatalf("centered vertical tab y = %d, want strip aligned to the top", bounds.Min.Y)
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

func TestTabsBottomPlacementMovesPanelBeforeTheStrip(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(300, 200)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	var got image.Rectangle
	panel := &tabsOverlayProbe{got: &got}
	items := []TabItem{{Key: "account", Label: "Account", Panel: panel}}

	frame.BeginFrameWithViewport(ctx, viewport)
	Tabs("settings-bottom", "account", items).Placement(TabsBottom).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	want := image.Rect(8, 8, 18, 18)
	if got != want {
		t.Fatalf("bottom panel anchor = %v, want %v", got, want)
	}
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

func TestTabsUncontrolledClickChangesSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsTestItems()
	widget := Tabs("uncontrolled-settings", "", items).DefaultSelectedKey("account")

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 200))
	securityBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Security")
	if !ok {
		t.Fatal("security tab was not exposed in semantics")
	}
	clickTabsAt(router, f32.Pt(
		float32(securityBounds.Min.X+securityBounds.Max.X)/2,
		float32(securityBounds.Min.Y+securityBounds.Max.Y)/2,
	))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 200))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 200))

	selected := selectedSemanticLabels(router.AppendSemantics(nil))
	if len(selected) != 1 || selected[0] != "Security" {
		t.Fatalf("uncontrolled click selected labels = %v, want [Security]", selected)
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

func TestTabsWorkbenchShortcutMovesSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	items := tabsTestItems()
	widgetForFrame := func() TabsWidget {
		return Tabs("shortcut-tabs", selected, items).OnChange(func(key string) { selected = key })
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	state := testComponentState[tabsState](ctx, "shortcut-tabs", stateSlotTabs)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["account"].clickable})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	router.Queue(key.Event{Name: key.NamePageDown, Modifiers: key.ModShortcut, State: key.Press})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(360, 180))
	if selected != "security" {
		t.Fatalf("shortcut selection = %q, want security", selected)
	}
}

func TestTabsManualActivationMovesFocusBeforeSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	items := tabsTestItems()
	widgetForFrame := func() TabsWidget {
		return Tabs("manual-activation", selected, items).
			Activation(TabsActivationManual).
			OnChange(func(key string) { selected = key })
	}

	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	state := testComponentState[tabsState](ctx, "manual-activation", stateSlotTabs)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["account"].clickable})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(360, 180))
	if selected != "account" {
		t.Fatalf("manual arrow selection = %q, want account", selected)
	}
	if !router.Source().Focused(&state.items["security"].clickable) {
		t.Fatal("manual arrow navigation did not move focus")
	}

	router.Queue(key.Event{Name: key.NameReturn, State: key.Press})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(3*time.Millisecond)), image.Pt(360, 180))
	if selected != "security" {
		t.Fatalf("manual activation selection = %q, want security", selected)
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

func TestTabsPanelFadeHonorsDurationAndMotion(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	items := tabsTestItems()
	widgetForFrame := func() TabsWidget {
		return Tabs("panel-fade", selected, items).PanelTransition(TabsPanelFade)
	}

	start := time.Unix(1, 0)
	layoutTabsFrame(ctx, router, widgetForFrame(), start, image.Pt(360, 180))
	state := testComponentState[tabsState](ctx, "panel-fade", stateSlotTabs)
	if got := state.panelOpacity.Current(); got != 1 {
		t.Fatalf("initial panel opacity = %v, want 1", got)
	}
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Millisecond), image.Pt(360, 180))
	if got := state.panelOpacity.Current(); got != 0 {
		t.Fatalf("panel transition first-frame opacity = %v, want 0", got)
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(100*time.Millisecond), image.Pt(360, 180))
	if got := state.panelOpacity.Current(); got <= 0 || got >= 1 {
		t.Fatalf("mid-transition panel opacity = %v, want between 0 and 1", got)
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(500*time.Millisecond), image.Pt(360, 180))
	if got := state.panelOpacity.Current(); got != 1 {
		t.Fatalf("completed panel opacity = %v, want 1", got)
	}

	theme := frame.ActiveTheme(ctx)
	theme.Motion.Enabled = false
	selected = "account"
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Second), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Second+time.Millisecond), image.Pt(360, 180))
	if got := state.panelOpacity.Current(); got != 1 {
		t.Fatalf("disabled-motion panel opacity = %v, want 1", got)
	}
}

func TestTabsAccessibleLabelReplacesVisualLabelInSemantics(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{{Key: "settings", Label: "", AccessibleLabel: "Settings", Panel: text.New("Settings")}}
	layoutTabsFrame(ctx, router, Tabs("accessible", "settings", items), time.Unix(1, 0), image.Pt(240, 120))
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Settings"); !ok {
		t.Fatal("accessible tab label was not exposed in semantics")
	}
}

func TestTabsForceRenderInitializesHiddenPanelsWithoutExposingThem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	account := &tabsStatePanelProbe{}
	security := &tabsStatePanelProbe{}
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: account},
		{Key: "security", Label: "Security", Panel: security},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("force-render", selected, items).ForceRender(true)
	}

	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	if account.layouts != 1 || security.layouts != 1 {
		t.Fatalf("force-render layouts = account %d security %d, want one each", account.layouts, security.layouts)
	}
	if descriptions := semanticDescriptions(router.AppendSemantics(nil)); containsString(descriptions, "Tab panel: Security") {
		t.Fatal("hidden force-rendered panel leaked semantic operations")
	}
	if security.key == "" {
		t.Fatal("force-rendered hidden panel did not initialize its state")
	}
	if _, ok := frame.PeekState[widget.Clickable](ctx, security.key, "probe"); !ok {
		t.Fatal("force-rendered hidden panel state was not retained")
	}

	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if account.layouts != 1 || security.layouts != 2 {
		t.Fatalf("force-render second frame layouts = account %d security %d, want 1 and 2", account.layouts, security.layouts)
	}
}

func TestTabsForceRenderYieldsToDestroyOnHidden(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: &tabsStatePanelProbe{}},
		{Key: "security", Label: "Security", Panel: &tabsStatePanelProbe{}},
	}
	widget := Tabs("force-render-destroy", "account", items).ForceRender(true).DestroyOnHidden(true)
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(360, 180))
	state := testComponentState[tabsState](ctx, "force-render-destroy", stateSlotTabs)
	if state == nil || len(state.renderedPanels) != 0 {
		t.Fatalf("rendered panels with DestroyOnHidden = %#v, want empty", state)
	}
}

func TestTabsKeepAliveRetainsHiddenPanelState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	account := &tabsStatePanelProbe{}
	security := &tabsStatePanelProbe{}
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: account},
		{Key: "security", Label: "Security", Panel: security},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("settings", selected, items).KeepAlive(true)
	}

	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	if account.state == nil || account.key == "" {
		t.Fatal("selected panel did not create its state")
	}
	first := account.state
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if security.layouts != 1 {
		t.Fatalf("selected security panel layouts = %d, want 1", security.layouts)
	}
	retained, ok := frame.PeekState[widget.Clickable](ctx, account.key, "probe")
	if !ok || retained != first {
		t.Fatal("hidden account panel state was not retained")
	}
}

func TestTabsDestroyOnHiddenReleasesPanelState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	account := &tabsStatePanelProbe{}
	security := &tabsStatePanelProbe{}
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: account},
		{Key: "security", Label: "Security", Panel: security},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("settings", selected, items).KeepAlive(true).DestroyOnHidden(true)
	}

	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if _, ok := frame.PeekState[widget.Clickable](ctx, account.key, "probe"); ok {
		t.Fatal("destroy-on-hidden panel state was retained")
	}
}

func TestTabsRemovedItemReleasesRetainedPanelState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	account := &tabsStatePanelProbe{}
	security := &tabsStatePanelProbe{}
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: account},
		{Key: "security", Label: "Security", Panel: security},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("removed-panel", selected, items).KeepAlive(true)
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if security.key == "" {
		t.Fatal("security panel did not create retained state")
	}
	items = items[:1]
	selected = "account"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(360, 180))
	if _, ok := frame.PeekState[widget.Clickable](ctx, security.key, "probe"); ok {
		t.Fatal("removed security panel state was retained")
	}
}

func TestTabsPanelKeysAreScopedPerItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	account := &tabsStatePanelProbe{}
	security := &tabsStatePanelProbe{}
	items := []TabItem{
		{Key: "account", Label: "Account", Panel: account},
		{Key: "security", Label: "Security", Panel: security},
	}
	widgetForFrame := func() TabsWidget { return Tabs("settings", selected, items) }
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if account.key == security.key {
		t.Fatalf("panel state keys collided: %q", account.key)
	}
}

func TestTabsSlotsAndClosableItemsExposeLayout(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	leading := &tabsSlotProbe{}
	trailing := &tabsSlotProbe{}
	closed := ""
	items := []TabItem{{Key: "account", Label: "Account", Closable: true, Panel: text.New("Panel")}}
	widget := Tabs("settings", "account", items).
		Leading(leading).
		Trailing(trailing).
		OnClose(func(key string) { closed = key })
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(360, 180))
	if leading.layouts != 1 || trailing.layouts != 1 {
		t.Fatalf("slot layouts = leading %d trailing %d, want 1 and 1", leading.layouts, trailing.layouts)
	}
	closeBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Close Account")
	if !ok {
		t.Fatal("closable tab did not expose a close action")
	}
	clickTabsAt(router, f32.Pt(float32(closeBounds.Min.X+closeBounds.Max.X)/2, float32(closeBounds.Min.Y+closeBounds.Max.Y)/2))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if closed != "account" {
		t.Fatalf("closed key = %q, want account", closed)
	}
}

func TestTabsCustomContentUsesMeasuredWidth(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	content := &tabsMeasurableProbe{}
	widget := Tabs("custom-content-width", "custom", []TabItem{{Key: "custom", Content: content}}).Fit()
	dims := widget.Layout(ctx, gtx)
	wantMin := 44 + 2*gtx.Dp(frame.ActiveTheme(ctx).Components.Tabs.TabPaddingX)
	if dims.Size.X < wantMin {
		t.Fatalf("custom content tabs width = %d, want at least %d", dims.Size.X, wantMin)
	}
}

func TestTabsCustomContentKeysAreScopedPerItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	account := &tabsScopedContentProbe{}
	security := &tabsScopedContentProbe{}
	items := []TabItem{
		{Key: "account", Content: account},
		{Key: "security", Content: security},
	}

	layoutTabsFrame(ctx, router, Tabs("content-scope", "account", items).Fit(), time.Unix(1, 0), image.Pt(320, 100))

	if account.measureKey == "" || account.layoutKey == "" || security.measureKey == "" {
		t.Fatalf("content keys were not observed: account %#v security %#v", account, security)
	}
	if account.measureKey != account.layoutKey {
		t.Fatalf("account measure/layout scopes differ: %q and %q", account.measureKey, account.layoutKey)
	}
	if account.measureKey == security.measureKey {
		t.Fatalf("tab content keys collided: %q", account.measureKey)
	}
}

func TestTabsHostSlotsUseDistinctKeyScopes(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	leading := &tabsScopedContentProbe{}
	trailing := &tabsScopedContentProbe{}
	widget := Tabs("slot-scope", "account", []TabItem{{Key: "account", Label: "Account"}}).
		Fit().
		Leading(leading).
		Trailing(trailing)

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(320, 100))

	if leading.layoutKey == "" || trailing.layoutKey == "" {
		t.Fatalf("slot keys were not observed: leading %q trailing %q", leading.layoutKey, trailing.layoutKey)
	}
	if leading.layoutKey == trailing.layoutKey {
		t.Fatalf("host slot keys collided: %q", leading.layoutKey)
	}
}

func TestFitTabsKeepCustomContentIntrinsic(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	content := &tabsConstraintProbe{}
	widget := Tabs("fit-content", "custom", []TabItem{{Key: "custom", Content: content}}).Fit()
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(400, 100))

	if content.constraints.Min.X != 0 {
		t.Fatalf("intrinsic content min width = %d, want 0", content.constraints.Min.X)
	}
	if content.constraints.Max.X != 44 {
		t.Fatalf("intrinsic content max width = %d, want 44", content.constraints.Max.X)
	}
}

func TestClosableTabsDoNotDoubleRightPadding(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	sizeStyle := tabsSizeStyleFor(frame.ActiveTheme(ctx), TabsMedium)
	item := TabItem{Key: "file", Label: "file.go"}
	plain := Tabs("plain-width", "file", []TabItem{item}).Fit()
	closable := Tabs("closable-width", "file", []TabItem{{Key: item.Key, Label: item.Label, Closable: true}}).
		Fit().OnClose(func(string) {})

	plainWidth := plain.measureTabWidth(ctx, gtx, item, sizeStyle)
	closableItem := closable.items[0]
	closableWidth := closable.measureTabWidth(ctx, gtx, closableItem, sizeStyle)
	closeSize, closeGap := closable.closeButtonMetrics(ctx)
	closeSlot := gtx.Dp(closeSize) + gtx.Dp(closeGap)
	wantWidth := plainWidth + max(closeSlot-gtx.Dp(sizeStyle.paddingX), 0)
	if closableWidth != wantWidth {
		t.Fatalf("closable tab width = %d, want %d (plain %d, close slot %d, padding %d)", closableWidth, wantWidth, plainWidth, closeSlot, gtx.Dp(sizeStyle.paddingX))
	}
}

func TestTabsCloseButtonDoesNotSelectAnotherTab(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	changed := ""
	closed := ""
	items := []TabItem{
		{Key: "account", Label: "Account", Closable: true, Panel: text.New("Account")},
		{Key: "security", Label: "Security", Closable: true, Panel: text.New("Security")},
	}
	widget := Tabs("settings-close", selected, items).
		OnChange(func(key string) { changed = key }).
		OnClose(func(key string) { closed = key })
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(400, 180))
	closeBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Close Security")
	if !ok {
		t.Fatal("security close action was not exposed")
	}
	clickTabsAt(router, f32.Pt(float32(closeBounds.Min.X+closeBounds.Max.X)/2, float32(closeBounds.Min.Y+closeBounds.Max.Y)/2))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(400, 180))
	if closed != "security" {
		t.Fatalf("closed key = %q, want security", closed)
	}
	if changed != "" {
		t.Fatalf("close click selected another tab: %q", changed)
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

func TestTabsSelectionVisibilityRechecksAfterResize(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("resize-selection", "Revenue", items)

	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(560, 120))
	state := testComponentState[tabsState](ctx, "resize-selection", stateSlotTabs)
	if state == nil {
		t.Fatal("tabs state was not initialized")
	}
	// Simulate the user scrolling back to the start after the initial layout.
	// The next frame has unchanged geometry, so selection visibility must not
	// fight that explicit scroll position.
	state.selectionPending = false
	state.list.Position.First = 0
	state.list.Position.Offset = 0
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(560, 120))

	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(2*time.Millisecond)), image.Pt(150, 120))
	index := tabsIndexByKey(items, "Revenue")
	if !state.selectionFullyVisible(index) {
		t.Fatalf("selected item remained clipped after resize: position=%+v", state.list.Position)
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

	layoutTabsFrameWithOverlays(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 160))
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

func TestTabsOverflowMenuMeasuresVisibleItemsAndExposesMoreTrigger(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("overflow-menu", "Overview", items).Overflow(TabsOverflowMenu)
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 160))
	state := testComponentState[tabsState](ctx, "overflow-menu", stateSlotTabs)
	if state.list.Position.Count == 0 || state.list.Position.Count >= len(items) {
		t.Fatalf("visible tab count = %d, want between 1 and %d", state.list.Position.Count, len(items)-1)
	}
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "More"); !ok {
		t.Fatal("overflow menu did not expose the More trigger")
	}
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Revenue"); ok {
		t.Fatal("hidden tab was laid out in the main tab strip")
	}
}

func TestTabsOverflowTriggerUsesMeasuredCustomContent(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	trigger := &tabsMeasurableProbe{}
	widget := Tabs("overflow-custom-trigger", "Overview", tabsOverflowItems()).
		Overflow(TabsOverflowMenu).
		OverflowTrigger(trigger).
		MoreLabel("More tabs")
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 160))
	if trigger.layouts != 1 {
		t.Fatalf("custom overflow trigger layouts = %d, want one visible layout", trigger.layouts)
	}
	bounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "More tabs")
	if !ok || bounds.Dx() <= 0 {
		t.Fatalf("custom overflow trigger semantics = %v, want labeled non-empty trigger", bounds)
	}
}

func TestTabsOverflowTriggerCanMeasureStatefulButton(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	trigger := buttonui.Button("custom-more-button", text.New("More")).Label("More tabs")
	widget := Tabs("overflow-button-trigger", "Overview", tabsOverflowItems()).
		Overflow(TabsOverflowMenu).
		OverflowTrigger(trigger)
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 160))
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "More tabs"); !ok {
		t.Fatal("stateful custom overflow trigger was not laid out after measurement")
	}
}

func TestTabsOverflowTriggersAreScopedPerTabs(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(130, 80)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Now:         time.Unix(1, 0),
		Ops:         new(op.Ops),
		Source:      router.Source(),
	}
	items := []TabItem{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "three", Label: "Three"}}
	trigger := buttonui.Button("shared-overflow-trigger", text.New("More"))

	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	Tabs("first", "one", items).Fit().Overflow(TabsOverflowMenu).OverflowTrigger(trigger).Layout(ctx, gtx)
	Tabs("second", "one", items).Fit().Overflow(TabsOverflowMenu).OverflowTrigger(trigger).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)
}

func TestTabsOverflowMenuSelectsHiddenItemAndRequestsVisibility(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "Overview"
	changed := ""
	items := tabsOverflowItems()
	widgetForFrame := func() TabsWidget {
		return Tabs("overflow-menu-select", selected, items).
			Overflow(TabsOverflowMenu).
			OnChange(func(key string) { changed = key; selected = key })
	}
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(300, 160))
	moreBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "More")
	if !ok {
		t.Fatal("overflow menu trigger was not found")
	}
	morePoint := f32.Pt(float32(moreBounds.Min.X+moreBounds.Max.X)/2, float32(moreBounds.Min.Y+moreBounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: morePoint})
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 160))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: morePoint})
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 160))
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(200*time.Millisecond)), image.Pt(300, 160))
	itemBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Reports")
	if !ok {
		t.Fatal("opening More did not expose a hidden item")
	}
	itemPoint := f32.Pt(float32(itemBounds.Min.X+itemBounds.Max.X)/2, float32(itemBounds.Min.Y+itemBounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: itemPoint})
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(201*time.Millisecond)), image.Pt(300, 160))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: itemPoint})
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(202*time.Millisecond)), image.Pt(300, 160))
	layoutTabsFrameWithOverlays(ctx, router, widgetForFrame(), time.Unix(1, int64(203*time.Millisecond)), image.Pt(300, 160))
	if changed != "Reports" || selected != "Reports" {
		t.Fatalf("hidden selection = changed %q selected %q, want Reports", changed, selected)
	}
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Reports"); !ok {
		t.Fatal("selected hidden tab was not promoted back into the visible strip")
	}
}

func TestTabsOverflowItemsKeepSelectedItemVisible(t *testing.T) {
	items := tabsOverflowItems()
	widget := TabsWidget{items: items}
	visible, hidden := widget.overflowItems(3, "Performance")
	if len(visible) != 3 || visible[2].Key != "Performance" {
		t.Fatalf("visible overflow items = %#v, want Performance in the last visible slot", visible)
	}
	if tabsIndexByKey(hidden, "Performance") >= 0 || tabsIndexByKey(hidden, "Reports") < 0 {
		t.Fatalf("hidden overflow items = %#v, selected item or displaced item has wrong visibility", hidden)
	}
}

func TestTabsOverflowSelectedLastVisibleGeometry(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("overflow-selected-last", "Conversions", items).
		Overflow(TabsOverflowMenu).
		OverflowTrigger(text.New("More"))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(640, 160))
	nodes := router.AppendSemantics(nil)
	selectedNode, ok := semanticNodeByDescription(nodes, "Selected tab")
	if !ok {
		t.Fatal("selected tab was not exposed in semantics")
	}
	selected := selectedNode.Desc.Bounds
	more, ok := semanticBoundsForLabel(nodes, "More")
	if !ok {
		t.Fatal("overflow trigger was not exposed in semantics")
	}
	if selected.Max.X > more.Min.X {
		t.Fatalf("selected tab %v overlaps overflow trigger %v", selected, more)
	}
}

func TestTabsOverflowMeasurementAllocatesSiblingsWithinExactConstraints(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(660, 160))
	widget := Tabs("overflow-measure-exact", "Conversions", tabsOverflowItems()).
		Overflow(TabsOverflowMenu).
		OverflowTrigger(text.New("More"))
	overflow := widget.measureOverflow(ctx, gtx, "Conversions")
	if overflow.moreSize.X <= 0 || overflow.moreSize.X >= gtx.Constraints.Max.X {
		t.Fatalf("More width = %d, want a content-sized child below %d", overflow.moreSize.X, gtx.Constraints.Max.X)
	}
	if total := overflow.listSize.X + overflow.gap + overflow.moreSize.X; total > gtx.Constraints.Max.X {
		t.Fatalf("overflow siblings width = %d, exceeds available %d", total, gtx.Constraints.Max.X)
	}
	if overflow.listSize.Y >= gtx.Constraints.Max.Y || overflow.moreSize.Y >= gtx.Constraints.Max.Y {
		t.Fatalf("overflow child heights = list %d, More %d; expected tab-row sizing below %d", overflow.listSize.Y, overflow.moreSize.Y, gtx.Constraints.Max.Y)
	}
}

func TestTabsOverflowMeasurementClampsTinyConstraints(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(0, 0))
	overflow := (TabsWidget{items: tabsOverflowItems(), overflowMode: TabsOverflowMenu}).measureOverflow(ctx, gtx, "Overview")
	if overflow.listSize.X < 0 || overflow.listSize.Y < 0 || overflow.moreSize.X < 0 || overflow.moreSize.Y < 0 {
		t.Fatalf("tiny overflow sizes became negative: list %v more %v", overflow.listSize, overflow.moreSize)
	}
}

func TestTabsCloseSelectedFallsBackToNextEnabledTab(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	changed := ""
	closed := ""
	items := []TabItem{
		{Key: "account", Label: "Account", Closable: true, Panel: text.New("Account")},
		{Key: "security", Label: "Security", Panel: text.New("Security")},
		{Key: "billing", Label: "Billing", Disabled: true, Panel: text.New("Billing")},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("close-fallback", selected, items).
			OnChange(func(key string) { changed = key; selected = key }).
			OnClose(func(key string) { closed = key })
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(400, 180))
	closeBounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Close Account")
	if !ok {
		t.Fatal("selected tab did not expose its close action")
	}
	clickTabsAt(router, f32.Pt(float32(closeBounds.Min.X+closeBounds.Max.X)/2, float32(closeBounds.Min.Y+closeBounds.Max.Y)/2))
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(400, 180))
	if closed != "account" || changed != "security" || selected != "security" {
		t.Fatalf("close fallback = closed %q changed %q selected %q", closed, changed, selected)
	}
}

func TestTabsOnAddExposesAddTabAction(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	added := 0
	widget := Tabs("add-tab", "account", []TabItem{{Key: "account", Label: "Account"}}).
		OnAdd(func() { added++ })
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(320, 120))
	bounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Add tab")
	if !ok {
		t.Fatal("OnAdd did not expose an Add tab action")
	}
	clickTabsAt(router, f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(320, 120))
	if added != 1 {
		t.Fatalf("add callback count = %d, want 1", added)
	}
}

func TestTabsInlineEditCommitsWithEnter(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	committedKey, committedLabel := "", ""
	editing := "account"
	widgetForFrame := func() TabsWidget {
		return Tabs("edit-tab", "account", []TabItem{{Key: "account", Label: "Account", Editable: true}}).
			EditingKey(editing).
			OnEdit(func(key, label string) { committedKey, committedLabel = key, label })
	}
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(320, 120))
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(320, 120))
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "Account"); !ok {
		t.Fatal("editing tab did not expose its editor label")
	}
	router.Queue(key.Event{Name: key.NameEnter, State: key.Press})
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(320, 120))
	if committedKey != "account" || committedLabel != "Account" {
		t.Fatalf("committed edit = %q/%q, want account/Account", committedKey, committedLabel)
	}
}

func TestTabsInlineEditCommitsWhenFocusLeavesEditor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	editing := "account"
	committed := ""
	items := []TabItem{
		{Key: "account", Label: "Account", Editable: true},
		{Key: "security", Label: "Security"},
	}
	widgetForFrame := func() TabsWidget {
		return Tabs("edit-blur", "account", items).
			EditingKey(editing).
			OnEdit(func(key, label string) {
				committed = key + ":" + label
				editing = ""
			})
	}
	start := time.Unix(1, 0)
	layoutTabsFrame(ctx, router, widgetForFrame(), start, image.Pt(360, 120))
	state := testComponentState[tabsState](ctx, "edit-blur", stateSlotTabs)
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Millisecond), image.Pt(360, 120))
	if !router.Source().Focused(&state.items["account"].editor) {
		t.Fatal("inline editor did not receive focus")
	}
	router.Source().Execute(key.FocusCmd{Tag: &state.items["security"].clickable})
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(2*time.Millisecond), image.Pt(360, 120))
	if committed != "account:Account" {
		t.Fatalf("blur commit = %q, want account:Account", committed)
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

func TestVerticalTabsOverflowMenuReservesMoreRow(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsOverflowItems()
	widget := Tabs("vertical-overflow-menu", "Overview", items).
		Vertical().Overflow(TabsOverflowMenu)
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(320, 120))
	state := testComponentState[tabsState](ctx, "vertical-overflow-menu", stateSlotTabs)
	if state.list.Position.Count == 0 || state.list.Position.Count >= len(items) {
		t.Fatalf("vertical visible tab count = %d, want between 1 and %d", state.list.Position.Count, len(items)-1)
	}
	if _, ok := semanticBoundsForLabel(router.AppendSemantics(nil), "More"); !ok {
		t.Fatal("vertical overflow menu did not expose the More trigger")
	}
}

func TestVerticalTabsOverflowMenuMeasuresCustomTriggerWidth(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(320, 120))
	trigger := &tabsMeasurableProbe{size: image.Pt(180, 24)}
	widget := Tabs("vertical-custom-more", "Overview", tabsOverflowItems()).
		Vertical().
		Overflow(TabsOverflowMenu).
		OverflowTrigger(trigger)

	overflow := widget.measureOverflow(ctx, gtx, "Overview")
	padding := gtx.Dp(tabsSizeStyleFor(frame.ActiveTheme(ctx), TabsMedium).paddingX)
	if want := trigger.size.X + padding*2; overflow.moreSize.X < want {
		t.Fatalf("vertical More width = %d, want at least %d", overflow.moreSize.X, want)
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

func TestTabsSemanticsExposeTabListAndPanelRoles(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutTabsFrame(ctx, router, Tabs("settings", "security", tabsTestItems()), time.Unix(1, 0), image.Pt(400, 200))

	descriptions := semanticDescriptions(router.AppendSemantics(nil))
	if !containsString(descriptions, "Tab list") {
		t.Fatalf("semantic descriptions = %v, missing tab list", descriptions)
	}
	if !containsString(descriptions, "Selected tab") {
		t.Fatalf("semantic descriptions = %v, missing selected tab", descriptions)
	}
	if !containsString(descriptions, "Tab panel: Security") {
		t.Fatalf("semantic descriptions = %v, missing selected panel", descriptions)
	}
}

func TestTabsPanelSemanticsExposeLabelAndEnabledState(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutTabsFrame(ctx, router, Tabs("panel-semantics", "security", tabsTestItems()), time.Unix(1, 0), image.Pt(400, 200))

	node, ok := semanticNodeByDescription(router.AppendSemantics(nil), "Tab panel: Security")
	if !ok {
		t.Fatal("selected panel semantic node was not found")
	}
	if node.Desc.Label != "Security" || node.Desc.Disabled {
		t.Fatalf("panel semantics = %#v, want label Security and enabled", node.Desc)
	}
}

func TestTabsRegistersWorkbenchSemanticRelationships(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutTabsFrame(ctx, router, Tabs("semantic-roles", "security", tabsTestItems()), time.Unix(1, 0), image.Pt(400, 200))
	nodes := frame.Semantics(ctx)
	var list, selectedTab, panel *frame.SemanticNode
	for index := range nodes {
		node := &nodes[index]
		switch node.Role {
		case frame.SemanticTabList:
			list = node
		case frame.SemanticTab:
			if node.Selected {
				selectedTab = node
			}
		case frame.SemanticTabPanel:
			panel = node
		}
	}
	if list == nil || selectedTab == nil || panel == nil {
		t.Fatalf("semantic roles = %#v", nodes)
	}
	if selectedTab.Controls != panel.Key || selectedTab.Owner != list.Key || panel.Owner != list.Key {
		t.Fatalf("semantic relations = tab %#v panel %#v list %#v", *selectedTab, *panel, *list)
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
	if want := image.Rect(94, 0, 214, 32); horizontal != want {
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

func layoutTabsFrameWithOverlays(ctx *frame.Context, router *input.Router, tabs TabsWidget, now time.Time, viewport image.Point) {
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
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

type tabsPanelProbe struct {
	layouts *int
}

type tabsStatePanelProbe struct {
	layouts int
	key     string
	state   *widget.Clickable
}

func (p *tabsStatePanelProbe) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	p.layouts++
	p.key = frame.FullKey(ctx, "panel-control")
	p.state = frame.UseState[widget.Clickable](ctx, p.key, "probe")
	return layout.Dimensions{Size: image.Pt(80, 32)}
}

type tabsSlotProbe struct {
	layouts int
}

type tabsMeasurableProbe struct {
	layouts int
	size    image.Point
}

type tabsScopedContentProbe struct {
	measureKey string
	layoutKey  string
}

type tabsConstraintProbe struct {
	constraints layout.Constraints
}

func (p *tabsConstraintProbe) Measure(_ *frame.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsConstraintProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsMeasurableProbe) Measure(_ *frame.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: p.intrinsicSize()}
}

func (p *tabsMeasurableProbe) Layout(_ *frame.Context, _ layout.Context) layout.Dimensions {
	p.layouts++
	return layout.Dimensions{Size: p.intrinsicSize()}
}

func (p *tabsMeasurableProbe) intrinsicSize() image.Point {
	if p.size != (image.Point{}) {
		return p.size
	}
	return image.Pt(44, 24)
}

func (p *tabsScopedContentProbe) Measure(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	p.measureKey = frame.FullKey(ctx, "content")
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsScopedContentProbe) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	p.layoutKey = frame.FullKey(ctx, "content")
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsSlotProbe) Layout(_ *frame.Context, _ layout.Context) layout.Dimensions {
	p.layouts++
	return layout.Dimensions{Size: image.Pt(24, 24)}
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

func semanticDescriptions(nodes []input.SemanticNode) []string {
	var descriptions []string
	for _, node := range nodes {
		if node.Desc.Description != "" {
			descriptions = append(descriptions, node.Desc.Description)
		}
		descriptions = append(descriptions, semanticDescriptions(node.Children)...)
	}
	return descriptions
}

func semanticNodeByDescription(nodes []input.SemanticNode, description string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Description == description {
			return node, true
		}
		if child, ok := semanticNodeByDescription(node.Children, description); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func containsString(values []string, want string) bool {
	return slices.Contains(values, want)
}
