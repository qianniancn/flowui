package input

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	gioinput "gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type inputGroupFixedWidget struct {
	size image.Point
}

func (w inputGroupFixedWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type inputGroupColorProbe struct {
	got color.NRGBA
}

func (p *inputGroupColorProbe) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	p.got = ctx.ForegroundColor()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func TestInputGroupOptionsAreImmutableAndInheritInput(t *testing.T) {
	field := Input("price", "10").Variant(InputSecondary).Disabled(true).Invalid(true).FullWidth()
	base := InputGroup(field)
	configured := base.
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 16)}).
		PrefixPadding(8, 4).
		SuffixPadding(6, 0).
		Variant(InputPrimary).
		Disabled(false).
		Invalid(false)

	if base.variant != InputSecondary || !base.disabled || !base.invalid || !base.fullWidth {
		t.Fatalf("inherited group options = %#v", base)
	}
	if base.prefix != nil || base.suffix != nil {
		t.Fatal("base group was mutated")
	}
	if configured.prefix == nil || configured.suffix == nil || !configured.hasPrefixPadding || !configured.hasSuffixPadding || configured.variant != InputPrimary || configured.disabled || configured.invalid {
		t.Fatalf("configured group = %#v", configured)
	}
}

func TestInputGroupTextAreaOptionsAreImmutableAndInherited(t *testing.T) {
	field := TextArea("prompt", "").
		Variant(TextAreaSecondary).
		Disabled(true).
		Invalid(true).
		FullWidth().
		Rows(5)
	base := InputGroupTextArea(field)
	configured := base.
		Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 16)}).
		Variant(InputPrimary).
		Disabled(false).
		Invalid(false)

	if !base.multiline || base.textArea.rows != 5 || base.variant != InputSecondary || !base.disabled || !base.invalid || !base.fullWidth {
		t.Fatalf("inherited textarea group options = %#v", base)
	}
	if base.prefix != nil || base.suffix != nil {
		t.Fatal("base textarea group was mutated")
	}
	if configured.prefix == nil || configured.suffix == nil || configured.variant != InputPrimary || configured.disabled || configured.invalid {
		t.Fatalf("configured textarea group = %#v", configured)
	}
}

func TestInputGroupDefaultAndFullWidthLayout(t *testing.T) {
	group := InputGroup(Input("email", "").Placeholder("name@email.com")).
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

func TestInputGroupTextAreaRowsControlHeight(t *testing.T) {
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
			area := TextArea(test.name, "")
			if test.rows > 0 {
				area = area.Rows(test.rows)
			}
			dims := InputGroupTextArea(area).
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

func TestInputGroupTextAreaConfiguresMultilineEditor(t *testing.T) {
	ctx := newContext(nil)
	InputGroupTextArea(TextArea("prompt", "First\nSecond").Rows(4)).Layout(ctx, testLayoutContext())
	editor := testComponentState[widget.Editor](ctx, "prompt", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing textarea group editor")
	}
	if editor.SingleLine || editor.Submit || editor.Text() != "First\nSecond" {
		t.Fatalf("textarea group editor = single line %v submit %v text %q", editor.SingleLine, editor.Submit, editor.Text())
	}
}

func TestInputGroupTextAreaAlignsSlotsToTop(t *testing.T) {
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

func TestInputGroupEditorInsetScalesOnceAtHighDPI(t *testing.T) {
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

func TestInputGroupTextAreaHeightDoesNotOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if got := inputGroupTextAreaContentHeight(20, maxInt, 200); got != 200 {
		t.Fatalf("clamped content height = %d, want 200", got)
	}
}

func TestInputGroupStylesMatchHeroUIStates(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	primaryDeclaration := inputGroupDefaultDeclaration(&activeTheme, InputPrimary, false, 0)
	secondaryDeclaration := inputGroupDefaultDeclaration(&activeTheme, InputSecondary, false, 0)
	primary := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{})
	hovered := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Hovered: true})
	focused := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Focused: true})
	secondary := resolveInputTestStyle(&activeTheme, secondaryDeclaration, flowstyle.StyleState{})
	secondaryHovered := resolveInputTestStyle(&activeTheme, secondaryDeclaration, flowstyle.StyleState{Hovered: true})
	invalid := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Invalid: true})
	disabled := resolveInputTestStyle(&activeTheme, primaryDeclaration, flowstyle.StyleState{Disabled: true})
	divider := resolveInputTestPart(&activeTheme, primaryDeclaration, flowstyle.PartIndicator, flowstyle.StyleState{})

	if resolvedBackground(primary) != activeTheme.Palette.FieldBackgroundColor() || primary.Paint == nil || len(primary.Paint.Shadows) != 3 || resolvedBackground(divider) != activeTheme.Palette.Border {
		t.Fatalf("primary style = %#v", primary)
	}
	wantPrimaryHover := color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff}
	if resolvedBackground(hovered) != wantPrimaryHover {
		t.Fatalf("primary hover = %#v, want %#v", resolvedBackground(hovered), wantPrimaryHover)
	}
	if resolvedBackground(focused) != activeTheme.Palette.FieldFocusColor() || focused.Paint.Outline.Color != (flowstyle.SolidColor{Color: activeTheme.Palette.Focus}) || focused.Paint.Outline.Width != 2 {
		t.Fatalf("focused style = %#v", focused)
	}
	if resolvedBackground(secondary) != activeTheme.Palette.DefaultColor() || secondary.Paint == nil || len(secondary.Paint.Shadows) != 0 {
		t.Fatalf("secondary style = %#v", secondary)
	}
	if resolvedBackground(secondaryHovered) != activeTheme.Palette.DefaultHoverColor() {
		t.Fatalf("secondary hover = %#v", secondaryHovered)
	}
	if invalid.Paint.Outline.Color != (flowstyle.SolidColor{Color: activeTheme.Palette.Danger}) || invalid.Paint.Outline.Width != 1 {
		t.Fatalf("invalid style = %#v", invalid)
	}
	if disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled opacity = %#v", disabled.Paint)
	}
	darkTheme := theme.DarkTheme()
	if dark := resolveInputTestStyle(&darkTheme, inputGroupDefaultDeclaration(&darkTheme, InputPrimary, false, 0), flowstyle.StyleState{}); resolvedBackground(dark) != darkTheme.Palette.FieldBackgroundColor() || len(dark.Paint.Shadows) != 0 {
		t.Fatalf("dark input group style = %#v", dark)
	}
}

