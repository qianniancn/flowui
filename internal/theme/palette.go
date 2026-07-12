package theme

import "image/color"

// ColorOr returns fallback when value is unset.
func ColorOr(value, fallback color.NRGBA) color.NRGBA {
	if value.A == 0 {
		return fallback
	}
	return value
}

func (p Palette) OverlayColor() color.NRGBA {
	return ColorOr(p.Overlay, p.Surface)
}

func (p Palette) OverlayForegroundColor() color.NRGBA {
	return ColorOr(p.OverlayForeground, p.Foreground)
}

func (p Palette) OverlayShadowColor() color.NRGBA {
	return ColorOr(p.OverlayShadow, p.Shadow)
}

func (p Palette) DefaultColor() color.NRGBA {
	return ColorOr(p.Default, p.SurfaceRaised)
}

func (p Palette) DefaultForegroundColor() color.NRGBA {
	return ColorOr(p.DefaultForeground, p.Foreground)
}

func (p Palette) DefaultHoverColor() color.NRGBA {
	return ColorOr(p.DefaultHover, p.SurfacePressed)
}

func (p Palette) SeparatorColor() color.NRGBA {
	return ColorOr(p.Separator, p.Border)
}

func (p Palette) SuccessSoftForegroundColor() color.NRGBA {
	return ColorOr(p.SuccessSoftForeground, p.Success)
}

func (p Palette) WarningSoftForegroundColor() color.NRGBA {
	return ColorOr(p.WarningSoftForeground, p.Warning)
}
