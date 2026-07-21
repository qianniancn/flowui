package selects

import (
	"image"
	"slices"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/listbox"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func testSetComponentState[T any](ctx *frame.Context, key, slot string, value *T) {
	frame.UseStateWith(ctx, key, slot, func() *T { return value })
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

func TestSelectOptions(t *testing.T) {
	changed := false
	selectionChanged := false
	openChanged := false
	selectWidget := Select("state", "ca", selectTestItems()).
		Placeholder("Choose a state").
		EmptyText("Nothing available").
		Label("State").
		Description("Shipping destination").
		ErrorMessage("State is required").
		ValueText("California").
		Indicator(text.New("v")).
		OnChange(func(string) { changed = true }).
		OnSelectionChange(func([]string) { selectionChanged = true }).
		OnOpenChange(func(bool) { openChanged = true }).
		Open(true).
		DefaultOpen(true).
		Placement(overlay.PopoverTopStart).
		ShouldFlip(false).
		AvoidOverflow(false).
		DisabledKeys([]string{"tx"}).
		Variant(SelectSecondary).
		Disabled(true).
		Invalid(true).
		Required(true).
		FullWidth()

	if selectWidget.key != "state" || selectWidget.selectedKey != "ca" || len(selectWidget.items) != 3 {
		t.Fatal("select constructor did not preserve its key, value, and items")
	}
	if selectWidget.placeholder != "Choose a state" || selectWidget.emptyText != "Nothing available" {
		t.Fatal("select placeholder options were not set")
	}
	if selectWidget.label != "State" || selectWidget.description != "Shipping destination" || selectWidget.errorMessage != "State is required" {
		t.Fatal("select field text options were not set")
	}
	if selectWidget.valueText != "California" || selectWidget.variant != SelectSecondary {
		t.Fatal("select visual options were not set")
	}
	if selectWidget.indicator == nil {
		t.Fatal("custom select indicator was not set")
	}
	if !selectWidget.hasOpen || !selectWidget.open || !selectWidget.hasDefaultOpen || !selectWidget.defaultOpen {
		t.Fatal("select open options were not set")
	}
	if selectWidget.placement != overlay.PopoverTopStart || selectWidget.flipEnabled() || selectWidget.overflowAvoidanceEnabled() {
		t.Fatal("select popover options were not set")
	}
	if !selectWidget.disabled || !selectWidget.invalid || !selectWidget.required || !selectWidget.fullWidth {
		t.Fatal("select state options were not set")
	}
	selectWidget.onChange("tx")
	selectWidget.onSelectionChange([]string{"ca"})
	selectWidget.onOpenChange(false)
	if !changed || !selectionChanged || !openChanged {
		t.Fatal("select callbacks were not retained")
	}
}

func TestSelectConstructorsAndDisplayValues(t *testing.T) {
	single := Select("language", "go", selectTestItems())
	if value, selected := single.displayValue(); value != "Go" || !selected {
		t.Fatalf("single display value = %q %v, want Go true", value, selected)
	}

	multiple := SelectMultiple("languages", []string{"rust", "go", "rust"}, selectTestItems())
	if value, selected := multiple.displayValue(); value != "Rust, Go" || !selected {
		t.Fatalf("multiple display value = %q %v, want Rust, Go true", value, selected)
	}

	empty := Select("empty", "missing", selectTestItems()).Placeholder("Choose one")
	if value, selected := empty.displayValue(); value != "Choose one" || selected {
		t.Fatalf("placeholder = %q %v, want Choose one false", value, selected)
	}

	custom := multiple.ValueText("2 languages")
	if value, selected := custom.displayValue(); value != "2 languages" || !selected {
		t.Fatalf("custom value = %q %v, want 2 languages true", value, selected)
	}
}

func TestSelectDataVersionCachesItems(t *testing.T) {
	widget := SelectSections("cached", "one", []SelectSection{{Title: "Group", Items: []SelectItem{{Key: "one", Label: "One"}}}}).DataVersion(1)
	state := new(selectState)
	items := state.itemsFor(widget)
	cachedItems := state.itemsFor(widget)
	if &items[0] != &cachedItems[0] {
		t.Fatal("unchanged Select data version did not reuse items")
	}
	updatedItems := state.itemsFor(widget.DataVersion(2))
	if &items[0] == &updatedItems[0] {
		t.Fatal("changed Select data version reused stale items")
	}
}

func TestSelectSectionConstructors(t *testing.T) {
	sections := []SelectSection{{Title: "Systems", Items: selectTestItems()[:2]}}
	single := SelectSections("language", "go", sections)
	multiple := SelectMultipleSections("languages", []string{"rust"}, sections)

	if len(single.sections) != 1 || len(single.allItems()) != 2 || single.selectionMode != SelectSelectionSingle {
		t.Fatal("single section constructor did not configure sections")
	}
	if len(multiple.sections) != 1 || len(multiple.allItems()) != 2 || multiple.selectionMode != SelectSelectionMultiple {
		t.Fatal("multiple section constructor did not configure sections")
	}
}

func TestSelectDefaultAndControlledOpen(t *testing.T) {
	state := new(selectState)
	if !state.isOpen(Select("language", "", nil).DefaultOpen(true)) {
		t.Fatal("default-open select did not initialize open")
	}

	called := 0
	controlled := Select("language", "", nil).Open(false).OnOpenChange(func(open bool) {
		if !open {
			t.Fatal("controlled open callback received false")
		}
		called++
	})
	if got := state.requestOpen(newContext(nil), controlled, true); got {
		t.Fatal("controlled select changed its effective state before model update")
	}
	if called != 1 {
		t.Fatalf("controlled callback calls = %d, want 1", called)
	}
	state.requestOpen(newContext(nil), controlled, false)
	if called != 1 {
		t.Fatal("controlled select emitted a callback for its current state")
	}
}

func TestSelectTriggerClickOpensUncontrolled(t *testing.T) {
	state := new(selectState)
	state.trigger.Click()
	ctx := newContext(nil)
	widget := Select("language", "go", selectTestItems())

	open := state.handleTrigger(ctx, testLayoutContext(), widget, false)

	if !open || !state.open {
		t.Fatal("trigger click did not open uncontrolled select")
	}
	if state.focusIntent != selectFocusSelected {
		t.Fatalf("focus intent = %v, want selected", state.focusIntent)
	}
}

func TestSelectOpeningAnotherClosesActiveSelect(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	start := time.Unix(1, 0)
	closed := 0
	first := Select("first", "go", selectTestItems()).
		DefaultOpen(true).
		OnOpenChange(func(open bool) {
			if !open {
				closed++
			}
		})
	second := Select("second", "rust", selectTestItems())

	layoutSelectPairTestFrame(ctx, router, first, second, start, false)
	firstState := testComponentState[selectState](ctx, "first", stateSlotSelect)
	secondState := testComponentState[selectState](ctx, "second", stateSlotSelect)
	if firstState == nil || secondState == nil || frame.ActiveExclusive(ctx, "select") != "first" {
		t.Fatal("default-open select did not become active")
	}

	secondState.trigger.Click()
	layoutSelectPairTestFrame(ctx, router, first, second, start.Add(time.Millisecond), false)

	if firstState.open || !secondState.open {
		t.Fatalf("select open states = first %v second %v, want false true", firstState.open, secondState.open)
	}
	if closed != 1 {
		t.Fatalf("first close callbacks = %d, want 1", closed)
	}
	if frame.ActiveExclusive(ctx, "select") != "second" {
		t.Fatalf("active select = %q, want second", frame.ActiveExclusive(ctx, "select"))
	}
}

func TestSelectOpeningAnotherRequestsControlledSelectClose(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	start := time.Unix(1, 0)
	controlledOpen := true
	closeRequests := 0
	controlled := func() SelectWidget {
		return Select("controlled", "go", selectTestItems()).
			Open(controlledOpen).
			OnOpenChange(func(open bool) {
				if !open {
					controlledOpen = false
					closeRequests++
				}
			})
	}
	second := Select("second", "rust", selectTestItems())

	layoutSelectPairTestFrame(ctx, router, controlled(), second, start, false)
	testComponentState[selectState](ctx, "second", stateSlotSelect).trigger.Click()
	layoutSelectPairTestFrame(ctx, router, controlled(), second, start.Add(time.Millisecond), true)
	layoutSelectPairTestFrame(ctx, router, controlled(), second, start.Add(2*time.Millisecond), false)

	if controlledOpen {
		t.Fatal("controlled select did not receive a close request")
	}
	if closeRequests != 1 {
		t.Fatalf("controlled close requests = %d, want 1", closeRequests)
	}
	if frame.ActiveExclusive(ctx, "select") != "second" {
		t.Fatalf("active select = %q, want second", frame.ActiveExclusive(ctx, "select"))
	}
}

func TestSelectPeerAndDismissRequestControlledCloseOnce(t *testing.T) {
	ctx := newContext(nil)
	closeRequests := 0
	controlled := Select("first", "go", selectTestItems()).
		Open(true).
		OnOpenChange(func(open bool) {
			if !open {
				closeRequests++
			}
		})
	first := &selectState{key: "first"}
	first.bind(controlled)
	second := &selectState{key: "second"}
	testSetComponentState(ctx, "first", stateSlotSelect, first)
	testSetComponentState(ctx, "second", stateSlotSelect, second)
	frame.BeginFrame(ctx)
	frame.RegisterExclusive(ctx, "select", "first", first.closeForPeer)
	frame.RegisterExclusive(ctx, "select", "second", second.closeForPeer)
	frame.ActivateExclusive(ctx, "select", "first")

	activateSelect(ctx, second)
	first.dismiss[0].Click()
	if open := first.handleOverlayEvents(ctx, testLayoutContext(), controlled, true); !open {
		t.Fatal("controlled select changed its effective state before its model updated")
	}
	if closeRequests != 1 {
		t.Fatalf("overlapping peer and dismiss close requests = %d, want 1", closeRequests)
	}

	first.dismiss[0].Click()
	first.handleOverlayEvents(ctx, testLayoutContext(), controlled, true)
	if closeRequests != 2 {
		t.Fatalf("later controlled close request was suppressed, calls = %d, want 2", closeRequests)
	}
}

func TestSelectPeerCloseDeduplicationExpiresNextFrame(t *testing.T) {
	ctx := newContext(nil)
	closeRequests := 0
	controlled := Select("first", "go", selectTestItems()).
		Open(true).
		OnOpenChange(func(open bool) {
			if !open {
				closeRequests++
			}
		})
	first := &selectState{key: "first"}
	first.bind(controlled)
	second := &selectState{key: "second"}
	testSetComponentState(ctx, "first", stateSlotSelect, first)
	testSetComponentState(ctx, "second", stateSlotSelect, second)
	frame.BeginFrame(ctx)
	frame.RegisterExclusive(ctx, "select", "first", first.closeForPeer)
	frame.RegisterExclusive(ctx, "select", "second", second.closeForPeer)
	frame.ActivateExclusive(ctx, "select", "first")

	activateSelect(ctx, second)
	frame.BeginFrame(ctx)
	first.dismiss[0].Click()
	first.handleOverlayEvents(ctx, testLayoutContext(), controlled, true)

	if closeRequests != 2 {
		t.Fatalf("next-frame close request remained deduplicated, calls = %d, want 2", closeRequests)
	}
}

func TestSelectPeerCloseDoesNotRestoreOldTriggerFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	start := time.Unix(1, 0)
	first := Select("first", "go", selectTestItems()).DefaultOpen(true)
	second := Select("second", "rust", selectTestItems())

	layoutSelectPairTestFrame(ctx, router, first, second, start, false)
	firstState := testComponentState[selectState](ctx, "first", stateSlotSelect)
	secondState := testComponentState[selectState](ctx, "second", stateSlotSelect)
	secondState.trigger.Click()
	layoutSelectPairTestFrame(ctx, router, first, second, start.Add(time.Millisecond), true)

	if !router.Source().Focused(&secondState.trigger) {
		t.Fatal("opening select trigger did not keep focus after closing its peer")
	}
	if router.Source().Focused(&firstState.trigger) {
		t.Fatal("peer close restored focus to the old select trigger")
	}
}

