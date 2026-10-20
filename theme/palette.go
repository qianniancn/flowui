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
	return p.Overlay
}

func (p Palette) OverlayForegroundColor() color.NRGBA {
	return p.OverlayForeground
}

func (p Palette) OverlayShadowColor() color.NRGBA {
	return p.OverlayShadow
}

func (p Palette) DefaultColor() color.NRGBA {
	return p.Default
}

func (p Palette) DefaultForegroundColor() color.NRGBA {
	return p.DefaultForeground
}

func (p Palette) DefaultHoverColor() color.NRGBA {
	return p.DefaultHover
}

func (p Palette) FieldBackgroundColor() color.NRGBA {
	return p.FieldBackground
}

func (p Palette) FieldHoverColor() color.NRGBA {
	return p.FieldHover
}

func (p Palette) FieldForegroundColor() color.NRGBA {
	return p.FieldForeground
}

func (p Palette) FieldPlaceholderColor() color.NRGBA {
	return p.FieldPlaceholder
}

func (p Palette) FieldFocusColor() color.NRGBA {
	return p.FieldFocus
}

func (p Palette) SeparatorColor() color.NRGBA {
	return p.Separator
}

func (p Palette) SuccessSoftForegroundColor() color.NRGBA {
	return p.SuccessSoftForeground
}

func (p Palette) WarningSoftForegroundColor() color.NRGBA {
	return p.WarningSoftForeground
}
