package theme

import "image/color"

// SoftStyle picks how a tone's low-emphasis background is produced.
type SoftStyle uint8

const (
	// SoftRamp takes a rung of the tone's own ramp.
	SoftRamp SoftStyle = iota
	// SoftAlpha keeps the fill colour and reduces its alpha instead.
	SoftAlpha
)

// ToneOptions says which rung each step of a tone's colour pair comes from.
type ToneOptions struct {
	Fill      Step
	Hover     Step
	Pressed   Step
	Soft      Step
	SoftHover Step
	OnSoft    Step

	Mode           SoftStyle
	SoftAlpha      float64
	SoftHoverAlpha float64

	OnFill color.NRGBA
}

// NeutralOptions describes how the flat surface roles are built from a canvas
// and an ink. Text and border roles composite ink over the canvas; raised
// fills composite it over the surface. Light and dark recipes use the same
// operations with different canvas, ink and alpha values.
type NeutralOptions struct {
	Canvas color.NRGBA
	Ink    color.NRGBA

	// Text roles, plus the two surfaces built straight off the canvas.
	ForegroundAlpha float64
	MutedAlpha      float64
	SurfaceAlpha    float64
	OverlayAlpha    float64

	// Fills laid over Surface rather than over the canvas, so a raised panel
	// reads as raised rather than as the canvas shaded.
	SurfaceSecondaryAlpha float64
	SurfaceTertiaryAlpha  float64
	SurfaceHoverAlpha     float64
	SurfacePressedAlpha   float64
	DisabledAlpha         float64

	// Border, separator and placeholder are the canvas shaded toward the ink,
	// so they composite straight over it.
	BorderAlpha      float64
	SeparatorAlpha   float64
	PlaceholderAlpha float64

	DisabledForegroundAlpha float64

	// FieldHoverStaysFlat pins the field's hover face to its resting one. A
	// style sets it when hover is signalled somewhere other than the field's
	// own background, such as its border, to avoid a second hover cue.
	FieldHoverStaysFlat bool
}

// DeriveOptions is the full derivation parameter set: the recipe a style feeds
// to Scales.Derive. Start from DefaultDeriveOptions and change what the style
// disagrees with — a zero value derives nothing useful.
type DeriveOptions struct {
	// Tones holds one entry per tone, indexed by Tone.
	Tones   [ToneCount]ToneOptions
	Neutral NeutralOptions

	SurfaceShadow color.NRGBA
	OverlayShadow color.NRGBA
}

// DefaultDeriveOptions is the brandless recipe: a light canvas with one hue
// family per tone, no style-specific deviation. It exists so a theme can be
// built without borrowing another style's values; a style package copies it and
// overrides the values its own design system changes.
func DefaultDeriveOptions() DeriveOptions {
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	base := ToneOptions{
		Fill: Step500, Hover: Step400, Pressed: Step600,
		Soft: Step50, SoftHover: Step100, OnSoft: Step700,
		Mode: SoftRamp, OnFill: white,
	}

	var o DeriveOptions
	for t := range o.Tones {
		o.Tones[t] = base
	}

	o.Neutral = NeutralOptions{
		Canvas: white,
		Ink:    color.NRGBA{A: 0xff},

		ForegroundAlpha: 0.88, // colorText
		MutedAlpha:      0.45, // colorTextTertiary
		SurfaceAlpha:    0,    // colorBgContainer sits on the canvas
		OverlayAlpha:    0,    // colorBgElevated likewise

		SurfaceSecondaryAlpha: 0.02, // colorFillQuaternary
		SurfaceTertiaryAlpha:  0.04, // colorFillTertiary
		SurfaceHoverAlpha:     0.02, // colorFillQuaternary
		SurfacePressedAlpha:   0.06, // colorFillSecondary
		// No disabled fill: a brandless recipe has no gray-scheme tokens to
		// swap in, so a disabled control fades instead of turning gray. A style
		// whose design system paints one sets this explicitly.
		DisabledAlpha: 0,

		BorderAlpha:      0.15, // colorBorder
		SeparatorAlpha:   0.06, // colorBorderSecondary
		PlaceholderAlpha: 0.25, // colorTextQuaternary

		DisabledForegroundAlpha: 0.25, // colorTextQuaternary
	}
	o.SurfaceShadow = color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x26}
	o.OverlayShadow = color.NRGBA{R: 0x0f, G: 0x17, B: 0x29, A: 0x5c}
	return o
}

