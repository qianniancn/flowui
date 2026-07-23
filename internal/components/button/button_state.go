package button

import "github.com/qianniancn/FlowUI/internal/theme"

func buttonPressedScale(theme *theme.Theme, size ButtonSize) float32 {
	switch size {
	case ButtonSmall:
		return theme.Components.Button.PressedScaleSmall
	case ButtonLarge:
		return theme.Components.Button.PressedScaleLarge
	default:
		return theme.Components.Button.PressedScaleMedium
	}
}
