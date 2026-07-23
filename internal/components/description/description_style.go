package description

import (
	"image/color"

	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type descriptionStyle struct {
	text color.NRGBA
}

func descriptionStyleFor(theme *theme.Theme, disabled bool) descriptionStyle {
	style := descriptionStyle{text: theme.Palette.MutedForeground}
	if disabled {
		style.text = theme.DisabledColor(style.text)
	}
	return style
}

func descriptionDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	return flowstyle.Style{}.
		FontSize(activeTheme.Components.Description.TextSize).
		TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.MutedForeground})

}

func descriptionStateDeclaration(activeTheme *theme.Theme, state flowstyle.StyleState) flowstyle.Style {
	if !state.Disabled {
		return flowstyle.Style{}
	}
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: activeTheme.DisabledColor(activeTheme.Palette.MutedForeground)})

}
