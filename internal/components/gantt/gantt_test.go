package gantt

import (
	"fmt"
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func ganttDate(day int) time.Time {
	return time.Date(2025, time.January, day, 0, 0, 0, 0, time.UTC)
}

func TestGanttCopiesTasksAndResolvesGroups(t *testing.T) {
	tasks := []Task{
		NewTask("plan", "Plan", ganttDate(1), ganttDate(4)).Group("planning").DependsOn("design"),
		NewTask("design", "Design", ganttDate(4), ganttDate(8)).Group("planning").Progress(.5),
	}
	widget := New("project", tasks)
	tasks[0].label = "Changed"
	tasks[0].dependsOn[0] = "changed"
	if widget.tasks[0].label != "Plan" || widget.tasks[0].dependsOn[0] != "design" {
		t.Fatal("GanttChart retained mutable task input")
	}
	activeTheme := theme.DefaultTheme()
	resolved := widget.resolveTasks(&activeTheme)
	if resolved[0].color != resolved[1].color || resolved[0].progress != 1 || resolved[1].progress != .5 {
		t.Fatalf("resolved tasks = %#v", resolved)
	}
}

func TestGanttCachesResolvedAndVisibleTasks(t *testing.T) {
	widget := New("project", []Task{
		NewTask("root", "Root", ganttDate(1), ganttDate(4)).Group("keep"),
		NewTask("child", "Child", ganttDate(2), ganttDate(3)).Parent("root"),
	})
	activeTheme := theme.DefaultTheme()
	state := &chartState{}
	first := widget.resolveTasksCached(state, &activeTheme)
	second := widget.resolveTasksCached(state, &activeTheme)
	if len(first) != 2 || len(second) != 2 || &first[0] != &second[0] {
		t.Fatal("resolved task cache was not reused")
	}
	rebuilt := New("project", []Task{
		NewTask("root", "Root", ganttDate(1), ganttDate(4)).Group("keep"),
		NewTask("child", "Child", ganttDate(2), ganttDate(3)).Parent("root"),
	})
	rebuiltResolved := rebuilt.resolveTasksCached(state, &activeTheme)
	if &rebuiltResolved[0] != &first[0] || state.resolvedTasksRevision != 1 {
		t.Fatal("equivalent declarative task data restarted the resolved cache")
	}
	visible := widget.visibleTasksCached(state, first)
	state.localCollapsed = map[string]bool{"root": true}
	state.visibilityRevision++
	collapsed := widget.visibleTasksCached(state, first)
	if len(visible) != 2 || len(collapsed) != 1 || collapsed[0].key != "root" {
		t.Fatalf("visible task cache did not invalidate: visible=%#v collapsed=%#v", visible, collapsed)
	}
	updated := New("project", []Task{
		NewTask("root", "Root", ganttDate(1), ganttDate(5)).Group("keep"),
		NewTask("child", "Child", ganttDate(2), ganttDate(3)).Parent("root"),
	})
	resolvedUpdated := updated.resolveTasksCached(state, &activeTheme)
	if &resolvedUpdated[0] == &first[0] || state.resolvedTasksRevision < 2 {
		t.Fatal("resolved task cache did not invalidate for new task data")
	}
}

func TestGanttMilestoneGeometryAndSelection(t *testing.T) {
	tasks := []Task{
		NewTask("plan", "Plan", ganttDate(1), ganttDate(4)),
		NewMilestone("review", "Review", ganttDate(5)).DependsOn("plan"),
	}
	activeTheme := theme.DefaultTheme()
	resolved := New("project", tasks).resolveTasks(&activeTheme)
	geometry := New("project", tasks).resolveGeometry(resolved, image.Rect(0, 0, 300, 100), ganttDate(1), ganttDate(8), 18)
	milestone := geometry.tasks[1]
	if !milestone.task.milestone || milestone.rect.Empty() || milestone.rect.Dx() != milestone.rect.Dy() {
		t.Fatalf("milestone geometry = %#v", milestone)
	}
	selection, ok := selectionAt(f32.Pt(float32(milestone.rect.Min.X+1), float32(milestone.rect.Min.Y+1)), true, geometry)
	if !ok || selection.task.key != "review" {
		t.Fatalf("milestone selection = %#v, %v", selection, ok)
	}
}

func TestGanttBaselineAndMarkerConfiguration(t *testing.T) {
	task := NewTask("plan", "Plan", ganttDate(1), ganttDate(6)).Baseline(ganttDate(2), ganttDate(5))
	widget := New("project", []Task{task}).Marker(NewTimeMarker(ganttDate(3)).Text("Status"))
	if !widget.tasks[0].hasBaseline || !widget.tasks[0].baselineStart.Equal(ganttDate(2)) || len(widget.markers) != 1 || widget.markers[0].label != "Status" {
		t.Fatalf("baseline and marker configuration = %#v, %#v", widget.tasks[0], widget.markers)
	}
	resolved := widget.resolveTasks(themePtr())
	geometry := widget.resolveGeometry(resolved, image.Rect(0, 0, 300, 100), ganttDate(1), ganttDate(8), 18)
	if geometry.tasks[0].baseline.Empty() || geometry.tasks[0].baseline.Dy() >= geometry.tasks[0].rect.Dy() {
		t.Fatalf("baseline geometry = %#v", geometry.tasks[0])
	}
}

func themePtr() *theme.Theme {
	value := theme.DefaultTheme()
	return &value
}

func TestGanttRejectsInvalidConfiguration(t *testing.T) {
	tests := []func(){
		func() { NewTask("", "Task", ganttDate(1), ganttDate(2)) },
		func() { NewTask("task", "Task", ganttDate(2), ganttDate(1)) },
		func() { New("chart", nil).Height(0) },
		func() { New("chart", nil).TimeRange(ganttDate(2), ganttDate(1)) },
		func() { New("chart", nil).TimeTicks(1) },
		func() { New("chart", nil).RowHeight(0) },
		func() { NewTask("task", "Task", ganttDate(1), ganttDate(2)).Progress(1.1) },
		func() {
			New("chart", []Task{NewTask("task", "Task", ganttDate(1), ganttDate(2)).DependsOn("missing")}).resolveTasks(themePtr())
		},
		func() {
			New("chart", []Task{NewTask("task", "Task", ganttDate(1), ganttDate(2)).DependsOn("task")}).resolveTasks(themePtr())
		},
		func() {
			New("chart", []Task{
				NewTask("dependency", "Dependency", ganttDate(1), ganttDate(2)),
				NewTask("task", "Task", ganttDate(2), ganttDate(3)).DependsOn("dependency", "dependency"),
			}).resolveTasks(themePtr())
		},
		func() {
			New("chart", []Task{
				NewTask("a", "A", ganttDate(1), ganttDate(2)).DependsOn("b"),
				NewTask("b", "B", ganttDate(2), ganttDate(3)).DependsOn("a"),
			}).resolveTasks(themePtr())
		},
	}
	for _, run := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid Gantt configuration did not panic")
				}
			}()
			run()
		}()
	}
}

