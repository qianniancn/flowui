package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// CardRequest describes the semantic card variant and optional instance
// declaration needed by a card visual provider.
type CardRequest struct {
	Variant  CardVariant
	Override *CardOverride
}

// CardOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type CardOverride struct {
	Declaration style.Style
	// Padding and Gap are optional instance geometry overrides. They travel
	// through the visual wrapper so the behavior widget does not retain
	// visual spacing state.
	Padding *unit.Dp
	Gap     *unit.Dp
}

// CardOverrideStyle extracts the provider-owned instance layer.
func CardOverrideStyle(request CardRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CardMetrics contains theme-owned layout values used by card behavior.
// Card padding is carried by the resolved root declaration; only the inter-
// child gap is a neutral layout fact that card behavior must read separately.
type CardMetrics struct {
	Gap unit.Dp
}

// CardVisual is the complete visual plan consumed by card behavior.
type CardVisual struct {
	Defaults style.Style
	Variant  style.Style
	Override style.Style
	Metrics  CardMetrics
}

// CardProvider supplies card declarations and layout metrics without taking
// ownership of child association, ordering, or overlay behavior.
type CardProvider interface {
	Card(active *Theme, request CardRequest) CardVisual
}

// ValidateCardVisual checks the declaration and layout metrics before card
// behavior stores a provider result in its cache.
func ValidateCardVisual(visual CardVisual) {
	if visual.Metrics.Gap < 0 {
		invalidVisual("Card", "Metrics.Gap", "must be at least 0")
	}

	resolved := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Override)
	if resolved.Paint == nil {
		invalidVisual("Card", "Resolved.Paint", "must be provided")
	}
	if resolved.Paint.Radius == nil {
		invalidVisual("Card", "Resolved.Paint.Radius", "must be provided")
	}
	if *resolved.Paint.Radius < 0 {
		invalidVisual("Card", "Resolved.Paint.Radius", "must be at least 0")
	}
	if resolved.Paint.Border == nil {
		invalidVisual("Card", "Resolved.Paint.Border", "must be provided")
	}
	if resolved.Paint.Border.Width == nil {
		invalidVisual("Card", "Resolved.Paint.Border.Width", "must be provided")
	}
	if !hasColor(resolved.Paint.Border.Color) {
		invalidVisual("Card", "Resolved.Paint.Border.Color", "must resolve to a solid color")
	}
	if resolved.Paint.Opacity == nil {
		invalidVisual("Card", "Resolved.Paint.Opacity", "must be provided")
	}
	if resolved.Box == nil {
		invalidVisual("Card", "Resolved.Box", "must be provided")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("Card", "Resolved.Box.Padding", "must be provided")
	}
	if resolved.Text == nil {
		invalidVisual("Card", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Card", "Resolved.Text.Color", "must resolve to a solid color")
	}
}
