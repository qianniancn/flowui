package testtheme

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Pagination(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	tokens := active.Components.Pagination
	metrics := theme.PaginationMetrics{
		Size:    tokens.MediumSize,
		ItemGap: tokens.ItemGap, ContentGap: tokens.ContentGap, CompactWidth: tokens.CompactWidth,
	}
	textSize := tokens.MediumTextSize
	switch request.Size {
	case theme.PaginationSmall:
		metrics.Size, textSize = tokens.SmallSize, tokens.SmallTextSize
	case theme.PaginationLarge:
		metrics.Size, textSize = tokens.LargeSize, tokens.LargeTextSize
	}
	return theme.PaginationVisual{
		Defaults: style.Style{}, Summary: style.Style{}.TextColor(style.Solid(active.Semantic().MutedForeground)).FontSize(textSize), Override: theme.PaginationOverrideStyle(request),
		Metrics: metrics, Motion: active.Motion,
		Button: func(gtx layout.Context, request theme.PaginationButtonRequest) layout.Dimensions {
			height := gtx.Constraints.Max.Y
			if height <= 0 {
				height = gtx.Constraints.Min.Y
			}
			width := gtx.Constraints.Max.X
			if width <= 0 || width > 96 {
				width = 32
				if request.Kind == theme.PaginationPrevious || request.Kind == theme.PaginationNext {
					width = 64
				}
			}
			return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, height))}
		},
	}
}
