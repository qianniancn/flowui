package ui

import "github.com/qianniancn/flowui/internal/components/tree"

type TreeWidget = tree.Widget
type TreeItem = tree.Item
type TreeVariant = tree.Variant
type TreeSize = tree.Size
type TreeGuideStyle = tree.GuideStyle
type TreeDropPosition = tree.DropPosition
type TreeDropEvent = tree.DropEvent
type TreeSelectionMode = tree.SelectionMode
type TreeChildrenState = tree.ChildrenState

const (
	TreeDefault     = tree.VariantDefault
	TreeSurface     = tree.VariantSurface
	TreeMedium      = tree.SizeMedium
	TreeSmall       = tree.SizeSmall
	TreeGuideSolid  = tree.GuideSolid
	TreeGuideDashed = tree.GuideDashed
	TreeDropBefore  = tree.DropBefore
	TreeDropInside  = tree.DropInside
	TreeDropAfter   = tree.DropAfter

	TreeSelectionSingle   = tree.SelectionSingle
	TreeSelectionMultiple = tree.SelectionMultiple
	TreeSelectionNone     = tree.SelectionNone

	TreeChildrenLoaded   = tree.ChildrenLoaded
	TreeChildrenUnloaded = tree.ChildrenUnloaded
	TreeChildrenLoading  = tree.ChildrenLoading
	TreeChildrenError    = tree.ChildrenError
)

func Tree(key, selectedKey string, items []TreeItem) TreeWidget {
	return tree.New(key, selectedKey, items)
}