func TestGanttTimeWindowNormalizesAndMaps(t *testing.T) {
	fullStart, fullEnd := ganttDate(1), ganttDate(11)
	window := normalizeTimeWindow(ganttDate(3), ganttDate(7), fullStart, fullEnd)
	if window.Start <= 0 || window.End >= 1 || window.End <= window.Start {
		t.Fatalf("normalized window = %#v", window)
	}
	start, end := mapTimeWindow(fullStart, fullEnd, window)
	if absDuration(start.Sub(ganttDate(3))) > 50*time.Millisecond || absDuration(end.Sub(ganttDate(7))) > 50*time.Millisecond {
		t.Fatalf("mapped window = %s to %s", start, end)
	}
}

func TestGanttTimeWindowKeepsGestureStateAcrossFrames(t *testing.T) {
	fullStart, fullEnd := ganttDate(1), ganttDate(11)
	widget := New("chart", nil).TimeWindow(ganttDate(3), ganttDate(7))
	state := &chartState{}
	initial := widget.effectiveTimeWindow(state, fullStart, fullEnd)
	state.timeWindow = chart.NewDataWindow(.4, .8)
	current := widget.effectiveTimeWindow(state, fullStart, fullEnd)
	if initial == current || current.Start != .4 || current.End != .8 {
		t.Fatalf("time window reset across frames: initial=%#v current=%#v", initial, current)
	}
}

