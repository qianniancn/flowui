// Package segmentedstyle contains the theme-owned track recipe for segmented
// controls. The widget keeps selection and keyboard behavior.
package segmentedstyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSegmented returns the provider-owned track surface and geometry.
func BuildSegmented(active *theme.Theme, request theme.SegmentedRequest) theme.SegmentedVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Segmented
	return theme.SegmentedVisual{
		Root: style.New().
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			Background(style.Color(active.Semantic().Muted)),
		Metrics:  theme.SegmentedMetrics(tokens),
		Override: theme.SegmentedOverrideStyle(request),
	}
}
