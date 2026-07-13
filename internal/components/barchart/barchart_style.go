package barchart

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type chartStyle struct {
	axis          color.NRGBA
	axisLabel     color.NRGBA
	grid          color.NRGBA
	categoryHover color.NRGBA
	barBackground color.NRGBA
	tooltip       color.NRGBA
	tooltipText   color.NRGBA
	tooltipBorder color.NRGBA
	focus         color.NRGBA
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
		tooltip:       activeTheme.Palette.Overlay,
		tooltipText:   activeTheme.Palette.OverlayForeground,
		tooltipBorder: activeTheme.Palette.Border,
		focus:         activeTheme.Palette.Focus,
		opacity:       opacity,
	}
}
