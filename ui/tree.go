package ui

import "github.com/qianniancn/flowui/internal/components/tree"

type TreeWidget = tree.Widget

// TreeItem describes one node in a tree.
type TreeItem = tree.Item

// TreeSimpleItem describes one flat node for TreeItemsFromSimple.
type TreeSimpleItem = tree.SimpleItem

// TreeVariant selects the tree surface treatment.
type TreeVariant = tree.Variant

// TreeSize selects the tree row height.
type TreeSize = tree.Size

// TreeGuideStyle controls how hierarchy guides are drawn.
type TreeGuideStyle = tree.GuideStyle

// TreeExpandAction controls whether branch rows expand on single or double click.
type TreeExpandAction = tree.ExpandAction

// TreeDropPosition identifies where a dragged item will be inserted.
type TreeDropPosition = tree.DropPosition

// TreeDropEvent describes a completed tree drop.
type TreeDropEvent = tree.DropEvent

// TreeDragEvent describes a Tree drag operation or valid drop target.
type TreeDragEvent = tree.DragEvent

// TreeLoadEvent describes an asynchronous child-load request.
type TreeLoadEvent = tree.LoadEvent

// TreeSelectionMode controls how many tree items may be selected.
type TreeSelectionMode = tree.SelectionMode

// TreeChildrenState reports the loading state of a node's children.
type TreeChildrenState = tree.ChildrenState

const (
	TreeDefault             = tree.VariantDefault
	TreeSurface             = tree.VariantSurface
	TreeMedium              = tree.SizeMedium
	TreeSmall               = tree.SizeSmall
	TreeGuideSolid          = tree.GuideSolid
	TreeGuideDashed         = tree.GuideDashed
	TreeExpandOnClick       = tree.ExpandActionClick
	TreeExpandOnDoubleClick = tree.ExpandActionDoubleClick
	TreeDropBefore          = tree.DropBefore
	TreeDropInside          = tree.DropInside
	TreeDropAfter           = tree.DropAfter

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

// NewTreeSimpleItem creates a flat Tree item for TreeItemsFromSimple.
func NewTreeSimpleItem(key, parentKey, label string) TreeSimpleItem {
	return tree.NewSimpleItem(key, parentKey, label)
}

// TreeItemsFromSimple builds a hierarchy from flat Tree items. Items whose
// ParentKey is empty or unknown become roots.
func TreeItemsFromSimple(items []TreeSimpleItem) []TreeItem {
	return tree.ItemsFromSimple(items)
}