func TestSelectActiveOwnershipClearsWhenStateIsSwept(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := Select("language", "go", selectTestItems()).DefaultOpen(true)

	layoutSelectTestFrame(ctx, router, selectWidget, time.Unix(1, 0))
	if frame.ActiveExclusive(ctx, "select") != "language" {
		t.Fatalf("active select = %q, want language", frame.ActiveExclusive(ctx, "select"))
	}

	frame.BeginFrame(ctx)
	frame.EndFrame(ctx)

	if frame.ActiveExclusive(ctx, "select") != "" {
		t.Fatalf("active select after sweep = %q, want empty", frame.ActiveExclusive(ctx, "select"))
	}
}

func TestSelectActiveOwnershipClearsWhenSelectCloses(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := Select("language", "go", selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)

	layoutSelectTestFrame(ctx, router, selectWidget, start)
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)
	state.trigger.Click()
	layoutSelectTestFrame(ctx, router, selectWidget, start.Add(time.Millisecond))

	if state.open {
		t.Fatal("select trigger did not close the active select")
	}
	if frame.ActiveExclusive(ctx, "select") != "" {
		t.Fatalf("active select after close = %q, want empty", frame.ActiveExclusive(ctx, "select"))
	}
}

func TestSelectArrowDownOpensAndFocusesFirstOption(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	widget := Select("language", "", selectTestItems()).DisabledKeys([]string{"go"})
	start := time.Unix(1, 0)
	layoutSelectTestFrame(ctx, router, widget, start)
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)
	if state == nil {
		t.Fatal("select state was not retained")
	}
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})

	layoutSelectTestFrame(ctx, router, widget, start)
	layoutSelectTestFrame(ctx, router, widget, start.Add(selectEnterDuration))

	if !state.open {
		t.Fatal("ArrowDown did not open select")
	}
	item, focus, ok := listbox.DerivedItem(ctx, "language", "options", "rust")
	if !ok {
		t.Fatal("select options were not laid out")
	}
	if !router.Source().Focused(item) {
		t.Fatal("ArrowDown did not focus first enabled option")
	}
	if !frame.FocusVisible(ctx, item, true) {
		t.Fatal("keyboard-opened select marked the option focus as pointer-originated")
	}
	layoutSelectTestFrame(ctx, router, widget, start.Add(selectEnterDuration+time.Millisecond))
	_, focus, _ = listbox.DerivedItem(ctx, "language", "options", "rust")
	if focus.TargetOpacity != 1 {
		t.Fatal("keyboard-opened select did not show the focused option ring")
	}
}

