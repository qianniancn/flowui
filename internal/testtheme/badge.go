package testtheme

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Badge(active *theme.Theme, request theme.BadgeRequest) theme.BadgeVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Badge
	minSize, radius, textSize, lineHeight := tokens.MediumMinSize, tokens.MediumRadius, tokens.MediumTextSize, tokens.MediumLineHeight
	switch request.Size {
	case theme.BadgeSizeSmall:
		minSize, radius, textSize, lineHeight = tokens.SmallMinSize, tokens.SmallRadius, tokens.SmallTextSize, tokens.SmallLineHeight
	case theme.BadgeSizeLarge:
		minSize, radius, textSize, lineHeight = tokens.LargeMinSize, tokens.LargeRadius, tokens.LargeTextSize, tokens.LargeLineHeight
	}
	sem := active.Semantic()
	return theme.BadgeVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(sem.Surface)).
			BorderWidth(tokens.BorderWidth).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium)).MaxLines(1)),
		Variant: style.Style{}.
			Background(style.Solid(sem.SurfaceRaised)).
			TextColor(style.Solid(sem.SurfaceForeground)),
		Size: style.Style{}.
			MinWidth(minSize).
			MinHeight(minSize).
			PaddingX(tokens.LabelPaddingX).
			Radius(radius).
			Part(style.PartLabel, style.Style{}.FontSize(textSize).LineHeight(lineHeight)),
		Override: theme.BadgeOverrideStyle(request),
		Metrics:  theme.BadgeMetrics{PlacementOffsetRatio: tokens.PlacementOffsetRatio},
	}
}
