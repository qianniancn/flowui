package theme

import (
	"github.com/qianniancn/flowui/style"
)

// BadgeRequest describes the semantic axes and surrounding surface needed by
// a badge visual provider. The widget owns placement and content behavior;
// the provider owns declarations for the badge root and label parts.
type BadgeRequest struct {
	Color    BadgeColor
	Variant  BadgeVariant
	Size     BadgeSize
	Override *BadgeOverride
}

// BadgeOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type BadgeOverride struct {
	Declaration style.Style
}

// BadgeOverrideStyle extracts the provider-owned instance layer.
func BadgeOverrideStyle(request BadgeRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// BadgeMetrics contains neutral geometry facts used by badge placement.
type BadgeMetrics struct {
	PlacementOffsetRatio float32
}

// BadgeVisual is the complete visual plan consumed by badge behavior.
type BadgeVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	Override style.Style
	Metrics  BadgeMetrics
}

// BadgeProvider supplies badge declarations and placement metrics without
// taking ownership of anchor, content, or placement behavior.
type BadgeProvider interface {
	Badge(active *Theme, request BadgeRequest) BadgeVisual
}

// ValidateBadgeVisual checks the declaration and placement contract before a
// widget stores a provider result in its cache.
func ValidateBadgeVisual(visual BadgeVisual) {
	if visual.Metrics.PlacementOffsetRatio < 0 {
		invalidVisual("Badge", "Metrics.PlacementOffsetRatio", "must be at least 0")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if root.Box == nil {
		invalidVisual("Badge", "Root.Box", "must be provided")
	}
	if root.Box.MinWidth == nil {
		invalidVisual("Badge", "Root.Box.MinWidth", "must be provided")
	}
	if root.Box.MinHeight == nil {
		invalidVisual("Badge", "Root.Box.MinHeight", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Badge", "Root.Paint", "must be provided")
	}
	if root.Paint.Background == nil {
		invalidVisual("Badge", "Root.Paint.Background", "must be provided")
	}
	if root.Paint.Border == nil {
		invalidVisual("Badge", "Root.Paint.Border", "must be provided")
	}

	label := style.CascadePart(style.State{}, style.PartLabel, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if label.Text == nil {
		invalidVisual("Badge", "Label.Text", "must be provided")
	}
	if label.Text.FontSize == nil {
		invalidVisual("Badge", "Label.Text.FontSize", "must be provided")
	}
	if *label.Text.FontSize <= 0 {
		invalidVisual("Badge", "Label.Text.FontSize", "must be greater than 0")
	}
	if label.Text.LineHeight == nil {
		invalidVisual("Badge", "Label.Text.LineHeight", "must be provided")
	}
	if *label.Text.LineHeight <= 0 {
		invalidVisual("Badge", "Label.Text.LineHeight", "must be greater than 0")
	}
	if !hasColor(label.Text.Color) {
		invalidVisual("Badge", "Label.Text.Color", "must resolve to a solid color")
	}
}