func TestSelectPointerOpenFocusesSelectedWithoutFocusRing(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	widget := Select("language", "go", selectTestItems())
	start := time.Unix(1, 0)
	layoutSelectTestFrame(ctx, router, widget, start)
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)
	state.trigger.Click()

	layoutSelectTestFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutSelectTestFrame(ctx, router, widget, start.Add(time.Millisecond+selectEnterDuration))
	item, focus, ok := listbox.DerivedItem(ctx, "language", "options", "go")
	if !ok {
		t.Fatal("pointer-opened select did not lay out its selected option")
	}
	if !router.Source().Focused(item) {
		t.Fatal("pointer-opened select did not focus its selected option")
	}
	if frame.FocusVisible(ctx, item, true) {
		t.Fatal("pointer-opened select did not preserve pointer focus modality")
	}

	layoutSelectTestFrame(ctx, router, widget, start.Add(2*time.Millisecond+selectEnterDuration))
	_, focus, _ = listbox.DerivedItem(ctx, "language", "options", "go")
	if focus.TargetOpacity != 0 {
		t.Fatal("pointer-opened select showed a focused option ring")
	}
}

func TestSelectPointerClickKeepsEmptyTriggerFocusHidden(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	widget := Select("language", "", nil).FullWidth()
	start := time.Unix(1, 0)
	layoutSelectTestFrame(ctx, router, widget, start)
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)

	position := f32.Pt(20, 20)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutSelectTestFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutSelectTestFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutSelectTestFrame(ctx, router, widget, start.Add(3*time.Millisecond))

	if !state.open || !router.Source().Focused(&state.trigger) {
		t.Fatal("pointer click did not open and focus the empty select trigger")
	}
	if frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer click exposed the empty select keyboard focus ring")
	}
}

