package theme

// Semantic returns the semantic colour layer for this theme.
//
// A theme built from a Seed carries its derived layer (UseSemantic) and that is
// what comes back. Everything else — the default theme and any style package
// that still hand-fills a palette — gets the v1 palette projected onto the
// semantic layer, which is why this stays a method rather than a plain field.
func (t *Theme) Semantic() Semantic {
	if t.derived != nil {
		return *t.derived
	}
	return SemanticFromPalette(t.Palette)
}

// UseSemantic installs an explicitly derived semantic layer and back-fills the
// palette from it, so consumers still reading the palette (material bridge,
// action presets, node graph) see the derived colours. The component board's
// palette-role slots follow too — a menu background is the overlay surface in
// every theme, so installing a palette without pointing those slots at it is
// how a theme ends up painting someone else's colours.
//
// Once installed the palette is a projection: mutating it no longer reaches the
// components. Change the Seed and re-derive instead of patching colours.
func (t *Theme) UseSemantic(s Semantic) {
	if t == nil {
		return
	}
	t.derived = &s
	t.Palette = PaletteFromSemantic(s)
	applyDerivedComponents(t)
}

// SemanticFromPalette maps the v1 palette onto the v2 semantic layer.
//
// This is a migration bridge, not the destination. The values are still v1's
// hand-filled ones and the read path is all that changes; Phase 4 replaces it
// with derivation from a Seed. It exists so components can move to the semantic
// layer in batches while every style package keeps working untouched.
//
// Roles v1 has no equivalent for borrow the closest one rather than inventing a
// value, and the borrow is noted at the call site.
func SemanticFromPalette(p Palette) Semantic {
	// The neutral tone is built first: three later roles are its own fill, so
	// reading them off one construction keeps the names in agreement.
	neutral := TonePair{
		Fill:      p.Default,
		OnFill:    p.DefaultForeground,
		Hover:     p.DefaultHover,
		Pressed:   p.DefaultHover,
		Soft:      p.Default,
		SoftHover: p.DefaultHover,
		OnSoft:    p.DefaultForeground,
	}

	var s Semantic

	s.Background = p.Background
	s.Foreground = p.Foreground

	s.Surface = p.Surface
	s.SurfaceForeground = p.SurfaceForeground
	s.SurfaceSecondary = p.SurfaceSecondary
	s.SurfaceSecondaryForeground = p.SurfaceSecondaryForeground
	s.SurfaceTertiary = p.SurfaceTertiary
	s.SurfaceTertiaryForeground = p.SurfaceTertiaryForeground
	s.SurfaceHover = p.SurfaceHover
	s.SurfacePressed = p.SurfacePressed
	s.SurfaceRaised = p.SurfaceRaised

	s.Overlay = p.Overlay
	s.OverlayForeground = p.OverlayForeground

	// v1 has no Muted surface; the tertiary one is the same intent.
	s.Muted = p.SurfaceTertiary
	s.MutedForeground = p.MutedForeground

	s.Border = p.Border
	s.Separator = p.Separator

	s.FieldBackground = p.FieldBackground
	s.FieldHover = p.FieldHover
	s.FieldForeground = p.FieldForeground
	s.FieldPlaceholder = p.FieldPlaceholder
	s.FieldFocus = p.FieldFocus

	s.Segment = p.Segment
	s.SegmentForeground = p.SegmentForeground

	// The neutral control face is the neutral control's own fill, and the
	// hand-filled palettes carry it in their Default slot — which is exactly what
	// ToneDefault reads, so both names resolve to the same value here.
	//
	// ⚠️ This path and derive_options still pick different *rungs* for it: the
	// palette's Default/DefaultHover here, Surface/SurfacePressed there, so
	// ToneDefault.Fill is #ebebec against a derived theme's #ffffff. Both are
	// self-consistent and both satisfy ControlSurface == ToneDefault.Fill, so the
	// role is unified while the rung is not. Unifying the rung changes the
	// shipped DefaultTheme's secondary button colour, so it is a separate visual
	// step rather than a cleanup.
	s.ControlSurface = neutral.Fill
	s.ControlSurfaceHover = neutral.Hover
	// The outlined stroke is the palette's border, and its label is the neutral
	// control's own label rather than the canvas foreground.
	s.ControlOutline = p.Border
	s.ControlOutlineForeground = p.DefaultForeground

	s.Ring = p.Focus
	s.Selection = p.Selection

	s.DisabledBackground = p.DisabledBackground
	s.DisabledForeground = p.DisabledForeground
	s.DisabledBorder = p.DisabledBorder

	s.SurfaceShadow = p.SurfaceShadow
	s.OverlayShadow = p.OverlayShadow

	// The palette's accent gradient is the brand surface's gradient; a hand-filled
	// theme may leave it nil, which means "draw no gradient".
	s.PrimaryGradient = p.AccentGradient

	s.Tones[ToneDefault] = neutral
	s.Tones[TonePrimary] = TonePair{
		Fill:      p.Accent,
		OnFill:    p.AccentForeground,
		Hover:     p.AccentHover,
		Pressed:   p.AccentPressed,
		Soft:      p.AccentSoft,
		SoftHover: p.AccentSoftHover,
		OnSoft:    p.AccentSoftForeground,
	}
	// v1 has a single accent hue, so the secondary tone mirrors primary.
	s.Tones[ToneSecondary] = s.Tones[TonePrimary]
	s.Tones[ToneHover] = TonePair{
		Fill:      p.SurfaceHover,
		OnFill:    p.Foreground,
		Hover:     p.SurfacePressed,
		Pressed:   p.SurfacePressed,
		Soft:      p.SurfaceHover,
		SoftHover: p.SurfacePressed,
		OnSoft:    p.Foreground,
	}
	// v1 gives warning and danger interaction steps but leaves success without
	// any, so success falls back to its resting fill until Phase 4 derives it.
	s.Tones[ToneSuccess] = TonePair{
		Fill:      p.Success,
		OnFill:    p.SuccessForeground,
		Hover:     p.Success,
		Pressed:   p.Success,
		Soft:      p.SuccessSoft,
		SoftHover: p.SuccessSoft,
		OnSoft:    p.SuccessSoftForeground,
	}
	s.Tones[ToneWarning] = TonePair{
		Fill:      p.Warning,
		OnFill:    p.WarningForeground,
		Hover:     ColorOr(p.WarningHover, p.Warning),
		Pressed:   ColorOr(p.WarningHover, p.Warning),
		Soft:      p.WarningSoft,
		SoftHover: ColorOr(p.WarningSoftHover, p.WarningSoft),
		OnSoft:    p.WarningSoftForeground,
	}
	s.Tones[ToneDanger] = TonePair{
		Fill:      p.Danger,
		OnFill:    p.DangerForeground,
		Hover:     ColorOr(p.DangerHover, p.Danger),
		Pressed:   ColorOr(p.DangerPressed, p.Danger),
		Soft:      p.DangerSoft,
		SoftHover: ColorOr(p.DangerSoftHover, p.DangerSoft),
		OnSoft:    p.DangerSoftForeground,
	}
	return s
}
