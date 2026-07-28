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
	thumbSize.X = max(thumbSize.X, 1)
	thumbSize.Y = max(thumbSize.Y, 1)
	edge = min(max(edge, 0), max(main/2, 0), axis.Convert(thumbSize).X/2)
	inner := max(main-2*edge, 0)
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
		if !rangeMode {
			center := geometry.centers[0].Y
			if center >= geometry.size.Y-geometry.edge {
				return image.Rectangle{}
			}
			if center <= geometry.edge {
				return image.Rect(0, 0, geometry.size.X, geometry.size.Y)
			}
			return image.Rect(0, center, geometry.size.X, geometry.size.Y)
		}
		start := geometry.centers[1].Y
		end := geometry.centers[0].Y
		if start <= geometry.edge {
			start = 0
		}
		if end >= geometry.size.Y-geometry.edge {
			end = geometry.size.Y
		}
		return image.Rect(0, min(start, end), geometry.size.X, max(start, end))
	}
	if !rangeMode {
		center := geometry.centers[0].X
		if center <= geometry.edge {
			return image.Rectangle{}
		}
		if center >= geometry.size.X-geometry.edge {
			return image.Rect(0, 0, geometry.size.X, geometry.size.Y)
		}
		return image.Rect(0, 0, center, geometry.size.Y)
	}
	start := geometry.centers[0].X
	end := geometry.centers[1].X
	if start <= geometry.edge {
		start = 0
	}
	if end >= geometry.size.X-geometry.edge {
		end = geometry.size.X
	}
	return image.Rect(min(start, end), 0, max(start, end), geometry.size.Y)
}