func TestSelectEscapeClosesFromOptionAndRestoresTriggerFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	widget := Select("language", "go", selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	layoutSelectTestFrame(ctx, router, widget, start)
	layoutSelectTestFrame(ctx, router, widget, start.Add(selectEnterDuration))
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)
	item, _, ok := listbox.DerivedItem(ctx, "language", "options", "go")
	if state == nil || !ok || !router.Source().Focused(item) {
		t.Fatal("open select did not focus selected option")
	}
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})

	layoutSelectTestFrame(ctx, router, widget, start.Add(selectEnterDuration+time.Millisecond))

	if state.open {
		t.Fatal("Escape did not close select")
	}
	if !router.Source().Focused(&state.trigger) {
		t.Fatal("Escape did not restore trigger focus")
	}
}

func TestSelectProgrammaticCloseRestoresTriggerFocus(t *testing.T) {
	ctx := newContext(nil)
	state := &selectState{wasOpen: true}
	gtx, router := selectFocusTestContext()
	state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(20, 20)}
	})

	state.observeOpen(ctx, false, true)
	frame.ApplyFrameCommands(ctx, gtx)
	router.Frame(gtx.Ops)

	if !router.Source().Focused(&state.trigger) {
		t.Fatal("programmatic close did not restore trigger focus")
	}
}

