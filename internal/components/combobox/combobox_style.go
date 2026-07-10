package combobox

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func comboBoxItemTextColor(theme *theme.Theme, disabled bool) color.NRGBA {
	col := theme.Palette.Foreground
	if disabled {
		return theme.DisabledColor(col)
	}
	return col
}

func comboBoxItemDescriptionColor(theme *theme.Theme, disabled bool) color.NRGBA {
	col := theme.Palette.MutedForeground
	if disabled {
		return theme.DisabledColor(col)
	}
	return col
}

func comboBoxItemStyleFor(theme *theme.Theme, hovered, pressed, selected, disabled bool) comboBoxItemStyle {
	style := comboBoxItemStyle{}
	if hovered || selected {
		style.bg = theme.Palette.SurfaceRaised
	}
	if pressed {
		style.bg = theme.Palette.SurfacePressed
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
	}
	return style
}

type comboBoxItemStyle struct {
	bg color.NRGBA
}
