package ui

import "github.com/qianniancn/FlowUI/internal/components/button"

type ButtonWidget = button.ButtonWidget
type ButtonVariant = button.ButtonVariant
type ButtonSize = button.ButtonSize

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
)

func Button(key string, child Widget) ButtonWidget {
	return button.Button(key, child)
}
