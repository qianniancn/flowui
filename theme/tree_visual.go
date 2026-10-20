package theme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// TreeVariant identifies the visual treatment of a tree container.
type TreeVariant uint8

const (
	TreeVariantDefault TreeVariant = iota
	TreeVariantSurface
)

// TreeSize selects a provider-owned density preset.
type TreeSize uint8

const (
	TreeSizeMedium TreeSize = iota
	TreeSizeSmall
)

// TreeGuideStyle identifies the provider-neutral indentation stroke pattern.
type TreeGuideStyle uint8

const (
	TreeGuideSolid TreeGuideStyle = iota
	TreeGuideDashed
)

// TreeDropPosition identifies the position of a drop indicator relative to a row.
type TreeDropPosition uint8

const (
	TreeDropBefore TreeDropPosition = iota
	TreeDropInside
	TreeDropAfter
)

// TreeRequest contains semantic choices made by tree behavior and an optional
// instance declaration. Selection, expansion, and drag state stay in widget.
type TreeRequest struct {
	Variant  TreeVariant
	Size     TreeSize
	Override *TreeOverride
}

// TreeOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type TreeOverride struct {
	Declaration style.Style
}

// TreeOverrideStyle extracts the provider-owned instance layer.
func TreeOverrideStyle(request TreeRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TreeMetrics contains neutral geometry and typography consumed by tree
// behavior layout. It contains no colors or semantic objects.
type TreeMetrics struct {
	Padding              unit.Dp
	Gap                  unit.Dp
	MaxHeight            unit.Dp
	RowHeight            unit.Dp
	DescriptionRowHeight unit.Dp
	RowRadius            unit.Dp
	RowPaddingX          unit.Dp
	RowPaddingY          unit.Dp
	Indent               unit.Dp
	ChevronSlotSize      unit.Dp
	ChevronIconSize      unit.Dp
	ContentGap           unit.Dp
	ItemTextSize         unit.Sp
	ItemDescriptionSize  unit.Sp
	ItemTextWeight       font.Weight
	FocusRingWidth       unit.Dp
	SurfaceRadius        unit.Dp
	DragPreviewOffset    unit.Dp
	DragPreviewMaxWidth  unit.Dp
	DragPreviewPaddingX  unit.Dp
	DragPreviewPaddingY  unit.Dp
	DragPreviewRadius    unit.Dp
}

// TreeColors contains all semantic colors used by tree behavior and provider
// renderers. The widget never derives a fallback color from the active theme.
type TreeColors struct {
	RootBackground        color.NRGBA
	RootForeground        color.NRGBA
	RowHover              color.NRGBA
	RowSelected           color.NRGBA
	RowSelectedHover      color.NRGBA
	Foreground            color.NRGBA
	Description           color.NRGBA
	SelectedForeground    color.NRGBA
	SelectedDescription   color.NRGBA
	Chevron               color.NRGBA
	Focus                 color.NRGBA
	Guides                color.NRGBA
	DropIndicator         color.NRGBA
	Retry                 color.NRGBA
	ToggleBackground      color.NRGBA
	DragPreviewBackground color.NRGBA
	DragPreviewForeground color.NRGBA
	EditorSelection       color.NRGBA
}

// TreeItemRequest contains behavior-owned state used to resolve one row's
// provider visual treatment.
type TreeItemRequest struct {
	Selected bool
	Hovered  bool
	Disabled bool
}

// TreeItemVisual is the provider-resolved state used by row composition.
type TreeItemVisual struct {
	Background  color.NRGBA
	Foreground  color.NRGBA
	Description color.NRGBA
	Chevron     color.NRGBA
	Focus       color.NRGBA
	Opacity     float32
}

type TreeGuideSegment struct {
	From   image.Point
	To     image.Point
	Extend bool
}

type TreeGuidesRequest struct {
	Segments []TreeGuideSegment
	Style    TreeGuideStyle
	Color    color.NRGBA
}

type TreeTextRequest struct {
	Text   string
	Size   unit.Sp
	Color  color.NRGBA
	Weight font.Weight
}

// TreeEditorStyle is the provider-owned bridge to Gio's stateful rename
// editor. Tree behavior owns the editor value and focus lifecycle; the theme
// owns its typography and colors.
type TreeEditorStyle struct {
	TextSize   unit.Sp
	Foreground color.NRGBA
	Selection  color.NRGBA
	FontWeight font.Weight
}

// TreeEditorStyleResolver resolves the provider-owned rename editor style for
// one row state.
type TreeEditorStyleResolver func(TreeItemRequest) TreeEditorStyle

type TreeDragPreviewRequest struct {
	Label string
}

type TreeRootRenderer func(layout.Context, image.Point, unit.Dp, color.NRGBA, bool)

type TreeRowRenderer func(layout.Context, image.Point, TreeItemVisual, float32, unit.Dp)

type TreeDropIndicatorRenderer func(layout.Context, image.Point, TreeDropPosition, int, TreeMetrics, color.NRGBA)

type TreeToggleRenderer func(layout.Context, image.Point, TreeMetrics, float32, bool, color.NRGBA, color.NRGBA)

type TreeRetryRenderer func(layout.Context, image.Point, TreeMetrics, color.NRGBA)

type TreeGuidesRenderer func(layout.Context, TreeGuidesRequest, unit.Dp)

type TreeTextRenderer func(layout.Context, TreeTextRequest) layout.Dimensions

type TreeDragPreviewRenderer func(layout.Context, TreeDragPreviewRequest, TreeMetrics) layout.Dimensions

// TreeVisual is the complete visual plan consumed by tree behavior.
type TreeVisual struct {
	Defaults        style.Style
	Override        style.Style
	Metrics         TreeMetrics
	Colors          TreeColors
	DisabledOpacity float32
	Motion          MotionTheme
	Item            func(TreeItemRequest) TreeItemVisual
	EditorStyle     TreeEditorStyleResolver
	Root            TreeRootRenderer
	Row             TreeRowRenderer
	DropIndicator   TreeDropIndicatorRenderer
	Toggle          TreeToggleRenderer
	Retry           TreeRetryRenderer
	Guides          TreeGuidesRenderer
	Text            TreeTextRenderer
	DragPreview     TreeDragPreviewRenderer
}

// TreeProvider supplies tree metrics, state visuals, colors, and renderers
// without taking ownership of hierarchy, selection, expansion, or drag state.
type TreeProvider interface {
	Tree(active *Theme, request TreeRequest) TreeVisual
}

// ValidateTreeVisual checks the complete tree renderer and metric contract
// before a widget stores a provider result in its visual cache.
func ValidateTreeVisual(visual TreeVisual) {
	m := visual.Metrics
	if !validNonNegativeDp(m.Padding) {
		invalidVisual("Tree", "Metrics.Padding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.Gap) {
		invalidVisual("Tree", "Metrics.Gap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.MaxHeight) {
		invalidVisual("Tree", "Metrics.MaxHeight", "must be finite and non-negative")
	}
	if !validPositiveDp(m.RowHeight) {
		invalidVisual("Tree", "Metrics.RowHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.DescriptionRowHeight) {
		invalidVisual("Tree", "Metrics.DescriptionRowHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.RowRadius) {
		invalidVisual("Tree", "Metrics.RowRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.RowPaddingX) {
		invalidVisual("Tree", "Metrics.RowPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.RowPaddingY) {
		invalidVisual("Tree", "Metrics.RowPaddingY", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.Indent) {
		invalidVisual("Tree", "Metrics.Indent", "must be finite and non-negative")
	}
	if !validPositiveDp(m.ChevronSlotSize) {
		invalidVisual("Tree", "Metrics.ChevronSlotSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ChevronIconSize) {
		invalidVisual("Tree", "Metrics.ChevronIconSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.ContentGap) {
		invalidVisual("Tree", "Metrics.ContentGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.ItemTextSize) {
		invalidVisual("Tree", "Metrics.ItemTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.ItemDescriptionSize) {
		invalidVisual("Tree", "Metrics.ItemDescriptionSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.FocusRingWidth) {
		invalidVisual("Tree", "Metrics.FocusRingWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.SurfaceRadius) {
		invalidVisual("Tree", "Metrics.SurfaceRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.DragPreviewOffset) {
		invalidVisual("Tree", "Metrics.DragPreviewOffset", "must be finite and non-negative")
	}
	if !validPositiveDp(m.DragPreviewMaxWidth) {
		invalidVisual("Tree", "Metrics.DragPreviewMaxWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.DragPreviewPaddingX) {
		invalidVisual("Tree", "Metrics.DragPreviewPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.DragPreviewPaddingY) {
		invalidVisual("Tree", "Metrics.DragPreviewPaddingY", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.DragPreviewRadius) {
		invalidVisual("Tree", "Metrics.DragPreviewRadius", "must be finite and non-negative")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("Tree", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if visual.Item == nil {
		invalidVisual("Tree", "Item", "must be provided")
	}
	if visual.EditorStyle == nil {
		invalidVisual("Tree", "EditorStyle", "must be provided")
	}
	if visual.Root == nil {
		invalidVisual("Tree", "Root", "must be provided")
	}
	if visual.Row == nil {
		invalidVisual("Tree", "Row", "must be provided")
	}
	if visual.DropIndicator == nil {
		invalidVisual("Tree", "DropIndicator", "must be provided")
	}
	if visual.Toggle == nil {
		invalidVisual("Tree", "Toggle", "must be provided")
	}
	if visual.Retry == nil {
		invalidVisual("Tree", "Retry", "must be provided")
	}
	if visual.Guides == nil {
		invalidVisual("Tree", "Guides", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Tree", "Text", "must be provided")
	}
	if visual.DragPreview == nil {
		invalidVisual("Tree", "DragPreview", "must be provided")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Tree", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Tree", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Tree", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}

	for _, request := range []TreeItemRequest{{}, {Hovered: true}, {Selected: true}, {Disabled: true}} {
		item := visual.Item(request)
		if !hasValidOpacityValue(item.Opacity) {
			invalidVisual("Tree", "Item.Opacity", "must be finite and between 0 and 1")
		}

		editor := visual.EditorStyle(request)
		if !validPositiveSp(editor.TextSize) {
			invalidVisual("Tree", "Editor.TextSize", "must be finite and greater than 0")
		}
		if editor.Foreground.A == 0 {
			invalidVisual("Tree", "Editor.Foreground.A", "must be nonzero")
		}
		if editor.Selection.A == 0 {
			invalidVisual("Tree", "Editor.Selection.A", "must be nonzero")
		}
	}
}
