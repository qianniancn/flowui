// Package cardstyle contains the theme-owned visual recipe for semantic
// content cards. Card behavior only applies the returned declaration and
// metrics while retaining child ordering and field association behavior.
package cardstyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/surfacestyle"
)

// BuildCard builds the provider-owned card visual plan.
func BuildCard(active *theme.Theme, request theme.CardRequest) theme.CardVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	surfaceVariant := cardSurfaceVariant(request.Variant)
	surfaceVisual := surfacestyle.BuildSurface(active, theme.SurfaceRequest{Variant: surfaceVariant})
	padding := active.Components.Card.Padding
	gap := active.Components.Card.Gap
	if request.Override != nil {
		if request.Override.Padding != nil {
			padding = max(*request.Override.Padding, 0)
		}
		if request.Override.Gap != nil {
			gap = max(*request.Override.Gap, 0)
		}
	}
	declaration := style.Join(
		surfaceVisual.Defaults,
		style.Style{}.Radius(active.Components.Card.Radius).Padding(padding),
	)
	if request.Variant != theme.CardTransparent {
		declaration = declaration.Shadow(style.ShadowSurface)
	}
	return theme.CardVisual{
		Defaults: declaration,
		Variant:  surfaceVisual.Variant,
		Override: theme.CardOverrideStyle(request),
		Metrics: theme.CardMetrics{
			Gap: gap,
		},
	}
}

func cardSurfaceVariant(variant theme.CardVariant) theme.SurfaceVariant {
	switch variant {
	case theme.CardSecondary:
		return theme.SurfaceSecondary
	case theme.CardTertiary:
		return theme.SurfaceTertiary
	case theme.CardTransparent:
		return theme.SurfaceTransparent
	default:
		return theme.SurfaceDefault
	}
}
