package testtheme

import (
	"time"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Image(active *theme.Theme, request theme.ImageRequest) theme.ImageVisual {
	return theme.ImageVisual{Override: theme.ImageOverrideStyle(request)}
}

func (behaviorProvider) Code(active *theme.Theme, request theme.CodeRequest) theme.CodeVisual {
	return theme.CodeVisual{
		Root: style.New().Padding(14).Radius(6).Background(style.Solid(active.CodeColors().Background)),
		Metrics: theme.CodeMetrics{
			Padding: 14, Radius: 6, HeaderGap: 10, LineNumberGap: 12, TextSize: 14,
		},
		Override: theme.CodeOverrideStyle(request),
	}
}

func (behaviorProvider) Skeleton(active *theme.Theme, request theme.SkeletonRequest) theme.SkeletonVisual {
	sem := active.Semantic()
	return theme.SkeletonVisual{
		Metrics:  theme.SkeletonMetrics{Height: 16, Radius: 6, Period: 1600 * time.Millisecond},
		Colors:   theme.SkeletonColors{Base: sem.Muted, Highlight: sem.Border},
		Motion:   active.Motion,
		Override: theme.SkeletonOverrideStyle(request),
	}
}

func (behaviorProvider) Segmented(active *theme.Theme, request theme.SegmentedRequest) theme.SegmentedVisual {
	return theme.SegmentedVisual{
		Root: style.New().Padding(3).Radius(8).Background(style.Color(active.Semantic().Muted)),
		Metrics: theme.SegmentedMetrics{
			Padding: 3, Radius: 8, Gap: 2,
		},
		Override: theme.SegmentedOverrideStyle(request),
	}
}
