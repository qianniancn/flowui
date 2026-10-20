package input

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	gioinput "gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
)

type inputGroupFixedWidget struct {
	size image.Point
}

func (w inputGroupFixedWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type inputGroupColorProbe struct {
	got color.NRGBA
}

func (p *inputGroupColorProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	p.got = ctx.ForegroundColor()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func TestGroupOptionsAreImmutableAndInheritInput(t *testing.T) {
	field := New("10").Key("price").Variant(Secondary).Disabled(true).Invalid(true).FullWidth()
	base := Group(field)
	configured := base.
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 16)}).
		Variant(Primary).
		Disabled(false).
		Invalid(false)

	if base.variant != Secondary || !base.disabled || !base.invalid || !base.fullWidth {
		t.Fatalf("inherited group options = %#v", base)
	}
	if base.prefix != nil || base.suffix != nil {
		t.Fatal("base group was mutated")
	}
	if configured.prefix == nil || configured.suffix == nil || configured.variant != Primary || configured.disabled || configured.invalid {
		t.Fatalf("configured group = %#v", configured)
	}
}

func TestGroupTextAreaOptionsAreImmutableAndInherited(t *testing.T) {
	field := NewTextArea("").Key("prompt").
		Variant(TextAreaSecondary).
		Disabled(true).
		Invalid(true).
		FullWidth().
		Rows(5)
	base := GroupTextArea(field)
	configured := base.
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 16)}).
		Variant(Primary).
		Disabled(false).
		Invalid(false)

	if !base.multiline || base.textArea.rows != 5 || base.variant != Secondary || !base.disabled || !base.invalid || !base.fullWidth {
		t.Fatalf("inherited textarea group options = %#v", base)
	}
	if base.prefix != nil || base.suffix != nil {
		t.Fatal("base textarea group was mutated")
	}
	if configured.prefix == nil || configured.suffix == nil || configured.variant != Primary || configured.disabled || configured.invalid {
		t.Fatalf("configured textarea group = %#v", configured)
	}
}

func TestGroupDefaultAndFullWidthLayout(t *testing.T) {
	group := Group(New("").Key("email").Placeholder("name@email.com")).
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 16)})
	dims := group.Layout(newContext(nil), testLayoutContext())
	if dims.Size.Y != 36 {
		t.Fatalf("input group height = %d, want 36", dims.Size.Y)
	}
	if dims.Size.X <= 16+24+48 {
		t.Fatalf("input group width = %d, want room for input and slots", dims.Size.X)
	}

	dims = group.FullWidth().Layout(newContext(nil), testLayoutContext())
	if dims.Size.X != 300 {
		t.Fatalf("full-width input group width = %d, want 300", dims.Size.X)
	}
}

func TestGroupPartsProvideResolvedLayoutMetrics(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	visual := (&inputGroupState{}).visualDeclarations(Group(New("").Key("query")).Variant(Primary), &activeTheme, 0, nil)
	declaration := visual.Defaults
	custom := style.Style{}.
		Part(style.PartContent, style.Style{}.
			PaddingLeft(20).
			PaddingRight(24).
			PaddingY(11).
			MinHeight(54).
			LineHeight(27)).
		Part(style.PartIndicator, style.Style{}.Width(3))
	resolved := resolveInputGroupStyle(ctx, gtx, "group", style.State{}, declaration, custom)
	editorStyle := visual.EditorStyle(resolved.content, resolved.placeholder, resolved.selection)
	left, right := resolvedHorizontalPadding(resolved.content)
	if left != 20 || right != 24 || resolvedVerticalPadding(resolved.content) != 11 || resolvedMinHeight(resolved.content) != 54 || editorStyle.LineHeight != 27 {
		t.Fatalf("content metrics = %v/%v, vertical %v, min-height %v, line-height %v", left, right, resolvedVerticalPadding(resolved.content), resolvedMinHeight(resolved.content), editorStyle.LineHeight)
	}
	if resolved.divider.Box == nil || resolved.divider.Box.Width == nil || *resolved.divider.Box.Width != 3 {
		t.Fatalf("divider width = %#v, want 3dp", resolved.divider.Box)
	}
}

func TestStyledGroupPartPaddingProvidesSlotMetrics(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	visual := (&inputGroupState{}).visualDeclarations(Group(New("").Key("query")), &activeTheme, 0, nil)
	custom := style.Style{}.Part(style.PartPrefix, style.Style{}.PaddingLeft(10).PaddingRight(8))
	resolved := resolveInputGroupStyle(ctx, gtx, "group-padding", style.State{}, visual.Defaults, custom)
	left, right, ok := resolvedExplicitHorizontalPadding(resolved.prefix)
	if !ok || left != 10 || right != 8 {
		t.Fatalf("prefix slot padding = %v/%v/%v, want 10/8/true", left, right, ok)
	}
}

