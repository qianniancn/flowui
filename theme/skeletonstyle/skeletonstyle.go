// Package skeletonstyle contains the theme-owned placeholder recipe. The
// skeleton widget keeps sizing, layout, and pulse animation.
package skeletonstyle

import (
	"github.com/qianniancn/flowui/theme"
)

// BuildSkeleton returns the provider-owned placeholder shape and pulse range.
func BuildSkeleton(active *theme.Theme, request theme.SkeletonRequest) theme.SkeletonVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := active.Semantic()
	tokens := active.Components.Skeleton
	return theme.SkeletonVisual{
		Metrics: theme.SkeletonMetrics(tokens),
		Colors: theme.SkeletonColors{
			Base:      sem.Muted,
			Highlight: sem.Border,
		},
		Motion:   active.Motion,
		Override: theme.SkeletonOverrideStyle(request),
	}
}
