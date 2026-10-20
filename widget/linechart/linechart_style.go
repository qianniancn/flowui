package linechart

import (
	"image/color"

	"github.com/qianniancn/flowui/theme"
)

type chartStyle struct {
	axis        color.NRGBA
	axisLabel   color.NRGBA
	grid        color.NRGBA
	crosshair   color.NRGBA
	background  color.NRGBA
	legendHover color.NRGBA
	opacity     float32
}

func lineChartStyleFor(visual theme.LineChartVisual, disabled bool) chartStyle {
	opacity := float32(1)
	if disabled {
		opacity = visual.DisabledOpacity
	}
	return chartStyle{
		axis:        visual.Colors.Axis,
		axisLabel:   visual.Colors.AxisLabel,
		grid:        visual.Colors.Grid,
		crosshair:   visual.Colors.Crosshair,
		background:  visual.Colors.Background,
		legendHover: visual.Colors.LegendHover,
		opacity:     opacity,
	}
}
