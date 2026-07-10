package radiogroup

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func radioStyleFor(theme *theme.Theme, variant RadioGroupVariant, hovered, pressed, disabled, invalid bool) radioStyle {
	style := radioStyle{
		bg:          theme.Palette.Surface,
		border:      theme.Palette.Border,
		selectedBg:  theme.Palette.Accent,
		dot:         theme.Palette.AccentForeground,
		fg:          theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		focusColor:  theme.Palette.Focus,
		pressed:     pressed,
	}
	if variant == RadioSecondary {
		style.bg = theme.Palette.SurfaceRaised
	}
	if hovered {
		style.border = theme.Palette.SurfacePressed
	}
	if pressed {
		style.selectedBg = theme.Palette.AccentHover
	}
	if invalid {
		style.border = theme.Palette.Danger
		style.selectedBg = theme.Palette.Danger
		style.dot = theme.Palette.DangerForeground
		style.focusColor = theme.Palette.Danger
		style.focusColor.A = theme.Palette.Focus.A
		if hovered || pressed {
			style.border = theme.Palette.DangerHover
			style.selectedBg = theme.Palette.DangerHover
		}
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.border = theme.DisabledColor(style.border)
		style.selectedBg = theme.DisabledColor(style.selectedBg)
		style.dot = theme.DisabledColor(style.dot)
		style.fg = theme.DisabledColor(style.fg)
		style.description = theme.DisabledColor(style.description)
		style.focusColor = color.NRGBA{}
	}
	return style
}

type radioStyle struct {
	bg          color.NRGBA
	border      color.NRGBA
	selectedBg  color.NRGBA
	dot         color.NRGBA
	fg          color.NRGBA
	description color.NRGBA
	focusColor  color.NRGBA
	selected    float32
	focus       float32
	pressed     bool
}
