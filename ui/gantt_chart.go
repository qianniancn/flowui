package ui

import (
	"time"

	"github.com/qianniancn/flowui/internal/components/gantt"
)

type GanttChartWidget = gantt.Widget
type GanttTask = gantt.Task
type GanttTimeMarker = gantt.TimeMarker
type GanttSelection = gantt.Selection
type GanttTaskChange = gantt.TaskChange

// GanttTimeWindow represents the visible time range (used for controlled scrolling).
type GanttTimeWindow struct {
	Start time.Time
	End   time.Time
}

func GanttChart(key string, tasks []GanttTask) GanttChartWidget {
	return gantt.New(key, tasks)
}

func NewGanttTask(key, label string, start, end time.Time) GanttTask {
	return gantt.NewTask(key, label, start, end)
}

func NewGanttMilestone(key, label string, at time.Time) GanttTask {
	return gantt.NewMilestone(key, label, at)
}

func NewGanttMarker(at time.Time) GanttTimeMarker {
	return gantt.NewTimeMarker(at)
}