func TestNaturallyDisabledSelectCloseDoesNotRestoreTriggerFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	background := new(widget.Clickable)
	start := time.Unix(1, 0)

	layoutSelectWithBackgroundFrame(ctx, router, Select("language", "go", selectTestItems()).Open(true), background, false, start)
	state := testComponentState[selectState](ctx, "language", stateSlotSelect)
	router.Source().Execute(key.FocusCmd{Tag: background})
	if !router.Source().Focused(background) {
		t.Fatal("background did not receive setup focus")
	}

	layoutSelectWithBackgroundFrame(ctx, router, Select("language", "go", selectTestItems()).Open(false), background, true, start.Add(time.Millisecond))
	if !router.Source().Focused(background) {
		t.Fatal("naturally disabled select close displaced existing focus")
	}
	if router.Source().Focused(&state.trigger) {
		t.Fatal("naturally disabled select restored trigger focus")
	}
}

func TestSelectDismissClosesWithoutForcingTriggerFocus(t *testing.T) {
	state := &selectState{open: true}
	state.dismiss[0].Click()
	ctx := newContext(nil)
	widget := Select("language", "go", selectTestItems())

	gtx, router := selectFocusTestContext()
	open := state.handleOverlayEvents(ctx, gtx, widget, true)
	state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(20, 20)}
	})
	frame.ApplyFrameCommands(ctx, gtx)
	router.Frame(gtx.Ops)
	if _, wake := router.WakeupTime(); !wake {
		t.Fatal("dismiss did not request a redraw")
	}

	if open || state.open {
		t.Fatal("dismiss click did not close select")
	}
	if router.Source().Focused(&state.trigger) {
		t.Fatal("dismiss forced focus back to the select trigger")
	}
}

func TestSelectPointerClickOutsideCloses(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := SelectMultiple("language", []string{"go"}, selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	var background widget.Clickable
	backgroundClicked := false

	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration))
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(290, 500),
	})
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+time.Millisecond))
	if testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("pointer press outside the select popover did not close it immediately")
	}
	closeStart := start.Add(selectEnterDuration + 2*time.Millisecond)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, closeStart)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, closeStart.Add(selectExitDuration))
	if progress := testComponentState[selectState](ctx, "language", stateSlotSelect).transition.Current(); progress != 0 {
		t.Fatalf("held pointer left exit progress at %v, want 0", progress)
	}
	router.Queue(
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  f32.Pt(290, 500),
		},
	)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, closeStart.Add(selectExitDuration+time.Millisecond))

	if testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("pointer click outside the select popover did not close it")
	}
	if !backgroundClicked {
		t.Fatal("select dismiss backdrop swallowed the background click")
	}
	if !router.Source().Focused(&background) {
		t.Fatal("select dismiss restored trigger focus instead of focusing the clicked background control")
	}
}

