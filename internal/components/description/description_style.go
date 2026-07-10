package description

import (
	"image/color"

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
