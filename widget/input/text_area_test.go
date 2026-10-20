package input

import (
	"image"
	"strings"
	"testing"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
)

func TestTextAreaSyncsValueAndConfiguresMultilineEditor(t *testing.T) {
	ctx := newContext(nil)
	NewTextArea("First line\nSecond line").Key("notes").
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
	base := NewTextArea("").Key("notes")
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
		{name: "default", area: NewTextArea("").Key("default"), want: 76},
		{name: "one", area: NewTextArea("").Key("one").Rows(1), want: 38},
		{name: "six", area: NewTextArea("").Key("six").Rows(6), want: 136},
	} {
		t.Run(test.name, func(t *testing.T) {
			dims := test.area.Layout(newContext(nil), testLayoutContext())
			if dims.Size.Y != test.want {
				t.Fatalf("textarea height = %d, want %d", dims.Size.Y, test.want)
			}
		})
	}
}

func TestTextAreaHeightUsesResolvedTypographyAndPadding(t *testing.T) {
	gtx := testLayoutContext()
	lineHeight := unit.Sp(30)
	resolved := style.ResolvedStyle{
		Box:  &style.BoxStyle{Padding: &style.Insets{Top: 7, Bottom: 9}},
		Text: &style.TextStyle{LineHeight: &lineHeight},
	}
	if got := textAreaHeight(gtx, resolved, unit.Sp(30), 2); got != 76 {
		t.Fatalf("resolved textarea height = %v, want 76", got)
	}
}

func TestTextAreaRowsLimitLongContentHeight(t *testing.T) {
	dims := NewTextArea(strings.Repeat("A line of text\n", 100)).Key("notes").
		Rows(6).
		Layout(newContext(nil), testLayoutContext())
	if dims.Size.Y != 136 {
		t.Fatalf("textarea height = %d, want 136", dims.Size.Y)
	}
}

func TestTextAreaFullWidth(t *testing.T) {
	dims := NewTextArea("").Key("notes").FullWidth().Layout(newContext(nil), testLayoutContext())
	if dims.Size.X != 300 {
		t.Fatalf("textarea width = %d, want 300", dims.Size.X)
	}
}

func TestTextAreaFrameConstrainsInnerGeometryToRows(t *testing.T) {
	var got layout.Constraints
	child := func(gtx layout.Context) layout.Dimensions {
		got = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}

	activeTheme := *testtheme.Behavior()
	declaration := (&textAreaState{}).visualDeclarations(NewTextArea("").Key("notes").Variant(TextAreaPrimary).FullWidth(), &activeTheme, 76, nil).Defaults
	resolved := resolveInputTestStyle(&activeTheme, declaration, style.State{})
	state := &textAreaState{}
	NewTextArea("").Key("notes").FullWidth().layoutFrame(newContext(nil), testLayoutContext(), state, resolved, true, child)
	if got.Min != image.Pt(276, 60) || got.Max != image.Pt(276, 60) {
		t.Fatalf("inner constraints = %#v, want fixed 276x60", got)
	}
}

func TestTextAreaParentDisabledClearsHover(t *testing.T) {
	ctx := newContext(nil)
	state := &textAreaState{}
	state.Hovered = true
	core.UseStateWith(ctx, "notes", "textAreaState", func() *textAreaState { return state })
	NewTextArea("").Key("notes").Layout(ctx, testLayoutContext().Disabled())
	if state.Hovered {
		t.Fatal("parent-disabled textarea kept its hover state")
	}
}
