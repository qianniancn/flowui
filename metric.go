package flowui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

func dpFloat(gtx layout.Context, dp unit.Dp) float32 {
	scale := gtx.Metric.PxPerDp
	if scale == 0 {
		scale = 1
	}
	return float32(dp) * scale
}