func TestSelectPointerClickInsideDoesNotDismiss(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := SelectMultiple("languages", nil, selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	var background widget.Clickable
	backgroundClicked := false

	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration))
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  f32.Pt(50, 60),
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  f32.Pt(50, 60),
		},
	)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+time.Millisecond))

	if !testComponentState[selectState](ctx, "languages", stateSlotSelect).open {
		t.Fatal("pointer click inside the select popover dismissed it")
	}
	if backgroundClicked {
		t.Fatal("select popover allowed an inside click to reach the background")
	}
}

func TestSelectPanelPaddingPressPreservesOptionFocus(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := Select("language", "go", selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	var background widget.Clickable
	backgroundClicked := false

	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration))
	item, _, ok := listbox.DerivedItem(ctx, "language", "options", "go")
	if !ok || !router.Source().Focused(item) {
		t.Fatal("open select did not focus its selected option")
	}

	position := f32.Pt(22, 44)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+time.Millisecond))
	if !router.Source().Focused(item) {
		t.Fatal("panel padding press cleared option focus")
	}
	if !testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("panel padding press closed the select")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+2*time.Millisecond))
	if backgroundClicked {
		t.Fatal("panel padding click reached the background")
	}
}

func TestSelectAnimatedPanelEdgeHasNoPointerHole(t *testing.T) {
	customTheme := DefaultTheme()
	customTheme.Components.Select.AnimationScale = .5
	ctx := frame.New(nil, &customTheme, locale.LanguageAuto)
	router := new(input.Router)
	selectWidget := Select("language", "go", selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	var background widget.Clickable
	backgroundClicked := false

	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start)
	midpoint := start.Add(selectEnterDuration / 2)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, midpoint)
	position := f32.Pt(20.5, 60)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, midpoint.Add(time.Millisecond))
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, midpoint.Add(2*time.Millisecond))

	if testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("click beside the animated panel did not dismiss the select")
	}
	if !backgroundClicked {
		t.Fatal("animated panel edge left a pointer hole instead of a pass-through dismiss area")
	}
}

func TestSelectPointerTriggerCanCloseAndReopen(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selectWidget := Select("language", "go", selectTestItems()).DefaultOpen(true)
	start := time.Unix(1, 0)
	var background widget.Clickable
	backgroundClicked := false

	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start)
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration))
	queueSelectPointerClick(router, f32.Pt(50, 20))
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+time.Millisecond))
	if testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("pointer click on the open select trigger did not close it")
	}

	queueSelectPointerClick(router, f32.Pt(50, 20))
	layoutNestedSelectTestFrame(ctx, router, selectWidget, &background, &backgroundClicked, start.Add(selectEnterDuration+2*time.Millisecond))
	if !testComponentState[selectState](ctx, "language", stateSlotSelect).open {
		t.Fatal("dismiss event closed the select after its trigger reopened it")
	}
}

func TestSelectSingleSelectionClosesAndMultipleStaysOpen(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		ctx := newContext(nil)
		listbox.EnsureDerivedItem(ctx, "language", "options", "rust").Click()
		state := &selectState{open: true}
		changed := ""
		widget := Select("language", "go", selectTestItems()).OnChange(func(key string) { changed = key })

		frame.BeginFrame(ctx)
		widget.listBox(ctx, state, true).Layout(ctx, testLayoutContext())

		if changed != "rust" {
			t.Fatalf("single changed = %q, want rust", changed)
		}
		if state.open {
			t.Fatal("single selection kept select open")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		ctx := newContext(nil)
		listbox.EnsureDerivedItem(ctx, "languages", "options", "rust").Click()
		state := &selectState{open: true}
		var changed []string
		widget := SelectMultiple("languages", []string{"go"}, selectTestItems()).OnSelectionChange(func(keys []string) { changed = keys })

		frame.BeginFrame(ctx)
		widget.listBox(ctx, state, true).Layout(ctx, testLayoutContext())

		if !slices.Equal(changed, []string{"go", "rust"}) {
			t.Fatalf("multiple changed = %#v, want [go rust]", changed)
		}
		if !state.open {
			t.Fatal("multiple selection closed select")
		}
	})
}

