package flowui

import "image/color"

type surfaceStyle struct {
	background color.NRGBA
	foreground color.NRGBA
}

func surfaceStyleFor(theme *Theme, variant SurfaceVariant) surfaceStyle {
	switch variant {
	case SurfaceSecondary:
		return surfaceStyle{
			background: paletteColor(theme.Palette.SurfaceSecondary, theme.Palette.SurfaceRaised),
			foreground: paletteColor(theme.Palette.SurfaceSecondaryForeground, theme.Palette.Foreground),
		}
	case SurfaceTertiary:
		return surfaceStyle{
			background: paletteColor(theme.Palette.SurfaceTertiary, theme.Palette.SurfacePressed),
			foreground: paletteColor(theme.Palette.SurfaceTertiaryForeground, theme.Palette.Foreground),
		}
	case SurfaceTransparent:
		return surfaceStyle{foreground: theme.Palette.Foreground}
	default:
		return surfaceStyle{
			background: theme.Palette.Surface,
			foreground: paletteColor(theme.Palette.SurfaceForeground, theme.Palette.Foreground),
		}
	}
}
