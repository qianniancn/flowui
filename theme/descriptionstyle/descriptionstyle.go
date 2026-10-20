// Package descriptionstyle contains the theme-owned visual recipe for
// supporting text. The description widget remains responsible for text
// content, field association, and layout behavior.
package descriptionstyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildDescription builds the complete provider-owned description visual.
func BuildDescription(active *theme.Theme, request theme.DescriptionRequest) theme.DescriptionVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}

	muted := active.Semantic().MutedForeground
	defaults := style.Style{}.
		FontSize(active.Components.Description.TextSize).
		TextColor(style.Solid(muted)).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.TextColor(style.Solid(active.DisabledColor(muted))),
		})
	return theme.DescriptionVisual{
		Defaults: defaults,
		Override: theme.DescriptionOverrideStyle(request),
	}
}