func TestGanttPointerCursorMatchesInteraction(t *testing.T) {
	state := &chartState{}
	if got := ganttPointerCursor(state, false); got != pointer.CursorDefault {
		t.Fatalf("passive chart cursor = %v, want default", got)
	}
	if got := ganttPointerCursor(state, true); got != pointer.CursorPointer {
		t.Fatalf("clickable task cursor = %v, want pointer", got)
	}
	state.windowGesture.Update(pointer.Event{
		Kind: pointer.Press, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20),
	}, image.Rect(0, 0, 100, 40), chart.NewDataWindow(.2, .8), false)
	if got := ganttPointerCursor(state, false); got != pointer.CursorDefault {
		t.Fatalf("pressed chart cursor = %v, want default", got)
	}
	state.windowGesture.Update(pointer.Event{
		Kind: pointer.Drag, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 20),
	}, image.Rect(0, 0, 100, 40), chart.NewDataWindow(.2, .8), false)
	if got := ganttPointerCursor(state, false); got != pointer.CursorGrabbing {
		t.Fatalf("panning chart cursor = %v, want grabbing", got)
	}
}

func TestGanttResizeCursorWinsOverDraggingCursor(t *testing.T) {
	state := &chartState{editDragging: true, editMode: taskEditStart}
	if got, ok := taskEditingCursor(state); !ok || got != pointer.CursorEastWestResize {
		t.Fatalf("resizing cursor = %v, %v; want east-west resize", got, ok)
	}
	state.editMode = taskEditEnd
	if got, ok := taskEditingCursor(state); !ok || got != pointer.CursorEastWestResize {
		t.Fatalf("end resize cursor = %v, %v; want east-west resize", got, ok)
	}
	state.editMode = taskEditMove
	if got, ok := taskEditingCursor(state); !ok || got != pointer.CursorGrabbing {
		t.Fatalf("move cursor = %v, %v; want grabbing", got, ok)
	}
}

func TestGanttEntryProgressStaggersTasks(t *testing.T) {
	if got := ganttEntryProgress(.2, 0, 3); got <= 0 {
		t.Fatalf("first entry progress = %v", got)
	}
	if got := ganttEntryProgress(.2, 2, 3); got != 0 {
		t.Fatalf("last entry progress = %v, want 0 during stagger", got)
	}
	if got := ganttEntryProgress(1, 2, 3); got != 1 {
		t.Fatalf("completed entry progress = %v", got)
	}
}

func TestGanttHiddenGroupsFilterVisibleTasks(t *testing.T) {
	tasks := []resolvedTask{
		{Task: NewTask("a", "A", ganttDate(1), ganttDate(2)).Group("keep")},
		{Task: NewTask("b", "B", ganttDate(2), ganttDate(3)).Group("hide")},
	}
	visible := New("chart", nil).HiddenGroups("hide").visibleTasks(&chartState{}, tasks)
	if len(visible) != 1 || visible[0].key != "a" {
		t.Fatalf("visible tasks = %#v", visible)
	}
}

func TestGanttHierarchyResolvesDepthAndCollapse(t *testing.T) {
	tasks := []Task{
		NewTask("root", "Root", ganttDate(1), ganttDate(8)).Collapsed(true),
		NewTask("child", "Child", ganttDate(2), ganttDate(5)).Parent("root"),
		NewTask("grandchild", "Grandchild", ganttDate(3), ganttDate(4)).Parent("child"),
	}
	widget := New("chart", tasks)
	resolved := widget.resolveTasks(themePtr())
	if resolved[0].depth != 0 || resolved[1].depth != 1 || resolved[2].depth != 2 {
		t.Fatalf("resolved depths = %d, %d, %d", resolved[0].depth, resolved[1].depth, resolved[2].depth)
	}
	visible := widget.visibleTasks(&chartState{}, resolved)
	if len(visible) != 1 || visible[0].key != "root" {
		t.Fatalf("collapsed hierarchy visible tasks = %#v", visible)
	}
	if !resolved[0].hasChildren || !resolved[1].hasChildren || !widget.taskCollapsed(&chartState{}, resolved[0]) {
		t.Fatalf("hierarchy metadata = root children=%v child children=%v collapsed=%v", resolved[0].hasChildren, resolved[1].hasChildren, widget.taskCollapsed(&chartState{}, resolved[0]))
	}
}

