package selects

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type selectStyle struct {
	field field.Style
	error color.NRGBA
}

func selectStyleFor(theme *theme.Theme, variant SelectVariant, hovered, focusVisible, disabled, invalid bool) selectStyle {
	style := selectStyle{
		field: field.ResolveStyle(theme, variant, hovered, focusVisible, disabled, invalid),
		error: theme.Palette.Danger,
	}
	if disabled {
		style.error = theme.DisabledColor(style.error)
	}
	return style
}
