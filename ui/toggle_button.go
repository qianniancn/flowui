package ui

import "github.com/qianniancn/FlowUI/internal/components/togglebutton"

type ToggleButtonWidget = togglebutton.ToggleButtonWidget
type ToggleButtonVariant = togglebutton.ToggleButtonVariant
type ToggleButtonSize = togglebutton.ToggleButtonSize

const (
	ToggleButtonDefault = togglebutton.ToggleButtonDefault
	ToggleButtonGhost   = togglebutton.ToggleButtonGhost

	ToggleButtonMedium = togglebutton.ToggleButtonMedium
	ToggleButtonSmall  = togglebutton.ToggleButtonSmall
	ToggleButtonLarge  = togglebutton.ToggleButtonLarge
)

func ToggleButton(key string, selected bool, child Widget) ToggleButtonWidget {
	return togglebutton.ToggleButton(key, selected, child)
}
