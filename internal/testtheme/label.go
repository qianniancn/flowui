package testtheme

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Label(active *theme.Theme, request theme.LabelRequest) theme.LabelVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	danger := active.Semantic().Tone(theme.ToneDanger).Fill
	defaults := style.Style{}.
		FontSize(active.Components.Label.TextSize).
		FontWeight(500).
		TextColor(style.Solid(request.Foreground)).
		Rules(
			style.Rule{When: style.Invalid, Set: style.Style{}.TextColor(style.Solid(danger))},
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
