package animation

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
)

func LerpFloat(from, to, progress float32) float32 {
	if progress == 0 || from == to {
		return from
	}
	if progress == 1 {
		return to
	}
	value := from*(1-progress) + to*progress
	if math.IsInf(float64(value), 1) {
		return math.MaxFloat32
	}
	if math.IsInf(float64(value), -1) {
		return -math.MaxFloat32
	}
	return value
}

func LerpFloat64(from, to float64, progress float32) float64 {
	if progress == 0 || from == to {
		return from
	}
	if progress == 1 {
		return to
	}
	ratio := float64(progress)
	value := from*(1-ratio) + to*ratio
	if math.IsInf(value, 1) {
		return math.MaxFloat64
	}
	if math.IsInf(value, -1) {
		return -math.MaxFloat64
	}
	return value
}

func LerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	if from.A == 0 {
		from.R, from.G, from.B = to.R, to.G, to.B
	}
	if to.A == 0 {
		to.R, to.G, to.B = from.R, from.G, from.B
	}
	return color.NRGBA{
		R: lerpByte(from.R, to.R, progress),
		G: lerpByte(from.G, to.G, progress),
		B: lerpByte(from.B, to.B, progress),
		A: lerpByte(from.A, to.A, progress),
	}
}

func LerpPoint(from, to f32.Point, progress float32) f32.Point {
	return f32.Pt(
		LerpFloat(from.X, to.X, progress),
		LerpFloat(from.Y, to.Y, progress),
	)
}

func LerpRect(from, to image.Rectangle, progress float32) image.Rectangle {
	return image.Rectangle{
		Min: image.Pt(
			lerpInt(from.Min.X, to.Min.X, progress),
			lerpInt(from.Min.Y, to.Min.Y, progress),
		),
		Max: image.Pt(
			lerpInt(from.Max.X, to.Max.X, progress),
			lerpInt(from.Max.Y, to.Max.Y, progress),
		),
	}
}

func LerpByte(from, to byte, progress float32) byte {
	return lerpByte(from, to, progress)
}

func lerpByte(from, to byte, progress float32) byte {
	value := LerpFloat(float32(from), float32(to), progress)
	return byte(min(max(value+0.5, 0), 255))
}

func lerpInt(from, to int, progress float32) int {
	return int(math.Round(float64(LerpFloat(float32(from), float32(to), progress))))
}
