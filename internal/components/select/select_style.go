package selects

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/theme"
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