func TestSelectFocusIntentSkipsDisabledOptions(t *testing.T) {
	items := selectTestItems()
	widget := Select("language", "rust", items).DisabledKeys([]string{"rust"})

	if index, ok := widget.focusOptionIndex(items, selectFocusSelected); !ok || index != 0 {
		t.Fatalf("selected focus fallback = %d %v, want first enabled", index, ok)
	}
	if index, ok := widget.focusOptionIndex(items, selectFocusLast); !ok || index != 0 {
		t.Fatalf("last enabled = %d %v, want index 0", index, ok)
	}
}

func TestSelectFocusPendingOptionRequestsListItemFocus(t *testing.T) {
	ctx := newContext(nil)
	listbox.EnsureDerivedItem(ctx, "language", "options", "go")
	rustItem := listbox.EnsureDerivedItem(ctx, "language", "options", "rust")
	state := &selectState{focusIntent: selectFocusSelected}
	widget := Select("language", "rust", selectTestItems())
	gtx, router := selectFocusTestContext()
	rustItem.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(20, 20)}
	})

	widget.focusPendingOption(ctx, state)
	frame.ApplyFrameCommands(ctx, gtx)
	router.Frame(gtx.Ops)

	if !router.Source().Focused(rustItem) {
		t.Fatal("selected item did not receive pending focus request")
	}
	if state.focusIntent != selectFocusNone {
		t.Fatal("focus intent was not consumed")
	}
}

func TestSelectLayoutAndFullWidth(t *testing.T) {
	ctx := newContext(nil)
	dims := Select("language", "go", selectTestItems()).Layout(ctx, testLayoutContext())
	if dims.Size.Y != 36 {
		t.Fatalf("select height = %d, want 36", dims.Size.Y)
	}

	ctx = newContext(nil)
	dims = Select("language", "go", selectTestItems()).FullWidth().Layout(ctx, testLayoutContext())
	if dims.Size.X != 300 {
		t.Fatalf("full-width select width = %d, want 300", dims.Size.X)
	}
}

func TestSelectMultipleValueWrapsAndGrowsTrigger(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Min: image.Pt(120, 0), Max: image.Pt(120, 200)}
	items := []SelectItem{
		{Key: "argentina", Label: "Argentina"},
		{Key: "new-zealand", Label: "New Zealand"},
		{Key: "thailand", Label: "Thailand"},
	}

	dims := SelectMultiple("countries", []string{"argentina", "new-zealand", "thailand"}, items).FullWidth().Layout(ctx, gtx)

	if dims.Size.Y <= gtx.Dp(frame.ActiveTheme(ctx).Components.Select.Height) {
		t.Fatalf("wrapped multiple-select height = %d, want greater than base height", dims.Size.Y)
	}
}

func TestSelectSupportMessageMatchesInvalidState(t *testing.T) {
	widget := Select("language", "", nil).Description("Pick one").Invalid(true)
	if message, _ := widget.supportMessage(); message != "" {
		t.Fatalf("invalid description = %q, want hidden", message)
	}
	if message, isError := widget.ErrorMessage("Required").supportMessage(); message != "Required" || !isError {
		t.Fatalf("invalid error = %q, %v", message, isError)
	}
}

func TestSelectStylesMatchPrimaryAndSecondaryFields(t *testing.T) {
	theme := DefaultTheme()
	primary := selectStyleFor(&theme, SelectPrimary, false, false, false, false)
	secondary := selectStyleFor(&theme, SelectSecondary, false, false, false, false)
	invalid := selectStyleFor(&theme, SelectPrimary, false, false, false, true)

	if primary.field.ShadowOpacity == 0 {
		t.Fatal("primary select should keep field shadow")
	}
	if secondary.field.ShadowOpacity != 0 || secondary.field.Background != theme.Palette.DefaultColor() {
		t.Fatal("secondary select does not match lower-emphasis field style")
	}
	if invalid.field.Border != theme.Palette.Danger {
		t.Fatal("invalid select does not use danger border")
	}
}

