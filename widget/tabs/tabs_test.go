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
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	buttonui "github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
)

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
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

func TestTabsItemPartAppliesStateOpacity(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	widget := Widget{key: "tabs"}
	resolve := func(key string, state style.State) style.ResolvedStyle {
		return widget.resolveItemStyle(ctx, gtx, key, state, color.NRGBA{A: 0xff})
	}

	hovered := resolve("hovered", style.State{Hovered: true})
	disabled := resolve("disabled", style.State{Disabled: true})
	if hovered.Paint == nil || hovered.Paint.Opacity == nil || *hovered.Paint.Opacity != 0.7 {
		t.Fatalf("hovered tab opacity = %#v, want 0.7", hovered.Paint)
	}
	if disabled.Paint == nil || disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != core.ActiveTheme(ctx).DisabledOpacityValue() {
		t.Fatalf("disabled tab opacity = %#v", disabled.Paint)
	}
}

func TestTabsUsesDefaultCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	tabs := New("cursor", "account", tabsTestItems())
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
	custom := style.Style{}.
		Part(style.PartItem, style.Style{}.
			Background(style.SolidColor{Color: base}).
			On(style.Hovered, style.Style{}.Background(style.SolidColor{Color: hoveredColor})).
			On(style.Selected, style.Style{}.Background(style.SolidColor{Color: selectedColor})))

	widget := Widget{key: "tabs"}
	ctx := newContext(nil)
	gtx := testLayoutContext()
	visual := widget.visual(ctx)
	visual.Override = custom
	resolveBackground := func(itemKey string, state style.State) color.NRGBA {
		resolved := widget.resolveItemStyle(ctx, gtx, itemKey, state, color.NRGBA{A: 0xff}, visual)
		brush, ok := computed.Brush(resolved.Paint.Background)
		if !ok {
			t.Fatalf("item %q background = %#v", itemKey, resolved.Paint)
		}
		return brush.ColorAt(.5)
	}

	if got := resolveBackground("hovered", style.State{Hovered: true}); got != hoveredColor {
		t.Fatalf("hovered item background = %#v", got)
	}
	if got := resolveBackground("selected", style.State{Selected: true}); got != selectedColor {
		t.Fatalf("selected item background = %#v", got)
	}
	if got := resolveBackground("idle", style.State{}); got != base {
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

	full := New("full", "daily", items).Layout(newContext(nil), gtx)
	fit := New("fit", "daily", items).Size(Small).Fit().Layout(newContext(nil), gtx)
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

func TestHorizontalGapMatchesIndicatorGeometry(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := []TabItem{
		{Key: "daily", Label: "Daily"},
		{Key: "weekly", Label: "Weekly"},
		{Key: "monthly", Label: "Monthly"},
	}
	layoutTabsFrame(ctx, router, New("horizontal-gap", "daily", items).Fit(), time.Unix(1, 0), image.Pt(400, 100))
	state := testComponentState[tabsState](ctx, "horizontal-gap", stateSlotTabs)
	wantGap := testLayoutContext().Dp(core.ActiveTheme(ctx).Components.Tabs.TabGap)
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
	widget := New("overflow-indicator-gap", selectedKey, tabsOverflowItems()).
		Overflow(OverflowMenu).
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
	tabHeight := gtx.Dp(tabsSizeStyleFor(core.ActiveTheme(ctx), widget.size).height)
	wantGap := gtx.Dp(core.ActiveTheme(ctx).Components.Tabs.TabGap)
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
	layoutTabsFrame(ctx, router, New("centered", "a", items).Centered(true), time.Unix(1, 0), image.Pt(400, 100))

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
	layoutTabsFrame(ctx, router, New("centered-vertical", "a", items).Vertical().Centered(true), time.Unix(1, 0), image.Pt(300, 240))

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

	core.BeginFrameWithViewport(ctx, viewport)
	New("settings", "account", items).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	want := image.Rect(8, 72, 18, 82)
	if got != want {
		t.Fatalf("panel anchor = %v, want %v", got, want)
	}
}

func TestBottomPlacementMovesPanelBeforeTheStrip(t *testing.T) {
	ctx := newContext(nil)
	viewport := image.Pt(300, 200)
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: new(op.Ops)}
	var got image.Rectangle
	panel := &tabsOverlayProbe{got: &got}
	items := []TabItem{{Key: "account", Label: "Account", Panel: panel}}

	core.BeginFrameWithViewport(ctx, viewport)
	New("settings-bottom", "account", items).Placement(Bottom).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	want := image.Rect(8, 8, 18, 18)
	if got != want {
		t.Fatalf("bottom panel anchor = %v, want %v", got, want)
	}
}

