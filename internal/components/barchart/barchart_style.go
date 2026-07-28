package barchart

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/theme"
)

type chartStyle struct {
	axis          color.NRGBA
	axisLabel     color.NRGBA
	grid          color.NRGBA
	categoryHover color.NRGBA
	barBackground color.NRGBA
	opacity       float32
}

func barChartStyleFor(activeTheme *theme.Theme, disabled bool) chartStyle {
	grid := activeTheme.Palette.Border
	grid.A = byte(float32(grid.A) * 0.8)
	categoryHover := activeTheme.Palette.Accent
	categoryHover.A = 18
	barBackground := activeTheme.Palette.MutedForeground
	barBackground.A = 28
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return chartStyle{
		axis:          activeTheme.Palette.Border,
		axisLabel:     activeTheme.Palette.MutedForeground,
		grid:          grid,
		categoryHover: categoryHover,
		barBackground: barBackground,
		opacity:       opacity,
	}
}
