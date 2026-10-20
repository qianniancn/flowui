package theme

import (
	"image/color"
	"math"
)

// Low-level constants for the ten-step ramp algorithm.
const (
	hueStep         = 2
	saturationStep  = 0.16
	saturationStep2 = 0.05
	brightnessStep1 = 0.05
	brightnessStep2 = 0.15
	lightColorCount = 5
	darkColorCount  = 4
)

// Ramp builds the ten-step ramp derived from one seed colour: five lighter
// rungs, the seed itself, four darker rungs. The seed lands on Step500.
func Ramp(seed color.NRGBA) ColorScale {
	base := toHSV(seed)
	var scale ColorScale

	i := 0
	for n := lightColorCount; n > 0; n-- {
		scale.Steps[i] = hsv{
			h: stepHue(base.h, n, true),
			s: stepSaturation(base.h, base.s, n, true),
			v: stepValue(base.v, n, true),
		}.nrgba()
		i++
	}
	scale.Steps[i] = color.NRGBA{R: seed.R, G: seed.G, B: seed.B, A: 0xff}
	i++
	for n := 1; n <= darkColorCount; n++ {
		scale.Steps[i] = hsv{
			h: stepHue(base.h, n, false),
			s: stepSaturation(base.h, base.s, n, false),
			v: stepValue(base.v, n, false),
		}.nrgba()
		i++
	}
	return scale
}

// darkRampBase is the background dark ramps are mixed into when the caller
// gives none. Note it is not the dark canvas (#000): the mix base is a constant
// of the ramp algorithm.
var darkRampBase = color.NRGBA{R: 0x14, G: 0x14, B: 0x14, A: 0xff}

// darkColorMap is the dark table: each dark rung blends one rung of the light
// ramp into darkRampBase. Index is the *light* ramp's, so dark ramps are built
// on top of Ramp rather than replacing it; Amount is percent, the unit Mix
// takes.
var darkColorMap = [10]struct {
	Index  int
	Amount float64
}{
	{7, 15}, {6, 25}, {5, 30}, {5, 45}, {5, 65},
	{5, 85}, {4, 90}, {3, 95}, {2, 97}, {1, 98},
}

// DarkRamp builds the ten-step ramp produced by the dark algorithm. The raw
// result runs dark to light, opposite to Ramp, so it is stored reversed: both
// constructors keep ColorScale's documented light-to-dark ordering and only the
// step picks differ between a light and a dark recipe.
func DarkRamp(seed color.NRGBA) ColorScale {
	light := Ramp(seed)
	var scale ColorScale
	for i, m := range darkColorMap {
		scale.Steps[len(darkColorMap)-1-i] = Mix(darkRampBase, light.Steps[m.Index], m.Amount)
	}
	return scale
}

// stepHue warms cool hues as they lighten and cools warm hues, so a ramp reads
// as one family instead of a flat hue rotation.
func stepHue(h float64, n int, light bool) float64 {
	r := math.Round(h)
	cool := r >= 60 && r <= 240
	var hue float64
	if light == cool {
		hue = h - hueStep*float64(n)
	} else {
		hue = h + hueStep*float64(n)
	}
	if hue >= 360 {
		hue -= 360
	}
	if hue < 0 {
		hue += 360
	}
	return hue
}

// stepSaturation lightens fast and darkens slow. Grey seeds keep their
// saturation; the lightest rung is clamped hard so it reads as a tint.
func stepSaturation(h, s float64, n int, light bool) float64 {
	if h == 0 && s == 0 {
		return s
	}
	var sat float64
	switch {
	case light:
		sat = s - saturationStep*float64(n)
	case n == darkColorCount:
		sat = s + saturationStep
	default:
		sat = s + saturationStep2*float64(n)
	}
	if sat > 1 {
		sat = 1
	}
	if light && n == lightColorCount && sat > 0.1 {
		sat = 0.1
	}
	if sat < 0.06 {
		sat = 0.06
	}
	return math.Round(sat*100) / 100
}

func stepValue(v float64, n int, light bool) float64 {
	value := v - brightnessStep2*float64(n)
	if light {
		value = v + brightnessStep1*float64(n)
	}
	if value > 1 {
		value = 1
	}
	return math.Round(value*100) / 100
}
