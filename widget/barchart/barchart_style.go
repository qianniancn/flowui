package barchart

import (
	"image/color"

	"github.com/qianniancn/flowui/theme"
)

type chartStyle struct {
	axis          color.NRGBA
	axisLabel     color.NRGBA
	grid          color.NRGBA
	categoryHover color.NRGBA
	barBackground color.NRGBA
	legendHover   color.NRGBA
	opacity       float32
}

func barChartStyleFor(visual theme.BarChartVisual, disabled bool) chartStyle {
	opacity := float32(1)
	if disabled {
		opacity = visual.DisabledOpacity
	}
	return chartStyle{
		axis:          visual.Colors.Axis,
		axisLabel:     visual.Colors.AxisLabel,
		grid:          visual.Colors.Grid,
		categoryHover: visual.Colors.CategoryHover,
		barBackground: visual.Colors.BarBackground,
		legendHover:   visual.Colors.LegendHover,
		opacity:       opacity,
	}
}
