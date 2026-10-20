// Package imagestyle contains the theme-owned visual recipe for images. The
// image widget keeps loading, scaling, opacity, and accessibility behavior.
package imagestyle

import (
	"github.com/qianniancn/flowui/theme"
)

// BuildImage returns the provider-owned shape for an image.
func BuildImage(active *theme.Theme, request theme.ImageRequest) theme.ImageVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ImageVisual{
		Metrics: theme.ImageMetrics{
			Radius: active.Components.Image.Radius,
		},
		Override: theme.ImageOverrideStyle(request),
	}
}
