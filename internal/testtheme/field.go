package testtheme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func behaviorColor(value style.ColorSource) (color.NRGBA, bool) {
	switch value := value.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}

func behaviorPaintColor(value style.PaintSource) (color.NRGBA, bool) {
	switch value := value.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	case style.Gradient:
		if len(value.Stops) > 0 {
			return behaviorColor(value.Stops[len(value.Stops)/2].Color)
		}
	case *style.Gradient:
		if value != nil && len(value.Stops) > 0 {
			return behaviorColor(value.Stops[len(value.Stops)/2].Color)
		}
	}
	return color.NRGBA{}, false
}

func behaviorResolvedText(value style.ResolvedStyle) color.NRGBA {
	if value.Text == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	result, ok := behaviorColor(value.Text.Color)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return result
}

func behaviorResolvedPaint(value style.ResolvedStyle) color.NRGBA {
	if value.Paint == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	result, ok := behaviorPaintColor(value.Paint.Background)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return result
}

func behaviorPartWidth(value style.ResolvedStyle) unit.Dp {
	if value.Box != nil && value.Box.Width != nil {
		return *value.Box.Width
	}
	return 16
}

func behaviorPartSize(gtx layout.Context, value style.ResolvedStyle) image.Point {
	if value.Box == nil {
		return image.Pt(gtx.Dp(16), gtx.Dp(16))
	}
	width, height := 0, 0
	if value.Box.Width != nil {
		width = max(gtx.Dp(*value.Box.Width), 0)
	}
	if value.Box.Height != nil {
		height = max(gtx.Dp(*value.Box.Height), 0)
	}
	if width == 0 {
		width = height
	}
	if height == 0 {
		height = width
	}
	return image.Pt(width, height)
}
