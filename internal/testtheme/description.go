package testtheme

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Description(active *theme.Theme, request theme.DescriptionRequest) theme.DescriptionVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	mutedForeground := active.Semantic().MutedForeground
	return theme.DescriptionVisual{
		Defaults: style.Style{}.
			FontSize(active.Components.Description.TextSize).
			TextColor(style.Solid(mutedForeground)).
			Rules(style.Rule{
				When: style.Disabled,
				Set:  style.Style{}.TextColor(style.Solid(active.DisabledColor(mutedForeground))),
			}),
		Override: theme.DescriptionOverrideStyle(request),
	}
}
