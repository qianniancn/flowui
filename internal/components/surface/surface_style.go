package surface

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type surfaceStyle struct {
	background color.NRGBA
	foreground color.NRGBA
}

func surfaceStyleFor(activeTheme *theme.Theme, variant SurfaceVariant) surfaceStyle {
	switch variant {
	case SurfaceSecondary:
		return surfaceStyle{
			background: theme.ColorOr(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.SurfaceRaised),
			foreground: theme.ColorOr(activeTheme.Palette.SurfaceSecondaryForeground, activeTheme.Palette.Foreground),
		}
	case SurfaceTertiary:
		return surfaceStyle{
			background: theme.ColorOr(activeTheme.Palette.SurfaceTertiary, activeTheme.Palette.SurfacePressed),
			foreground: theme.ColorOr(activeTheme.Palette.SurfaceTertiaryForeground, activeTheme.Palette.Foreground),
		}
	case SurfaceTransparent:
		return surfaceStyle{foreground: activeTheme.Palette.Foreground}
	default:
		return surfaceStyle{
			background: activeTheme.Palette.Surface,
			foreground: theme.ColorOr(activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Foreground),
		}
	}
}
