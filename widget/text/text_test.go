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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type countingTextProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

func (provider countingTextProvider) Text(active *theme.Theme, request theme.TextRequest) theme.TextVisual {
	(*provider.calls)++
	return provider.ComponentVisualProvider.Text(active, request)
}

func textThemeWithProvider(provider theme.ComponentVisualProvider) *theme.Theme {
	value := *testtheme.Behavior()
	value.VisualProvider = provider
	return &value
}

func TestTextVisualCacheReusesProviderResult(t *testing.T) {
	calls := 0
	ctx := core.New(nil, textThemeWithProvider(countingTextProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}), locale.LanguageEnglish)
	value := New("FlowUI")

	first := value.visualWithCache(ctx)
	second := value.visualWithCache(ctx)
	if calls != 1 {
		t.Fatalf("text provider calls = %d, want 1", calls)
	}
	firstResolved := first.Defaults.Resolve(style.State{})
	secondResolved := second.Defaults.Resolve(style.State{})
	if first.Selection != second.Selection || firstResolved.Text == nil || secondResolved.Text == nil ||
		*firstResolved.Text.FontSize != *secondResolved.Text.FontSize ||
		firstResolved.Text.Color != secondResolved.Text.Color ||
		*firstResolved.Text.Typeface != *secondResolved.Text.Typeface {
		t.Fatal("cached text visual changed between identical requests")
	}
	stats := value.visualCache.stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 0 {
		t.Fatalf("text visual cache stats = %#v, want one hit and one miss", stats)
	}
}

func TestTextVisualCacheInvalidatesForegroundAndTheme(t *testing.T) {
	calls := 0
	active := textThemeWithProvider(countingTextProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	})
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("FlowUI")

	value.visualWithCache(ctx)
	restore := core.PushColors(ctx, color.NRGBA{R: 0x99, A: 0xff}, color.NRGBA{})
	value.visualWithCache(ctx)
	restore()
	if calls != 2 {
		t.Fatalf("foreground change provider calls = %d, want 2", calls)
	}

	core.ReplaceTheme(ctx, *textThemeWithProvider(countingTextProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}))
	value.visualWithCache(ctx)
	if calls != 3 {
		t.Fatalf("theme change provider calls = %d, want 3", calls)
	}
	stats := value.visualCache.stats()
	if stats.KeyInvalidations != 2 || stats.ThemeInvalidations != 1 || stats.Misses != 3 {
		t.Fatalf("text visual invalidation stats = %#v, want two key and one theme invalidation", stats)
	}
}

func TestTextTransparentColor(t *testing.T) {
	text := New("hidden").Color(color.NRGBA{})
	if !text.hasColor {
		t.Fatal("transparent color should still count as explicit")
	}
}

func TestTextStyleAppliesCommonBoxModel(t *testing.T) {
	ctx := textTestContext()
	base := layoutTextFrame(ctx, new(input.Router), New("FlowUI"), time.Unix(1, 0), image.Pt(300, 100))
	padded := layoutTextFrame(ctx, new(input.Router), New("FlowUI").Style(style.Style{}.Padding(6)), time.Unix(2, 0), image.Pt(300, 100))
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
	label := configured.labelStyle(ctx, configured.resolveStyleStatic(ctx, style.State{}))
	if label.Font != (font.Font{Typeface: "monospace", Style: font.Italic, Weight: font.Medium}) {
		t.Fatalf("Gio label font = %#v", label.Font)
	}
	if label.Alignment != giotext.End || label.MaxLines != 3 || label.Truncator != "---" || label.WrapPolicy != giotext.WrapGraphemes || label.LineHeight != 22 || label.LineHeightScale != 1.1 {
		t.Fatalf("Gio label layout = %#v", label)
	}
	if label.SelectionColor != testtheme.Behavior().Semantic().Selection {
		t.Fatalf("selection color = %#v, want %#v", label.SelectionColor, testtheme.Behavior().Semantic().Selection)
	}
}

func TestTextExplicitNormalWeightOverridesComponentDefault(t *testing.T) {
	resolved := textPropertyDeclaration(New("Regular").Weight(font.Normal)).Resolve(style.State{})
	if resolved.Text == nil || resolved.Text.FontWeight == nil || font.Weight(*resolved.Text.FontWeight) != font.Normal {
		t.Fatalf("explicit normal weight declaration = %#v", resolved.Text)
	}
}

