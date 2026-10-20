// Package statusbarstyle contains the theme-owned visual recipe for the
// compact application status bar. The widget owns slot order and width
// constraints; this package owns chrome geometry and separator treatment.
package statusbarstyle

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/surfacestyle"
)

// Config contains the visual geometry for one status-bar style family.
type Config struct {
	Height      unit.Dp
	Padding     unit.Dp
	Gap         unit.Dp
	BorderWidth unit.Dp
}

// DefaultConfig is the neutral chrome geometry used by shipped style
// providers until a style family chooses a distinct status-bar treatment.
func DefaultConfig() Config {
	return Config{Height: 28, Padding: 10, Gap: 8, BorderWidth: 1}
}

// BuildStatusBar builds a complete provider-owned status-bar visual plan.
func BuildStatusBar(active *theme.Theme, request theme.StatusBarRequest, config Config) theme.StatusBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	surface := surfacestyle.BuildSurface(active, theme.SurfaceRequest{Variant: theme.SurfaceSecondary})
	defaults := style.Join(
		surface.Defaults,
		style.Style{}.
			Height(config.Height).
			PaddingX(config.Padding).
			BorderBottom(config.BorderWidth, style.Solid(active.Semantic().Separator)),
	)
	return theme.StatusBarVisual{
		Defaults: defaults,
		Variant:  surface.Variant,
		Override: theme.StatusBarOverrideStyle(request),
		Metrics: theme.StatusBarMetrics{
			Height: config.Height,
			Gap:    config.Gap,
		},
	}
}
