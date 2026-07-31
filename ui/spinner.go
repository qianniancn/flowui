package ui

import "github.com/qianniancn/flowui/internal/components/spinner"

type SpinnerWidget = spinner.SpinnerWidget

// SpinnerColor selects the spinner's semantic color.
type SpinnerColor = spinner.SpinnerColor

// SpinnerSize selects the spinner's diameter.
type SpinnerSize = spinner.SpinnerSize

const (
	SpinnerAccent  = spinner.SpinnerAccent
	SpinnerCurrent = spinner.SpinnerCurrent
	SpinnerSuccess = spinner.SpinnerSuccess
	SpinnerWarning = spinner.SpinnerWarning
	SpinnerDanger  = spinner.SpinnerDanger

	SpinnerMedium     = spinner.SpinnerMedium
	SpinnerSmall      = spinner.SpinnerSmall
	SpinnerLarge      = spinner.SpinnerLarge
	SpinnerExtraLarge = spinner.SpinnerExtraLarge
)

// Spinner creates an indeterminate progress spinner.
func Spinner() SpinnerWidget {
	return spinner.Spinner()
}