func TestGroupTextAreaRowsControlHeight(t *testing.T) {
	for _, test := range []struct {
		name string
		rows int
		want int
	}{
		{name: "one", rows: 1, want: 38},
		{name: "default", rows: 0, want: 76},
		{name: "six", rows: 6, want: 136},
	} {
		t.Run(test.name, func(t *testing.T) {
			area := NewTextArea("").Key(test.name)
			if test.rows > 0 {
				area = area.Rows(test.rows)
			}
			dims := GroupTextArea(area).
				Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
				Suffix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
				FullWidth().
				Layout(newContext(nil), testLayoutContext())
			if dims.Size != image.Pt(300, test.want) {
				t.Fatalf("textarea group size = %v, want 300x%d", dims.Size, test.want)
			}
		})
	}
}

func TestGroupTextAreaConfiguresMultilineEditor(t *testing.T) {
	ctx := newContext(nil)
	GroupTextArea(NewTextArea("First\nSecond").Key("prompt").Rows(4)).Layout(ctx, testLayoutContext())
	editor := testComponentState[widget.Editor](ctx, "prompt", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing textarea group editor")
	}
	if editor.SingleLine || editor.Submit || editor.Text() != "First\nSecond" {
		t.Fatalf("textarea group editor = single line %v submit %v text %q", editor.SingleLine, editor.Submit, editor.Text())
	}
}

func TestGroupTextAreaAlignsSlotsToTop(t *testing.T) {
	if got := inputGroupChildY(76, 16, 8, true); got != 8 {
		t.Fatalf("multiline slot y = %d, want 8", got)
	}
	if got := inputGroupChildY(76, 16, 8, false); got != 30 {
		t.Fatalf("single-line slot y = %d, want 30", got)
	}
	if got := inputGroupSlotHeight(16, 8, true, true); got != 24 {
		t.Fatalf("multiline slot height = %d, want 24", got)
	}
}

func TestGroupEditorInsetScalesOnceAtHighDPI(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Metric = unit.Metric{PxPerDp: 2, PxPerSp: 2}
	gtx.Constraints = layout.Exact(image.Pt(200, 100))
	var got layout.Constraints
	child := insetInputGroupEditor(func(gtx layout.Context) layout.Dimensions {
		got = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, 12, 12, 8)
	dims := child(gtx)
	if got != layout.Exact(image.Pt(152, 68)) || dims.Size != image.Pt(200, 100) {
		t.Fatalf("high-DPI inset = constraints %#v size %v", got, dims.Size)
	}
}

func TestGroupTextAreaHeightDoesNotOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if got := inputGroupTextAreaContentHeight(20, maxInt, 200); got != 200 {
		t.Fatalf("clamped content height = %d, want 200", got)
	}
}

func TestGroupSlotsUseMutedForeground(t *testing.T) {
	prefix := new(inputGroupColorProbe)
	suffix := new(inputGroupColorProbe)
	Group(New("").Key("email")).Prefix(prefix).Suffix(suffix).Layout(newContext(nil), testLayoutContext())
	if prefix.got == (color.NRGBA{}) || prefix.got != suffix.got {
		t.Fatalf("slot colors = prefix %#v suffix %#v, want the same non-transparent provider color", prefix.got, suffix.got)
	}
}

func TestGroupActionUsesStableSize(t *testing.T) {
	action := GroupAction("clear", "Clear value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)})
	if got := action.Layout(newContext(nil), testLayoutContext()).Size; got != image.Pt(24, 24) {
		t.Fatalf("action size = %v, want 24x24", got)
	}
}

func TestGroupActionTextAndCustomContentRemainAccessibleAndClickable(t *testing.T) {
	for _, custom := range []bool{false, true} {
		active := base.Light()
		ctx := core.New(nil, &active, locale.LanguageAuto)
		router := new(gioinput.Router)
		clicks := 0
		action := GroupAction("copy", "Copy value").OnClick(func() { clicks++ })
		if custom {
			action = action.Content(inputGroupFixedWidget{size: image.Pt(16, 16)}).Label("Copy token")
		}
		frame := func() layout.Dimensions {
			gtx := testLayoutContext()
			gtx.Source = router.Source()
			core.BeginFrame(ctx)
			dims := action.Layout(ctx, gtx)
			core.ApplyFrameCommands(ctx, gtx)
			core.EndFrame(ctx)
			router.Frame(gtx.Ops)
			return dims
		}
		dims := frame()
		wantLabel := "Copy value"
		if custom {
			wantLabel = "Copy token"
		} else if dims.Size.X <= dims.Size.Y {
			t.Fatal("action text did not contribute to layout width")
		}
		found := false
		for _, node := range router.AppendSemantics(nil) {
			if node.Desc.Class == semantic.Button && node.Desc.Label == wantLabel {
				found = true
			}
		}
		if !found {
			t.Fatalf("action semantics missing label %q", wantLabel)
		}
		position := f32.Pt(float32(dims.Size.X)/2, float32(dims.Size.Y)/2)
		router.Queue(
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: position},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: position},
		)
		frame()
		if clicks != 1 {
			t.Fatalf("custom=%v action clicks = %d, want 1", custom, clicks)
		}
	}
}