func TestTextWithResolvedTextUsesSharedDeclaration(t *testing.T) {
	ctx := textTestContext()
	resolved := style.Style{}.
		FontSize(21).
		FontWeight(int(font.Bold)).
		TextColor(style.RGB(0x123456)).
		Resolve(style.State{}).Text
	if resolved == nil {
		t.Fatal("resolved text declaration is nil")
	}

	value := WithResolvedText(New("FlowUI").Style(style.Style{}.
		FontSize(99).
		TextColor(style.RGB(0xff0000))), resolved)
	restore := core.PushInheritedStyle(ctx, style.Style{}.
		FontSize(18).
		TextColor(style.RGB(0x00ff00)))
	defer restore()

	layoutResolved := value.resolveLayoutStyle(ctx, layout.Context{Ops: new(op.Ops)})
	if layoutResolved.Text != resolved {
		t.Fatalf("layout resolved text pointer = %p, want %p", layoutResolved.Text, resolved)
	}
	measureResolved := value.resolveMeasureStyle(ctx)
	if measureResolved.Text != resolved {
		t.Fatalf("measure resolved text pointer = %p, want %p", measureResolved.Text, resolved)
	}
	staticResolved := ResolveStatic(ctx, value)
	if staticResolved.Text != resolved {
		t.Fatalf("static resolved text pointer = %p, want %p", staticResolved.Text, resolved)
	}
	label := value.labelStyle(ctx, layoutResolved)
	if label.TextSize != 21 || label.Font.Weight != font.Bold {
		t.Fatalf("label from shared declaration = size %v weight %v", label.TextSize, label.Font.Weight)
	}
}

func TestTextStyleCascadesScopeBeforeInstance(t *testing.T) {
	ctx := textTestContext()
	restore := core.PushStyle(ctx, style.Style{}.FontSize(18).TextColor(style.RGB(0x112233)))
	defer restore()
	value := New("FlowUI").
		Weight(font.Bold).
		Style(style.Style{}.
			FontSize(20).
			FontWeight(int(font.Normal)).
			Typeface("monospace").
			FontStyle(font.Italic).
			Wrap(giotext.WrapGraphemes).
			Truncator("...").
			LineHeightScale(1.25),
		)
	label := value.labelStyle(ctx, value.resolveStyleStatic(ctx, style.State{}))
	if label.TextSize != 20 || label.Color != style.RGB(0x112233).Color {
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
	from := style.RGB(0x102030).Color
	to := style.RGB(0x8090a0).Color
	resolveAt := func(now time.Time, active bool) color.NRGBA {
		textStyle := style.Style{}.TextColor(style.SolidColor{Color: from}).
			Transition(style.PropTextColor, 100*time.Millisecond)
		if active {
			textStyle = textStyle.TextColor(style.SolidColor{Color: to})
		}
		core.BeginFrame(ctx)
		resolved := New("Status").Style(textStyle).
			resolveLayoutStyle(ctx, layout.Context{Ops: new(op.Ops), Now: now})
		core.EndFrame(ctx)
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

	stateValue, ok := core.PeekState[selectableTextState](ctx, "message", stateSlotSelectableText)
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

func TestTextSelectionColorOverrideAndTextColor(t *testing.T) {
	ctx := textTestContext()
	customSelection := color.NRGBA{R: 0xff, G: 0x88, B: 0x00, A: 0xff}
	customText := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}

	selectable := Selectable("styled-selection", "FlowUI").
		SelectionColor(customSelection).
		Color(customText)
	label := selectable.labelStyle(ctx, selectable.resolveStyleStatic(ctx, style.State{}))
	if label.SelectionColor != customSelection {
		t.Fatalf("selection color = %#v, want %#v", label.SelectionColor, customSelection)
	}
	if label.Color != customText {
		t.Fatalf("text color = %#v, want %#v", label.Color, customText)
	}

	// Without an override both colours follow the active theme.
	plain := New("FlowUI")
	defaultLabel := plain.labelStyle(ctx, plain.resolveStyleStatic(ctx, style.State{}))
	if defaultLabel.SelectionColor != testtheme.Behavior().Semantic().Selection {
		t.Fatalf("default selection color = %#v", defaultLabel.SelectionColor)
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

	stateValue, ok := core.PeekState[selectableTextState](ctx, "pointer-selection", stateSlotSelectableText)
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
	stateValue, ok := core.PeekState[selectableTextState](ctx, "summary", stateSlotSelectableText)
	if !ok || !stateValue.selectable.Truncated() {
		t.Fatal("single-line SelectableText was not truncated")
	}
}

func textTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
}

func layoutTextFrame(ctx *core.Context, router *input.Router, value Widget, now time.Time, maximum image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	dims := value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
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

func TestTextStyleJoinsSeveralDeclarations(t *testing.T) {
	ctx := textTestContext()
	value := New("FlowUI").Style(
		style.Style{}.FontSize(21),
		style.Style{}.TextColor(style.RGB(0x123456)),
		style.Style{}.FontSize(23),
	)
	resolved := value.resolveLayoutStyle(ctx, layout.Context{Ops: new(op.Ops)})
	if resolved.Text == nil {
		t.Fatal("text declaration is nil")
	}
	if resolved.Text.FontSize == nil || *resolved.Text.FontSize != 23 {
		t.Fatalf("joined font size = %v, want the last declaration (23)", resolved.Text.FontSize)
	}
	if resolved.Text.Color == nil {
		t.Fatal("joined text colour is unset")
	}
	solid, ok := resolved.Text.Color.(style.SolidColor)
	if !ok || solid.Color != (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}) {
		t.Fatalf("joined text colour = %v, want 0x123456", resolved.Text.Color)
	}
}
