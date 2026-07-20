package surface

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type surfaceStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	border      color.NRGBA
	borderWidth unit.Dp
}

func surfaceStyleFor(activeTheme *theme.Theme, variant SurfaceVariant) surfaceStyle {
	style := surfaceStyle{
		border:      activeTheme.Palette.Border,
		borderWidth: activeTheme.Components.Surface.BorderWidth,
	}
	switch variant {
	case SurfaceSecondary:
		style.background = theme.ColorOr(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.SurfaceRaised)
		style.foreground = theme.ColorOr(activeTheme.Palette.SurfaceSecondaryForeground, activeTheme.Palette.Foreground)
	case SurfaceTertiary:
		style.background = theme.ColorOr(activeTheme.Palette.SurfaceTertiary, activeTheme.Palette.SurfacePressed)
		style.foreground = theme.ColorOr(activeTheme.Palette.SurfaceTertiaryForeground, activeTheme.Palette.Foreground)
	case SurfaceTransparent:
		style.foreground = activeTheme.Palette.Foreground
	default:
		style.background = activeTheme.Palette.Surface
		style.foreground = theme.ColorOr(activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Foreground)
	}
	return style
}
