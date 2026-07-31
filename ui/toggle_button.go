package ui

import "github.com/qianniancn/flowui/internal/components/togglebutton"

type ToggleButtonWidget = togglebutton.ToggleButtonWidget

// ToggleButtonVariant selects a toggle button's visual treatment.
type ToggleButtonVariant = togglebutton.ToggleButtonVariant

// ToggleButtonSize selects a toggle button's control height and padding.
type ToggleButtonSize = togglebutton.ToggleButtonSize

const (
	ToggleButtonDefault = togglebutton.ToggleButtonDefault
	ToggleButtonGhost   = togglebutton.ToggleButtonGhost

	ToggleButtonMedium = togglebutton.ToggleButtonMedium
	ToggleButtonSmall  = togglebutton.ToggleButtonSmall
	ToggleButtonLarge  = togglebutton.ToggleButtonLarge
)

// ToggleButton creates a keyed button initialized with selected.
func ToggleButton(key string, selected bool, child Widget) ToggleButtonWidget {
	return togglebutton.ToggleButton(key, selected, child)
}
