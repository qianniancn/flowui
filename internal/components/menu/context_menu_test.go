package menu

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type contextMenuFixedWidget struct {
	size image.Point
}

type contextMenuFocusWidget struct {
	size   image.Point
	target event.Tag
}

func (w contextMenuFocusWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	for {
		if _, ok := gtx.Event(key.FocusFilter{Target: w.target}); !ok {
			break
		}
	}
	dims := layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
	event.Op(gtx.Ops, w.target)
	return dims
}

func (w contextMenuFixedWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func contextMenuTestWidget(items []Item) ContextMenuWidget {
	return ContextMenu(
		"row-menu",
		contextMenuFixedWidget{size: image.Pt(160, 100)},
		Menu("actions", items),
	)
}

func TestContextMenuOpensAtSecondaryClickPosition(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
	start := time.Unix(1, 0)
	layoutContextMenuFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: f32.Pt(42, 31)})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	if !state.open || state.anchor != image.Rect(42, 31, 43, 32) {
		t.Fatalf("context menu = open %v anchor %v", state.open, state.anchor)
	}
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+time.Millisecond))
	if !frame.HasTopOverlay(ctx) {
		t.Fatal("context menu did not register a popup overlay")
	}
	menuState := contextRootMenuState(t, ctx, state)
	if !router.Source().Focused(&menuState.item("copy").clickable) {
		t.Fatal("opening context menu did not focus first action")
	}
}

func TestContextMenuShiftF10OpensAtTriggerCenter(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
	start := time.Unix(1, 0)
	layoutContextMenuFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if !router.Source().Focused(&state.trigger) {
		t.Fatal("context menu trigger did not accept keyboard focus")
	}
	router.Queue(key.Event{Name: key.NameF10, Modifiers: key.ModShift, State: key.Press})
	layoutContextMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if !state.open || state.anchor != image.Rect(80, 50, 81, 51) {
		t.Fatalf("Shift+F10 context menu = open %v anchor %v", state.open, state.anchor)
	}
}

func TestContextMenuShiftF10UsesAdditionalFocusTarget(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	target := new(int)
	widget := ContextMenu(
		"child-focus-menu",
		contextMenuFocusWidget{size: image.Pt(160, 100), target: target},
		Menu("actions", []Item{{Key: "copy", Label: "Copy"}}),
	).FocusTargets(target)
	start := time.Unix(1, 0)
	layoutContextMenuFrame(ctx, router, widget, start)
	router.Source().Execute(key.FocusCmd{Tag: target})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if !router.Source().Focused(target) {
		t.Fatal("additional context menu focus target did not accept focus")
	}

	router.Queue(key.Event{Name: key.NameF10, Modifiers: key.ModShift, State: key.Press})
	layoutContextMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	state, _ := frame.PeekState[contextMenuState](ctx, "child-focus-menu", stateSlotContextMenu)
	if !state.open {
		t.Fatal("Shift+F10 on an additional focus target did not open the context menu")
	}
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+3*time.Millisecond))

	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+4*time.Millisecond))
	if !router.Source().Focused(target) {
		t.Fatal("closing the context menu did not restore its additional focus target")
	}
}

func TestControlledContextMenuRequestsOpenWithoutMutatingInternalState(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	requested := make([]bool, 0, 1)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}}).
		Open(false).
		OnOpenChange(func(open bool) { requested = append(requested, open) })
	start := time.Unix(1, 0)
	layoutContextMenuFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: f32.Pt(42, 31)})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	if state.open || len(requested) != 1 || !requested[0] {
		t.Fatalf("controlled open request = internal %v requests %v", state.open, requested)
	}

	layoutContextMenuFrame(ctx, router, widget.Open(true), start.Add(2*time.Millisecond))
	if !state.isOpen(widget.Open(true)) {
		t.Fatal("controlled context menu did not follow the supplied open state")
	}
}

func TestContextMenuLongPressOpensAndMovementCancels(t *testing.T) {
	t.Run("open", func(t *testing.T) {
		ctx := menuTestContext()
		router := new(input.Router)
		widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
		start := time.Unix(1, 0)
		layoutContextMenuFrame(ctx, router, widget, start)
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 7, Position: f32.Pt(50, 40)})
		layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
		layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuLongPressDelay+2*time.Millisecond))
		state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
		if !state.open || state.anchor.Dx() != 10 || state.anchor.Dy() != 10 {
			t.Fatalf("long press = open %v anchor %v", state.open, state.anchor)
		}
	})

	t.Run("cancel movement", func(t *testing.T) {
		ctx := menuTestContext()
		router := new(input.Router)
		widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
		start := time.Unix(1, 0)
		layoutContextMenuFrame(ctx, router, widget, start)
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 9, Position: f32.Pt(20, 20)})
		layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
		router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Touch, PointerID: 9, Position: f32.Pt(40, 20)})
		layoutContextMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
		layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuLongPressDelay+3*time.Millisecond))
		state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
		if state.open {
			t.Fatal("moved touch opened context menu")
		}
	})
}

