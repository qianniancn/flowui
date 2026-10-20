// Package labelstyle contains the theme-owned visual recipe for form labels.
// The label widget remains responsible for field association, focus, click,
// required-state behavior, and text layout.
package labelstyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildLabel builds the complete provider-owned label visual plan.
func BuildLabel(active *theme.Theme, request theme.LabelRequest) theme.LabelVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}

	sem := active.Semantic()
	danger := sem.Tone(theme.ToneDanger).Fill
	defaults := style.Style{}.
		FontSize(active.Components.Label.TextSize).
		FontWeight(500).
		TextColor(style.Solid(request.Foreground)).
		Rules(
			style.Rule{
				When: style.Invalid,
				Set:  style.Style{}.TextColor(style.Solid(danger)),
			},
			style.Rule{
				Rank: 1,
				When: style.Disabled,
				Set: style.Style{}.
					TextColor(style.Solid(active.DisabledColor(request.Foreground))).
					On(style.Invalid, style.Style{}.
						TextColor(style.Solid(active.DisabledColor(danger)))),
			},
		)
	required := style.Style{}.
		TextColor(style.Solid(danger)).
		Rules(style.Rule{
			When: style.Disabled,
			Set:  style.Style{}.TextColor(style.Solid(active.DisabledColor(danger))),
		})
	return theme.LabelVisual{
		Defaults:        defaults,
		Required:        required,
		Override:        theme.LabelOverrideStyle(request),
		RequiredMarkGap: active.Components.Label.RequiredMarkOffset,
	}
}
