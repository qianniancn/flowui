package style

import (
	"image/color"
	"math"
	"time"
)

// RGB converts a packed 0xRRGGBB value to an opaque color.
func RGB(value uint32) SolidColor {
	return SolidColor{Color: color.NRGBA{
		R: uint8(value >> 16),
		G: uint8(value >> 8),
		B: uint8(value),
		A: 0xff,
	}}
}

// RGBA converts a packed 0xRRGGBBAA value to a color.
func RGBA(value uint32) SolidColor {
	return SolidColor{Color: color.NRGBA{
		R: uint8(value >> 24),
		G: uint8(value >> 16),
		B: uint8(value >> 8),
		A: uint8(value),
	}}
}

// Color converts any standard library color to non-premultiplied RGBA.
func Color(value color.Color) SolidColor {
	if value == nil {
		return SolidColor{}
	}
	return SolidColor{Color: color.NRGBAModel.Convert(value).(color.NRGBA)}
}

// WithAlpha replaces the resolved color's alpha channel. Alpha is clamped to
// [0, 1], so theme tokens remain reusable for translucent paint.
func WithAlpha(value ColorSource, alpha float32) AlphaColor {
	return AlphaColor{Source: value, Alpha: uint8(clamp01(alpha)*255 + .5)}
}

// LinearGradient constructs a downward gradient. Angle uses CSS degrees:
// zero points upward and 90 points right.
func LinearGradient(stops ...StyleGradientStop) StyleGradient {
	return StyleGradient{AngleDegrees: 180, Stops: append([]StyleGradientStop(nil), stops...)}
}

func ColorStop(offset float32, value ColorSource) StyleGradientStop {
	return StyleGradientStop{Offset: offset, Color: value}
}

func (g StyleGradient) Angle(degrees float32) StyleGradient {
	if finite(degrees) {
		g.AngleDegrees = degrees
	}
	return g
}

func TransitionDelay(value time.Duration) TransitionOption {
	return func(transition *Transition) {
		transition.Delay = max(value, 0)
	}
}

func TransitionEase(value EaseFunc) TransitionOption {
	return func(transition *Transition) {
		transition.Ease = value
	}
}

func clamp01(value float32) float32 {
	if math.IsNaN(float64(value)) {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func finite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