func TestTabsClickChangesSelection(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	widget := func() Widget {
		return New("settings", selected, tabsTestItems()).OnChange(func(key string) {
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
	widget := New("uncontrolled-settings", "", items).DefaultSelectedKey("account")

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
	widget := New("settings", selected, items).OnChange(func(key string) {
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
	widget := func() Widget {
		return New("settings", selected, items).OnChange(func(key string) {
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
	widgetForFrame := func() Widget {
		return New("shortcut-tabs", selected, items).OnChange(func(key string) { selected = key })
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
	widgetForFrame := func() Widget {
		return New("manual-activation", selected, items).
			Activation(ActivationManual).
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
	widget := func() Widget {
		return New("settings", selected, tabsTestItems()).Vertical().OnChange(func(key string) {
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
	widget := func() Widget {
		return New("settings", selected, items).OnChange(func(key string) {
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

	New("settings", "security", items).Layout(newContext(nil), testLayoutContext())
	if accountLayouts != 0 || securityLayouts != 1 {
		t.Fatalf("panel layouts = account %d security %d, want 0 and 1", accountLayouts, securityLayouts)
	}
}

func TestPanelFadeHonorsDurationAndMotion(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "account"
	items := tabsTestItems()
	widgetForFrame := func() Widget {
		return New("panel-fade", selected, items).PanelTransition(PanelFade)
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

	theme := core.ActiveTheme(ctx)
	theme.Motion.Enabled = false
	selected = "account"
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Second), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), start.Add(time.Second+time.Millisecond), image.Pt(360, 180))
	if got := state.panelOpacity.Current(); got != 1 {
		t.Fatalf("disabled-motion panel opacity = %v, want 1", got)
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
	widgetForFrame := func() Widget {
		return New("force-render", selected, items).ForceRender(true)
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
	if _, ok := core.PeekState[widget.Clickable](ctx, security.key, "probe"); !ok {
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
	widget := New("force-render-destroy", "account", items).ForceRender(true).DestroyOnHidden(true)
	layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(360, 180))
	state := testComponentState[tabsState](ctx, "force-render-destroy", stateSlotTabs)
	if state == nil || len(state.renderedPanels) != 0 {
		t.Fatalf("rendered panels with DestroyOnHidden = %#v, want empty", state)
	}
}

func TestTabsHiddenPanelStateLifecycle(t *testing.T) {
	for _, test := range []struct {
		name            string
		destroyOnHidden bool
		wantRetained    bool
	}{
		{name: "keep alive", wantRetained: true},
		{name: "destroy on hidden", destroyOnHidden: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			selected := "account"
			account := &tabsStatePanelProbe{}
			security := &tabsStatePanelProbe{}
			items := []TabItem{
				{Key: "account", Label: "Account", Panel: account},
				{Key: "security", Label: "Security", Panel: security},
			}
			widgetForFrame := func() Widget {
				return New("settings", selected, items).KeepAlive(true).DestroyOnHidden(test.destroyOnHidden)
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
			retained, ok := core.PeekState[widget.Clickable](ctx, account.key, "probe")
			if ok != test.wantRetained || ok && retained != first {
				t.Fatalf("hidden panel state retained = %v, want %v", ok, test.wantRetained)
			}
		})
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
	widgetForFrame := func() Widget {
		return New("removed-panel", selected, items).KeepAlive(true)
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
	if _, ok := core.PeekState[widget.Clickable](ctx, security.key, "probe"); ok {
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
	widgetForFrame := func() Widget { return New("settings", selected, items) }
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, 0), image.Pt(360, 180))
	selected = "security"
	layoutTabsFrame(ctx, router, widgetForFrame(), time.Unix(1, int64(time.Millisecond)), image.Pt(360, 180))
	if account.key == security.key {
		t.Fatalf("panel state keys collided: %q", account.key)
	}
}

func TestTabsCustomContentUsesMeasuredWidth(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	content := &tabsMeasurableProbe{}
	widget := New("custom-content-width", "custom", []TabItem{{Key: "custom", Content: content}}).Fit()
	dims := widget.Layout(ctx, gtx)
	wantMin := 44 + 2*gtx.Dp(core.ActiveTheme(ctx).Components.Tabs.TabPaddingX)
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

	layoutTabsFrame(ctx, router, New("content-scope", "account", items).Fit(), time.Unix(1, 0), image.Pt(320, 100))

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
	widget := New("slot-scope", "account", []TabItem{{Key: "account", Label: "Account"}}).
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
	widget := New("fit-content", "custom", []TabItem{{Key: "custom", Content: content}}).Fit()
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
	sizeStyle := tabsSizeStyleFor(core.ActiveTheme(ctx), Medium)
	item := TabItem{Key: "file", Label: "file.go"}
	plain := New("plain-width", "file", []TabItem{item}).Fit()
	closable := New("closable-width", "file", []TabItem{{Key: item.Key, Label: item.Label, Closable: true}}).
		Fit().OnClose(func(string) {})

	plainWidth := plain.measureTabWidth(ctx, gtx, item, sizeStyle)
	closableItem := closable.items[0]
	closableWidth := closable.measureTabWidth(ctx, gtx, closableItem, sizeStyle)
	closeSize, closeGap := closable.closeButtonMetrics(closable.visual(ctx))
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
	widget := New("settings-close", selected, items).
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
	layoutTabsFrame(ctx, router, New("settings", "billing", tabsTestItems()), time.Unix(1, 0), image.Pt(500, 200))

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
	widget := New("resize-selection", "Revenue", items)

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

func TestOverflowScrollButtonsAdvanceList(t *testing.T) {
	for _, test := range []struct {
		name     string
		vertical bool
		viewport image.Point
		label    string
	}{
		{name: "horizontal", viewport: image.Pt(300, 160), label: "Scroll tabs right"},
		{name: "vertical", vertical: true, viewport: image.Pt(320, 120), label: "Scroll tabs down"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			items := tabsOverflowItems()
			widget := New("overflow", "Overview", items)
			if test.vertical {
				widget = widget.Vertical()
			}

			layoutTabsFrameWithOverlays(ctx, router, widget, time.Unix(1, 0), test.viewport)
			state := testComponentState[tabsState](ctx, "overflow", stateSlotTabs)
			if !state.canScrollNext(len(items)) {
				t.Fatal("overflowing tabs did not expose forward scrolling")
			}
			bounds, ok := semanticBoundsForLabel(router.AppendSemantics(nil), test.label)
			if !ok {
				t.Fatalf("missing %q scroll button semantics", test.label)
			}
			clickTabsAt(router, f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
			layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), test.viewport)
			if state.list.Position.First == 0 {
				t.Fatal("forward scroll button did not advance the tab list")
			}
			layoutTabsFrame(ctx, router, widget, time.Unix(1, int64(2*time.Millisecond)), test.viewport)
			if state.list.Position.First == 0 {
				t.Fatal("controlled selection pulled scrolling back to the first tab")
			}
		})
	}
}

func TestOverflowMenuReservesTrigger(t *testing.T) {
	for _, test := range []struct {
		name     string
		vertical bool
		viewport image.Point
	}{
		{name: "horizontal", viewport: image.Pt(300, 160)},
		{name: "vertical", vertical: true, viewport: image.Pt(320, 120)},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			items := tabsOverflowItems()
			widget := New("overflow-menu", "Overview", items).Overflow(OverflowMenu)
			if test.vertical {
				widget = widget.Vertical()
			}
			layoutTabsFrame(ctx, router, widget, time.Unix(1, 0), test.viewport)
			state := testComponentState[tabsState](ctx, "overflow-menu", stateSlotTabs)
			if state.list.Position.Count == 0 || state.list.Position.Count >= len(items) {
				t.Fatalf("visible tab count = %d, want between 1 and %d", state.list.Position.Count, len(items)-1)
			}
			nodes := router.AppendSemantics(nil)
			if _, ok := semanticBoundsForLabel(nodes, "More"); !ok {
				t.Fatal("overflow menu did not expose the More trigger")
			}
			if _, ok := semanticBoundsForLabel(nodes, "Revenue"); ok {
				t.Fatal("hidden tab was laid out in the main tab strip")
			}
		})
	}
}

func TestTabsOverflowTriggerUsesMeasuredCustomContent(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	trigger := &tabsMeasurableProbe{}
	widget := New("overflow-custom-trigger", "Overview", tabsOverflowItems()).
		Overflow(OverflowMenu).
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
	trigger := buttonui.New("More").Label("More tabs")
	widget := New("overflow-button-trigger", "Overview", tabsOverflowItems()).
		Overflow(OverflowMenu).
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
	trigger := buttonui.New("More")

	core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	New("first", "one", items).Fit().Overflow(OverflowMenu).OverflowTrigger(trigger).Layout(ctx, gtx)
	New("second", "one", items).Fit().Overflow(OverflowMenu).OverflowTrigger(trigger).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
}

func TestOverflowMenuSelectsHiddenItemAndRequestsVisibility(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "Overview"
	changed := ""
	items := tabsOverflowItems()
	widgetForFrame := func() Widget {
		return New("overflow-menu-select", selected, items).
			Overflow(OverflowMenu).
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
	widget := Widget{items: items}
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
	widget := New("overflow-selected-last", "Conversions", items).
		Overflow(OverflowMenu).
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
	widget := New("overflow-measure-exact", "Conversions", tabsOverflowItems()).
		Overflow(OverflowMenu).
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
	overflow := (Widget{items: tabsOverflowItems(), overflowMode: OverflowMenu}).measureOverflow(ctx, gtx, "Overview")
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
	widgetForFrame := func() Widget {
		return New("close-fallback", selected, items).
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
	widget := New("add-tab", "account", []TabItem{{Key: "account", Label: "Account"}}).
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
	widgetForFrame := func() Widget {
		return New("edit-tab", "account", []TabItem{{Key: "account", Label: "Account", Editable: true}}).
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
	widgetForFrame := func() Widget {
		return New("edit-blur", "account", items).
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

func TestVerticalOverflowMenuMeasuresCustomTriggerWidth(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Exact(image.Pt(320, 120))
	trigger := &tabsMeasurableProbe{size: image.Pt(180, 24)}
	widget := New("vertical-custom-more", "Overview", tabsOverflowItems()).
		Vertical().
		Overflow(OverflowMenu).
		OverflowTrigger(trigger)

	overflow := widget.measureOverflow(ctx, gtx, "Overview")
	padding := gtx.Dp(tabsSizeStyleFor(core.ActiveTheme(ctx), Medium).paddingX)
	if want := trigger.size.X + padding*2; overflow.moreSize.X < want {
		t.Fatalf("vertical More width = %d, want at least %d", overflow.moreSize.X, want)
	}
}

func TestTabsScrollButtonLabelsMatchOrientation(t *testing.T) {
	tests := []struct {
		name      string
		tabs      Widget
		direction int
		want      string
	}{
		{name: "horizontal previous", tabs: Widget{}, direction: -1, want: "Scroll tabs left"},
		{name: "horizontal next", tabs: Widget{}, direction: 1, want: "Scroll tabs right"},
		{name: "vertical previous", tabs: Widget{orientation: Vertical}, direction: -1, want: "Scroll tabs up"},
		{name: "vertical next", tabs: Widget{orientation: Vertical}, direction: 1, want: "Scroll tabs down"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			if got := test.tabs.scrollButtonLabel(ctx, test.direction); got != test.want {
				t.Fatalf("label = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTabsScrollButtonLabelsUseWindowLanguage(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageChinese)
	if got := (Widget{}).scrollButtonLabel(ctx, -1); got != "向左滚动标签页" {
		t.Fatalf("Chinese horizontal label = %q", got)
	}
	if got := (Widget{orientation: Vertical}).scrollButtonLabel(ctx, 1); got != "向下滚动标签页" {
		t.Fatalf("Chinese vertical label = %q", got)
	}
}

func TestTabsSemanticContracts(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := tabsTestItems()
	items[1].AccessibleLabel = "Security settings"
	layoutTabsFrame(ctx, router, New("semantic-roles", "security", items), time.Unix(1, 0), image.Pt(400, 200))

	gioNodes := router.AppendSemantics(nil)
	selected := selectedSemanticLabels(gioNodes)
	if len(selected) != 1 || selected[0] != "Security settings" {
		t.Fatalf("selected semantic labels = %v, want [Security settings]", selected)
	}
	descriptions := semanticDescriptions(gioNodes)
	for _, want := range []string{"Tab list", "Selected tab", "Tab panel: Security settings"} {
		if !containsString(descriptions, want) {
			t.Fatalf("semantic descriptions = %v, missing %q", descriptions, want)
		}
	}
	panelNode, ok := semanticNodeByDescription(gioNodes, "Tab panel: Security settings")
	if !ok || panelNode.Desc.Label != "Security settings" || panelNode.Desc.Disabled {
		t.Fatalf("panel semantics = %#v, want labeled and enabled", panelNode.Desc)
	}

	nodes := core.Semantics(ctx)
	var list, selectedTab, panel *core.SemanticNode
	for index := range nodes {
		node := &nodes[index]
		switch node.Role {
		case core.SemanticTabList:
			list = node
		case core.SemanticTab:
			if node.Selected {
				selectedTab = node
			}
		case core.SemanticTabPanel:
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
	if got := state.transition(gtx, "first", Horizontal, first); got != first {
		t.Fatalf("initial indicator = %v, want %v", got, first)
	}
	if got := state.transition(gtx, "second", Horizontal, second); got != first {
		t.Fatalf("indicator at transition start = %v, want %v", got, first)
	}

	gtx.Now = start.Add(tabsIndicatorDuration / 2)
	middle := state.transition(gtx, "second", Horizontal, second)
	if middle.Min.X <= first.Min.X || middle.Min.X >= second.Min.X {
		t.Fatalf("indicator midpoint X = %d, want between %d and %d", middle.Min.X, first.Min.X, second.Min.X)
	}
	if middle.Dx() <= first.Dx() || middle.Dx() >= second.Dx() {
		t.Fatalf("indicator midpoint width = %d, want between %d and %d", middle.Dx(), first.Dx(), second.Dx())
	}

	gtx.Now = start.Add(tabsIndicatorDuration)
	if got := state.transition(gtx, "second", Horizontal, second); got != second {
		t.Fatalf("indicator at transition end = %v, want %v", got, second)
	}
}

func TestTabsIndicatorTracksScrollingWithoutRestarting(t *testing.T) {
	state := new(tabsIndicatorState)
	gtx := testLayoutContext()
	gtx.Now = time.Unix(1, 0)
	initial := image.Rect(80, 0, 180, 32)
	shifted := initial.Add(image.Pt(-24, 0))

	state.transition(gtx, "selected", Horizontal, initial)
	if got := state.transition(gtx, "selected", Horizontal, shifted); got != shifted {
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

	state.transition(gtx, "first", Horizontal, first, motion)
	if got := state.transition(gtx, "second", Horizontal, second, motion); got != second {
		t.Fatalf("disabled motion indicator = %v, want %v", got, second)
	}
}

func TestTabsItemRectAccountsForListOffsetAndOrientation(t *testing.T) {
	position := layout.Position{First: 1, Offset: 10}
	widths := []int{80, 100, 120}

	horizontal := Widget{}.tabRect(position, widths, 2, 32, 4)
	if want := image.Rect(94, 0, 214, 32); horizontal != want {
		t.Fatalf("horizontal rect = %v, want %v", horizontal, want)
	}
	vertical := (Widget{orientation: Vertical}).tabRect(position, widths, 2, 32, 4)
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

func layoutTabsFrame(ctx *core.Context, router *input.Router, tabs Widget, now time.Time, viewport image.Point) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	tabs.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutTabsFrameWithOverlays(ctx *core.Context, router *input.Router, tabs Widget, now time.Time, viewport image.Point) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	tabs.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)
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

func (p *tabsStatePanelProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	p.layouts++
	p.key = core.FullKey(ctx, "panel-control")
	p.state = core.UseState[widget.Clickable](ctx, p.key, "probe")
	return layout.Dimensions{Size: image.Pt(80, 32)}
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

func (p *tabsConstraintProbe) Measure(_ *core.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsConstraintProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsMeasurableProbe) Measure(_ *core.Context, _ layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: p.intrinsicSize()}
}

func (p *tabsMeasurableProbe) Layout(_ *core.Context, _ layout.Context) layout.Dimensions {
	p.layouts++
	return layout.Dimensions{Size: p.intrinsicSize()}
}

func (p *tabsMeasurableProbe) intrinsicSize() image.Point {
	if p.size != (image.Point{}) {
		return p.size
	}
	return image.Pt(44, 24)
}

func (p *tabsScopedContentProbe) Measure(ctx *core.Context, _ layout.Context) layout.Dimensions {
	p.measureKey = core.FullKey(ctx, "content")
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

func (p *tabsScopedContentProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	p.layoutKey = core.FullKey(ctx, "content")
	return layout.Dimensions{Size: image.Pt(44, 24)}
}

type tabsOverlayProbe struct {
	got *image.Rectangle
}

func (p *tabsOverlayProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
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

func (p *tabsPanelProbe) Layout(_ *core.Context, _ layout.Context) layout.Dimensions {
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

func TestUncontrolledPreservesEmptySelectionWithoutControl(t *testing.T) {
	widget := New("tabs", "", nil).Uncontrolled()
	if widget.hasSelectedKey || !widget.hasDefaultSelected || widget.defaultSelectedKey != "" {
		t.Fatalf("Uncontrolled() = hasSelectedKey=%v hasDefault=%v default=%q", widget.hasSelectedKey, widget.hasDefaultSelected, widget.defaultSelectedKey)
	}
}

func TestSelectedKeyAfterDefaultSelectedKeyWinsControl(t *testing.T) {
	widget := New("tabs", "", nil).DefaultSelectedKey("seed").SelectedKey("live")
	if !widget.hasSelectedKey || widget.selectedKey != "live" {
		t.Fatalf("SelectedKey after DefaultSelectedKey = hasSelectedKey=%v selectedKey=%q", widget.hasSelectedKey, widget.selectedKey)
	}
}
