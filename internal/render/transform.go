package render

import (
	"image"

	"gioui.org/f32"
	"gioui.org/op"
)

// Scale returns a transform that scales around the center of size.
func Scale(size image.Point, scale float32) op.TransformOp {
	return ScaleXY(size, scale, scale)
}

// ScaleXY returns a transform that scales each axis around the center of size.
func ScaleXY(size image.Point, x, y float32) op.TransformOp {
	origin := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	factor := f32.Pt(x, y)
	return op.Affine(f32.AffineId().Scale(origin, factor))
}
