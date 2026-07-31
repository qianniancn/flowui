package ui

import (
	"time"

	"github.com/qianniancn/flowui/internal/components/heatmap"
)

type HeatmapWidget = heatmap.Widget

// HeatmapValue stores one dated value in a heatmap.
type HeatmapValue = heatmap.CalendarValue

// Heatmap creates a calendar heatmap over the supplied time range.
func Heatmap(key string, start, end time.Time, values []HeatmapValue) HeatmapWidget {
	return heatmap.New(key, start, end, values)
}
