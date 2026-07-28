package menu

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/frame"
)

type contextMenuFixedWidget struct {
	size image.Point
}

type contextMenuFocusWidget struct {
	size   image.Point
	target event.Tag
}

type contextMenuOverflowWidget struct {
	color color.NRGBA
}

type menuAlignmentWidget struct {
	color color.NRGBA
}

func (w contextMenuOverflowWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, w.color, clip.Rect{Min: image.Pt(0, -1), Max: image.Pt(1, 1)}.Op())
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(10, 10))}
}

func (w menuAlignmentWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, w.color, clip.Rect{Max: image.Pt(2, 2)}.Op())
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(2, 2))}
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

func TestContextMenuUsesMenuRootBackground(t *testing.T) {
	ctx := menuTestContext()
	router := new(input.Router)
	got := frame.ActiveTheme(ctx).Palette.Background
	content := Menu("actions", []Item{{Key: "copy", Label: "Copy"}}).
		BeforeContent(frame.WidgetFunc(func(ctx *frame.Context, _ layout.Context) layout.Dimensions {
			got = ctx.BackgroundColor()
			return layout.Dimensions{}
		}))
	widget := ContextMenu("row-menu", contextMenuFixedWidget{size: image.Pt(160, 100)}, content)

	openContextMenuForTest(ctx, router, widget)
	if want := menuBackgroundColor(frame.ActiveTheme(ctx)); got != want {
		t.Fatalf("context menu background = %#v, want %#v", got, want)
	}
}

func TestContextMenuDoesNotClipTriggerVisual(t *testing.T) {
	window, err := headless.NewWindow(14, 14)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	ctx := menuTestContext()
	var router input.Router
	var ops op.Ops
	want := color.NRGBA{R: 0x24, G: 0x68, B: 0xf2, A: 0xff}
	frame.BeginFrameWithViewport(ctx, image.Pt(14, 14))
	offset := op.Offset(image.Pt(2, 2)).Push(&ops)
	ContextMenu("overflow", contextMenuOverflowWidget{color: want}, Menu("actions", nil)).Layout(ctx, layout.Context{
		Constraints: layout.Exact(image.Pt(10, 10)), Source: router.Source(), Ops: &ops,
	})
	offset.Pop()
	frame.EndFrame(ctx)

	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 14, 14))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := color.NRGBAModel.Convert(pixels.At(2, 1)).(color.NRGBA); got != want {
		t.Fatalf("trigger overflow pixel = %#v, want %#v", got, want)
	}
}

func TestMenuItemContentIsVerticallyCentered(t *testing.T) {
	window, err := headless.NewWindow(80, 30)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	ctx := menuTestContext()
	tokens := &frame.ActiveTheme(ctx).Components.Menu
	tokens.Width = 80
	tokens.Padding = 0
	tokens.ItemMinHeight = 30
	tokens.ItemPaddingX = 0
	tokens.ItemPaddingY = 0
	var router input.Router
	var ops op.Ops
	want := color.NRGBA{R: 0xff, A: 0xff}
	frame.BeginFrameWithViewport(ctx, image.Pt(80, 30))
	item := Item{Key: "probe", Label: "Copy", Leading: menuAlignmentWidget{color: want}}
	menu := Menu("alignment", []Item{item})
	menu.layoutItem(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(80, 30)}, Source: router.Source(), Ops: &ops,
	}, menu.stateFor(ctx), entry{item: item}, true)
	frame.EndFrame(ctx)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 80, 30))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	bounds := image.Rectangle{}
	for y := range 30 {
		for x := range 80 {
			if color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA) == want {
				point := image.Rect(x, y, x+1, y+1)
				if bounds.Empty() {
					bounds = point
				} else {
					bounds = bounds.Union(point)
				}
			}
		}
	}
	if bounds != image.Rect(0, 14, 2, 16) {
		t.Fatalf("menu item content bounds = %v, want vertically centered %v", bounds, image.Rect(0, 14, 2, 16))
	}
}

func TestMenuIndicatorUsesThemeColor(t *testing.T) {
	window, err := headless.NewWindow(40, 30)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	ctx := menuTestContext()
	tokens := &frame.ActiveTheme(ctx).Components.Menu
	tokens.IndicatorColor = color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	var router input.Router
	var ops op.Ops
	frame.BeginFrameWithViewport(ctx, image.Pt(40, 30))
	item := Item{Key: "selected", Kind: ItemRadio, Checked: true}
	menu := Menu("indicator-color", []Item{item})
	menu.layoutItem(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(40, 30)}, Source: router.Source(), Ops: &ops,
	}, menu.stateFor(ctx), entry{item: item}, true)
	frame.EndFrame(ctx)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 40, 30))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	for y := range 30 {
		for x := range 40 {
			if got := color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA); got == tokens.IndicatorColor {
				return
			}
		}
	}
	t.Fatalf("menu indicator did not use theme color %#v", tokens.IndicatorColor)
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
