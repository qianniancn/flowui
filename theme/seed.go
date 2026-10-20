package theme

import (
	"image/color"
	"time"
)

// Seed is the single input a v2 theme is built from; everything else is
// derived. The zero value is valid: a grey, motionless theme.
type Seed struct {
	Primary color.NRGBA // brand hue, any hue angle
	Neutral color.NRGBA // neutral base; decides the cool / warm cast
	Danger  color.NRGBA // an unset hue takes the primary's ramp
	Success color.NRGBA
	Warning color.NRGBA
	Info    color.NRGBA

	Motion MotionSeed
}

// MotionSeed is the motion half of a Seed. The zero value turns animation off.
type MotionSeed struct {
	Enabled  bool
	Duration time.Duration // state-feedback baseline
	Ease     EaseKind
}

// The feedback hues are stated rather than left unset, because an unset hue
// takes the primary's ramp: with no brand hue in the seed, danger, success and
// warning would come out the same grey as everything else. They are one family
// — four hue angles at a shared saturation and value — so the four read as
// siblings rather than as four borrowed brand palettes.
const (
	feedbackSaturation = 0.72
	feedbackValue      = 0.86
)

// DefaultSeed is the framework's own starting point: no brand hue, so nothing
// in it is borrowed from a style package, and motion at the 200ms baseline. A
// style package does not build on it — published styles state their own seeds —
// and an app that wants a brand passes its own Primary instead.
func DefaultSeed() Seed {
	return Seed{
		Danger:  hsv{h: 4, s: feedbackSaturation, v: feedbackValue}.nrgba(),
		Success: hsv{h: 145, s: feedbackSaturation, v: feedbackValue}.nrgba(),
		Warning: hsv{h: 40, s: feedbackSaturation, v: feedbackValue}.nrgba(),
		Info:    hsv{h: 212, s: feedbackSaturation, v: feedbackValue}.nrgba(),

		Motion: MotionSeed{Enabled: true, Duration: 200 * time.Millisecond},
	}
}
