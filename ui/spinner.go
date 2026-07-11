package ui

import "github.com/qianniancn/FlowUI/internal/components/spinner"

type SpinnerWidget = spinner.SpinnerWidget
type SpinnerColor = spinner.SpinnerColor
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

func Spinner() SpinnerWidget {
	return spinner.Spinner()
}
