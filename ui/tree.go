package ui

import "github.com/qianniancn/flowui/internal/components/tree"

type TreeWidget = tree.Widget

// TreeItem describes one node in a tree.
type TreeItem = tree.Item

// TreeVariant selects the tree surface treatment.
type TreeVariant = tree.Variant

// TreeSize selects the tree row height.
type TreeSize = tree.Size

// TreeGuideStyle controls how hierarchy guides are drawn.
type TreeGuideStyle = tree.GuideStyle

// TreeDropPosition identifies where a dragged item will be inserted.
type TreeDropPosition = tree.DropPosition

// TreeDropEvent describes a completed tree drop.
type TreeDropEvent = tree.DropEvent

// TreeSelectionMode controls how many tree items may be selected.
type TreeSelectionMode = tree.SelectionMode

// TreeChildrenState reports the loading state of a node's children.
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

// Tree creates a hierarchical tree with one selected key.
func Tree(key, selectedKey string, items []TreeItem) TreeWidget {
	return tree.New(key, selectedKey, items)
}
