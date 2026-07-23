package slider

import (
	"image"

	"gioui.org/layout"
)

type sliderGeometry struct {
	size       image.Point
	axis       layout.Axis
	edge       int
	inner      int
	centers    [2]image.Point
	thumbRects [2]image.Rectangle
}

func newSliderGeometry(size image.Point, axis layout.Axis, edge int, lower, upper float32, rangeMode bool, thumbSize image.Point) sliderGeometry {
	main := axis.Convert(size).X
	edge = min(max(edge, 0), max(main/2, 0))
	inner := max(main-2*edge, 0)
	thumbSize.X = max(thumbSize.X, 1)
	thumbSize.Y = max(thumbSize.Y, 1)
	geometry := sliderGeometry{size: size, axis: axis, edge: edge, inner: inner}
	ratios := [2]float32{min(max(lower, 0), 1), min(max(upper, 0), 1)}
	if !rangeMode {
		ratios[1] = ratios[0]
	}
	for index, ratio := range ratios {
		mainCenter := edge + int(float32(inner)*ratio+0.5)
		if axis == layout.Vertical {
			mainCenter = main - mainCenter
		}
		center := axis.Convert(image.Pt(mainCenter, axis.Convert(size).Y/2))
		geometry.centers[index] = center
		geometry.thumbRects[index] = image.Rectangle{
			Min: center.Sub(thumbSize.Div(2)),
			Max: center.Sub(thumbSize.Div(2)).Add(thumbSize),
		}
	}
	return geometry
}

func sizeCross(axis layout.Axis, size image.Point) int {
	return axis.Convert(size).Y
}

func sliderFillRect(geometry sliderGeometry, rangeMode bool) image.Rectangle {
	if geometry.axis == layout.Vertical {
		start := geometry.centers[0].Y
		end := geometry.size.Y
		if rangeMode {
			start = geometry.centers[1].Y
			end = geometry.centers[0].Y
		}
		return image.Rect(0, min(start, end), geometry.size.X, max(start, end))
	}
	start := 0
	end := geometry.centers[0].X
	if rangeMode {
		start = geometry.centers[0].X
		end = geometry.centers[1].X
	}
	return image.Rect(min(start, end), 0, max(start, end), geometry.size.Y)
}
