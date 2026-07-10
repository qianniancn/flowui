package radiogroup

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
	"github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
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

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func TestRadioGroupOptions(t *testing.T) {
	items := []RadioItem{{Key: "basic", Label: "Basic"}}
	var changed string
	radio := RadioGroup("plan", "basic", items).
		Horizontal().
		Variant(RadioSecondary).
		Disabled(true).
		Invalid(true).
		OnChange(func(key string) {
			changed = key
		})

	if radio.key != "plan" || radio.selectedKey != "basic" || len(radio.items) != 1 {
		t.Fatal("radio group constructor did not set fields")
	}
	if !radio.horizontal || radio.variant != RadioSecondary || !radio.disabled || !radio.invalid {
		t.Fatal("radio group options were not set")
	}
	radio.onChange("pro")
	if changed != "pro" {
		t.Fatalf("on change = %q, want pro", changed)
	}
}

func TestRadioGroupClickSelectsDifferentItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	items := radioTestItems()
	selected := "basic"
	radio := func() RadioGroupWidget {
		return RadioGroup("plan", selected, items).
			OnChange(func(key string) {
				selected = key
			})
	}

	layoutRadioGroupFrame(ctx, router, radio())
	item := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["pro"]
	if item == nil {
		t.Fatal("missing pro radio item")
	}
	item.clickable.Click()
	layoutRadioGroupFrame(ctx, router, radio())

	if selected != "pro" {
		t.Fatalf("selected = %q, want pro", selected)
	}
}

func TestRadioGroupDoesNotNotifyForSelectedItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var calls int
	radio := RadioGroup("plan", "basic", radioTestItems()).
		OnChange(func(string) {
			calls++
		})

	layoutRadioGroupFrame(ctx, router, radio)
	testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["basic"].clickable.Click()
	layoutRadioGroupFrame(ctx, router, radio)

	if calls != 0 {
		t.Fatalf("on change calls = %d, want 0", calls)
	}
}

func TestRadioGroupDisabledItemDoesNotNotify(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var selected string
	items := []RadioItem{
		{Key: "basic", Label: "Basic"},
		{Key: "pro", Label: "Pro", Disabled: true},
	}
	radio := RadioGroup("plan", "basic", items).
		OnChange(func(key string) {
			selected = key
		})

	layoutRadioGroupFrame(ctx, router, radio)
	testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["pro"].clickable.Click()
	layoutRadioGroupFrame(ctx, router, radio)

	if selected != "" {
		t.Fatalf("selected = %q, want no change", selected)
	}
}

func TestRadioGroupStateSweepsRemovedItems(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutRadioGroupFrame(ctx, router, RadioGroup("plan", "basic", radioTestItems()))
	layoutRadioGroupFrame(ctx, router, RadioGroup("plan", "basic", []RadioItem{{Key: "basic", Label: "Basic"}}))

	if testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["pro"] != nil {
		t.Fatal("removed radio item state was kept")
	}
}

func TestRadioGroupRejectsDuplicateItemKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected duplicate item key panic")
		}
	}()
	RadioGroup("plan", "", []RadioItem{
		{Key: "basic", Label: "Basic"},
		{Key: "basic", Label: "Duplicate"},
	}).Layout(newContext(nil), testLayoutContext())
}

func TestRadioGroupRejectsEmptyItemKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected empty item key panic")
		}
	}()
	RadioGroup("plan", "", []RadioItem{{Label: "Missing key"}}).Layout(newContext(nil), testLayoutContext())
}

func TestRadioStyleSecondaryAndInvalid(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceRaised = color.NRGBA{R: 7, G: 8, B: 9, A: 255}
	style := radioStyleFor(&theme, RadioSecondary, false, false, false, false)
	if style.bg != theme.Palette.SurfaceRaised {
		t.Fatalf("secondary bg = %#v, want %#v", style.bg, theme.Palette.SurfaceRaised)
	}

	style = radioStyleFor(&theme, RadioPrimary, true, false, false, true)
	if style.border != theme.Palette.DangerHover {
		t.Fatalf("invalid hover border = %#v, want %#v", style.border, theme.Palette.DangerHover)
	}
	if style.selectedBg != theme.Palette.DangerHover {
		t.Fatalf("invalid selected bg = %#v, want %#v", style.selectedBg, theme.Palette.DangerHover)
	}
}

