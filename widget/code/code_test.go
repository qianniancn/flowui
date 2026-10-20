package code

import (
	"image"
	"strings"
	"testing"

	"gioui.org/f32"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/highlight"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/uitest"
)

func lineText(line []segment) string {
	var value strings.Builder
	for _, seg := range line {
		value.WriteString(seg.text)
	}
	return value.String()
}

func TestSplitLinesKeepsKindsAcrossNewlines(t *testing.T) {
	tokens := []highlight.Token{
		{Text: "// note ", Kind: highlight.KindComment},
		{Text: "more\nfunc", Kind: highlight.KindComment},
		{Text: " main", Kind: highlight.KindKeyword},
		{Text: "\n", Kind: highlight.KindPlain},
	}
	lines := splitLines(tokens)
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	if got := lineText(lines[0]); got != "// note more" {
		t.Fatalf("line 0 text = %q", got)
	}
	if got := lines[0][0].kind; got != highlight.KindComment {
		t.Fatalf("line 0 kind = %d, want comment", got)
	}
	if got := lines[1][0]; got.text != "func" || got.kind != highlight.KindComment {
		t.Fatalf("line 1 keeps the comment kind across the break: %#v", got)
	}
	if got := lines[1][1]; got.text != " main" || got.kind != highlight.KindKeyword {
		t.Fatalf("line 1 second segment = %#v", got)
	}
	if len(lines[2]) != 0 {
		t.Fatalf("trailing empty line = %#v", lines[2])
	}
}

func TestHitTestMapsPositionToRune(t *testing.T) {
	// Two lines of three glyphs, ten pixels wide each.
	makeRow := func(top int, start int) row {
		line := row{top: top, bottom: top + 20, start: start, stop: start + 3}
		for i := range 3 {
			line.boxes = append(line.boxes, glyphBox{
				rect:  image.Rect(i*10, top, i*10+10, top+20),
				start: start + i,
				stop:  start + i + 1,
			})
		}
		line.width = 30
		return line
	}
	rows := []row{makeRow(0, 0), makeRow(20, 4)}

	for _, test := range []struct {
		name string
		at   f32.Point
		want int
	}{
		{name: "before first glyph", at: f32.Pt(-5, 10), want: 0},
		{name: "first glyph", at: f32.Pt(2, 10), want: 0},
		{name: "second glyph", at: f32.Pt(12, 10), want: 1},
		{name: "glyph midpoint", at: f32.Pt(15, 10), want: 2},
		{name: "past line end", at: f32.Pt(200, 10), want: 3},
		{name: "second line", at: f32.Pt(2, 30), want: 4},
		{name: "below all lines", at: f32.Pt(5, 500), want: 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := hitTest(rows, test.at); got != test.want {
				t.Fatalf("hitTest = %d, want %d", got, test.want)
			}
		})
	}
}

func TestXAtTracksGlyphEdges(t *testing.T) {
	line := row{
		start: 0, stop: 2, width: 20,
		boxes: []glyphBox{
			{rect: image.Rect(0, 0, 10, 20), start: 0, stop: 1},
			{rect: image.Rect(10, 0, 20, 20), start: 1, stop: 2},
		},
	}
	for _, test := range []struct {
		index, want int
	}{
		{0, 0},
		{1, 10},
		{2, 20},
		{9, 20},
	} {
		if got := xAt(line, test.index); got != test.want {
			t.Fatalf("xAt(%d) = %d, want %d", test.index, got, test.want)
		}
	}
}

func TestCodeBlockRenders(t *testing.T) {
	h := uitest.New(image.Pt(600, 400), base.Light())
	source := "func main() {\n\t// 你好\n\tfmt.Println(\"hi\")\n}\n"
	dims := h.Frame(layout.Scroll(New("sample", "go", source)).Vertical())
	if dims.Size.X != 600 || dims.Size.Y <= 0 {
		t.Fatalf("code block dimensions = %v", dims.Size)
	}
}

func TestCodeBlockWithoutLanguageFallsBackToPlain(t *testing.T) {
	h := uitest.New(image.Pt(600, 400), base.Light())
	dims := h.Frame(New("plain", "", "SELECT 1;"))
	if dims.Size.Y <= 0 {
		t.Fatalf("plain block dimensions = %v", dims.Size)
	}
}