// DarkDeriveOptions is the brandless dark recipe: the light one with the canvas
// and the ink swapped and every rung read from the other end of the ramp.
//
// The neutral alphas do not carry over, because the roles they build do not sit
// in the same place: on a light canvas the container *is* the canvas and its
// alpha is zero, while on a black one it has to lift off it to read as a panel.
// The tone rungs move for the same reason: a pale fill reads as low-emphasis
// against a light canvas and as a highlight against a black one, so the soft
// pair swaps to the dark end and the fill steps walk one rung toward the seed.
func DarkDeriveOptions() DeriveOptions {
	o := DefaultDeriveOptions()

	for t := range o.Tones {
		o.Tones[t] = darkTone()
	}

	o.Neutral = NeutralOptions{
		Canvas: color.NRGBA{A: 0xff},
		Ink:    color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},

		ForegroundAlpha: 0.85,
		MutedAlpha:      0.45,
		SurfaceAlpha:    0.08,
		OverlayAlpha:    0.12,

		SurfaceSecondaryAlpha: 0.04,
		SurfaceTertiaryAlpha:  0.08,
		SurfaceHoverAlpha:     0.04,
		SurfacePressedAlpha:   0.12,
		DisabledAlpha:         0, // no gray-scheme tokens to swap in

		BorderAlpha:      0.26,
		SeparatorAlpha:   0.19,
		PlaceholderAlpha: 0.25,

		DisabledForegroundAlpha: 0.25,
	}
	// A black drop shadow reads as nothing on a black canvas: the surface drops
	// its shadow entirely and the overlay carries a white one instead.
	o.SurfaceShadow = color.NRGBA{}
	o.OverlayShadow = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x4d}
	return o
}

// darkTone is the rung set every tone reads under DarkDeriveOptions. It is the
// light set walked two rungs toward the seed: the fill pair moves up so a filled
// control still reads as the hue's mid tone against black, and the soft pair
// moves down to the dark end where a low-emphasis fill belongs.
//
// On-fill stays white because the fill still sits above the canvas either way;
// on-soft moves with the soft pair so the label keeps its contrast.
func darkTone() ToneOptions {
	return ToneOptions{
		Fill: Step400, Hover: Step300, Pressed: Step500,
		Soft: Step700, SoftHover: Step600, OnSoft: Step400,
		Mode: SoftRamp, OnFill: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
	}
}

