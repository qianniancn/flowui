package main

import (
	"fmt"
	"html"
	"strings"

	"gioui.org/font"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/x/richtext"
	"gioui.org/x/styledtext"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	widgetcode "github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/yuin/goldmark/ast"
	tableast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/util"
)

func (d markdownDocument) blocks(key string, parent ast.Node) ui.Widget {
	var children []ui.Widget
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		children = append(children, d.block(key, node))
	}
	return layout.Column(children...).Gap(12)
}

func (d markdownDocument) block(key string, node ast.Node) ui.Widget {
	switch node := node.(type) {
	case *ast.Heading:
		return text.New(markdownPlain(d.source, node)).Size(16).Weight(font.Bold)
	case *ast.Paragraph, *ast.TextBlock:
		return markdownInline(inlineKey(key, node), markdownSpans(d.source, node), false)
	case *ast.FencedCodeBlock:
		return markdownCode(fmt.Sprintf("%s-code-%d", key, node.Pos()),
			string(node.Language(d.source)), node.Lines().Value(d.source))
	case *ast.CodeBlock:
		return markdownCode(fmt.Sprintf("%s-code-%d", key, node.Pos()),
			"", node.Lines().Value(d.source))
	case *tableast.Table:
		return d.table(key, node)
	case *ast.List:
		var items []ui.Widget
		index := node.Start
		for item := node.FirstChild(); item != nil; item = item.NextSibling() {
			marker := "•"
			if node.IsOrdered() {
				marker = fmt.Sprintf("%d.", index)
			}
			items = append(items, layout.Row(
				text.New(marker).Size(13), layout.Expanded(d.blocks(key, item)),
			).Gap(8))
			index++
		}
		return layout.Column(items...).Gap(12)
	case *ast.Blockquote:
		return layout.Box(d.blocks(key, node)).Style(style.New().PaddingLeft(16))
	case *ast.ThematicBreak:
		return layout.Divider()
	default:
		// Display unsupported block source (such as HTML) as text.
		return text.New(string(node.Lines().Value(d.source))).Size(13)
	}
}

func (d markdownDocument) table(key string, node *tableast.Table) ui.Widget {
	var rows [][]ast.Node
	for row := node.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []ast.Node
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, cell)
		}
		rows = append(rows, cells)
	}
	// The state key must identify this table, not just the document: a document
	// may hold several tables, and reused keys collide within a frame.
	tableKey := fmt.Sprintf("%s-table-%d", key, node.Pos())
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		// A grid needs width; a narrow container turns each row into a block of
		// labelled fields instead, which stays readable at phone widths.
		if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(680)) {
			return d.tableBlocks(tableKey, rows).Layout(ctx, gtx)
		}
		return d.tableFlow(tableKey, rows).Layout(ctx, gtx)
	})
}

// tableFlow renders a wide table with the Table component in flow mode, so the
// grid grows inside the page scroll instead of owning a viewport of its own.
func (d markdownDocument) tableFlow(key string, rows [][]ast.Node) ui.Widget {
	header := rows[0]
	columns := make([]table.Column, len(header))
	for index, cell := range header {
		weight := float32(1)
		if index == len(header)-1 {
			// The description column carries the prose, so it takes the spare
			// width; the first three hold short identifiers.
			weight = 3
		}
		columns[index] = table.Column{
			Key:    fmt.Sprintf("column-%d", index),
			Label:  markdownPlain(d.source, cell),
			Header: markdownInline(fmt.Sprintf("%s-head-%d", key, index), markdownSpans(d.source, cell), true),
			Weight: weight,
		}
	}
	body := make([]table.Row, 0, len(rows))
	for rowIndex, row := range rows[1:] {
		cells := make([]table.Cell, len(columns))
		for index := range columns {
			if index >= len(row) {
				continue
			}
			cells[index] = table.Cell{Content: markdownInline(
				fmt.Sprintf("%s-cell-%d-%d", key, rowIndex, index),
				markdownSpans(d.source, row[index]), false)}
		}
		label := ""
		if len(row) > 0 {
			label = markdownPlain(d.source, row[0])
		}
		body = append(body, table.Row{Key: fmt.Sprintf("row-%d", rowIndex), Label: label, Cells: cells})
	}
	return table.New(key, columns, body).Flow().HeaderHeight(36).RowHeight(36)
}

