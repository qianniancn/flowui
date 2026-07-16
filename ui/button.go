package ui

import "github.com/qianniancn/FlowUI/internal/components/button"

type ButtonWidget = button.ButtonWidget
type ButtonVariant = button.ButtonVariant
type ButtonSize = button.ButtonSize
type ButtonGroupWidget = button.ButtonGroupWidget
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

func Button(key string, child Widget) ButtonWidget {
	return button.Button(key, child)
}

func ButtonGroup(buttons ...ButtonWidget) ButtonGroupWidget {
	return button.ButtonGroup(buttons...)
}