func TestGanttRejectsInvalidHierarchy(t *testing.T) {
	tests := []func(){
		func() {
			New("chart", []Task{NewTask("child", "Child", ganttDate(1), ganttDate(2)).Parent("missing")}).resolveTasks(themePtr())
		},
		func() {
			New("chart", []Task{NewTask("root", "Root", ganttDate(1), ganttDate(2)).Parent("root")}).resolveTasks(themePtr())
		},
		func() {
			New("chart", []Task{
				NewTask("a", "A", ganttDate(1), ganttDate(2)).Parent("b"),
				NewTask("b", "B", ganttDate(2), ganttDate(3)).Parent("a"),
			}).resolveTasks(themePtr())
		},
	}
	for _, run := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid Gantt hierarchy did not panic")
				}
			}()
			run()
		}()
	}
}

func TestGanttTaskEditHitAndInterval(t *testing.T) {
	widget := New("chart", []Task{NewTask("task", "Task", ganttDate(2), ganttDate(6))})
	geometry := widget.resolveGeometry(widget.resolveTasks(themePtr()), image.Rect(100, 0, 500, 40), ganttDate(1), ganttDate(9), 18, 1)
	item, mode, ok := taskEditHit(f32.Pt(float32(geometry.tasks[0].rect.Min.X+2), float32(geometry.tasks[0].rect.Min.Y+2)), geometry, 7)
	if !ok || item.task.key != "task" || mode != taskEditStart {
		t.Fatalf("start edge hit = %#v, %v, %v", item, mode, ok)
	}
	_, mode, ok = taskEditHit(f32.Pt(float32(geometry.tasks[0].rect.Max.X-2), float32(geometry.tasks[0].rect.Min.Y+2)), geometry, 7)
	if !ok || mode != taskEditEnd {
		t.Fatalf("end edge hit = %v, %v", mode, ok)
	}
	_, mode, ok = taskEditHit(f32.Pt(float32((geometry.tasks[0].rect.Min.X+geometry.tasks[0].rect.Max.X)/2), float32(geometry.tasks[0].rect.Min.Y+2)), geometry, 7)
	if !ok || mode != taskEditMove {
		t.Fatalf("body hit = %v, %v", mode, ok)
	}
	start, end := editedInterval(taskEditMove, ganttDate(2), ganttDate(6), 200, 250, geometry)
	if !start.Equal(ganttDate(3)) || !end.Equal(ganttDate(7)) {
		t.Fatalf("moved interval = %s to %s", start, end)
	}
	start, end = editedInterval(taskEditStart, ganttDate(2), ganttDate(6), 200, 300, geometry)
	if !start.Before(end) {
		t.Fatalf("resized interval is invalid = %s to %s", start, end)
	}
}

func TestGanttClippedTaskEdgeMovesWithoutResizing(t *testing.T) {
	widget := New("chart", []Task{NewTask("task", "Task", ganttDate(1), ganttDate(6))})
	geometry := widget.resolveGeometry(widget.resolveTasks(themePtr()), image.Rect(100, 0, 500, 40), ganttDate(2), ganttDate(7), 18, 1)
	task := geometry.tasks[0]
	if task.unclippedRect.Min.X >= geometry.plot.Min.X || task.rect.Min.X != geometry.plot.Min.X {
		t.Fatalf("task was not clipped at the window start: unclipped=%v visible=%v", task.unclippedRect, task.rect)
	}
	_, mode, ok := taskEditHit(f32.Pt(float32(task.rect.Min.X+2), float32(task.rect.Min.Y+2)), geometry, 7)
	if !ok || mode != taskEditMove {
		t.Fatalf("clipped start edge hit = %v, %v; want move", mode, ok)
	}
	moveMode := mode
	_, mode, ok = taskEditHit(f32.Pt(float32(task.rect.Max.X-2), float32(task.rect.Min.Y+2)), geometry, 7)
	if !ok || mode != taskEditEnd {
		t.Fatalf("visible end edge hit = %v, %v; want end resize", mode, ok)
	}
	start, end := editedInterval(moveMode, task.task.start, task.task.end, 200, 160, geometry)
	if end.Sub(start) != task.task.end.Sub(task.task.start) {
		t.Fatalf("moving clipped task changed duration from %v to %v", task.task.end.Sub(task.task.start), end.Sub(start))
	}
	if !start.Equal(task.task.start) || !end.Equal(task.task.end) {
		t.Fatalf("clipped task moved farther outside the window: %v to %v", start, end)
	}
}