func TestInputGroupSlotsUseMutedForeground(t *testing.T) {
	prefix := new(inputGroupColorProbe)
	suffix := new(inputGroupColorProbe)
	InputGroup(Input("email", "")).Prefix(prefix).Suffix(suffix).Layout(newContext(nil), testLayoutContext())
	want := theme.DefaultTheme().Palette.MutedForeground
	if prefix.got != want || suffix.got != want {
		t.Fatalf("slot colors = prefix %#v suffix %#v, want %#v", prefix.got, suffix.got, want)
	}
}

func TestInputGroupActionUsesStableSize(t *testing.T) {
	action := InputGroupAction("clear", "Clear value", inputGroupFixedWidget{size: image.Pt(16, 16)})
	if got := action.Layout(newContext(nil), testLayoutContext()).Size; got != image.Pt(24, 24) {
		t.Fatalf("action size = %v, want 24x24", got)
	}
}

func TestInputGroupActionHasKeyboardFocusStyle(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	base := resolveInputTestStyle(&activeTheme, inputGroupActionStyle(), flowstyle.StyleState{})
	focused := resolveInputTestStyle(&activeTheme, inputGroupActionStyle(), flowstyle.StyleState{FocusVisible: true})
	if base.Paint == nil || base.Paint.Outline == nil || base.Paint.Outline.Width != 2 {
		t.Fatalf("action base outline = %#v, want a 2dp transparent ring", base.Paint)
	}
	baseColor, ok := styleruntime.Color(base.Paint.Outline.Color)
	if !ok || baseColor.A != 0 {
		t.Fatalf("action base outline color = %#v, want transparent", base.Paint.Outline.Color)
	}
	if focused.Paint == nil || focused.Paint.Outline == nil || focused.Paint.Outline.Width != 2 {
		t.Fatalf("action focus outline = %#v, want a 2dp ring", focused.Paint)
	}
	focusColor, ok := styleruntime.Color(focused.Paint.Outline.Color)
	if !ok || focusColor != activeTheme.Palette.Focus {
		t.Fatalf("action focus outline color = %#v, want %#v", focused.Paint.Outline.Color, activeTheme.Palette.Focus)
	}
}

func TestInputGroupContentPartControlsEditorCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := InputGroup(Input("query", "flowui")).
		Style(flowstyle.Style{}.
			Part(flowstyle.PartContent, flowstyle.Style{}.Cursor(pointer.CursorCrosshair))).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(120, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorCrosshair {
		t.Fatalf("content cursor = %v, want %v", got, pointer.CursorCrosshair)
	}
}

func TestInputGroupDisabledUsesDefaultCursor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := InputGroup(Input("query", "flowui")).Disabled(true).FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(120, 18)})
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if got := router.Cursor(); got != pointer.CursorNotAllowed {
		t.Fatalf("disabled editor cursor = %v, want %v", got, pointer.CursorNotAllowed)
	}
}

func TestInputGroupActionCanFocusEditorWhenEnabled(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := InputGroupAction("copy", "Copy value", inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := InputGroup(Input("token", "flow_live_123")).
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

func TestInputGroupPrefixPressFocusesInput(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := InputGroup(Input("email", "")).
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

func TestInputGroupPrefixActionRemainsInteractive(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := InputGroupAction("filter", "Filter values", inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := InputGroup(Input("query", "flowui")).
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

func TestInputGroupTextAreaPrefixPressFocusesEditor(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	group := InputGroupTextArea(TextArea("prompt", "").Rows(3)).
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

func TestInputGroupSuffixButtonRemainsInteractive(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	suffix := InputGroupAction("copy", "Copy value", inputGroupFixedWidget{size: image.Pt(16, 16)}).
		OnClick(func() { clicked = true })
	group := InputGroup(Input("token", "flow_live_123")).
		SuffixAction(suffix).
		SuffixPadding(12, 0).
		FullWidth()
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

func layoutInputGroupTestFrame(ctx *frame.Context, router *gioinput.Router, group InputGroupWidget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	group.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
