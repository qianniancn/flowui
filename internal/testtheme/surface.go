package testtheme

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Surface(active *theme.Theme, request theme.SurfaceRequest) theme.SurfaceVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := active.Semantic()
	background, foreground := sem.Surface, sem.SurfaceForeground
	switch request.Variant {
	case theme.SurfaceSecondary:
		background, foreground = sem.SurfaceSecondary, sem.SurfaceSecondaryForeground
	case theme.SurfaceTertiary:
		background, foreground = sem.SurfaceTertiary, sem.SurfaceTertiaryForeground
	case theme.SurfaceTransparent:
		background, foreground = color.NRGBA{}, sem.Foreground
	}
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
