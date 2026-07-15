package text

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	giotext "gioui.org/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestTextTransparentColor(t *testing.T) {
	text := New("hidden").Color(color.NRGBA{})
	if !text.hasColor {
		t.Fatal("transparent color should still count as explicit")
	}
}

func TestTextOptionsUseValueSemantics(t *testing.T) {
	base := New("FlowUI")
	configured := base.
		Size(18).
		Color(color.NRGBA{R: 1, G: 2, B: 3, A: 4}).
		Typeface("serif").
		Style(font.Italic).
		Weight(font.Bold).
		Align(giotext.Middle).
		MaxLines(2).
		Truncator("...").
		Wrap(giotext.WrapWords).
		LineHeight(24).
		LineHeightScale(1.2)

	if base.size != 0 || base.hasColor || base.hasTypeface || base.hasStyle || base.hasWeight || base.maxLines != 0 || base.lineHeight != 0 || base.lineHeightScale != 0 {
		t.Fatalf("base Text was mutated: %#v", base)
	}
	if configured.size != 18 || configured.color != (color.NRGBA{R: 1, G: 2, B: 3, A: 4}) || configured.font != (font.Font{Typeface: "serif", Style: font.Italic, Weight: font.Bold}) {
		t.Fatalf("configured Text appearance = %#v", configured)
	}
	if configured.alignment != giotext.Middle || configured.maxLines != 2 || configured.truncator != "..." || configured.wrapPolicy != giotext.WrapWords || configured.lineHeight != 24 || configured.lineHeightScale != 1.2 {
		t.Fatalf("configured Text layout = %#v", configured)
	}
}

func TestTextFontAndLayoutOptionsMapToGioLabel(t *testing.T) {
	ctx := textTestContext()
	configured := New("FlowUI").
		Font(font.Font{Typeface: "monospace", Style: font.Italic, Weight: font.Medium}).
		Align(giotext.End).
		MaxLines(3).
		Truncator("---").
		Wrap(giotext.WrapGraphemes).
		LineHeight(22).
		LineHeightScale(1.1)
	label := configured.labelStyle(ctx)
	if label.Font != (font.Font{Typeface: "monospace", Style: font.Italic, Weight: font.Medium}) {
		t.Fatalf("Gio label font = %#v", label.Font)
	}
	if label.Alignment != giotext.End || label.MaxLines != 3 || label.Truncator != "---" || label.WrapPolicy != giotext.WrapGraphemes || label.LineHeight != 22 || label.LineHeightScale != 1.1 {
		t.Fatalf("Gio label layout = %#v", label)
	}
	if label.SelectionColor != frame.ActiveTheme(ctx).Palette.Selection {
		t.Fatalf("selection color = %#v, want %#v", label.SelectionColor, frame.ActiveTheme(ctx).Palette.Selection)
	}
}

func TestTextExplicitNormalWeightOverridesComponentDefault(t *testing.T) {
	if got := New("Regular").Weight(font.Normal).DefaultWeight(font.Bold).ConfiguredWeight(); got != font.Normal {
		t.Fatalf("explicit normal weight = %v, want normal", got)
	}
	if got := New("Default").DefaultWeight(font.Bold).ConfiguredWeight(); got != font.Bold {
		t.Fatalf("default weight = %v, want bold", got)
	}
}

func TestTextClampsNegativeLayoutValues(t *testing.T) {
	configured := New("FlowUI").MaxLines(-1).LineHeight(-1).LineHeightScale(-1)
	if configured.maxLines != 0 || configured.lineHeight != 0 || configured.lineHeightScale != 0 {
		t.Fatalf("negative layout values = lines %d height %v scale %v", configured.maxLines, configured.lineHeight, configured.lineHeightScale)
	}
}

func TestSelectableTextUsesGioSelectionAndClipboard(t *testing.T) {
	ctx := textTestContext()
	router := new(input.Router)
	selectable := Selectable("message", "Copy this text").MaxLines(1)
	start := time.Unix(1, 0)
	layoutTextFrame(ctx, router, selectable, start, image.Pt(240, 80))
	if !textSemanticTreeContains(router.AppendSemantics(nil), "Copy this text") {
		t.Fatal("SelectableText did not expose its text to semantics")
	}

	stateValue, ok := frame.PeekState[selectableTextState](ctx, "message", stateSlotSelectableText)
	if !ok || stateValue.selectable.Text() != "Copy this text" || stateValue.selectable.MaxLines != 1 {
		t.Fatalf("SelectableText state = %#v", stateValue)
	}
	stateValue.selectable.SetCaret(0, 4)
	router.Source().Execute(key.FocusCmd{Tag: &stateValue.selectable})
	layoutTextFrame(ctx, router, selectable, start.Add(time.Millisecond), image.Pt(240, 80))
	if !stateValue.selectable.Focused() {
		t.Fatal("SelectableText did not receive focus")
	}

	router.Queue(key.Event{Name: "C", Modifiers: key.ModShortcut, State: key.Press})
	layoutTextFrame(ctx, router, selectable, start.Add(2*time.Millisecond), image.Pt(240, 80))
	mime, content, ok := router.WriteClipboard()
	if !ok || mime != "application/text" || string(content) != "Copy" {
		t.Fatalf("clipboard = ok %v mime %q content %q", ok, mime, content)
	}
}

func TestSelectableTextPointerDragKeepsFocusAndSelection(t *testing.T) {
	ctx := textTestContext()
	router := new(input.Router)
	selectable := Selectable("pointer-selection", "Select this text")
	start := time.Unix(2, 0)
	layoutTextFrame(ctx, router, selectable, start, image.Pt(240, 80))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(2, 10)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(80, 10)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(80, 10)},
	)
	layoutTextFrame(ctx, router, selectable, start.Add(time.Millisecond), image.Pt(240, 80))
	layoutTextFrame(ctx, router, selectable, start.Add(2*time.Millisecond), image.Pt(240, 80))

	stateValue, ok := frame.PeekState[selectableTextState](ctx, "pointer-selection", stateSlotSelectableText)
	if !ok {
		t.Fatal("pointer SelectableText state is missing")
	}
	if !stateValue.selectable.Focused() || stateValue.selectable.SelectionLen() == 0 {
		t.Fatalf("pointer selection = focused %v length %d", stateValue.selectable.Focused(), stateValue.selectable.SelectionLen())
	}

	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(220, 70)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(220, 70)},
	)
	layoutTextFrame(ctx, router, selectable, start.Add(3*time.Millisecond), image.Pt(240, 80))
	layoutTextFrame(ctx, router, selectable, start.Add(4*time.Millisecond), image.Pt(240, 80))
	if stateValue.selectable.Focused() {
		t.Fatal("clicking outside SelectableText did not clear focus")
	}
}

func TestSelectableTextReportsTruncation(t *testing.T) {
	ctx := textTestContext()
	router := new(input.Router)
	selectable := Selectable("summary", "A long line of text that cannot fit in the available width").MaxLines(1)
	layoutTextFrame(ctx, router, selectable, time.Unix(2, 0), image.Pt(100, 80))
	stateValue, ok := frame.PeekState[selectableTextState](ctx, "summary", stateSlotSelectableText)
	if !ok || !stateValue.selectable.Truncated() {
		t.Fatal("single-line SelectableText was not truncated")
	}
}

func textTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutTextFrame(ctx *frame.Context, router *input.Router, value Widget, now time.Time, maximum image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := value.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func textSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || textSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
