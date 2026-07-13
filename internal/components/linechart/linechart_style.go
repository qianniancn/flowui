package linechart

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type chartStyle struct {
	axis      color.NRGBA
	axisLabel color.NRGBA
	grid      color.NRGBA
	crosshair color.NRGBA
	opacity   float32
}

func lineChartStyleFor(activeTheme *theme.Theme, disabled bool) chartStyle {
	grid := activeTheme.Palette.Border
	grid.A = byte(float32(grid.A) * 0.8)
	crosshair := activeTheme.Palette.MutedForeground
	crosshair.A = byte(float32(crosshair.A) * 0.75)
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return chartStyle{
		axis:      activeTheme.Palette.Border,
		axisLabel: activeTheme.Palette.MutedForeground,
		grid:      grid,
		crosshair: crosshair,
		opacity:   opacity,
	}
}
