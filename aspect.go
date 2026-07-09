package flowui

import (
	"image"
	"math"

	"gioui.org/layout"
)

type AspectRatioWidget struct {
	ratio float32
	child Widget
}

func AspectRatio(ratio float32, child Widget) AspectRatioWidget {
	if ratio <= 0 {
		panic("flowui: aspect ratio must be positive")
	}
	return AspectRatioWidget{
		ratio: ratio,
		child: child,
	}
}

func (a AspectRatioWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	size := aspectRatioSize(gtx.Constraints, a.ratio)
	gtx.Constraints = layout.Exact(size)
	dims := a.child.Layout(ctx, gtx)
	dims.Size = size
	return dims
}

func aspectRatioSize(constraints layout.Constraints, ratio float32) image.Point {
	width := constraints.Max.X
	height := ratioHeight(width, ratio)
	if height > constraints.Max.Y {
		height = constraints.Max.Y
		width = ratioWidth(height, ratio)
	}
	if width < constraints.Min.X {
		width = constraints.Min.X
		height = ratioHeight(width, ratio)
	}
	if height < constraints.Min.Y {
		height = constraints.Min.Y
		width = ratioWidth(height, ratio)
	}
	return constraints.Constrain(image.Pt(width, height))
}

func ratioWidth(height int, ratio float32) int {
	return int(math.Round(float64(float32(height) * ratio)))
}

func ratioHeight(width int, ratio float32) int {
	return int(math.Round(float64(float32(width) / ratio)))
}
