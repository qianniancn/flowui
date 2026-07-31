package ui

import "github.com/qianniancn/flowui/internal/components/progress"

type MeterWidget = progress.MeterWidget

// MeterColor selects the meter's semantic color.
type MeterColor = progress.MeterColor

// MeterSize selects the meter's size.
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

// Meter creates a compact meter indicator for value.
func Meter(key string, value float64) MeterWidget {
	return progress.Meter(key, value)
}
