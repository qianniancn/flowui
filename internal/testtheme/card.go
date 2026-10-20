package testtheme

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Card(active *theme.Theme, request theme.CardRequest) theme.CardVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	surfaceVariant := theme.SurfaceDefault
	switch request.Variant {
	case theme.CardSecondary:
		surfaceVariant = theme.SurfaceSecondary
	case theme.CardTertiary:
		surfaceVariant = theme.SurfaceTertiary
	case theme.CardTransparent:
		surfaceVariant = theme.SurfaceTransparent
	}
	sem := active.Semantic()
	background, foreground := sem.Surface, sem.SurfaceForeground
	switch surfaceVariant {
	case theme.SurfaceSecondary:
		background, foreground = sem.SurfaceSecondary, sem.SurfaceSecondaryForeground
	case theme.SurfaceTertiary:
		background, foreground = sem.SurfaceTertiary, sem.SurfaceTertiaryForeground
	case theme.SurfaceTransparent:
		background, foreground = color.NRGBA{}, sem.Foreground
	}
	variant := style.Style{}.TextColor(style.Solid(foreground))
	if surfaceVariant != theme.SurfaceTransparent {
		variant = variant.Background(style.Solid(background))
	}
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
	declaration := style.Style{}.
		BorderColor(style.Solid(sem.Border)).
		BorderWidth(active.Components.Surface.BorderWidth).
		Opacity(1).
		Radius(active.Components.Card.Radius).
		Padding(padding)
	if request.Variant != theme.CardTransparent {
		declaration = declaration.Shadow(style.ShadowSurface)
	}
	return theme.CardVisual{
		Defaults: declaration,
		Variant:  variant,
		Override: theme.CardOverrideStyle(request),
		Metrics: theme.CardMetrics{
			Gap: gap,
		},
	}
}
