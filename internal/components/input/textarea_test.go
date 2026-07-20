package input

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestTextAreaSyncsValueAndConfiguresMultilineEditor(t *testing.T) {
	ctx := newContext(nil)
	TextArea("notes", "First line\nSecond line").
		ReadOnly(true).
		MaxLength(120).
		Layout(ctx, testLayoutContext())

	editor := testComponentState[widget.Editor](ctx, "notes", stateSlotEditor)
	if editor == nil {
		t.Fatal("missing editor state")
	}
	if got := editor.Text(); got != "First line\nSecond line" {
		t.Fatalf("editor text = %q", got)
	}
	if editor.SingleLine || editor.Submit || !editor.ReadOnly || editor.MaxLen != 120 {
		t.Fatalf("editor configuration = single line %v submit %v read only %v max %d", editor.SingleLine, editor.Submit, editor.ReadOnly, editor.MaxLen)
	}
	if editor.Mask != 0 || editor.InputHint != key.HintText || editor.Filter != "" {
		t.Fatalf("editor text config = mask %q hint %v filter %q", editor.Mask, editor.InputHint, editor.Filter)
	}
}

func TestTextAreaOptionsAreImmutable(t *testing.T) {
	base := TextArea("notes", "")
	configured := base.
		Placeholder("Meeting notes").
		OnChange(func(string) {}).
		Variant(TextAreaSecondary).
		Disabled(true).
		Invalid(true).
		FullWidth().
		ReadOnly(true).
		MaxLength(240).
		Rows(6).
		Label("Notes")

	if configured.hint != "Meeting notes" || configured.onChange == nil || configured.variant != TextAreaSecondary || !configured.disabled || !configured.invalid || !configured.fullWidth || !configured.readOnly || configured.maxLength != 240 || configured.rows != 6 || configured.label != "Notes" {
		t.Fatalf("configured textarea = %#v", configured)
	}
	if base.hint != "" || base.onChange != nil || base.variant != TextAreaPrimary || base.disabled || base.invalid || base.fullWidth || base.readOnly || base.maxLength != 0 || base.rows != 0 || base.label != "" {
		t.Fatalf("base textarea was mutated: %#v", base)
	}
	if got := base.MaxLength(-1).maxLength; got != 0 {
		t.Fatalf("negative max length = %d, want 0", got)
	}
	if got := base.Rows(0).rows; got != 1 {
		t.Fatalf("non-positive rows = %d, want 1", got)
	}
}

func TestTextAreaRowsControlHeight(t *testing.T) {
	for _, test := range []struct {
		name string
		area TextAreaWidget
		want int
	}{
		{name: "default", area: TextArea("default", ""), want: 76},
		{name: "one", area: TextArea("one", "").Rows(1), want: 38},
		{name: "six", area: TextArea("six", "").Rows(6), want: 136},
	} {
		t.Run(test.name, func(t *testing.T) {
			dims := test.area.Layout(newContext(nil), testLayoutContext())
			if dims.Size.Y != test.want {
				t.Fatalf("textarea height = %d, want %d", dims.Size.Y, test.want)
			}
		})
	}
}

func TestTextAreaFullWidth(t *testing.T) {
	dims := TextArea("notes", "").FullWidth().Layout(newContext(nil), testLayoutContext())
	if dims.Size.X != 300 {
		t.Fatalf("textarea width = %d, want 300", dims.Size.X)
	}
}

func TestTextAreaFrameKeepsInnerGeometry(t *testing.T) {
	var got layout.Constraints
	child := func(gtx layout.Context) layout.Dimensions {
		got = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}

	TextArea("notes", "").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), new(inputState), inputStyle{Opacity: 1}, true, child)
	if got.Min != image.Pt(276, 60) || got.Max != image.Pt(276, 60) {
		t.Fatalf("inner constraints = %#v, want 276x60", got)
	}
}

func TestTextAreaHeroUIDefaultTheme(t *testing.T) {
	tokens := theme.DefaultTheme().Components.TextArea
	if tokens.MinHeight != 38 || tokens.Radius != 12 || tokens.PaddingX != 12 || tokens.PaddingY != 8 || tokens.TextSize != 14 || tokens.LineHeight != 20 {
		t.Fatalf("textarea geometry = %#v", tokens)
	}
	if tokens.FocusRingWidth != 2 || tokens.InvalidOutlineWidth != 1 || tokens.ShadowOpacity != 1 || tokens.ShadowStrength != 1.5 || tokens.ShadowColor != (color.NRGBA{A: 0xff}) {
		t.Fatalf("textarea state tokens = %#v", tokens)
	}
	darkTokens := theme.DarkTheme().Components.TextArea
	if darkTokens.ShadowOpacity != 0 || darkTokens.ShadowStrength != 1.5 {
		t.Fatalf("dark textarea shadow = %#v", darkTokens)
	}
}

func TestTextAreaStyleUsesTextAreaThemeTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.TextArea.FocusRingWidth = 4
	activeTheme.Components.TextArea.InvalidOutlineWidth = 3
	activeTheme.Components.TextArea.ShadowOpacity = 0.25

	primary := textAreaStyleFor(&activeTheme, TextAreaPrimary, false, false, false, false)
	focused := textAreaStyleFor(&activeTheme, TextAreaPrimary, false, true, false, false)
	invalid := textAreaStyleFor(&activeTheme, TextAreaPrimary, false, false, false, true)
	secondary := textAreaStyleFor(&activeTheme, TextAreaSecondary, false, false, false, false)
	if primary.ShadowOpacity != 0.25 || focused.RingWidth != 4 || invalid.RingWidth != 3 || secondary.ShadowOpacity != 0 {
		t.Fatalf("textarea styles = primary %#v focused %#v invalid %#v secondary %#v", primary, focused, invalid, secondary)
	}
	darkTheme := theme.DarkTheme()
	if dark := textAreaStyleFor(&darkTheme, TextAreaPrimary, false, false, false, false); dark.Background != darkTheme.Palette.FieldBackgroundColor() || dark.ShadowOpacity != 0 {
		t.Fatalf("dark textarea style = %#v", dark)
	}
}

func TestTextAreaParentDisabledClearsHover(t *testing.T) {
	ctx := newContext(nil)
	state := &inputState{}
	state.Hovered = true
	frame.UseStateWith(ctx, "notes", stateSlotInput, func() *inputState { return state })
	TextArea("notes", "").Layout(ctx, testLayoutContext().Disabled())
	if state.Hovered {
		t.Fatal("parent-disabled textarea kept its hover state")
	}
}
