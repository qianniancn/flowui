package render

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

// DpFloat converts dp to pixels without rounding to an integer pixel.
func DpFloat(gtx layout.Context, dp unit.Dp) float32 {
	scale := gtx.Metric.PxPerDp
	if scale == 0 {
		scale = 1
	}
	return float32(dp) * scale
}
