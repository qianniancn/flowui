package ui

import "github.com/qianniancn/flowui/internal/components/button"

type ButtonWidget = button.ButtonWidget

// ButtonVariant selects a button's visual treatment.
type ButtonVariant = button.ButtonVariant

// ButtonSize selects a button's control height and padding.
type ButtonSize = button.ButtonSize

type ButtonGroupWidget = button.ButtonGroupWidget

// ButtonGroupOrientation controls the direction of a button group.
type ButtonGroupOrientation = button.ButtonGroupOrientation

const (
	ButtonPrimary    = button.ButtonPrimary
	ButtonSecondary  = button.ButtonSecondary
	ButtonTertiary   = button.ButtonTertiary
	ButtonGhost      = button.ButtonGhost
	ButtonOutline    = button.ButtonOutline
	ButtonDanger     = button.ButtonDanger
	ButtonDangerSoft = button.ButtonDangerSoft

	ButtonMedium = button.ButtonMedium
	ButtonSmall  = button.ButtonSmall
	ButtonLarge  = button.ButtonLarge

	ButtonGroupHorizontal = button.ButtonGroupHorizontal
	ButtonGroupVertical   = button.ButtonGroupVertical
)

// Button creates a keyed clickable button around child.
func Button(key string, child Widget) ButtonWidget {
	return button.Button(key, child)
}

// ButtonGroup groups buttons into one themed control group.
func ButtonGroup(buttons ...ButtonWidget) ButtonGroupWidget {
	return button.ButtonGroup(buttons...)
}
