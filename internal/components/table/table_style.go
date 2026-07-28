package table

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

type tableStyle struct {
	root            color.NRGBA
	header          color.NRGBA
	body            color.NRGBA
	foreground      color.NRGBA
	muted           color.NRGBA
	columnSeparator color.NRGBA
	headerSeparator color.NRGBA
	rowSeparator    color.NRGBA
	border          color.NRGBA
	focus           color.NRGBA
}

func tableStyleFor(activeTheme *theme.Theme, variant Variant) tableStyle {
	columnSeparator := activeTheme.Palette.SeparatorColor()
	headerSeparator := columnSeparator
	headerSeparator.A = byte(float32(headerSeparator.A)*0.5 + 0.5)
	rowSeparator := render.LerpColor(activeTheme.Palette.Surface, activeTheme.Palette.Foreground, 0.19)
	rowSeparator.A = byte(float32(rowSeparator.A)*0.5 + 0.5)
	headerSurface := activeTheme.Palette.SurfaceTertiary
	style := tableStyle{
		header:          headerSurface,
		foreground:      activeTheme.Palette.Foreground,
		muted:           activeTheme.Palette.MutedForeground,
		columnSeparator: columnSeparator,
		headerSeparator: headerSeparator,
		rowSeparator:    rowSeparator,
		border:          activeTheme.Palette.Border,
		focus:           activeTheme.Palette.Focus,
	}
	if variant == VariantPrimary {
		style.root = headerSurface
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

func tableRowStyleFor(activeTheme *theme.Theme, variant Variant, stripe color.NRGBA, selected, hovered, pressed, disabled bool) tableRowStyle {
	style := tableRowStyle{
		background: stripe,
		foreground: activeTheme.Palette.Foreground,
		focus:      activeTheme.Palette.Focus,
		opacity:    1,
	}
	base := activeTheme.Palette.Background
	hoverBackground := render.LerpColor(base, activeTheme.Palette.DefaultColor(), 0.5)
	selectedBackground := render.LerpColor(base, activeTheme.Palette.Surface, 0.1)
	if variant == VariantPrimary {
		hoverBackground = render.LerpColor(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.Surface, 0.4)
		selectedBackground = render.LerpColor(activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.Surface, 0.1)
	}
	if hovered || pressed {
		style.background = hoverBackground
	}
	if selected {
		style.background = selectedBackground
	}
	if disabled {
		style.focus = color.NRGBA{}
		style.opacity = activeTheme.DisabledOpacityValue()
	}
	return style
}