func TestGanttTaskMoveStopsAtVisibleWindowEdges(t *testing.T) {
	widget := New("chart", []Task{NewTask("task", "Task", ganttDate(3), ganttDate(5))})
	geometry := widget.resolveGeometry(widget.resolveTasks(themePtr()), image.Rect(100, 0, 500, 40), ganttDate(1), ganttDate(7), 18, 1)
	originalStart, originalEnd := ganttDate(3), ganttDate(5)

	start, end := editedInterval(taskEditMove, originalStart, originalEnd, 300, -100, geometry)
	if !start.Equal(geometry.start) || end.Sub(start) != originalEnd.Sub(originalStart) {
		t.Fatalf("left-bounded move = %v to %v", start, end)
	}
	start, end = editedInterval(taskEditMove, originalStart, originalEnd, 300, 900, geometry)
	if !end.Equal(geometry.end) || end.Sub(start) != originalEnd.Sub(originalStart) {
		t.Fatalf("right-bounded move = %v to %v", start, end)
	}
}

func TestGanttProgressUsesActualCompletionTimeWhenClipped(t *testing.T) {
	widget := New("chart", []Task{
		NewTask("task", "Task", ganttDate(1), ganttDate(9)).Progress(.5),
	})
	geometry := widget.resolveGeometry(widget.resolveTasks(themePtr()), image.Rect(100, 0, 500, 40), ganttDate(3), ganttDate(7), 18, 1)
	item := geometry.tasks[0]
	if item.rect.Min.X != geometry.plot.Min.X || item.progressRect.Empty() {
		t.Fatalf("clipped task geometry = %#v", item)
	}
	want := int(math.Round(float64(mapTimeUnclamped(ganttDate(5), geometry.start, geometry.end, geometry.plot))))
	if item.progressRect.Max.X != want {
		t.Fatalf("progress endpoint = %d, want %d (rect=%v)", item.progressRect.Max.X, want, item.progressRect)
	}
}

func TestGanttDependencyRoutePointsArrowTowardOverlappingTarget(t *testing.T) {
	geometry := geometry{plot: image.Rect(0, 0, 400, 100)}
	source := taskGeometry{rect: image.Rect(180, 10, 300, 30)}
	target := taskGeometry{rect: image.Rect(120, 50, 240, 70)}
	from, to, _, forward := dependencyRoute(source, target, geometry.plot)
	if forward || int(from.X) != source.rect.Min.X || int(to.X) != target.rect.Max.X {
		t.Fatalf("backward dependency route = from %v to %v forward=%v", from, to, forward)
	}
}

func TestGanttDoubleClickResetsWindowWithoutTaskClickCallback(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	requested := false
	widget := New("chart", []Task{NewTask("task", "Task", ganttDate(1), ganttDate(8))}).
		Animation(false).
		TimeWindow(ganttDate(2), ganttDate(6)).
		OnTimeWindowChange(func(start, end time.Time) {
			requested = start.Equal(ganttDate(1).Add(-24*time.Hour)) && end.After(ganttDate(8))
		})
	layoutGanttFrame(ctx, router, widget, time.Unix(1, 0))
	queueGanttClick(router, 1, f32.Pt(300, 80))
	layoutGanttFrame(ctx, router, widget, time.Unix(1, 1))
	queueGanttClick(router, 1, f32.Pt(300, 80))
	layoutGanttFrame(ctx, router, widget, time.Unix(1, 2))
	if !requested {
		t.Fatal("Gantt double-click did not request a full time window without OnTaskClick")
	}
}

