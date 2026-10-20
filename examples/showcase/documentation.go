package main

import (
	"embed"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	markdowntext "github.com/yuin/goldmark/text"
)

//go:embed docs/*.md
var documentationFS embed.FS

type markdownDocument struct {
	source []byte
	root   ast.Node
}

// Parse embedded documents once. Widgets are built per view so their visual
// caches and selectable state remain local to each window.
var componentDocuments = sync.OnceValue(func() map[string]markdownDocument {
	files, err := fs.Glob(documentationFS, "docs/*.md")
	if err != nil {
		panic(err)
	}
	documents := make(map[string]markdownDocument, len(files))
	for _, file := range files {
		source, err := documentationFS.ReadFile(file)
		if err != nil {
			panic(err)
		}
		key := strings.TrimSuffix(path.Base(file), ".md")
		documents[key] = parseDocumentation(source)
	}
	return documents
})

func parseDocumentation(source []byte) markdownDocument {
	parser := goldmark.New(goldmark.WithExtensions(extension.Table)).Parser()
	return markdownDocument{source: source, root: parser.Parse(markdowntext.NewReader(source))}
}

func componentDocumentation(key string) ui.Widget {
	document, ok := componentDocuments()[key]
	if !ok {
		return nil
	}
	return document.view("docs-" + key)
}

func (d markdownDocument) view(key string) ui.Widget {
	var sections, blocks []ui.Widget
	var title string
	flush := func() {
		if title != "" || len(blocks) > 0 {
			body := layout.Column(blocks...).Gap(16)
			if title != "" {
				sections = append(sections, section(title, body))
			} else {
				sections = append(sections, body)
			}
		}
		blocks = nil
	}
	for node := d.root.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*ast.Heading); ok {
			switch heading.Level {
			case 1:
				// The catalog already provides the page title above the demo.
				continue
			case 2:
				flush()
				title = markdownPlain(d.source, node)
				continue
			}
		}
		blocks = append(blocks, d.block(key, node))
	}
	flush()
	return layout.Column(sections...).Gap(24)
}