func TestContextMenuEscapeClosesAndRestoresTriggerFocus(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
	start := openContextMenuForTest(ctx, router, widget)
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	if state.open || !router.Source().Focused(&state.trigger) {
		t.Fatalf("escape = open %v trigger focused %v", state.open, router.Source().Focused(&state.trigger))
	}
	if !frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("ContextMenu Escape did not restore keyboard-visible trigger focus")
	}
}

func TestContextMenuProgrammaticOpenUsesTriggerCenter(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}}).Open(true)
	layoutContextMenuFrame(ctx, router, widget, time.Unix(1, 0))
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	if !state.hasAnchor || state.anchor != image.Rect(80, 50, 81, 51) {
		t.Fatalf("programmatic anchor = %v", state.anchor)
	}
}

func TestContextMenuOutsideClickClosesWithoutRestoringTriggerFocus(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{Key: "copy", Label: "Copy"}})
	start := openContextMenuForTest(ctx, router, widget)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 3, Buttons: pointer.ButtonPrimary, Position: f32.Pt(420, 320)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 3, Position: f32.Pt(420, 320)},
	)
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	if state.open || router.Source().Focused(&state.trigger) {
		t.Fatalf("outside click = open %v trigger focused %v", state.open, router.Source().Focused(&state.trigger))
	}
	if frame.HasTopOverlay(ctx) {
		t.Fatal("outside click retained context menu input ownership for the dismissal frame")
	}
	layoutContextMenuFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if frame.HasTopOverlay(ctx) {
		t.Fatal("outside click left the context menu overlay active during its exit animation")
	}
}

func TestContextMenuActionClosesAndRestoresTriggerFocus(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	action := ""
	widget := ContextMenu(
		"row-menu",
		contextMenuFixedWidget{size: image.Pt(160, 100)},
		Menu("actions", []Item{{Key: "copy", Label: "Copy"}}).OnAction(func(key string) { action = key }),
	)
	start := openContextMenuForTest(ctx, router, widget)
	contextState, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	rootState := contextRootMenuState(t, ctx, contextState)
	rootState.item("copy").clickable.Click()
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if action != "copy" || contextState.open || !router.Source().Focused(&contextState.trigger) {
		t.Fatalf("action = %q open %v trigger focused %v", action, contextState.open, router.Source().Focused(&contextState.trigger))
	}
	if frame.FocusVisible(ctx, &contextState.trigger, true) {
		t.Fatal("pointer-selected ContextMenu item exposed keyboard-visible trigger focus")
	}
}

func TestContextMenuSubmenuOpensAndEscapeReturnsToParent(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{
		Key: "share", Label: "Share", Kind: ItemSubmenu,
		Children: []Item{{Key: "copy-link", Label: "Copy link"}, {Key: "email", Label: "Email"}},
	}})
	start := openContextMenuForTest(ctx, router, widget)
	contextState, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	rootState := contextRootMenuState(t, ctx, contextState)
	rootState.item("share").clickable.Click()
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	if rootState.openSubmenu != "share" {
		t.Fatal("submenu trigger did not open submenu")
	}
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+2*time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+3*time.Millisecond))
	childState := contextSubmenuState(t, ctx, rootState, "share")
	if !router.Source().Focused(&childState.item("copy-link").clickable) {
		t.Fatal("submenu did not focus first child")
	}
	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+4*time.Millisecond))
	if rootState.openSubmenu != "" || !contextState.open || !router.Source().Focused(&rootState.item("share").clickable) {
		t.Fatalf("submenu escape = child %q root open %v parent focused %v", rootState.openSubmenu, contextState.open, router.Source().Focused(&rootState.item("share").clickable))
	}
	if !menuItemFocusVisible(ctx, rootState.item("share"), true) {
		t.Fatal("submenu Escape did not restore keyboard-visible parent focus")
	}
}

func TestContextMenuSubmenuActionClosesRoot(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	action := ""
	widget := ContextMenu(
		"row-menu",
		contextMenuFixedWidget{size: image.Pt(160, 100)},
		Menu("actions", []Item{{
			Key: "share", Label: "Share", Kind: ItemSubmenu,
			Children: []Item{{Key: "copy-link", Label: "Copy link"}},
		}}).OnAction(func(key string) { action = key }),
	)
	start := openContextMenuForTest(ctx, router, widget)
	contextState, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	rootState := contextRootMenuState(t, ctx, contextState)
	rootState.item("share").clickable.Click()
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+2*time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+3*time.Millisecond))
	childState := contextSubmenuState(t, ctx, rootState, "share")
	childState.item("copy-link").clickable.Click()
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+4*time.Millisecond))
	if action != "copy-link" || contextState.open {
		t.Fatalf("submenu action = %q root open %v", action, contextState.open)
	}
}

