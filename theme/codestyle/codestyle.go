// Package codestyle contains the theme-owned surface recipe for code blocks.
// The code widget keeps selection, clipboard, highlighting, and layout.
package codestyle

import (
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildCode returns the provider-owned surface and geometry for a code block.
func BuildCode(active *theme.Theme, request theme.CodeRequest) theme.CodeVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Code
	surface := style.New().
		FillWidth().
		Padding(tokens.Padding).
		Radius(tokens.Radius).
		Background(style.Solid(active.CodeColors().Background))
	return theme.CodeVisual{
		Root:     surface,
		Metrics:  theme.CodeMetrics(tokens),
		Override: theme.CodeOverrideStyle(request),
	}
}
