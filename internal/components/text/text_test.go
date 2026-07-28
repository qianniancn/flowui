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
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

func TestTextTransparentColor(t *testing.T) {
	text := New("hidden").Color(color.NRGBA{})
	if !text.hasColor {
		t.Fatal("transparent color should still count as explicit")
	}
}

func TestTextStyleAppliesCommonBoxModel(t *testing.T) {
	ctx := textTestContext()
	base := layoutTextFrame(ctx, new(input.Router), New("FlowUI"), time.Unix(1, 0), image.Pt(300, 100))
	padded := layoutTextFrame(ctx, new(input.Router), New("FlowUI").Style(flowstyle.Style{}.Padding(6)), time.Unix(2, 0), image.Pt(300, 100))
	if padded.Size.X != base.Size.X+12 || padded.Size.Y != base.Size.Y+12 {
		t.Fatalf("padded text = %v, base %v", padded.Size, base.Size)
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
	label := configured.labelStyle(ctx, configured.resolveStyleStatic(ctx, flowstyle.StyleState{}))
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
	resolved := textPropertyDeclaration(New("Regular").Weight(font.Normal)).Resolve(flowstyle.StyleState{})
	if resolved.Text == nil || resolved.Text.FontWeight == nil || font.Weight(*resolved.Text.FontWeight) != font.Normal {
		t.Fatalf("explicit normal weight declaration = %#v", resolved.Text)
	}
}

func TestTextStyleCascadesScopeBeforeInstance(t *testing.T) {
	ctx := textTestContext()
	restore := frame.PushStyle(ctx, flowstyle.Style{}.FontSize(18).TextColor(flowstyle.RGB(0x112233)))
	defer restore()
	value := New("FlowUI").
		Weight(font.Bold).
		Style(flowstyle.Style{}.
			FontSize(20).
			FontWeight(int(font.Normal)).
			Typeface("monospace").
			FontStyle(font.Italic).
			Wrap(giotext.WrapGraphemes).
			Truncator("...").
			LineHeightScale(1.25),
		)
	label := value.labelStyle(ctx, value.resolveStyleStatic(ctx, flowstyle.StyleState{}))
	if label.TextSize != 20 || label.Color != flowstyle.RGB(0x112233).Color {
		t.Fatalf("resolved label = size %v color %#v", label.TextSize, label.Color)
	}
	if label.Font.Weight != font.Normal {
		t.Fatalf("resolved weight = %v, want instance normal", label.Font.Weight)
	}
	if label.Font.Typeface != "monospace" || label.Font.Style != font.Italic || label.WrapPolicy != giotext.WrapGraphemes || label.Truncator != "..." || label.LineHeightScale != 1.25 {
		t.Fatalf("resolved common text style = %#v", label)
	}
}

func TestTextConditionalStyleTransition(t *testing.T) {
	ctx := textTestContext()
	start := time.Unix(1, 0)
	from := flowstyle.RGB(0x102030).Color
	to := flowstyle.RGB(0x8090a0).Color
	resolveAt := func(now time.Time, active bool) color.NRGBA {
		frame.BeginFrame(ctx)
		resolved := New("Status").Style(
			flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: from}).
				Transition(flowstyle.PropTextColor, 100*time.Millisecond).
				WhenIf(active, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: to})),
		).resolveLayoutStyle(ctx, layout.Context{Ops: new(op.Ops), Now: now})
		frame.EndFrame(ctx)
		got, ok := styleColor(resolved.Text.Color)
		if !ok {
			t.Fatalf("resolved text color = %#v", resolved.Text.Color)
		}
		return got
	}

	if got := resolveAt(start, false); got != from {
		t.Fatalf("initial text color = %#v, want %#v", got, from)
	}
	if got := resolveAt(start, true); got != from {
		t.Fatalf("transition start = %#v, want %#v", got, from)
	}
	middle := resolveAt(start.Add(50*time.Millisecond), true)
	if middle == from || middle == to {
		t.Fatalf("transition midpoint = %#v", middle)
	}
	if got := resolveAt(start.Add(100*time.Millisecond), true); got != to {
		t.Fatalf("transition end = %#v, want %#v", got, to)
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
