// Package surfacestyle contains the theme-owned visual recipe for semantic
// content surfaces. Surface behavior only resolves declarations and lays out
// its child under the returned scope.
package surfacestyle

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSurface builds the provider-owned surface visual plan.
func BuildSurface(active *theme.Theme, request theme.SurfaceRequest) theme.SurfaceVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}

	sem := active.Semantic()
	background, foreground := surfaceColors(sem, request.Variant)
	variant := style.Style{}.TextColor(style.Solid(foreground))
	if request.Variant != theme.SurfaceTransparent {
		variant = variant.Background(style.Solid(background))
	}
	return theme.SurfaceVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(active.Components.Surface.BorderWidth).
			Opacity(1),
		Variant:  variant,
		Override: theme.SurfaceOverrideStyle(request),
	}
}

func surfaceColors(sem theme.Semantic, variant theme.SurfaceVariant) (background, foreground color.NRGBA) {
	switch variant {
	case theme.SurfaceSecondary:
		return sem.SurfaceSecondary, sem.SurfaceSecondaryForeground
	case theme.SurfaceTertiary:
		return sem.SurfaceTertiary, sem.SurfaceTertiaryForeground
	case theme.SurfaceTransparent:
		return color.NRGBA{}, sem.Foreground
	default:
		return sem.Surface, sem.SurfaceForeground
	}
}
