package surface

import (
	"image/color"

	"gioui.org/unit"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type surfaceStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	border      color.NRGBA
	borderWidth unit.Dp
}

func surfaceDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	return flowstyle.Style{}.
		BorderColor(flowstyle.SolidColor{Color: activeTheme.Palette.Border}).
		BorderWidth(activeTheme.Components.Surface.BorderWidth).
		Opacity(1)

}

func surfaceVariantDeclaration(activeTheme *theme.Theme, variant SurfaceVariant) flowstyle.Style {
	resolved := surfaceStyleFor(activeTheme, variant)
	builder := flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: resolved.foreground})
	if variant != SurfaceTransparent {
		builder = builder.Background(flowstyle.SolidColor{Color: resolved.background})
	}
	return builder
}

func surfaceStyleFor(activeTheme *theme.Theme, variant SurfaceVariant) surfaceStyle {
	style := surfaceStyle{
		border:      activeTheme.Palette.Border,
		borderWidth: activeTheme.Components.Surface.BorderWidth,
	}
	switch variant {
	case SurfaceSecondary:
		style.background = activeTheme.Palette.SurfaceSecondary
		style.foreground = activeTheme.Palette.SurfaceSecondaryForeground
	case SurfaceTertiary:
		style.background = activeTheme.Palette.SurfaceTertiary
		style.foreground = activeTheme.Palette.SurfaceTertiaryForeground
	case SurfaceTransparent:
		style.foreground = activeTheme.Palette.Foreground
	default:
		style.background = activeTheme.Palette.Surface
		style.foreground = activeTheme.Palette.SurfaceForeground
	}
	return style
}
