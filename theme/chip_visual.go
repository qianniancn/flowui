package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ChipRequest describes the semantic axes and caller-owned declaration needed
// by a chip visual provider. The widget owns label/content layout and semantic
// behavior; the provider owns declarations and compact spacing metrics.
type ChipRequest struct {
	Color    ChipColor
	Variant  ChipVariant
	Size     ChipSize
	Override *ChipOverride
}

// ChipOverride carries caller customization across the neutral provider
// boundary. It is deliberately provider-owned rather than stored on Widget.
type ChipOverride struct {
	Declaration style.Style
}

// ChipOverrideStyle extracts the provider-owned instance layer.
func ChipOverrideStyle(request ChipRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ChipMetrics contains neutral spacing facts used by chip content layout.
type ChipMetrics struct {
	LabelPaddingX unit.Dp
	ContentGap    unit.Dp
}

// ChipVisual is the complete visual plan consumed by chip behavior.
type ChipVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	Override style.Style
	Metrics  ChipMetrics
}

// ChipProvider supplies chip declarations and content spacing without taking
// ownership of the label or accessory widget behavior.
type ChipProvider interface {
	Chip(active *Theme, request ChipRequest) ChipVisual
}

// ValidateChipVisual checks the declaration and spacing contract before a
// widget stores a provider result in its visual cache. Chip behavior may
// measure and place accessories, but it must not invent their spacing or
// recover missing typography from a local recipe.
func ValidateChipVisual(visual ChipVisual) {
	if !validNonNegativeDp(visual.Metrics.LabelPaddingX) {
		invalidVisual("Chip", "Metrics.LabelPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(visual.Metrics.ContentGap) {
		invalidVisual("Chip", "Metrics.ContentGap", "must be finite and non-negative")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if root.Box == nil {
		invalidVisual("Chip", "Root.Box", "must be provided")
	}
	if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
		invalidVisual("Chip", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
	}
	if !validOptionalPadding(root.Box.Padding) {
		invalidVisual("Chip", "Root.Box.Padding", "padding sides must be finite and non-negative")
	}
	if root.Paint == nil {
		invalidVisual("Chip", "Root.Paint", "must be provided")
	}
	if root.Paint.Radius == nil {
		invalidVisual("Chip", "Root.Paint.Radius", "must be provided")
	}
	if !validNonNegativeDp(*root.Paint.Radius) {
		invalidVisual("Chip", "Root.Paint.Radius", "must be finite and non-negative")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Chip", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}

	label := style.CascadePart(style.State{}, style.PartLabel, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if label.Text == nil {
		invalidVisual("Chip", "Label.Text", "must be provided")
	}
	if label.Text.FontSize == nil {
		invalidVisual("Chip", "Label.Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*label.Text.FontSize) {
		invalidVisual("Chip", "Label.Text.FontSize", "must be finite and greater than 0")
	}
	if label.Text.LineHeight == nil {
		invalidVisual("Chip", "Label.Text.LineHeight", "must be provided")
	}
	if !validPositiveSp(*label.Text.LineHeight) {
		invalidVisual("Chip", "Label.Text.LineHeight", "must be finite and greater than 0")
	}
	if !hasColor(label.Text.Color) {
		invalidVisual("Chip", "Label.Text.Color", "must resolve to a solid color")
	}
}
