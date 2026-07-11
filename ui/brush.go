package ui

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/render"
)

type Brush = render.Brush
type GradientStop = render.GradientStop

func SolidBrush(col color.NRGBA) Brush {
	return render.SolidBrush(col)
}

func LinearGradient(stops ...GradientStop) Brush {
	return render.LinearGradient(stops...)
}

func ColorStop(offset float32, col color.NRGBA) GradientStop {
	return GradientStop{Offset: offset, Color: col}
}
