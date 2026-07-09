package render

import (
	"image/color"
	"time"
)

func Progress(elapsed, duration time.Duration) float32 {
	if elapsed <= 0 {
		return 0
	}
	if elapsed >= duration {
		return 1
	}
	return float32(elapsed) / float32(duration)
}

func Ease(progress float32) float32 {
	return progress * progress * (3 - 2*progress)
}

func Lerp(from, to, progress float32) float32 {
	return from + (to-from)*progress
}

func LerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	if from.A == 0 {
		from.R = to.R
		from.G = to.G
		from.B = to.B
	}
	if to.A == 0 {
		to.R = from.R
		to.G = from.G
		to.B = from.B
	}
	return color.NRGBA{
		R: LerpByte(from.R, to.R, progress),
		G: LerpByte(from.G, to.G, progress),
		B: LerpByte(from.B, to.B, progress),
		A: LerpByte(from.A, to.A, progress),
	}
}

func LerpByte(from, to byte, progress float32) byte {
	return byte(Lerp(float32(from), float32(to), progress) + 0.5)
}
