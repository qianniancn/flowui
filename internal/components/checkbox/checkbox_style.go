package checkbox

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func checkboxStyleFor(theme *theme.Theme, variant CheckboxVariant, hovered, pressed, disabled, invalid bool) checkboxStyle {
	danger := theme.Palette.Danger
	dangerHover := theme.Palette.DangerHover
	dangerFg := theme.Palette.DangerForeground
	style := checkboxStyle{
		bg:          theme.Palette.Surface,
		border:      theme.Palette.Border,
		accent:      theme.Palette.Accent,
		accentHover: theme.Palette.AccentHover,
		accentFg:    theme.Palette.AccentForeground,
		fg:          theme.Palette.Foreground,
		focusColor:  theme.Palette.Focus,
		shadow:      theme.Components.Checkbox.ShadowOpacity,
	}
	if variant == CheckboxSecondary {
		style.bg = theme.Palette.SurfaceTertiary
		style.shadow = 0
	}
	if invalid {
		style.border = danger
		style.accent = danger
		style.accentHover = dangerHover
		style.accentFg = dangerFg
		style.focusColor = danger
		style.focusColor.A = theme.Palette.Focus.A
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.border = theme.DisabledColor(style.border)
		style.accent = theme.DisabledColor(style.accent)
		style.accentHover = theme.DisabledColor(style.accentHover)
		style.accentFg = theme.DisabledColor(style.accentFg)
		style.fg = theme.DisabledColor(style.fg)
		style.focusColor = color.NRGBA{}
		style.shadow = 0
		return style
	}
	if hovered || pressed {
		if invalid {
			style.border = dangerHover
		} else {
			style.border = theme.Palette.SurfacePressed
		}
		style.accent = style.accentHover
	}
	return style
}

type checkboxStyle struct {
	bg          color.NRGBA
	border      color.NRGBA
	accent      color.NRGBA
	accentHover color.NRGBA
	accentFg    color.NRGBA
	fg          color.NRGBA
	focusColor  color.NRGBA
	selected    float32
	focus       float32
	shadow      float32
}

func checkboxLabelColor(activeTheme *theme.Theme, disabled bool) color.NRGBA {
	if disabled {
		return activeTheme.DisabledColor(activeTheme.Palette.Foreground)
	}
	return activeTheme.Palette.Foreground
}

func checkboxSupportingColor(activeTheme *theme.Theme, value color.NRGBA, disabled bool) color.NRGBA {
	if disabled {
		return activeTheme.DisabledColor(value)
	}
	return value
}
