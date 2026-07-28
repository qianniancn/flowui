package label

import (
	"image/color"

	"gioui.org/font"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type labelStyle struct {
	text     color.NRGBA
	required color.NRGBA
}

func labelStyleFor(theme *theme.Theme, foreground color.NRGBA, disabled, invalid bool) labelStyle {
	style := labelStyle{
		text:     foreground,
		required: theme.Palette.Danger,
	}
	if invalid {
		style.text = theme.Palette.Danger
	}
	if disabled {
		style.text = theme.DisabledColor(style.text)
		style.required = theme.DisabledColor(style.required)
	}
	return style
}

func labelDefaultDeclaration(activeTheme *theme.Theme, foreground color.NRGBA) flowstyle.Style {
	return flowstyle.Style{}.
		FontSize(activeTheme.Components.Label.TextSize).
		FontWeight(int(font.Medium)).
		TextColor(flowstyle.SolidColor{Color: foreground})

}

func labelStateDeclaration(activeTheme *theme.Theme, foreground color.NRGBA, state flowstyle.StyleState) flowstyle.Style {
	resolved := labelStyleFor(activeTheme, foreground, state.Disabled, state.Invalid)
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: resolved.text})

}

func labelRequiredColor(activeTheme *theme.Theme, disabled bool) color.NRGBA {
	if disabled {
		return activeTheme.DisabledColor(activeTheme.Palette.Danger)
	}
	return activeTheme.Palette.Danger
}