func TestCopyButtonReportsResult(t *testing.T) {
	h := uitest.New(image.Pt(420, 240), base.Light())
	root := layout.Box(New("copy", "go", "func main() {}")).
		Style(style.New().FillWidth().Padding(8))
	h.Frame(root)
	h.Frame(root)

	node, ok := uitest.SemanticByLabel(h.Semantics(), "复制代码（go）")
	if !ok {
		node, ok = uitest.SemanticByLabel(h.Semantics(), "复制")
	}
	if !ok {
		t.Fatal("copy button is missing")
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	h.Frame(root)
	h.Frame(root)
	if _, ok := uitest.SemanticByLabel(h.Semantics(), "已复制"); !ok {
		t.Fatal("copy button did not report the copy")
	}
}

func TestDragSelectsSourceRange(t *testing.T) {
	h := uitest.New(image.Pt(420, 240), base.Light())
	root := layout.Box(New("drag", "", "abc\ndef")).
		Style(style.New().FillWidth().Padding(8))
	h.Frame(root)
	h.Frame(root)

	node, ok := uitest.SemanticByLabel(h.Semantics(), "abc\ndef")
	if !ok {
		t.Fatal("code block has no accessible text")
	}
	bounds := node.Desc.Bounds
	// Start inside the first glyph, then drag far past the end of the line.
	from := f32.Pt(float32(bounds.Min.X+1), float32(bounds.Min.Y+6))
	to := f32.Pt(float32(bounds.Max.X+40), float32(bounds.Min.Y+6))
	h.TouchDown(from)
	h.Frame(root)
	h.TouchMove(to)
	h.Frame(root)
	h.TouchUp(to)
	h.Frame(root)

	state, ok := core.PeekState[blockState](h.Context(), core.FullKey(h.Context(), "drag"), stateSlot)
	if !ok {
		t.Fatal("code block state is missing")
	}
	start, stop, selected := state.selection()
	if !selected {
		t.Fatal("drag did not create a selection")
	}
	if start != 0 || stop != 3 {
		t.Fatalf("selection = [%d,%d), want the whole first line [0,3)", start, stop)
	}
}

func TestDragAcrossLinesSelectsBoth(t *testing.T) {
	h := uitest.New(image.Pt(420, 240), base.Light())
	root := layout.Box(New("drag-lines", "", "abc\ndef")).
		Style(style.New().FillWidth().Padding(8))
	h.Frame(root)
	h.Frame(root)

	node, ok := uitest.SemanticByLabel(h.Semantics(), "abc\ndef")
	if !ok {
		t.Fatal("code block has no accessible text")
	}
	bounds := node.Desc.Bounds
	from := f32.Pt(float32(bounds.Min.X+1), float32(bounds.Min.Y+6))
	to := f32.Pt(float32(bounds.Max.X+40), float32(bounds.Max.Y-2))
	h.TouchDown(from)
	h.Frame(root)
	h.TouchMove(to)
	h.Frame(root)
	h.TouchUp(to)
	h.Frame(root)

	state, ok := core.PeekState[blockState](h.Context(), core.FullKey(h.Context(), "drag-lines"), stateSlot)
	if !ok {
		t.Fatal("code block state is missing")
	}
	start, stop, selected := state.selection()
	if !selected {
		t.Fatal("drag did not create a selection")
	}
	if start != 0 || stop != len([]rune("abc\ndef")) {
		t.Fatalf("selection = [%d,%d), want the full source [0,7)", start, stop)
	}
}

func TestLongLinesWrapIntoVisualRows(t *testing.T) {
	h := uitest.New(image.Pt(120, 300), base.Light())
	long := "abcdefghijklmnopqrstuvwxyz0123456789"
	root := layout.Box(New("wrap", "", long)).Style(style.New().FillWidth().Padding(8))
	dims := h.Frame(root)
	if dims.Size.Y <= 2*20 {
		t.Fatalf("block height = %d, want several wrapped rows", dims.Size.Y)
	}
}

func TestWrappedSelectionReachesTheEnd(t *testing.T) {
	h := uitest.New(image.Pt(120, 300), base.Light())
	long := "abcdefghijklmnopqrstuvwxyz0123456789"
	root := layout.Box(New("wrap-drag", "", long)).Style(style.New().FillWidth().Padding(8))
	h.Frame(root)
	h.Frame(root)

	node, ok := uitest.SemanticByLabel(h.Semantics(), long)
	if !ok {
		t.Fatal("code block has no accessible text")
	}
	bounds := node.Desc.Bounds
	h.TouchDown(f32.Pt(float32(bounds.Min.X+1), float32(bounds.Min.Y+2)))
	h.Frame(root)
	h.TouchMove(f32.Pt(float32(bounds.Max.X-1), float32(bounds.Max.Y-1)))
	h.Frame(root)
	h.TouchUp(f32.Pt(float32(bounds.Max.X-1), float32(bounds.Max.Y-1)))
	h.Frame(root)

	state, ok := core.PeekState[blockState](h.Context(), core.FullKey(h.Context(), "wrap-drag"), stateSlot)
	if !ok {
		t.Fatal("code block state is missing")
	}
	start, stop, selected := state.selection()
	if !selected || start != 0 || stop != len([]rune(long)) {
		t.Fatalf("selection = [%d,%d) selected=%v, want the whole wrapped source", start, stop, selected)
	}
}

func TestExpandTabs(t *testing.T) {
	for _, test := range []struct{ name, in, want string }{
		{"none", "a b", "a b"},
		{"leading", "\tfmt", "    fmt"},
		{"multiple", "\tx\n\ty", "    x\n    y"},
		{"midline", "ab\tc", "ab  c"},
		{"mixed", "a\tb\n\t\tc", "a   b\n        c"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := expandTabs(test.in); got != test.want {
				t.Fatalf("expandTabs(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}

func TestLineNumbersDefaultToMultiLine(t *testing.T) {
	tests := []struct {
		name string
		code Widget
		key  string
		want bool
	}{
		{"multi default", New("numbers-multi", "go", "func main() {\n}"), "numbers-multi", true},
		{"single default", New("numbers-single", "go", "func main() {}"), "numbers-single", false},
		{"multi off", New("numbers-off", "go", "func main() {\n}").LineNumbers(false), "numbers-off", false},
		{"single on", New("numbers-on", "go", "func main() {}").LineNumbers(true), "numbers-on", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := uitest.New(image.Pt(600, 300), base.Light())
			h.Frame(test.code)
			state, ok := core.PeekState[blockState](h.Context(), core.FullKey(h.Context(), test.key), stateSlot)
			if !ok {
				t.Fatal("code block state is missing")
			}
			if got := state.measureGutter > 0; got != test.want {
				t.Fatalf("gutter visible = %v, want %v (gutter=%d)", got, test.want, state.measureGutter)
			}
		})
	}
}

func TestWrappedRowsKeepOneLineNumber(t *testing.T) {
	h := uitest.New(image.Pt(180, 500), base.Light())
	source := strings.Repeat("word ", 60)
	root := layout.Box(New("wrapped-numbers", "", source).LineNumbers(true)).
		Style(style.New().FillWidth().Padding(8))
	h.Frame(root)
	state, ok := core.PeekState[blockState](h.Context(), core.FullKey(h.Context(), "wrapped-numbers"), stateSlot)
	if !ok {
		t.Fatal("code block state is missing")
	}
	if state.measureGutter == 0 {
		t.Fatal("wrapped block has no line-number gutter")
	}
	if len(state.measureRows) < 2 {
		t.Fatalf("visual rows = %d, want a wrapped line", len(state.measureRows))
	}
	if !state.measureRows[0].first || state.measureRows[0].line != 1 {
		t.Fatalf("first row = %+v, want line 1, first", state.measureRows[0])
	}
	if state.measureRows[1].first || state.measureRows[1].line != 1 {
		t.Fatalf("continuation row = %+v, want line 1, not first", state.measureRows[1])
	}
}

type countingHighlighter struct {
	calls int
}

func (h *countingHighlighter) Tokenize(_, source string) []highlight.Token {
	h.calls++
	return []highlight.Token{{Text: source, Kind: highlight.KindPlain}}
}

func TestTokenizationIsCachedAcrossFrames(t *testing.T) {
	h := uitest.New(image.Pt(300, 300), base.Light())
	counter := &countingHighlighter{}
	root := New("cached-tokens", "", "one\ntwo\nthree")
	root.highlighter = counter
	h.Frame(root)
	h.Frame(root)
	h.Frame(root)
	if counter.calls != 1 {
		t.Fatalf("tokenize calls = %d, want 1", counter.calls)
	}
}

func TestNewNormalizesNewlines(t *testing.T) {
	got := New("crlf", "", "one\r\ntwo\rthree").source
	if got != "one\ntwo\nthree" {
		t.Fatalf("source = %q, want normalized newlines", got)
	}
}