func TestSelectEnterAndExitDurations(t *testing.T) {
	state := new(selectState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start
	if got := state.progress(gtx, true); got != 0 {
		t.Fatalf("enter start = %v, want 0", got)
	}
	gtx.Now = start.Add(selectEnterDuration)
	if got := state.progress(gtx, true); got != 1 {
		t.Fatalf("enter end = %v, want 1", got)
	}
	if got := state.progress(gtx, false); got != 1 {
		t.Fatalf("exit start = %v, want 1", got)
	}
	gtx.Now = gtx.Now.Add(selectExitDuration)
	if got := state.progress(gtx, false); got != 0 {
		t.Fatalf("exit end = %v, want 0", got)
	}
}

func TestSelectPanelUsesTriggerWidth(t *testing.T) {
	ctx := newContext(nil)
	state := new(selectState)
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(300, 200)}
	widget := Select("language", "go", selectTestItems())

	dims := widget.layoutOverlay(ctx, gtx, state, image.Rect(0, 0, 180, 36), true, 1, true)

	if dims.Size != image.Pt(300, 200) {
		t.Fatalf("overlay dimensions = %v, want bounds", dims.Size)
	}
	if !listbox.HasDerivedState(ctx, "language", "options") {
		t.Fatal("select panel did not reuse ListBox")
	}
}

func TestSelectOptionsIdentityDoesNotCollideWithUserKey(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	widget := Select("language", "go", selectTestItems())

	frame.BeginFrame(ctx)
	widget.listBox(ctx, &selectState{key: "language"}, true).Layout(ctx, gtx)
	listbox.ListBox("language:options", "", nil).Layout(ctx, gtx)

	if !listbox.HasDerivedState(ctx, "language", "options") {
		t.Fatal("missing Select's derived ListBox state")
	}
	if !listbox.HasState(ctx, "language:options") {
		t.Fatal("missing independently keyed user ListBox state")
	}
}

func TestSelectPopoverAnchorsToTriggerInsteadOfDescription(t *testing.T) {
	ctx := newContext(nil)
	state := new(selectState)
	widget := Select("language", "go", selectTestItems()).
		Label("Language").
		Description("A deliberately longer description than the trigger")

	dims := widget.layout(ctx, testLayoutContext(), state, false)

	if state.triggerRect.Min.Y <= 0 {
		t.Fatalf("trigger y = %d, want below label", state.triggerRect.Min.Y)
	}
	if state.triggerRect.Max.Y >= dims.Size.Y {
		t.Fatalf("trigger bottom = %d, field height = %d; description should be outside anchor", state.triggerRect.Max.Y, dims.Size.Y)
	}
}

func selectTestItems() []SelectItem {
	return []SelectItem{
		{Key: "go", Label: "Go"},
		{Key: "rust", Label: "Rust"},
		{Key: "swift", Label: "Swift", Disabled: true},
	}
}

func selectFocusTestContext() (layout.Context, *input.Router) {
	router := new(input.Router)
	ops := new(op.Ops)
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         ops,
	}, router
}

func layoutSelectTestFrame(ctx *frame.Context, router *input.Router, widget SelectWidget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 240)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutSelectPairTestFrame(ctx *frame.Context, router *input.Router, first, second SelectWidget, now time.Time, reverse bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 240)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	if reverse {
		second.Layout(ctx, gtx)
		first.Layout(ctx, gtx)
	} else {
		first.Layout(ctx, gtx)
		second.Layout(ctx, gtx)
	}
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutNestedSelectTestFrame(ctx *frame.Context, router *input.Router, selectWidget SelectWidget, background *widget.Clickable, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(300, 600)
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	presses := state.ActivePresses(background.History())
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	frame.FocusOnPress(ctx, background, background.History(), presses)
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	childGtx := gtx
	childGtx.Constraints = layout.Constraints{Max: image.Pt(120, viewport.Y)}
	layoutui.Box(selectWidget).PaddingLeft(20).Layout(ctx, childGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutSelectWithBackgroundFrame(ctx *frame.Context, router *input.Router, selectWidget SelectWidget, background *widget.Clickable, naturallyDisabled bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(300, 240)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	selectGtx := gtx
	selectGtx.Constraints = layout.Constraints{Max: image.Pt(120, viewport.Y)}
	if naturallyDisabled {
		selectGtx = selectGtx.Disabled()
	}
	selectWidget.Layout(ctx, selectGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func queueSelectPointerClick(router *input.Router, position f32.Point) {
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