func TestGanttVirtualizesRowsAndScrollsVisibleWindow(t *testing.T) {
	tasks := make([]Task, 100)
	for index := range tasks {
		start := ganttDate(1).AddDate(0, 0, index)
		tasks[index] = NewTask(fmt.Sprintf("task-%d", index), fmt.Sprintf("Task %d", index), start, start.Add(12*time.Hour))
	}
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	widget := New("chart", tasks).Animation(false)
	layoutGanttFrame(ctx, router, widget, time.Unix(3, 0))
	state, ok := frame.PeekState[chartState](ctx, "chart", "gantt-chart")
	if !ok || state == nil {
		t.Fatal("Gantt row state was not retained")
	}
	if state.rowList.Position.Count >= len(tasks) || state.rowList.Position.Length < len(tasks)*20 {
		t.Fatalf("Gantt rows were not virtualized: position=%#v", state.rowList.Position)
	}
	before := state.rowList.Position
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(400, 150)})
	layoutGanttFrame(ctx, router, widget, time.Unix(3, 1))
	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(400, 150), Scroll: f32.Pt(0, -120)})
	layoutGanttFrame(ctx, router, widget, time.Unix(3, 2))
	if state.rowList.Position.First == before.First && state.rowList.Position.Offset == before.Offset {
		t.Fatalf("Gantt row list did not scroll: before=%#v after=%#v", before, state.rowList.Position)
	}
}

func TestGanttVisibleRowRangeClampsStaleScrollPosition(t *testing.T) {
	first, last := visibleRowRange(layout.Position{First: 80, Count: 8}, 3)
	if first != 0 || last != 3 {
		t.Fatalf("stale row range = %d:%d, want 0:3", first, last)
	}
	first, last = visibleRowRange(layout.Position{First: 2, Count: 3}, 10)
	if first != 2 || last != 5 {
		t.Fatalf("normal row range = %d:%d, want 2:5", first, last)
	}
}

func TestGanttFitsCompactTaskListsWithoutClipping(t *testing.T) {
	if got := resolvedRowHeight(242, 8, 36, 24); got != 30 {
		t.Fatalf("compact row height = %d, want 30", got)
	}
	if got := resolvedRowHeight(242, 100, 36, 24); got != 36 {
		t.Fatalf("virtualized row height = %d, want 36", got)
	}
}

func TestGanttTimeWindowWheelKeepsPriorityOverRowScroll(t *testing.T) {
	tasks := make([]Task, 10)
	for index := range tasks {
		start := ganttDate(1).AddDate(0, 0, index)
		tasks[index] = NewTask(fmt.Sprintf("task-%d", index), fmt.Sprintf("Task %d", index), start, start.Add(12*time.Hour))
	}
	windowChanges := 0
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	widget := New("chart", tasks).Animation(false).OnTimeWindowChange(func(time.Time, time.Time) { windowChanges++ })
	layoutGanttFrame(ctx, router, widget, time.Unix(4, 0))
	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(400, 150), Scroll: f32.Pt(0, -1)})
	layoutGanttFrame(ctx, router, widget, time.Unix(4, 1))
	if windowChanges == 0 {
		t.Fatal("Gantt row virtualization consumed the time-window wheel gesture")
	}
}

func layoutGanttFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(520, 320)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func queueGanttClick(router *input.Router, id pointer.ID, position f32.Point) {
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: id, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: id, Position: position},
	)
}

func TestGanttLocalCollapseOverride(t *testing.T) {
	task := NewTask("root", "Root", ganttDate(1), ganttDate(2))
	state := &chartState{localCollapsed: map[string]bool{"root": true}}
	if !New("chart", nil).taskCollapsed(state, resolvedTask{Task: task}) {
		t.Fatal("local collapsed override was ignored")
	}
	state.localCollapsed["root"] = false
	if New("chart", nil).taskCollapsed(state, resolvedTask{Task: task}) {
		t.Fatal("local expanded override was ignored")
	}
}

func TestGanttToggleClickUsesLocalState(t *testing.T) {
	tasks := New("chart", []Task{
		NewTask("root", "Root", ganttDate(1), ganttDate(4)),
		NewTask("child", "Child", ganttDate(2), ganttDate(3)).Parent("root"),
	}).resolveTasks(themePtr())
	widget := New("chart", nil)
	state := &chartState{}
	state.beginToggleFrame()
	state.toggleItem("root").Click()
	if !widget.handleTaskToggleClicks(layout.Context{}, state, tasks, true) || !state.localCollapsed["root"] {
		t.Fatalf("toggle click did not collapse root: %#v", state.localCollapsed)
	}
	state.endToggleFrame()
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
