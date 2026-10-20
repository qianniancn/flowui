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

// Solid wraps a concrete color for use in declarations. A declaration reads a
// resolved value directly, e.g. Background(style.Solid(sem.Tone(theme.TonePrimary).Fill))
// with sem := t.Semantic(). This is a value constructor -- the color is captured
// now, not looked up later.
func Solid(value color.NRGBA) SolidColor {
	return SolidColor{Color: value}
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

// WithAlpha returns a copy of value with the alpha channel replaced. Alpha is
// clamped to [0, 1], which keeps a palette color reusable for translucent
// paint without wrapping it in another type.
func WithAlpha(value color.NRGBA, alpha float32) SolidColor {
	value.A = uint8(clamp01(alpha)*255 + .5)
	return SolidColor{Color: value}
}

// LinearGradient constructs a downward gradient. Angle uses CSS degrees:
// zero points upward and 90 points right.
func LinearGradient(stops ...GradientStop) Gradient {
	return Gradient{AngleDegrees: 180, Stops: append([]GradientStop(nil), stops...)}
}

func ColorStop(offset float32, value ColorSource) GradientStop {
	return GradientStop{Offset: offset, Color: value}
}

func (g Gradient) Angle(degrees float32) Gradient {
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

// ScaleAlpha multiplies an alpha channel by a factor in [0, 1].
//
// It is exported because the built-in components' disabled colors use the very
// same expression. One definition means Dim(DisabledOpacity) is bit-identical
// to the disabled color a fluent declaration would have produced, which is
// what lets the two be compared with reflect.DeepEqual.
//
// The truncation is deliberate: rounding here would make the two disagree on
// values like Accent's 255 * 0.5.
func ScaleAlpha(alpha uint8, factor float32) uint8 {
	return uint8(float32(alpha) * clamp01(factor))
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
