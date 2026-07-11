package spinner

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type spinnerStyle struct {
	color color.NRGBA
}

type spinnerSizeStyle struct {
	diameter    unit.Dp
	strokeRatio float32
	insetRatio  float32
}

func spinnerStyleFor(activeTheme *theme.Theme, spinnerColor SpinnerColor) spinnerStyle {
	style := spinnerStyle{color: activeTheme.Palette.Accent}
	switch spinnerColor {
	case SpinnerCurrent:
		style.color = activeTheme.Palette.Foreground
	case SpinnerSuccess:
		style.color = activeTheme.Palette.Success
	case SpinnerWarning:
		style.color = activeTheme.Palette.Warning
	case SpinnerDanger:
		style.color = activeTheme.Palette.Danger
	}
	return style
}

func spinnerSizeStyleFor(activeTheme *theme.Theme, size SpinnerSize) spinnerSizeStyle {
	tokens := activeTheme.Components.Spinner
	diameter := tokens.MediumSize
	switch size {
	case SpinnerSmall:
		diameter = tokens.SmallSize
	case SpinnerLarge:
		diameter = tokens.LargeSize
	case SpinnerExtraLarge:
		diameter = tokens.ExtraLargeSize
	}
	return spinnerSizeStyle{
		diameter:    diameter,
		strokeRatio: tokens.StrokeRatio,
		insetRatio:  tokens.InsetRatio,
	}
}
