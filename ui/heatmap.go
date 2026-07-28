package ui

import (
	"time"

	"github.com/qianniancn/flowui/internal/components/heatmap"
)

type HeatmapWidget = heatmap.Widget
type HeatmapValue = heatmap.CalendarValue

func Heatmap(key string, start, end time.Time, values []HeatmapValue) HeatmapWidget {
	return heatmap.New(key, start, end, values)
}
