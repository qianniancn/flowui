package ui

import "github.com/qianniancn/flowui/internal/components/progress"

type ProgressCircleWidget = progress.ProgressCircleWidget

// ProgressCircleColor selects the circle's semantic color.
type ProgressCircleColor = progress.ProgressCircleColor

// ProgressCircleSize selects the circle's diameter.
type ProgressCircleSize = progress.ProgressCircleSize

const (
	ProgressCircleAccent  = progress.ProgressCircleAccent
	ProgressCircleDefault = progress.ProgressCircleDefault
	ProgressCircleSuccess = progress.ProgressCircleSuccess
	ProgressCircleWarning = progress.ProgressCircleWarning
	ProgressCircleDanger  = progress.ProgressCircleDanger

	ProgressCircleMedium = progress.ProgressCircleMedium
	ProgressCircleSmall  = progress.ProgressCircleSmall
	ProgressCircleLarge  = progress.ProgressCircleLarge
)

// ProgressCircle creates a circular progress indicator for value.
func ProgressCircle(key string, value float64) ProgressCircleWidget {
	return progress.ProgressCircle(key, value)
}
