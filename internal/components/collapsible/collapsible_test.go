package collapsible

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
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestCollapsibleOptions(t *testing.T) {
	onChange := func(bool) {}
	widget := Collapsible("details", true, "Details", text.New("Body")).
		Leading(text.New("L")).
		Trailing(text.New("T")).
		Disabled(true).
		OnExpandedChange(onChange)
	if widget.key != "details" || !widget.expanded || widget.label != "Details" || widget.content == nil || widget.leading == nil || widget.trailing == nil || !widget.disabled || widget.onExpandedChange == nil {
		t.Fatal("collapsible options were not retained")
	}

	group := CollapsibleGroup("group", []string{"one"}, testItems()).
		AllowMultipleExpanded(true).
		Disabled(true).
		OnExpandedChange(func([]string) {})
	if group.key != "group" || len(group.expandedKeys) != 1 || len(group.items) != 3 || !group.allowMultipleExpanded || !group.disabled || group.onExpandedChange == nil {
		t.Fatal("collapsible group options were not retained")
	}
}

func TestCollapsibleThemeMatchesHeroUI(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Collapsible
	if tokens.BodyPadding != 8 || tokens.IndicatorSize != 16 || tokens.IndicatorStroke != 1.7 {
		t.Fatalf("collapsible metrics = padding %v indicator %v stroke %v", tokens.BodyPadding, tokens.IndicatorSize, tokens.IndicatorStroke)
	}
	if tokens.ContentDuration != 200*time.Millisecond || tokens.IndicatorDuration != 250*time.Millisecond {
		t.Fatalf("collapsible durations = content %v indicator %v", tokens.ContentDuration, tokens.IndicatorDuration)
	}
}

func TestToggleExpandedKeys(t *testing.T) {
	if got := toggleExpandedKeys([]string{"one"}, "two", false); !slices.Equal(got, []string{"two"}) {
		t.Fatalf("single expansion = %v, want [two]", got)
	}
	if got := toggleExpandedKeys([]string{"one"}, "two", true); !slices.Equal(got, []string{"one", "two"}) {
		t.Fatalf("multiple expansion = %v, want [one two]", got)
	}
	if got := toggleExpandedKeys([]string{"one", "two"}, "one", true); !slices.Equal(got, []string{"two"}) {
		t.Fatalf("collapsed keys = %v, want [two]", got)
	}
}

func TestCollapsibleClickRequestsControlledChange(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	expanded := false
	widget := func() Widget {
		return Collapsible("details", expanded, "Details", text.New("Body")).OnExpandedChange(func(next bool) {
			expanded = next
		})
	}

	layoutCollapsibleFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(320, 160))
	clickAt(router, f32.Pt(20, 20))
	layoutCollapsibleFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(320, 160))
	if !expanded {
		t.Fatal("click did not request expansion")
	}
}

func TestCollapsibleContentAnimatesOpen(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	viewport := image.Pt(320, 160)
	start := time.Unix(1, 0)
	collapsed := layoutCollapsibleFrame(ctx, router, Collapsible("details", false, "Details", text.New("Body")), start, viewport)
	opening := layoutCollapsibleFrame(ctx, router, Collapsible("details", true, "Details", text.New("Body")), start.Add(time.Millisecond), viewport)
	opened := layoutCollapsibleFrame(ctx, router, Collapsible("details", true, "Details", text.New("Body")), start.Add(201*time.Millisecond), viewport)
	if opening.Size.Y != collapsed.Size.Y {
		t.Fatalf("opening height = %d, want initial collapsed height %d", opening.Size.Y, collapsed.Size.Y)
	}
	if opened.Size.Y <= collapsed.Size.Y {
		t.Fatalf("opened height = %d, want greater than %d", opened.Size.Y, collapsed.Size.Y)
	}
}

func TestCollapsibleGroupArrowKeysMoveFocus(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	router := new(input.Router)
	group := CollapsibleGroup("sections", nil, testItems())
	start := time.Unix(1, 0)
	layoutGroupFrame(ctx, router, group, start, image.Pt(320, 200))
	state, _ := frame.PeekState[collapsibleState](ctx, "sections", stateSlotCollapsible)
	router.Source().Execute(key.FocusCmd{Tag: &state.items["one"].clickable})
	layoutGroupFrame(ctx, router, group, start.Add(time.Millisecond), image.Pt(320, 200))
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutGroupFrame(ctx, router, group, start.Add(2*time.Millisecond), image.Pt(320, 200))
	if !router.Source().Focused(&state.items["three"].clickable) {
		t.Fatal("Down Arrow did not skip the disabled item")
	}
}

func TestCollapsibleGroupRejectsInvalidKeys(t *testing.T) {
	mustPanic(t, func() { validateItems([]Item{{Label: "Missing"}}) })
	mustPanic(t, func() { validateItems([]Item{{Key: "same"}, {Key: "same"}}) })
}

func testItems() []Item {
	return []Item{
		{Key: "one", Label: "One", Content: text.New("First")},
		{Key: "two", Label: "Two", Content: text.New("Second"), Disabled: true},
		{Key: "three", Label: "Three", Content: text.New("Third")},
	}
}

func layoutCollapsibleFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time, viewport image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func layoutGroupFrame(ctx *frame.Context, router *input.Router, widget GroupWidget, now time.Time, viewport image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func clickAt(router *input.Router, position f32.Point) {
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position},
	)
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
