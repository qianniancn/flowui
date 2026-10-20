package theme

import (
	"image/color"
	"math"
)

// Alpha replaces a colour's alpha, matching fast-color's setA.
func Alpha(c color.NRGBA, a float64) color.NRGBA {
	c.A = clamp8(a * 255)
	return c
}

// ContrastRatio returns the WCAG contrast ratio between two opaque colours.
// It is used to decide whether a label stays legible on a coloured fill.
func ContrastRatio(a, b color.NRGBA) float64 {
	x, y := relativeLuminance(a), relativeLuminance(b)
	return (max(x, y) + .05) / (min(x, y) + .05)
}

func relativeLuminance(c color.NRGBA) float64 {
	linear := func(channel uint8) float64 {
		v := float64(channel) / 255
		if v <= .04045 {
			return v / 12.92
		}
		return math.Pow((v+.055)/1.055, 2.4)
	}
	return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B)
}

// Composite lays fg over bg at the given alpha and returns the opaque result.
// Text may be stored with alpha; flattening here keeps the derived value
// independent of whatever surface an app later paints it on.
func Composite(bg, fg color.NRGBA, a float64) color.NRGBA {
	lerp := func(b, f uint8) uint8 {
		return clamp8(float64(b) + (float64(f)-float64(b))*a)
	}
	return color.NRGBA{R: lerp(bg.R, fg.R), G: lerp(bg.G, fg.G), B: lerp(bg.B, fg.B), A: 0xff}
}

// Mix blends b into a; amount is b's weight in percent, matching fast-color's
// mix. Alpha is taken from a.
func Mix(a, b color.NRGBA, amount float64) color.NRGBA {
	p := amount / 100
	lerp := func(x, y uint8) uint8 {
		return clamp8(float64(x) + (float64(y)-float64(x))*p)
	}
	return color.NRGBA{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B), A: a.A}
}
