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
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
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

func TestRadioGroupOptions(t *testing.T) {
	items := []RadioItem{{Key: "basic", Label: "Basic"}}
	var changed string
	radio := New(items).Key("plan").SelectedKey("basic").
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
	radio := func() Widget {
		return New(items).Key("plan").SelectedKey(selected).
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
	radio := New(radioTestItems()).Key("plan").SelectedKey("basic").
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
	radio := New(items).Key("plan").SelectedKey("basic").
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
	layoutRadioGroupFrame(ctx, router, New(radioTestItems()).Key("plan").SelectedKey("basic"))
	layoutRadioGroupFrame(ctx, router, New([]RadioItem{{Key: "basic", Label: "Basic"}}).Key("plan").SelectedKey("basic"))

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
	New([]RadioItem{
		{Key: "basic", Label: "Basic"},
		{Key: "basic", Label: "Duplicate"},
	}).Key("plan").Layout(newContext(nil), testLayoutContext())
}

func TestRadioGroupRejectsEmptyItemKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected empty item key panic")
		}
	}()
	New([]RadioItem{{Label: "Missing key"}}).Key("plan").Layout(newContext(nil), testLayoutContext())
}

func TestRadioPartsResolveCompleteStyles(t *testing.T) {
	off := color.NRGBA{R: 1, A: 0xff}
	on := color.NRGBA{G: 2, A: 0xff}
	dot := color.NRGBA{B: 3, A: 0xff}
	label := color.NRGBA{R: 4, A: 0xff}
	description := color.NRGBA{G: 5, A: 0xff}
	custom := style.Style{}.
		Part(style.PartIndicator, style.Style{}.
			Background(style.SolidColor{Color: off}).
			BorderWidth(3).
			Radius(8).
			BoxShadow(1, 2, 3, 4, style.RGBA(0x01020380)).
			Translate(2, 3).
			TextColor(style.SolidColor{Color: dot}).
			Opacity(.5).
			On(style.Checked, style.Style{}.Background(style.LinearGradient(
				style.ColorStop(0, style.SolidColor{Color: off}),
				style.ColorStop(1, style.SolidColor{Color: on}),
			)))).
		Part(style.PartLabel, style.Style{}.TextColor(style.SolidColor{Color: label}).FontSize(22)).
		Part(style.PartDescription, style.Style{}.TextColor(style.SolidColor{Color: description}).FontSize(13))

	ctx := newContext(nil)
	provider, _ := theme.ComponentProvider(core.ActiveTheme(ctx))
	visual := provider.RadioGroup(core.ActiveTheme(ctx), theme.RadioGroupRequest{Variant: theme.RadioPrimary, Override: &theme.RadioGroupOverride{Declaration: custom}})
	widget := New(radioTestItems()).Key("plan").SelectedKey("pro")
	resolved := widget.resolveItemStyle(ctx, layout.Context{Ops: new(op.Ops)}, "plan:pro", style.State{Checked: true, Selected: true}, visual)
	if resolved.indicatorOn.Paint == nil || resolved.indicatorOn.Paint.Opacity == nil || *resolved.indicatorOn.Paint.Opacity != .5 || resolved.indicatorOn.Paint.Border == nil || *resolved.indicatorOn.Paint.Border.Width != 3 || *resolved.indicatorOn.Paint.Radius != 8 || len(resolved.indicatorOn.Paint.Shadows) != 1 {
		t.Fatalf("indicator paint = %#v", resolved.indicatorOn.Paint)
	}
	if _, ok := resolved.indicatorOn.Paint.Background.(style.Gradient); !ok {
		t.Fatalf("selected background = %T, want gradient", resolved.indicatorOn.Paint.Background)
	}
	if resolved.indicatorOn.Trans == nil || *resolved.indicatorOn.Trans.TranslateX != 2 || *resolved.indicatorOn.Trans.TranslateY != 3 {
		t.Fatalf("indicator transform = %#v", resolved.indicatorOn.Trans)
	}
	if got, ok := computed.Color(resolved.indicatorOn.Text.Color); !ok || got != dot {
		t.Fatalf("dot color = %#v, ok %v", got, ok)
	}
	if got, ok := computed.Color(resolved.label.Text.Color); !ok || got != label || *resolved.label.Text.FontSize != 22 {
		t.Fatalf("label style = %#v, color %#v, ok %v", resolved.label.Text, got, ok)
	}
	if got, ok := computed.Color(resolved.description.Text.Color); !ok || got != description || *resolved.description.Text.FontSize != 13 {
		t.Fatalf("description style = %#v, color %#v, ok %v", resolved.description.Text, got, ok)
	}
}

