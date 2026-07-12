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
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func TestInputGroupHeroUIDefaultTheme(t *testing.T) {
	tokens := theme.DefaultTheme().Components.InputGroup
	if tokens.MinHeight != 36 || tokens.Radius != 12 || tokens.PaddingX != 12 || tokens.DividerWidth != 0 {
		t.Fatalf("input group geometry = %#v", tokens)
	}
	if tokens.TextSize != 14 || tokens.LineHeight != 20 || tokens.FocusRingWidth != 2 || tokens.InvalidOutlineWidth != 1 || tokens.ShadowOpacity != 1 {
		t.Fatalf("input group state tokens = %#v", tokens)
	}
	if got := theme.DarkTheme().Components.InputGroup.ShadowOpacity; got != 0 {
		t.Fatalf("dark input group shadow opacity = %v, want 0", got)
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

func TestInputGroupStylesMatchHeroUIStates(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	primary := inputGroupStyleFor(&activeTheme, InputPrimary, false, false, false, false)
	hovered := inputGroupStyleFor(&activeTheme, InputPrimary, true, false, false, false)
	focused := inputGroupStyleFor(&activeTheme, InputPrimary, true, true, false, false)
	secondary := inputGroupStyleFor(&activeTheme, InputSecondary, false, false, false, false)
	secondaryHovered := inputGroupStyleFor(&activeTheme, InputSecondary, true, false, false, false)
	invalid := inputGroupStyleFor(&activeTheme, InputPrimary, false, false, false, true)
	disabled := inputGroupStyleFor(&activeTheme, InputPrimary, false, false, true, false)

	if primary.Background != activeTheme.Palette.Surface || primary.ShadowOpacity != 1 || primary.Divider != activeTheme.Palette.Border {
		t.Fatalf("primary style = %#v", primary)
	}
	wantPrimaryHover := color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff}
	if hovered.Background != wantPrimaryHover {
		t.Fatalf("primary hover = %#v, want %#v", hovered.Background, wantPrimaryHover)
	}
	if focused.Background != activeTheme.Palette.Surface || focused.Ring != activeTheme.Palette.Focus || focused.RingWidth != 2 {
		t.Fatalf("focused style = %#v", focused)
	}
	if secondary.Background != activeTheme.Palette.SurfacePressed || secondary.ShadowOpacity != 0 {
		t.Fatalf("secondary style = %#v", secondary)
	}
	if secondaryHovered.Background != activeTheme.Palette.Border {
		t.Fatalf("secondary hover = %#v", secondaryHovered)
	}
	if invalid.Ring != activeTheme.Palette.Danger || invalid.RingWidth != 1 {
		t.Fatalf("invalid style = %#v", invalid)
	}
	if disabled.Opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled opacity = %v", disabled.Opacity)
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

func TestInputGroupSuffixButtonRemainsInteractive(t *testing.T) {
	ctx := newContext(nil)
	router := new(gioinput.Router)
	clicked := false
	suffix := button.Button("copy", inputGroupFixedWidget{size: image.Pt(16, 16)}).
		Variant(button.ButtonGhost).
		Size(button.ButtonSmall).
		IconOnly().
		OnClick(func() { clicked = true })
	group := InputGroup(Input("token", "flow_live_123")).
		Suffix(suffix).
		SuffixPadding(12, 0).
		FullWidth()
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, 0))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(284, 18)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(284, 18)},
	)
	layoutInputGroupTestFrame(ctx, router, group, time.Unix(1, int64(time.Millisecond)))
	if !clicked {
		t.Fatal("suffix button did not receive the pointer click")
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
