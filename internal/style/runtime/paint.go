package runtime

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/style"
)

func Brush(source style.PaintSource) (render.Brush, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return render.SolidBrush(value.Color), true
	case *style.SolidColor:
		if value != nil {
			return render.SolidBrush(value.Color), true
		}
	case style.StyleGradient:
		return gradientBrush(value), true
	case *style.StyleGradient:
		if value != nil {
			return gradientBrush(*value), true
		}
	}
	return render.Brush{}, false
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

func gradientBrush(value style.StyleGradient) render.Brush {
	stops := make([]render.GradientStop, len(value.Stops))
	for index, stop := range value.Stops {
		col, _ := Color(stop.Color)
		stops[index] = render.GradientStop{Offset: stop.Offset, Color: col}
	}
	return render.LinearGradient(stops...).Angle(value.AngleDegrees)
}