// Derive builds the semantic colour layer from a set of ramps.
func (s Scales) Derive(o DeriveOptions) Semantic {
	n := o.Neutral
	var sem Semantic

	sem.Background = n.Canvas
	sem.Foreground = Composite(n.Canvas, n.Ink, n.ForegroundAlpha)

	sem.Surface = Composite(n.Canvas, n.Ink, n.SurfaceAlpha)
	sem.SurfaceForeground = sem.Foreground
	sem.SurfaceSecondary = Composite(sem.Surface, n.Ink, n.SurfaceSecondaryAlpha)
	sem.SurfaceSecondaryForeground = sem.Foreground
	sem.SurfaceTertiary = Composite(sem.Surface, n.Ink, n.SurfaceTertiaryAlpha)
	sem.SurfaceTertiaryForeground = sem.Foreground
	sem.SurfaceHover = Composite(sem.Surface, n.Ink, n.SurfaceHoverAlpha)
	sem.SurfacePressed = Composite(sem.Surface, n.Ink, n.SurfacePressedAlpha)
	sem.SurfaceRaised = sem.Surface

	sem.Overlay = Composite(n.Canvas, n.Ink, n.OverlayAlpha)
	sem.OverlayForeground = sem.Foreground

	sem.Muted = sem.SurfaceTertiary
	sem.MutedForeground = Composite(n.Canvas, n.Ink, n.MutedAlpha)

	sem.Border = Composite(n.Canvas, n.Ink, n.BorderAlpha)
	sem.Separator = Composite(n.Canvas, n.Ink, n.SeparatorAlpha)

	sem.FieldBackground = sem.Surface
	// The field's hover face is the hover surface: a brandless field signals
	// hover the way every other surface does. It used to repeat the resting
	// background, which left a borderless field (the default board's Input has
	// no BorderWidth) resolving the identical style on hover — no feedback at
	// all. A style whose field signals hover elsewhere sets FieldHoverStaysFlat.
	sem.FieldHover = sem.SurfaceHover
	if n.FieldHoverStaysFlat {
		sem.FieldHover = sem.FieldBackground
	}
	sem.FieldForeground = sem.Foreground
	sem.FieldPlaceholder = Composite(n.Canvas, n.Ink, n.PlaceholderAlpha)
	// Focus keeps the resting face: the ring is the focus cue, and shading the
	// box as well would double-signal it.
	sem.FieldFocus = sem.FieldBackground

	sem.Segment = sem.SurfaceTertiary
	sem.SegmentForeground = sem.MutedForeground

	// The neutral control face is the neutral tone's own fill — a neutral control
	// *is* the neutral tone, so the role and the tone agree by construction.
	// Reading a different surface rung here (as this used to) made the role mean
	// one thing in a derived theme and another in a hand-filled one.
	sem.ControlSurface = sem.Surface
	sem.ControlSurfaceHover = sem.SurfacePressed

	// A neutral outlined control strokes with the separator-weight border and
	// labels in the neutral control's own foreground.
	sem.ControlOutline = sem.Border
	sem.ControlOutlineForeground = sem.Foreground

	// DisabledBackground is the one role whose absence is meaningful: at alpha
	// zero the field falls back to the family-wide opacity fade rather than
	// painting a gray fill. Composite always returns an opaque colour, so the
	// fill is composited only when the recipe asks for one.
	if n.DisabledAlpha > 0 {
		sem.DisabledBackground = Composite(sem.Surface, n.Ink, n.DisabledAlpha)
	}
	sem.DisabledForeground = Alpha(n.Ink, n.DisabledForegroundAlpha)
	sem.DisabledBorder = sem.Border

	sem.SurfaceShadow = o.SurfaceShadow
	sem.OverlayShadow = o.OverlayShadow

	for t := Tone(0); int(t) < ToneCount; t++ {
		sem.Tones[t] = s.DeriveTone(t, o.Tones[t])
	}
	// A neutral tone has no ramp, so its low-emphasis pair is the surface it
	// already fills with: Soft mirrors Fill and SoftHover mirrors Hover.
	//
	// The two roles are otherwise unrelated — Soft is a tint or a pale ramp rung
	// for a hue tone — and letting a neutral tone read a *different* surface
	// rung (as this used to) made ToneDefault.Soft and ToneHover.Soft mean
	// something no consumer could predict, because the palette bridge and this
	// path picked different rungs for the same slot.
	sem.Tones[ToneDefault] = neutralPair(sem.Surface, sem.Foreground, sem.SurfacePressed)
	sem.Tones[ToneHover] = neutralPair(sem.SurfaceSecondary, sem.Foreground, sem.SurfacePressed)

	// Focus rings and selection use the primary tone's low-emphasis fill,
	// so a dark recipe can move both by changing that tone's soft rung.
	sem.Ring = sem.Tones[TonePrimary].Soft
	sem.Selection = sem.Ring
	return sem
}

// DeriveTone resolves one tone from its ramp. Derive calls it for every tone;
// External themes with published colour scales can use the same tone mapping
// as seed-derived themes without duplicating the fill and soft rules.
func (s Scales) DeriveTone(t Tone, o ToneOptions) TonePair {
	return tonePair(s.Scale(t), o)
}

func tonePair(scale ColorScale, o ToneOptions) TonePair {
	p := TonePair{
		Fill:    scale.At(o.Fill),
		OnFill:  o.OnFill,
		Hover:   scale.At(o.Hover),
		Pressed: scale.At(o.Pressed),
		OnSoft:  scale.At(o.OnSoft),
	}
	if o.Mode == SoftAlpha {
		p.Soft = Alpha(p.Fill, o.SoftAlpha)
		p.SoftHover = Alpha(p.Fill, o.SoftHoverAlpha)
		return p
	}
	p.Soft = scale.At(o.Soft)
	p.SoftHover = scale.At(o.SoftHover)
	return p
}

// neutralPair builds a tone out of surface roles instead of a hue ramp. The
// neutral tones are the one place a tone does not map to a ten-step scale.
//
// Soft mirrors Fill and SoftHover mirrors Hover. Fill and Hover sit one surface
// level apart in every style, which is what makes the pair usable as a control
// face (resting plus hover); the Soft pair is deliberately the same two values
// rather than a third rung, so a consumer that reaches for either name gets the
// same surface.
func neutralPair(fill, onFill, hover color.NRGBA) TonePair {
	return TonePair{
		Fill:      fill,
		OnFill:    onFill,
		Hover:     hover,
		Pressed:   hover,
		Soft:      fill,
		SoftHover: hover,
		OnSoft:    onFill,
	}
}
