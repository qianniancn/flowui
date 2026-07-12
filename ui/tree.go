package ui

import "github.com/qianniancn/FlowUI/internal/components/tree"

type TreeWidget = tree.Widget
type TreeItem = tree.Item
type TreeVariant = tree.Variant
type TreeSelectionMode = tree.SelectionMode

const (
	TreeDefault = tree.VariantDefault
	TreeSurface = tree.VariantSurface

	TreeSelectionSingle = tree.SelectionSingle
	TreeSelectionNone   = tree.SelectionNone
)

func Tree(key, selectedKey string, items []TreeItem) TreeWidget {
	return tree.New(key, selectedKey, items)
}
