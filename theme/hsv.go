package theme

import (
	"image/color"
	"math"
)

// hsv is the color space the ramp algorithm uses: h in degrees, s and v in 0..1.
// Working in this space keeps generated ramps on stable hex values rather than
// close approximations.
type hsv struct {
	h, s, v float64
}

func toHSV(c color.NRGBA) hsv {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s float64
	if max != 0 {
		s = delta / max
	}
	if delta != 0 {
		switch max {
		case r:
			h = (g - b) / delta
		case g:
			h = (b-r)/delta + 2
		default:
			h = (r-g)/delta + 4
		}
		h *= 60
		if h < 0 {
			h += 360
		}
	}
	return hsv{h: h, s: s, v: max}
}

// nrgba converts back. The result is always opaque: ramps carry no alpha.
func (x hsv) nrgba() color.NRGBA {
	h := math.Mod(x.h, 360)
	if h < 0 {
		h += 360
	}
	sector := min(int(math.Floor(h/60)), 5)
	f := h/60 - math.Floor(h/60)

	vs := [4]float64{x.v, x.v * (1 - x.s), x.v * (1 - f*x.s), x.v * (1 - (1-f)*x.s)}
	perm := [6][3]int{{0, 3, 1}, {2, 0, 1}, {1, 0, 3}, {1, 2, 0}, {3, 1, 0}, {0, 1, 2}}[sector]
	return color.NRGBA{
		R: clamp8(vs[perm[0]] * 255),
		G: clamp8(vs[perm[1]] * 255),
		B: clamp8(vs[perm[2]] * 255),
		A: 0xff,
	}
}

func clamp8(x float64) uint8 {
	n := math.Round(x)
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}
