package testtheme

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) StatusBar(active *theme.Theme, request theme.StatusBarRequest) theme.StatusBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := active.Semantic()
	return theme.StatusBarVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(active.Components.Surface.BorderWidth).
			Opacity(1).
			Height(28).
			PaddingX(10).
			BorderBottom(1, style.Solid(sem.Separator)),
		Variant: style.Style{}.
			Background(style.Solid(sem.SurfaceSecondary)).
			TextColor(style.Solid(sem.SurfaceSecondaryForeground)),
		Override: theme.StatusBarOverrideStyle(request),
		Metrics: theme.StatusBarMetrics{
			Height: 28,
			Gap:    8,
		},
	}
}
