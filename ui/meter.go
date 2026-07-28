package ui

import "github.com/qianniancn/flowui/internal/components/progress"

type MeterWidget = progress.MeterWidget
type MeterColor = progress.MeterColor
type MeterSize = progress.MeterSize

const (
	MeterAccent  = progress.MeterAccent
	MeterDefault = progress.MeterDefault
	MeterSuccess = progress.MeterSuccess
	MeterWarning = progress.MeterWarning
	MeterDanger  = progress.MeterDanger

	MeterMedium = progress.MeterMedium
	MeterSmall  = progress.MeterSmall
	MeterLarge  = progress.MeterLarge
)

func Meter(key string, value float64) MeterWidget {
	return progress.Meter(key, value)
}
