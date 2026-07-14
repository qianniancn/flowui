package ui

import "github.com/qianniancn/FlowUI/internal/components/progress"

type ProgressCircleWidget = progress.ProgressCircleWidget
type ProgressCircleColor = progress.ProgressCircleColor
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

func ProgressCircle(key string, value float64) ProgressCircleWidget {
	return progress.ProgressCircle(key, value)
}
