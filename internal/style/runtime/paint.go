package runtime

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/style"
)

// ResolveColor resolves a color source against the active frame theme.
func ResolveColor(ctx *frame.Context, source style.ColorSource) (color.NRGBA, bool) {
	if !validColorSource(source) {
		return color.NRGBA{}, false
	}
	if ctx == nil {
		return Color(resolveColor(source, nil))
	}
	return Color(resolveColor(source, frame.ActiveTheme(ctx)))
}

// ResolveBrush resolves a paint source against the active frame theme.
func ResolveBrush(ctx *frame.Context, source style.PaintSource) (render.Brush, bool) {
	if source == nil {
		return render.Brush{}, false
	}
	if ctx == nil {
		return Brush(resolvePaint(source, nil))
	}
	return Brush(resolvePaint(source, frame.ActiveTheme(ctx)))
}

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

func validColorSource(source style.ColorSource) bool {
	switch value := source.(type) {
	case style.SolidColor, style.ThemeColor:
		return true
	case *style.SolidColor:
		return value != nil
	case *style.ThemeColor:
		return value != nil
	case style.AlphaColor:
		return validColorSource(value.Source)
	case *style.AlphaColor:
		return value != nil && validColorSource(value.Source)
	default:
		return false
	}
}

func gradientBrush(value style.StyleGradient) render.Brush {
	stops := make([]render.GradientStop, len(value.Stops))
	for index, stop := range value.Stops {
		col, _ := Color(stop.Color)
		stops[index] = render.GradientStop{Offset: stop.Offset, Color: col}
	}
	return render.LinearGradient(stops...).Angle(value.AngleDegrees)
}
