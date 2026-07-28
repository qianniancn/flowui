package ui

import "github.com/qianniancn/flowui/internal/components/progress"

type ProgressBarWidget = progress.ProgressBarWidget
type ProgressBarColor = progress.ProgressBarColor
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

func ProgressBar(key string, value float64) ProgressBarWidget {
	return progress.ProgressBar(key, value)
}
