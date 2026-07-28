package selects

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/theme"
)

type selectStyle struct {
	field field.Colors
	error color.NRGBA
}

func selectStyleFor(theme *theme.Theme, disabled bool) selectStyle {
	style := selectStyle{error: theme.Palette.Danger}
	if disabled {
		style.error = theme.DisabledColor(style.error)
	}
	return style
}
