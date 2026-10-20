# 15 - Gantt chart

`GanttChart` displays tasks on a time axis with progress, dependencies,
milestones, hierarchy, and an editable time window.

## Tasks

```go
tasks := []gantt.Task{
	gantt.NewTask("design", "Design", day(0), day(5)).Progress(0.8),
	gantt.NewTask("build", "Build", day(4), day(12)).Parent("release"),
	gantt.NewMilestone("launch", "Launch", day(12)),
}

chart := gantt.New("release-plan", tasks).
	TimeWindow(day(0), day(20))
```

Task keys are stable identities. Parent keys must refer to another task in the
same data set. Use dependencies to express ordering rather than relying on row
position.

## Selection and editing

Keep the selected task and edit callback in the model. `GanttTaskChange` reports
dragged or resized task boundaries; apply the change in `Update` and pass the
updated task list back to the chart.

The component tests under `widget/gantt` exercise
hierarchy, dependencies, progress, baselines, time markers, and editing.
