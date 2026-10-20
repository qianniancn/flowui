package computed

import (
	"image/color"

	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
)

func Brush(source style.PaintSource) (draw.Brush, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return draw.SolidBrush(value.Color), true
	case *style.SolidColor:
		if value != nil {
			return draw.SolidBrush(value.Color), true
		}
	case style.Gradient:
		return gradientBrush(value), true
	case *style.Gradient:
		if value != nil {
			return gradientBrush(*value), true
		}
	}
	return draw.Brush{}, false
}

func Color(source style.ColorSource) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

func gradientBrush(value style.Gradient) draw.Brush {
	stops := make([]draw.GradientStop, len(value.Stops))
	for index, stop := range value.Stops {
		col, _ := Color(stop.Color)
		stops[index] = draw.GradientStop{Offset: stop.Offset, Color: col}
	}
	return draw.LinearGradient(stops...).Angle(value.AngleDegrees)
}