func TestSubmenuExitKeepsParentMenuHoverInteractive(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := ContextMenu(
		"row-menu",
		contextMenuFixedWidget{size: image.Pt(160, 100)},
		Menu("actions", []Item{
			{Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}}},
			{Key: "rename", Label: "Rename"},
		}),
	)
	start := openContextMenuForTest(ctx, router, widget)
	contextState, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	rootState := contextRootMenuState(t, ctx, contextState)

	hoverStart := start.Add(time.Millisecond)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(60, 50)})
	layoutContextMenuFrame(ctx, router, widget, hoverStart)
	layoutContextMenuFrame(ctx, router, widget, hoverStart.Add(menuSubmenuOpenDelay+time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, hoverStart.Add(menuSubmenuOpenDelay+contextMenuEnterDuration+2*time.Millisecond))
	if rootState.openSubmenu != "share" {
		t.Fatalf("hovered submenu = %q, want share", rootState.openSubmenu)
	}

	switchAt := hoverStart.Add(menuSubmenuOpenDelay + contextMenuEnterDuration + 3*time.Millisecond)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(60, 90)})
	layoutContextMenuFrame(ctx, router, widget, switchAt)
	if rootState.openSubmenu != "" || !rootState.item("rename").clickable.Hovered() || !rootState.submenuActive {
		t.Fatalf("submenu switch = open %q rename hovered %v exit active %v", rootState.openSubmenu, rootState.item("rename").clickable.Hovered(), rootState.submenuActive)
	}

	layoutContextMenuFrame(ctx, router, widget, switchAt.Add(contextMenuExitDuration/2))
	if !rootState.item("rename").clickable.Hovered() {
		t.Fatal("parent Menu hover stopped during submenu exit animation")
	}
}

func TestClickingHoveredSubmenuParentKeepsPointerFocusHidden(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	widget := contextMenuTestWidget([]Item{{
		Key: "share", Label: "Share", Children: []Item{{Key: "copy-link", Label: "Copy link"}},
	}})
	start := openContextMenuForTest(ctx, router, widget)
	contextState, _ := frame.PeekState[contextMenuState](ctx, "row-menu", stateSlotContextMenu)
	rootState := contextRootMenuState(t, ctx, contextState)

	hoverStart := start.Add(time.Millisecond)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(60, 50)})
	layoutContextMenuFrame(ctx, router, widget, hoverStart)
	layoutContextMenuFrame(ctx, router, widget, hoverStart.Add(menuSubmenuOpenDelay+time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, hoverStart.Add(menuSubmenuOpenDelay+contextMenuEnterDuration+2*time.Millisecond))
	if rootState.openSubmenu != "share" {
		t.Fatalf("hovered submenu = %q, want share", rootState.openSubmenu)
	}

	clickAt := hoverStart.Add(menuSubmenuOpenDelay + contextMenuEnterDuration + 3*time.Millisecond)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 3, Buttons: pointer.ButtonPrimary, Position: f32.Pt(60, 50)})
	layoutContextMenuFrame(ctx, router, widget, clickAt)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 3, Position: f32.Pt(60, 50)})
	layoutContextMenuFrame(ctx, router, widget, clickAt.Add(time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, clickAt.Add(2*time.Millisecond))

	parent := rootState.item("share")
	if !router.Source().Focused(&parent.clickable) {
		t.Fatal("clicked submenu parent did not receive focus")
	}
	if menuItemFocusVisible(ctx, parent, true) {
		t.Fatal("pointer dismissal overwrote submenu parent with keyboard-visible focus")
	}
}

func contextRootMenuState(t *testing.T, ctx *frame.Context, state *contextMenuState) *menuState {
	t.Helper()
	key := frame.DerivedKey(ctx, state.key, "menu")
	value, ok := frame.PeekState[menuState](ctx, key, stateSlotMenu)
	if !ok {
		t.Fatal("missing root menu state")
	}
	return value
}

func contextSubmenuState(t *testing.T, ctx *frame.Context, parent *menuState, itemKey string) *menuState {
	t.Helper()
	key := frame.DerivedKey(ctx, parent.key, "submenu:"+itemKey)
	value, ok := frame.PeekState[menuState](ctx, key, stateSlotMenu)
	if !ok {
		t.Fatal("missing submenu state")
	}
	return value
}

func openContextMenuForTest(ctx *frame.Context, router *input.Router, widget ContextMenuWidget) time.Time {
	start := time.Unix(2, 0)
	layoutContextMenuFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: f32.Pt(40, 30)})
	layoutContextMenuFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutContextMenuFrame(ctx, router, widget, start.Add(contextMenuEnterDuration+time.Millisecond))
	return start.Add(contextMenuEnterDuration + time.Millisecond)
}

func layoutContextMenuFrame(ctx *frame.Context, router *input.Router, widget ContextMenuWidget, now time.Time) {
	viewport := image.Pt(480, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Constraints{Max: image.Pt(160, 100)}
	widget.Layout(ctx, triggerGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
