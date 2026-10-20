package testtheme

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Toolbar(active *theme.Theme, request theme.ToolbarRequest) theme.ToolbarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Toolbar
	sem := active.Semantic()
	variant := style.Style{}
	if request.Attached {
		variant = style.Style{}.
			Background(style.Solid(sem.Surface)).
			TextColor(style.Solid(sem.SurfaceForeground)).
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			Shadow(style.ShadowOverlay)
	}
	return theme.ToolbarVisual{
		Variant:  variant,
		Override: theme.ToolbarOverrideStyle(request),
		Separator: style.Style{}.
			Width(tokens.SeparatorWidth).
			Height(tokens.SeparatorLength).
			Background(style.Solid(sem.Separator)),
		VerticalSeparator: style.Style{}.
			Width(tokens.SeparatorLength * 2).
			Height(tokens.SeparatorWidth).
			Background(style.Solid(sem.Separator)),
		Metrics: theme.ToolbarMetrics{
			Gap: tokens.Gap,
		},
	}
}
