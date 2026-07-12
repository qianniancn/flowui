package table

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type tableStyle struct {
	root       color.NRGBA
	header     color.NRGBA
	body       color.NRGBA
	foreground color.NRGBA
	muted      color.NRGBA
	separator  color.NRGBA
	focus      color.NRGBA
}

func tableStyleFor(activeTheme *theme.Theme, variant Variant) tableStyle {
	separator := activeTheme.Palette.Border
	separator.A = byte(float32(separator.A)*0.55 + 0.5)
	style := tableStyle{
		header:     activeTheme.Palette.SurfaceSecondary,
		foreground: activeTheme.Palette.Foreground,
		muted:      activeTheme.Palette.MutedForeground,
		separator:  separator,
		focus:      activeTheme.Palette.Focus,
	}
	if variant == VariantPrimary {
		style.root = activeTheme.Palette.SurfaceSecondary
		style.body = activeTheme.Palette.Surface
	}
	return style
}

type tableRowStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	focus      color.NRGBA
	opacity    float32
}

func tableRowStyleFor(activeTheme *theme.Theme, variant Variant, selected, hovered, pressed, disabled bool) tableRowStyle {
	style := tableRowStyle{
		foreground: activeTheme.Palette.Foreground,
		focus:      activeTheme.Palette.Focus,
		opacity:    1,
	}
	base := activeTheme.Palette.Background
	if variant == VariantPrimary {
		base = activeTheme.Palette.Surface
	}
	if selected {
		style.background = render.LerpColor(base, activeTheme.Palette.SurfaceTertiary, 0.62)
	}
	if hovered {
		style.background = render.LerpColor(base, activeTheme.Palette.SurfaceHover, 0.8)
		if selected {
			style.background = render.LerpColor(base, activeTheme.Palette.SurfaceTertiary, 0.78)
		}
	}
	if pressed {
		style.background = render.LerpColor(base, activeTheme.Palette.SurfacePressed, 0.72)
	}
	if disabled {
		style.focus = color.NRGBA{}
		style.opacity = activeTheme.DisabledOpacityValue()
	}
	return style
}