func TestRadioSelectionAnimation(t *testing.T) {
	state := new(radioItemState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.selection(gtx, false); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.selection(gtx, true); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(radioSelectDuration / 2)
	mid := state.selection(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(radioSelectDuration)
	if got := state.selection(gtx, true); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestRadioFocusesOnPress(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutRadioGroupFrame(ctx, router, RadioGroup("plan", "basic", radioTestItems()))
	item := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["basic"]

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(10, 10),
	})
	layoutRadioGroupFrame(ctx, router, RadioGroup("plan", "basic", radioTestItems()))

	if !router.Source().Focused(&item.clickable) {
		t.Fatal("radio item did not gain focus after press")
	}
}

func TestRadioGroupArrowKeysSelectAndFocusNextItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "basic"
	radio := func() RadioGroupWidget {
		return RadioGroup("plan", selected, radioTestItems()).
			OnChange(func(key string) {
				selected = key
			})
	}

	layoutRadioGroupFrame(ctx, router, radio())
	state := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["basic"].clickable})
	layoutRadioGroupFrame(ctx, router, radio())

	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutRadioGroupFrame(ctx, router, radio())

	if selected != "pro" {
		t.Fatalf("selected = %q, want pro", selected)
	}
	if !router.Source().Focused(&state.items["pro"].clickable) {
		t.Fatal("next radio item did not gain focus")
	}
}

func TestRadioGroupArrowKeysSkipDisabledItems(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "basic"
	items := []RadioItem{
		{Key: "basic", Label: "Basic"},
		{Key: "pro", Label: "Pro", Disabled: true},
		{Key: "team", Label: "Team"},
	}
	radio := func() RadioGroupWidget {
		return RadioGroup("plan", selected, items).
			OnChange(func(key string) {
				selected = key
			})
	}

	layoutRadioGroupFrame(ctx, router, radio())
	state := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["basic"].clickable})
	layoutRadioGroupFrame(ctx, router, radio())

	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutRadioGroupFrame(ctx, router, radio())

	if selected != "team" {
		t.Fatalf("selected = %q, want team", selected)
	}
}

func TestRadioGroupHomeEndKeysSelectEdges(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "pro"
	radio := func() RadioGroupWidget {
		return RadioGroup("plan", selected, radioTestItems()).
			OnChange(func(key string) {
				selected = key
			})
	}

	layoutRadioGroupFrame(ctx, router, radio())
	state := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["pro"].clickable})
	layoutRadioGroupFrame(ctx, router, radio())

	router.Queue(key.Event{Name: key.NameHome, State: key.Press})
	layoutRadioGroupFrame(ctx, router, radio())
	if selected != "basic" {
		t.Fatalf("selected after Home = %q, want basic", selected)
	}

	router.Queue(key.Event{Name: key.NameEnd, State: key.Press})
	layoutRadioGroupFrame(ctx, router, radio())
	if selected != "pro" {
		t.Fatalf("selected after End = %q, want pro", selected)
	}
}

func TestRadioGroupHorizontalLayoutWraps(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints.Max = image.Pt(100, 200)
	children := []layout.Widget{
		fixedRadioChild(70, 10),
		fixedRadioChild(70, 10),
	}

	dims := layoutui.LayoutItems(gtx, true, 10, 5, children)

	if dims.Size != image.Pt(70, 25) {
		t.Fatalf("wrapped size = %v, want (70,25)", dims.Size)
	}
}

func radioTestItems() []RadioItem {
	return []RadioItem{
		{Key: "basic", Label: "Basic", Description: "For small teams"},
		{Key: "pro", Label: "Pro", Description: "For growing teams"},
	}
}

func fixedRadioChild(width, height int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
	}
}

func layoutRadioGroupFrame(ctx *frame.Context, router *input.Router, radio RadioGroupWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(400, 240)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	radio.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
