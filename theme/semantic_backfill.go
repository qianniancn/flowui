package theme

// PaletteFromSemantic projects the semantic layer back onto the v1 palette.
//
// v2 makes the semantic layer the source of truth: a style package derives it
// from a Seed and the palette becomes a projection. Consumers that have not
// moved to the semantic layer yet — the Gio material bridge, the action
// presets, the node graph theme — keep reading the palette, so it has to carry
// the derived colours instead of hand-filled ones.
//
// The projection is lossy by construction: the palette has no Secondary ramp
// and no separate soft fill for the neutral tone, so those do not survive a
// Semantic → Palette → Semantic round trip. Read Semantic directly where they
// matter.
func PaletteFromSemantic(s Semantic) Palette {
	neutral := s.Tone(ToneDefault)
	primary := s.Tone(TonePrimary)
	success := s.Tone(ToneSuccess)
	warning := s.Tone(ToneWarning)
	danger := s.Tone(ToneDanger)

	// The palette's Default/DefaultHover slots are the neutral control face, so
	// they carry ControlSurface rather than the neutral tone's own fill. The two
	// agree in a hand-filled theme and differ in a derived one, where the tone
	// fill is the canvas and the control face sits a level above it.
	//
	// The fallback keeps a role an older construction path never set from
	// collapsing the slot to transparent.
	controlSurface := s.ControlSurface
	if controlSurface.A == 0 {
		controlSurface = neutral.Fill
	}
	controlSurfaceHover := s.ControlSurfaceHover
	if controlSurfaceHover.A == 0 {
		controlSurfaceHover = neutral.Hover
	}

	return Palette{
		Background:                 s.Background,
		Surface:                    s.Surface,
		SurfaceForeground:          s.SurfaceForeground,
		SurfaceSecondary:           s.SurfaceSecondary,
		SurfaceSecondaryForeground: s.SurfaceSecondaryForeground,
		SurfaceTertiary:            s.SurfaceTertiary,
		SurfaceTertiaryForeground:  s.SurfaceTertiaryForeground,
		SurfaceHover:               s.SurfaceHover,
		SurfacePressed:             s.SurfacePressed,
		SurfaceRaised:              s.SurfaceRaised,
		Overlay:                    s.Overlay,
		OverlayForeground:          s.OverlayForeground,
		Foreground:                 s.Foreground,
		MutedForeground:            s.MutedForeground,
		Border:                     s.Border,
		Separator:                  s.Separator,

		Default:           controlSurface,
		DefaultForeground: neutral.OnFill,
		DefaultHover:      controlSurfaceHover,

		FieldBackground:  s.FieldBackground,
		FieldHover:       s.FieldHover,
		FieldForeground:  s.FieldForeground,
		FieldPlaceholder: s.FieldPlaceholder,
		FieldFocus:       s.FieldFocus,

		Segment:           s.Segment,
		SegmentForeground: s.SegmentForeground,

		Accent:               primary.Fill,
		AccentHover:          primary.Hover,
		AccentPressed:        primary.Pressed,
		AccentForeground:     primary.OnFill,
		AccentSoft:           primary.Soft,
		AccentSoftHover:      primary.SoftHover,
		AccentSoftForeground: primary.OnSoft,
		// The palette's AccentGradient is the semantic PrimaryGradient. Keeping
		// the projection total here is what lets a style package set the
		// gradient in its semantic recipe instead of patching the palette after
		// installing one.
		AccentGradient: s.PrimaryGradient,

		Success:               success.Fill,
		SuccessForeground:     success.OnFill,
		SuccessSoft:           success.Soft,
		SuccessSoftForeground: success.OnSoft,

		Warning:               warning.Fill,
		WarningForeground:     warning.OnFill,
		WarningHover:          warning.Hover,
		WarningSoft:           warning.Soft,
		WarningSoftHover:      warning.SoftHover,
		WarningSoftForeground: warning.OnSoft,

		Danger:               danger.Fill,
		DangerHover:          danger.Hover,
		DangerPressed:        danger.Pressed,
		DangerForeground:     danger.OnFill,
		DangerSoft:           danger.Soft,
		DangerSoftHover:      danger.SoftHover,
		DangerSoftForeground: danger.OnSoft,

		Focus:     s.Ring,
		Selection: s.Selection,

		DisabledBackground: s.DisabledBackground,
		DisabledForeground: s.DisabledForeground,
		DisabledBorder:     s.DisabledBorder,

		SurfaceShadow: s.SurfaceShadow,
		OverlayShadow: s.OverlayShadow,
	}
}