// tableBlocks is the narrow fallback: the header row becomes per-field labels,
// so a three-column reference table still reads at 480dp.
func (d markdownDocument) tableBlocks(key string, rows [][]ast.Node) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		var children []ui.Widget
		for rowIndex, row := range rows[1:] {
			var fields []ui.Widget
			for index, cell := range row {
				value := markdownInline(fmt.Sprintf("%s-block-%d-%d", key, rowIndex, index),
					markdownSpans(d.source, cell), index == 0)
				if index > 0 && index < len(rows[0]) {
					value = layout.Row(
						text.New(markdownPlain(d.source, rows[0][index])+"：").Size(12),
						layout.Expanded(value),
					).Gap(6)
				}
				fields = append(fields, value)
			}
			if len(children) > 0 {
				children = append(children, layout.Divider())
			}
			children = append(children, layout.Box(layout.Column(fields...).Gap(8)).
				Style(style.New().FillWidth().Padding(10)))
		}
		return layout.Column(children...).Layout(ctx, gtx)
	})
}

// markdownCode renders a code block through the reusable code component, so
// documentation snippets are highlighted and still selectable.
func markdownCode(key, lang string, lines []byte) ui.Widget {
	return widgetcode.New(key, lang, strings.TrimSuffix(string(lines), "\n"))
}

type markdownSpan struct {
	content string
	font    font.Font
	code    bool
}

func markdownSpans(source []byte, node ast.Node) []markdownSpan {
	var spans []markdownSpan
	var visit func(ast.Node, markdownSpan)
	visit = func(node ast.Node, format markdownSpan) {
		appendText := func(value string) {
			format.content = value
			spans = append(spans, format)
		}
		switch node := node.(type) {
		case *ast.Text:
			value := string(node.Value(source))
			if !node.IsRaw() && !format.code {
				value = html.UnescapeString(string(util.UnescapePunctuations([]byte(value))))
			}
			if format.code {
				value = strings.ReplaceAll(value, "\n", " ")
			} else if node.HardLineBreak() {
				value += "\n"
			} else if node.SoftLineBreak() {
				value += " "
			}
			appendText(value)
			return
		case *ast.String:
			appendText(string(node.Value))
			return
		case *ast.CodeSpan:
			format.code = true
		case *ast.Emphasis:
			if node.Level == 2 {
				format.font.Weight = font.Bold
			} else {
				format.font.Style = font.Italic
			}
		case *ast.AutoLink:
			appendText(string(node.Label(source)))
			return
		case *ast.RawHTML:
			appendText(string(node.Segments.Value(source)))
			return
		}
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			visit(child, format)
		}
		if link, ok := node.(*ast.Link); ok {
			appendText(" (" + string(link.Destination) + ")")
		}
	}
	visit(node, markdownSpan{})
	return spans
}

func markdownPlain(source []byte, node ast.Node) string {
	var value strings.Builder
	for _, span := range markdownSpans(source, node) {
		value.WriteString(span.content)
	}
	return value.String()
}

// inlineKey derives the retained state key for one inline text block. Rich text
// keeps pointer and selection state between frames, so the key must stay stable
// for as long as the document is unchanged.
func inlineKey(key string, node ast.Node) string {
	return fmt.Sprintf("%s-inline-%d", key, node.Pos())
}

// markdownInline renders one block of mixed-style text. key identifies the
// retained rich-text state and must be unique and stable while the document is
// unchanged, the same contract every keyed FlowUI widget follows.
func markdownInline(key string, spans []markdownSpan, bold bool) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		active := ctx.Theme()
		state := ui.UseState[richtext.InteractiveText](ctx, key)
		styles := make([]richtext.SpanStyle, 0, len(spans))
		var label strings.Builder
		for _, span := range spans {
			face := span.font
			face.Typeface = active.Typography.Typeface
			if span.code {
				face.Typeface = active.Typography.MonoTypeface
			}
			if bold {
				face.Weight = font.Bold
			}
			styles = append(styles, richtext.SpanStyle{
				Content: span.content, Font: face, Size: 13, Color: active.Semantic().Foreground,
			})
			label.WriteString(span.content)
		}
		// Reuse the window's shaper: Theme() is a detached token snapshot.
		body := richtext.Text(state, ctx.TextShaper(), styles...)
		// WrapWords matches the framework text default: grapheme wrapping splits
		// Latin identifiers mid-word, which is wrong for API tables.
		body.WrapPolicy = styledtext.WrapWords
		body.LineHeightScale = 1.4
		dimensions := body.Layout(gtx)
		// Scope the accessible text to the paragraph box, so assistive tooling
		// and component tests read the text at the bounds it occupies.
		restore := clip.Rect{Max: dimensions.Size}.Push(gtx.Ops)
		semantic.LabelOp(label.String()).Add(gtx.Ops)
		restore.Pop()
		return dimensions
	})
}
