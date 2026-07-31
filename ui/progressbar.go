package ui

import "github.com/qianniancn/flowui/internal/components/progress"

type ProgressBarWidget = progress.ProgressBarWidget

// ProgressBarColor selects the bar's semantic color.
type ProgressBarColor = progress.ProgressBarColor

// ProgressBarSize selects the bar's height.
type ProgressBarSize = progress.ProgressBarSize

const (
	ProgressBarAccent  = progress.ProgressBarAccent
	ProgressBarDefault = progress.ProgressBarDefault
	ProgressBarSuccess = progress.ProgressBarSuccess
	ProgressBarWarning = progress.ProgressBarWarning
	ProgressBarDanger  = progress.ProgressBarDanger

	ProgressBarMedium = progress.ProgressBarMedium
	ProgressBarSmall  = progress.ProgressBarSmall
	ProgressBarLarge  = progress.ProgressBarLarge
)

// ProgressBar creates a linear progress indicator for value.
func ProgressBar(key string, value float64) ProgressBarWidget {
	return progress.ProgressBar(key, value)
}