func TestRadioIndicatorPartControlsGeometry(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	gtx := testLayoutContext()
	override := style.Style{}.Part(style.PartIndicator, style.Style{}.Width(12).Height(10))
	widget := New(nil).Key("compact")
	provider, _ := theme.ComponentProvider(&activeTheme)
	visual := provider.RadioGroup(&activeTheme, theme.RadioGroupRequest{Variant: theme.RadioPrimary, Override: &theme.RadioGroupOverride{Declaration: override}})
	resolved := widget.resolveItemStyle(ctx, gtx, "compact:one", style.State{}, visual)
	dims := drawRadio(ctx, gtx, visual, resolved, 1)
	focusSpace := max(gtx.Dp(visual.Metrics.FocusSpace), 1)
	want := image.Pt(gtx.Dp(12)+focusSpace*2, gtx.Dp(10)+focusSpace*2)
	if dims.Size != want {
		t.Fatalf("styled radio size = %v, want %v", dims.Size, want)
	}
}

func TestRadioGroupDelegatesIndicatorDrawingToProvider(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	var got theme.RadioIndicatorRenderRequest
	activeTheme.VisualProvider = incompleteRadioProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.RadioGroupVisual) {
			visual.IndicatorRenderer = func(_ layout.Context, request theme.RadioIndicatorRenderRequest) layout.Dimensions {
				got = request
				return layout.Dimensions{Size: request.Size}
			}
		},
	}
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	gtx := testLayoutContext()
	provider, _ := theme.ComponentProvider(&activeTheme)
	visual := provider.RadioGroup(&activeTheme, theme.RadioGroupRequest{Variant: theme.RadioPrimary})
	widget := New(radioTestItems()).Key("provider-renderer").SelectedKey("pro")
	resolved := widget.resolveItemStyle(ctx, gtx, "provider-renderer:pro", style.State{Checked: true, Selected: true}, visual)
	resolved.selected = 1
	if dims := drawRadio(ctx, gtx, visual, resolved, 1); dims.Size == (image.Point{}) {
		t.Fatal("radio indicator layout returned an empty size")
	}
	if got.Selection != 1 || got.DotScale != visual.Metrics.DotScale || got.DotPressedScale != visual.Metrics.DotPressedScale {
		t.Fatalf("indicator request = %#v, want provider metrics and selected state", got)
	}
}

func TestRadioLabelAndDescriptionPartsUseCommonRenderer(t *testing.T) {
	items := []RadioItem{{Key: "one", Label: "One", Description: "Details"}}
	base := New(items).Key("base").Layout(newContext(nil), testLayoutContext())
	styled := (New(items).Key("styled")).Style(style.Style{}.
		Part(style.PartLabel, style.Style{}.PaddingY(5)).
		Part(style.PartDescription, style.Style{}.PaddingY(4)))
	styledDims := styled.Layout(newContext(nil), testLayoutContext())
	if styledDims.Size.Y <= base.Size.Y {
		t.Fatalf("styled radio height = %d, want greater than %d", styledDims.Size.Y, base.Size.Y)
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
	layoutRadioGroupFrame(ctx, router, New(radioTestItems()).Key("plan").SelectedKey("basic"))
	item := testComponentState[radioGroupState](ctx, "plan", stateSlotRadioGroup).items["basic"]

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(10, 10),
	})
	layoutRadioGroupFrame(ctx, router, New(radioTestItems()).Key("plan").SelectedKey("basic"))

	if !router.Source().Focused(&item.clickable) {
		t.Fatal("radio item did not gain focus after press")
	}
}

func TestRadioGroupArrowKeysSelectAndFocusNextItem(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	selected := "basic"
	radio := func() Widget {
		return New(radioTestItems()).Key("plan").SelectedKey(selected).
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
	radio := func() Widget {
		return New(items).Key("plan").SelectedKey(selected).
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
	radio := func() Widget {
		return New(radioTestItems()).Key("plan").SelectedKey(selected).
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

	dims := layoutui.Items(newContext(nil), gtx, true, 10, 5, children)

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

func layoutRadioGroupFrame(ctx *core.Context, router *input.Router, radio Widget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(400, 240)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	radio.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
