package ui

import (
	"time"

	"github.com/qianniancn/flowui/internal/components/gantt"
)

type GanttChartWidget = gantt.Widget

// GanttTask describes a task or milestone on a Gantt chart.
type GanttTask = gantt.Task

// GanttTimeMarker marks a point in time on a Gantt chart.
type GanttTimeMarker = gantt.TimeMarker

// GanttSelection describes a selected task range.
type GanttSelection = gantt.Selection

// GanttTaskChange reports a task move or resize.
type GanttTaskChange = gantt.TaskChange

// GanttTimeWindow represents the visible time range (used for controlled scrolling).
type GanttTimeWindow struct {
	// Start marks the beginning of the visible time range.
	Start time.Time
	// End is the end of the visible time range.
	End time.Time
}

// GanttChart creates a chart whose tasks are identified by their keys.
func GanttChart(key string, tasks []GanttTask) GanttChartWidget {
	return gantt.New(key, tasks)
}

// NewGanttTask creates a task spanning start through end.
func NewGanttTask(key, label string, start, end time.Time) GanttTask {
	return gantt.NewTask(key, label, start, end)
}

// NewGanttMilestone creates a zero-duration milestone at at.
func NewGanttMilestone(key, label string, at time.Time) GanttTask {
	return gantt.NewMilestone(key, label, at)
}

// NewGanttMarker creates a vertical time marker at at.
func NewGanttMarker(at time.Time) GanttTimeMarker {
	return gantt.NewTimeMarker(at)
}
