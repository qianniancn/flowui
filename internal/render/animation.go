package render

import (
	"image/color"
	"math"
	"time"
)

func Progress(elapsed, duration time.Duration) float32 {
	if elapsed <= 0 {
		return 0
	}
	if duration <= 0 || elapsed >= duration {
		return 1
	}
	return float32(elapsed) / float32(duration)
}

func Ease(progress float32) float32 {
	if math.IsNaN(float64(progress)) {
		return 0
	}
	progress = min(max(progress, 0), 1)
	return progress * progress * (3 - 2*progress)
}

func Lerp(from, to, progress float32) float32 {
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

func LerpByte(from, to byte, progress float32) byte {
	value := Lerp(float32(from), float32(to), progress)
	return byte(min(max(value+0.5, 0), 255))
}

func lerpByte(from, to byte, progress float32) byte {
	return LerpByte(from, to, progress)
}
