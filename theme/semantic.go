package theme

import "image/color"

// TonePair is the colour set one tone resolves to. Every field is a static
// value fixed at derive time — nothing here is computed per frame.
//
// The pair is deliberately wider than the five values v1 exposed per tone
// (Accent / AccentHover / AccentPressed / AccentSoft / AccentSoftForeground),
// because a soft foreground can use a different ramp step from its fill.
type TonePair struct {
	Fill      color.NRGBA // resting fill
	OnFill    color.NRGBA // text drawn on Fill
	Hover     color.NRGBA // fill while hovered
	Pressed   color.NRGBA // fill while pressed
	Soft      color.NRGBA // low-emphasis background
	SoftHover color.NRGBA // soft background while hovered
	OnSoft    color.NRGBA // text drawn on Soft
}

// Semantic is the v2 colour layer: flat surface roles plus one TonePair per
// tone. It replaces Palette, whose per-tone fields are folded into the pairs.
type Semantic struct {
	Background color.NRGBA
	Foreground color.NRGBA

	Surface                    color.NRGBA
	SurfaceForeground          color.NRGBA
	SurfaceSecondary           color.NRGBA
	SurfaceSecondaryForeground color.NRGBA
	SurfaceTertiary            color.NRGBA
	SurfaceTertiaryForeground  color.NRGBA
	SurfaceHover               color.NRGBA
	SurfacePressed             color.NRGBA
	SurfaceRaised              color.NRGBA

	Overlay           color.NRGBA
	OverlayForeground color.NRGBA

	Muted           color.NRGBA
	MutedForeground color.NRGBA

	Border    color.NRGBA
	Separator color.NRGBA

	FieldBackground  color.NRGBA
	FieldHover       color.NRGBA
	FieldForeground  color.NRGBA
	FieldPlaceholder color.NRGBA
	FieldFocus       color.NRGBA

	Segment           color.NRGBA
	SegmentForeground color.NRGBA

	Ring      color.NRGBA
	Selection color.NRGBA

	// ControlSurface and ControlSurfaceHover are the neutral control face: the
	// resting and hovered fill of a secondary or outlined control, and the hover
	// fill of a ghost one.
	//
	// They are named roles rather than a tone because no tone can carry them.
	// A neutral tone's Fill is its own surface, and a hue tone's Soft is a tint
	// or a pale ramp rung — neither is "the neutral surface a control sits on".
	// Providers can choose different surface rungs for the same role without
	// changing component logic.
	ControlSurface      color.NRGBA
	ControlSurfaceHover color.NRGBA

	// ControlOutline and ControlOutlineForeground are the neutral outlined
	// control: its resting stroke and the label drawn in that stroke's colour.
	//
	// A neutral outlined control is the one place a control's label takes the
	// stroke rather than the canvas foreground, so it needs its own role pair.
	// A toned outlined control does not use these — it strokes and labels in its
	// own tone's fill.
	ControlOutline           color.NRGBA
	ControlOutlineForeground color.NRGBA

	DisabledBackground color.NRGBA
	DisabledForeground color.NRGBA
	DisabledBorder     color.NRGBA

	SurfaceShadow color.NRGBA
	OverlayShadow color.NRGBA

	// PrimaryGradient, when set, is the brand surface's optional gradient. It
	// replaces the flat primary fill on the controls that opt in — the primary
	// button today — and is nil for a style that draws no gradient.
	//
	// It is a role a style *chooses* rather than a colour its ramp can produce,
	// so it belongs in this layer with the rest of the resolved values. Leaving
	// it on the palette left exactly one value that still answered palette
	// mutation after a derived layer was installed, which is the asymmetry v2
	// exists to remove.
	PrimaryGradient *Gradient

	Tones [ToneCount]TonePair
}

// Tone returns the colour set for a tone. An out-of-range tone yields
// ToneDefault rather than a transparent hole.
func (s Semantic) Tone(t Tone) TonePair {
	if int(t) >= ToneCount {
		return s.Tones[ToneDefault]
	}
	return s.Tones[t]
}

// HoverSurface is the neutral interaction surface — shadcn's `--accent`, which
// is deliberately *not* the brand colour. v1 called the brand colour Accent, so
// v2 keeps the two apart by never reusing the name.
func (s Semantic) HoverSurface() color.NRGBA {
	return s.Tones[ToneHover].Fill
}
