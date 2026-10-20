// Package toolbarstyle contains the theme-owned visual recipe for toolbar
// behavior. The widget owns child order, focus navigation, and semantics;
// this package owns surfaces, spacing, and separator declarations.
package toolbarstyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildToolbar builds a complete provider-owned toolbar visual plan.
func BuildToolbar(active *theme.Theme, request theme.ToolbarRequest) theme.ToolbarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Toolbar
	sem := active.Semantic()
	root := style.Style{}
	variant := style.Style{}
	if request.Attached {
		variant = style.Style{}.
			Background(style.Solid(sem.Surface)).
			TextColor(style.Solid(sem.SurfaceForeground)).
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			Shadow(style.ShadowOverlay)
	}
	separator := style.Style{}.
		Width(tokens.SeparatorWidth).
		Height(tokens.SeparatorLength).
		Background(style.Solid(sem.Separator))
	verticalSeparator := style.Style{}.
		Width(tokens.SeparatorLength * 2).
		Height(tokens.SeparatorWidth).
		Background(style.Solid(sem.Separator))
	return theme.ToolbarVisual{
		Defaults:          root,
		Variant:           variant,
		Override:          theme.ToolbarOverrideStyle(request),
		Separator:         separator,
		VerticalSeparator: verticalSeparator,
		Metrics: theme.ToolbarMetrics{
			Gap: tokens.Gap,
		},
	}
}
