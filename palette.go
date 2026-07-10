package flowui

import "image/color"

func paletteColor(value, fallback color.NRGBA) color.NRGBA {
	if value.A == 0 {
		return fallback
	}
	return value
}

func (p Palette) overlayColor() color.NRGBA {
	return paletteColor(p.Overlay, p.Surface)
}

func (p Palette) overlayForegroundColor() color.NRGBA {
	return paletteColor(p.OverlayForeground, p.Foreground)
}

func (p Palette) overlayShadowColor() color.NRGBA {
	return paletteColor(p.OverlayShadow, p.Shadow)
}