func TestGroupActionSlotPaddingComesFromProvider(t *testing.T) {
	group := Group(New("").Key("query")).SuffixAction(
		GroupAction("clear", "Clear value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}),
	)
	baseTheme := *testtheme.Behavior()
	baseContext := core.New(nil, &baseTheme, locale.LanguageAuto)
	base := group.Layout(baseContext, testLayoutContext()).Size

	customTheme := *testtheme.Behavior()
	customTheme.VisualProvider = slotPaddingInputVisualProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	customContext := core.New(nil, &customTheme, locale.LanguageAuto)
	custom := group.Layout(customContext, testLayoutContext()).Size
	if custom.X-base.X != 10 {
		t.Fatalf("provider action padding delta = %d, want 10 (5dp per side at 2x)", custom.X-base.X)
	}
}

func TestGroupContentPartControlsEditorCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := (Group(New("flowui").Key("query")).FullWidth()).Style(style.Style{}.
		Part(style.PartContent, style.Style{}.Cursor(pointer.CursorCrosshair)))
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(120, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorCrosshair {
		t.Fatalf("content cursor = %v, want %v", got, pointer.CursorCrosshair)
	}
}

func TestGroupDisabledUsesDefaultCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := Group(New("flowui").Key("query")).Disabled(true).FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(120, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorNotAllowed {
		t.Fatalf("disabled editor cursor = %v, want %v", got, pointer.CursorNotAllowed)
	}
}

func TestGroupActionCanFocusEditorWhenEnabled(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := GroupAction("copy", "Copy value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := Group(New("flow_live_123").Key("token")).
		SuffixAction(action).
		FocusOnActionPress(true).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  f32.Pt(284, 18),
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: 1,
			Position:  f32.Pt(284, 18),
		},
	)
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	editor := testComponentState[widget.Editor](ctx, "token", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing input group editor")
	}
	if !clicked {
		t.Fatal("action press did not reach the action")
	}
	if !router.Source().Focused(editor) {
		t.Fatal("action press did not focus the input when enabled")
	}
}

func TestGroupPrefixPressFocusesInput(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := Group(New("").Key("email")).
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	editor := testComponentState[widget.Editor](ctx, "email", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing input group editor")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(8, 18),
	})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if !router.Source().Focused(editor) {
		t.Fatal("pressing the prefix did not focus the input")
	}
}

func TestGroupPrefixActionRemainsInteractive(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := GroupAction("filter", "Filter values").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := Group(New("flowui").Key("query")).
		PrefixAction(action).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(12, 18)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(12, 18)},
	)
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if !clicked {
		t.Fatal("prefix action did not receive the pointer click")
	}
	editor := testComponentState[widget.Editor](ctx, "query", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing input group editor")
	}
	if router.Source().Focused(editor) {
		t.Fatal("pressing the prefix action focused the input")
	}
}

func TestGroupTextAreaPrefixPressFocusesEditor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := GroupTextArea(NewTextArea("").Key("prompt").Rows(3)).
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	editor := testComponentState[widget.Editor](ctx, "prompt", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing textarea group editor")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(8, 18),
	})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if !router.Source().Focused(editor) {
		t.Fatal("pressing the prefix did not focus the textarea")
	}
}

func TestGroupSuffixButtonRemainsInteractive(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	suffix := GroupAction("copy", "Copy value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := (Group(New("flow_live_123").Key("token")).
		SuffixAction(suffix).
		FullWidth()).Style(style.Style{}.
		Part(style.PartSuffix, style.Style{}.PaddingLeft(12)))
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(120, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorText {
		t.Fatalf("editor cursor = %v, want %v", got, pointer.CursorText)
	}
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(284, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 2*int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorPointer {
		t.Fatalf("suffix cursor = %v, want %v", got, pointer.CursorPointer)
	}
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(284, 18)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(284, 18)},
	)
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 2*int64(time.Millisecond)))
	if !clicked {
		t.Fatal("suffix button did not receive the pointer click")
	}
	editor := testComponentState[widget.Editor](ctx, "token", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing input group editor")
	}
	if router.Source().Focused(editor) {
		t.Fatal("pressing the suffix button focused the input")
	}
}

func layoutInputGroupTestFrame(ctx *core.Context, router *gioinput.Router, group core.Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	group.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
