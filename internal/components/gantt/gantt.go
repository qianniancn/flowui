package gantt

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/chart"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/tooltip"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

// Task is one scheduled interval. Construct it with NewTask.
type Task struct {
	key                        string
	label                      string
	group                      string
	start, end                 time.Time
	baselineStart, baselineEnd time.Time
	hasBaseline                bool
	progress                   float32
	hasProgress                bool
	color                      color.NRGBA
	hasColor                   bool
	dependsOn                  []string
	parentKey                  string
	collapsed                  bool
	milestone                  bool
}

func NewTask(key, label string, start, end time.Time) Task {
	if key == "" {
		panic("flowui: empty gantt task key")
	}
	if start.IsZero() || end.IsZero() || !end.After(start) {
		panic("flowui: gantt task end must be after start")
	}
	if label == "" {
		label = key
	}
	return Task{key: key, label: label, start: start, end: end}
}

// NewMilestone creates a zero-duration project milestone.
func NewMilestone(key, label string, at time.Time) Task {
	if key == "" {
		panic("flowui: empty gantt task key")
	}
	if at.IsZero() {
		panic("flowui: gantt milestone time is required")
	}
	if label == "" {
		label = key
	}
	return Task{key: key, label: label, start: at, end: at, milestone: true}
}

func (t Task) Group(value string) Task      { t.group = value; return t }
func (t Task) Color(value color.NRGBA) Task { t.color, t.hasColor = value, true; return t }

// Baseline adds the planned interval behind the current task interval.
func (t Task) Baseline(start, end time.Time) Task {
	if t.milestone || start.IsZero() || end.IsZero() || !end.After(start) {
		panic("flowui: gantt baseline end must be after start")
	}
	t.baselineStart, t.baselineEnd, t.hasBaseline = start, end, true
	return t
}
func (t Task) Progress(value float32) Task {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 || value > 1 {
		panic("flowui: gantt task progress must be between 0 and 1")
	}
	t.progress, t.hasProgress = value, true
	return t
}

// TimeMarker is a labeled vertical reference on the chart time axis.
type TimeMarker struct {
	at    time.Time
	label string
	color color.NRGBA
}

func NewTimeMarker(at time.Time) TimeMarker {
	if at.IsZero() {
		panic("flowui: gantt marker time is required")
	}
	return TimeMarker{at: at}
}
func (m TimeMarker) Text(value string) TimeMarker       { m.label = value; return m }
func (m TimeMarker) Color(value color.NRGBA) TimeMarker { m.color = value; return m }
func (t Task) DependsOn(keys ...string) Task            { t.dependsOn = append([]string(nil), keys...); return t }

// Parent assigns a parent task for hierarchical schedules.
func (t Task) Parent(key string) Task {
	if key == "" {
		panic("flowui: gantt parent task key must not be empty")
	}
	t.parentKey = key
	return t
}

// Collapsed controls whether this task's descendants are visible.
func (t Task) Collapsed(value bool) Task { t.collapsed = value; return t }

type Selection struct {
	Key       string
	Label     string
	Group     string
	Parent    string
	Depth     int
	Start     time.Time
	End       time.Time
	Progress  float32
	Color     color.NRGBA
	Milestone bool
	Collapsed bool
}

// TaskChange is emitted after a task interval is moved or resized.
// The widget remains immutable; the owner should rebuild the task list with
// the returned interval.
type TaskChange struct {
	Key           string
	Start         time.Time
	End           time.Time
	PreviousStart time.Time
	PreviousEnd   time.Time
}

type Widget struct {
	key                string
	tasks              []Task
	height             unit.Dp
	showGrid           bool
	showTooltip        bool
	showDependencies   bool
	showLegend         bool
	showTaskLabels     bool
	hiddenGroups       []string
	onLegendChange     func(string, bool)
	onTaskToggle       func(string, bool)
	onTaskChange       func(TaskChange)
	taskEditing        bool
	disabled           bool
	label              string
	emptyText          string
	taskAxis           string
	timeAxis           string
	markers            []TimeMarker
	timeStart, timeEnd time.Time
	hasTimeRange       bool
	tickCount          int
	rowHeight          unit.Dp
	formatTime         func(time.Time) string
	animation          bool
	animationDuration  time.Duration
	animationEasing    animation.Easing
	updateDuration     time.Duration
	updateEasing       animation.Easing
	timeWindowStart    time.Time
	timeWindowEnd      time.Time
	hasTimeWindow      bool
	onTimeWindowChange func(time.Time, time.Time)
	onTaskClick        func(Selection)
	tooltipContent     func(Selection) frame.Widget
	customStyle        flowstyle.Style
}

func New(key string, tasks []Task) Widget {
	if key == "" {
		panic("flowui: empty gantt chart key")
	}
	copyTasks := make([]Task, len(tasks))
	for index, task := range tasks {
		copyTasks[index] = task
		copyTasks[index].dependsOn = append([]string(nil), task.dependsOn...)
	}
	return Widget{key: key, tasks: copyTasks, showGrid: true, showTooltip: true, showDependencies: true, tickCount: 6, animation: true, animationDuration: 700 * time.Millisecond, animationEasing: animation.EaseCubicOut, updateDuration: 450 * time.Millisecond, updateEasing: animation.EaseCubicInOut}
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: gantt chart height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}
func (w Widget) Grid(show bool) Widget         { w.showGrid = show; return w }
func (w Widget) Tooltip(show bool) Widget      { w.showTooltip = show; return w }
func (w Widget) Dependencies(show bool) Widget { w.showDependencies = show; return w }
func (w Widget) Legend(show bool) Widget       { w.showLegend = show; return w }
func (w Widget) TaskLabels(show bool) Widget   { w.showTaskLabels = show; return w }
func (w Widget) TaskAxis(value string) Widget  { w.taskAxis = value; return w }
func (w Widget) TimeAxis(value string) Widget  { w.timeAxis = value; return w }

// HiddenGroups controls which grouped tasks are omitted from the plot.
func (w Widget) HiddenGroups(groups ...string) Widget {
	w.hiddenGroups = append([]string(nil), groups...)
	return w
}

// OnLegendChange registers a controlled group visibility callback.
func (w Widget) OnLegendChange(fn func(string, bool)) Widget {
	w.onLegendChange = fn
	return w
}

// OnTaskToggle registers a callback for parent expand/collapse changes.
// Without a callback the component keeps the collapsed state locally.
func (w Widget) OnTaskToggle(fn func(string, bool)) Widget {
	w.onTaskToggle = fn
	return w
}

// Editable enables pointer editing when OnTaskChange is also configured.
func (w Widget) Editable(enabled bool) Widget {
	w.taskEditing = enabled
	return w
}

// OnTaskChange registers the commit callback for task drag/resize edits.
func (w Widget) OnTaskChange(fn func(TaskChange)) Widget {
	w.onTaskChange = fn
	return w
}
func (w Widget) Marker(value TimeMarker) Widget {
	w.markers = append(append([]TimeMarker(nil), w.markers...), value)
	return w
}
func (w Widget) TimeRange(start, end time.Time) Widget {
	if start.IsZero() || end.IsZero() || !end.After(start) {
		panic("flowui: gantt chart time range end must be after start")
	}
	w.timeStart, w.timeEnd, w.hasTimeRange = start, end, true
	return w
}

// TimeWindow sets the initially visible portion of the configured time range.
// It is controlled when OnTimeWindowChange is registered.
func (w Widget) TimeWindow(start, end time.Time) Widget {
	if start.IsZero() || end.IsZero() || !end.After(start) {
		panic("flowui: gantt time window end must be after start")
	}
	w.timeWindowStart, w.timeWindowEnd, w.hasTimeWindow = start, end, true
	return w
}

// OnTimeWindowChange enables wheel zoom, drag pan, and double-click reset.
func (w Widget) OnTimeWindowChange(fn func(time.Time, time.Time)) Widget {
	w.onTimeWindowChange = fn
	return w
}
func (w Widget) TimeTicks(count int) Widget {
	if count < 2 {
		panic("flowui: gantt chart time tick count must be at least 2")
	}
	w.tickCount = count
	return w
}
func (w Widget) RowHeight(dp int) Widget {
	if dp <= 0 {
		panic("flowui: gantt chart row height must be positive")
	}
	w.rowHeight = unit.Dp(dp)
	return w
}
func (w Widget) FormatTime(fn func(time.Time) string) Widget { w.formatTime = fn; return w }

// Animation enables the G2-style task entrance and update transition.
func (w Widget) Animation(enabled bool) Widget { w.animation = enabled; return w }
func (w Widget) AnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "gantt chart")
	w.animationDuration = duration
	return w
}
func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "gantt chart")
	w.animationEasing = easing
	return w
}
func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "gantt chart update")
	w.updateDuration = duration
	return w
}
func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "gantt chart update")
	w.updateEasing = easing
	return w
}
func (w Widget) Label(value string) Widget             { w.label = value; return w }
func (w Widget) EmptyText(value string) Widget         { w.emptyText = value; return w }
func (w Widget) OnTaskClick(fn func(Selection)) Widget { w.onTaskClick = fn; return w }
func (w Widget) TooltipContent(fn func(Selection) frame.Widget) Widget {
	w.tooltipContent = fn
	return w
}
func (w Widget) Disabled(value bool) Widget         { w.disabled = value; return w }
func (w Widget) Style(value flowstyle.Style) Widget { w.customStyle = value; return w }

type resolvedTask struct {
	Task
	color       color.NRGBA
	progress    float32
	depth       int
	hasChildren bool
	parentIndex int
}
type timeTick struct {
	at    time.Time
	label string
	pixel float32
}
type taskGeometry struct {
	task          resolvedTask
	rect          image.Rectangle
	unclippedRect image.Rectangle
	progressRect  image.Rectangle
	baseline      image.Rectangle
	index         int
	rowTop        int
}
type geometry struct {
	plot       image.Rectangle
	start, end time.Time
	ticks      []timeTick
	tasks      []taskGeometry
	rowBand    float32
}
type chartState struct {
	click        gesture.Click
	pointerTag   struct{}
	rowScrollTag struct{}
	// rowList owns the vertical viewport; only rows laid out by it are drawn.
	rowList               layout.List
	rowBar                widget.Scrollbar
	hovered               bool
	pointer               f32.Point
	windowGesture         chart.DataWindowGesture
	timeWindow            chart.DataWindow
	timeWindowReady       bool
	timeWindowConfigStart time.Time
	timeWindowConfigEnd   time.Time
	animationReady        bool
	animationRevision     uint64
	animationSignature    string
	// resolvedTasksCache avoids rebuilding hierarchy and dependency metadata on
	// every animation, hover, or scroll frame. The content signature remains
	// stable when a declarative view rebuilds an equivalent task slice each frame.
	resolvedTasksCache          []resolvedTask
	resolvedTasksReady          bool
	resolvedTasksLen            int
	resolvedTaskInputSignature  uint64
	resolvedThemeSignature      uint64
	resolvedTasksRevision       uint64
	visibleTasksCache           []resolvedTask
	visibleTasksReady           bool
	visibleTasksResolvedRev     uint64
	visibleTasksVisibilityRev   uint64
	visibleTasksHiddenSig       uint64
	visibilityRevision          uint64
	legendGroupsCache           []resolvedTask
	legendGroupsReady           bool
	legendGroupsResolvedRev     uint64
	taskLabelWidthReady         bool
	taskLabelWidthValue         int
	taskLabelWidthResolvedRev   uint64
	taskLabelWidthVisibilityRev uint64
	taskLabelWidthHiddenSig     uint64
	taskLabelWidthMax           int
	taskLabelWidthMetric        unit.Metric
	taskLabelWidthTextPx        int
	legendItems                 map[string]*chart.LegendItem
	legendFrame                 map[string]struct{}
	localHiddenGroups           map[string]bool
	localCollapsed              map[string]bool
	toggleItems                 map[string]*chart.LegendItem
	toggleFrame                 map[string]struct{}
	editTag                     struct{}
	editMode                    taskEditMode
	editTaskKey                 string
	editPointerID               pointer.ID
	editOriginStart             time.Time
	editOriginEnd               time.Time
	editAnchorX                 float32
	editPreviewStart            time.Time
	editPreviewEnd              time.Time
	editDragging                bool
	clickPressValid             bool
	clickPressSelection         Selection
	tooltipTransition           tooltip.PopupTransition
	tooltipSelection            Selection
}

func stateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindGanttChart, key)
	return frame.UseState[chartState](ctx, key, "gantt-chart")
}

func (s *chartState) beginLegendFrame() { stateutil.BeginFrameMap(&s.legendFrame) }
func (s *chartState) endLegendFrame()   { stateutil.SweepFrameMap(s.legendItems, s.legendFrame) }
func (s *chartState) legendItem(key string) *chart.LegendItem {
	return stateutil.UseFrameMap(&s.legendItems, &s.legendFrame, key)
}
func (s *chartState) beginToggleFrame() { stateutil.BeginFrameMap(&s.toggleFrame) }
func (s *chartState) endToggleFrame()   { stateutil.SweepFrameMap(s.toggleItems, s.toggleFrame) }
func (s *chartState) toggleItem(key string) *chart.LegendItem {
	return stateutil.UseFrameMap(&s.toggleItems, &s.toggleFrame, key)
}

type taskEditMode uint8

const (
	taskEditNone taskEditMode = iota
	taskEditMove
	taskEditStart
	taskEditEnd
)

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(frame.WidgetFunc(w.layout)).Style(w.customStyle).Layout(ctx, gtx)
}

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := stateFor(ctx, w.key)
	state.beginLegendFrame()
	state.beginToggleFrame()
	defer func() {
		state.endToggleFrame()
		state.endLegendFrame()
	}()
	enabled := gtx.Enabled() && !w.disabled
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.GanttChart
	allTasks := w.resolveTasksCached(state, activeTheme)
	if w.handleLegendClicks(gtx, state, allTasks, enabled) {
		state.hovered = false
		if gtx.Ops != nil {
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	tasks := w.visibleTasksCached(state, allTasks)
	rowHeight := gtx.Dp(tokens.RowHeight)
	if w.rowHeight > 0 {
		rowHeight = gtx.Dp(w.rowHeight)
	}
	height := gtx.Dp(tokens.Height)
	if w.height > 0 {
		height = gtx.Dp(w.height)
	}
	height = max(height, gtx.Dp(tokens.PlotPaddingTop)+gtx.Dp(tokens.PlotPaddingBottom)+rowHeight)
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, height))
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}
	if len(allTasks) == 0 {
		w.drawEmpty(ctx, gtx, image.Rect(0, 0, size.X, size.Y))
		return layout.Dimensions{Size: size}
	}
	left := w.taskLabelWidth(ctx, gtx, state, tasks, max(size.X/2, 1)) + max(gtx.Dp(tokens.TickLabelGap), 0) + 6
	left = max(left, gtx.Dp(tokens.PlotPaddingLeft))
	left = min(left, max(size.X/2, 0))
	right := min(max(gtx.Dp(tokens.PlotPaddingRight), 0), max(size.X-left-1, 0))
	plotTop := max(gtx.Dp(tokens.PlotPaddingTop), 0)
	style := ganttStyle(activeTheme, !enabled)
	legend := recordedBlock{}
	if w.showLegend {
		legend = w.recordLegend(ctx, gtx, state, allTasks, style, max(size.X-left-right, 1), enabled)
		plotTop += legend.dims.Size.Y
		if legend.dims.Size.Y > 0 {
			plotTop += max(gtx.Dp(tokens.LegendGap), 0)
		}
	}
	if len(w.markers) > 0 {
		plotTop += gtx.Sp(tokens.AxisTextSize) + max(gtx.Dp(tokens.MarkerLabelGap), 0)
	}
	if w.taskAxis != "" {
		plotTop += gtx.Sp(tokens.AxisTextSize) + max(gtx.Dp(tokens.AxisNameGap), 0)
	}
	plotBottom := size.Y - gtx.Dp(tokens.PlotPaddingBottom)
	if w.timeAxis != "" {
		plotBottom -= gtx.Sp(tokens.AxisTextSize) + max(gtx.Dp(tokens.AxisNameGap), 0)
	}
	plot := image.Rect(left, plotTop, max(left+1, size.X-right), max(plotTop+1, plotBottom))
	fullStart, fullEnd := w.timeExtent(allTasks)
	window := w.effectiveTimeWindow(state, fullStart, fullEnd)
	start, end := mapTimeWindow(fullStart, fullEnd, window)
	windowChanged := func(next chart.DataWindow) {
		state.timeWindow, state.timeWindowReady = next, true
		if w.onTimeWindowChange != nil {
			w.onTimeWindowChange(unmapTime(next.Start, fullStart, fullEnd), unmapTime(next.End, fullStart, fullEnd))
		}
	}
	reveal := w.animationProgress(ctx, gtx, state, tasks, enabled)
	listGtx := gtx
	listGtx.Constraints = layout.Exact(image.Pt(max(plot.Dx(), 1), max(plot.Dy(), rowHeight)))
	visibleRowHeight := rowHeight
	if w.rowHeight <= 0 {
		visibleRowHeight = resolvedRowHeight(plot.Dy(), len(tasks), rowHeight, max(gtx.Dp(tokens.BarHeight)+6, 1))
	}
	rowsScrollable := enabled && len(tasks) > 1 && w.onTimeWindowChange == nil && visibleRowHeight*len(tasks) > plot.Dy()
	w.updateRowScroll(gtx, state, visibleRowHeight, rowsScrollable)
	state.rowList.Axis = layout.Vertical
	state.rowList.Gap = 0
	rowOffsetOp := op.Offset(plot.Min).Push(gtx.Ops)
	layoutui.LayoutTrackedScrollbar(ctx, listGtx, &state.rowList, &state.rowBar, len(tasks), !enabled, true, func(rowGtx layout.Context, _ int) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(max(plot.Dx(), 1), visibleRowHeight)}
	})
	rowOffsetOp.Pop()
	firstRow, lastRow := visibleRowRange(state.rowList.Position, len(tasks))
	visibleTasks := tasks[firstRow:lastRow]
	geo := w.resolveVisibleGeometry(visibleTasks, plot, start, end, gtx.Dp(tokens.BarHeight), visibleRowHeight, firstRow, state.rowList.Position.Offset, len(tasks), reveal)
	editingPointer := w.updateTaskEditing(gtx, state, geo, enabled)
	if state.editDragging {
		geo = applyEditPreview(geo, state)
	}
	var onWindowChange func(chart.DataWindow)
	if enabled && w.onTimeWindowChange != nil && !editingPointer && !state.editDragging && state.editTaskKey == "" {
		onWindowChange = windowChanged
	}
	chart.UpdatePointer(gtx, enabled, plot, window, false, onWindowChange, &state.pointerTag, &state.hovered, &state.pointer, &state.windowGesture)
	selection, selected := selectionAt(state.pointer, state.hovered, geo)
	activated, resetWindow := w.updateTaskClicks(gtx, enabled, state, geo)
	if resetWindow && onWindowChange != nil {
		windowChanged(chart.FullDataWindow())
	}
	if activated && w.onTaskClick != nil {
		if state.clickPressValid && state.clickPressSelection.Key != "" {
			w.onTaskClick(state.clickPressSelection)
		}
	}
	if w.handleTaskToggleClicks(gtx, state, allTasks, enabled) {
		state.hovered = false
		if gtx.Ops != nil {
			gtx.Execute(op.InvalidateCmd{})
		}
	}
	tooltipVisible := enabled && w.showTooltip && selected
	if tooltipVisible {
		state.tooltipSelection = exportSelection(selection.task)
	}
	progress := state.tooltipTransition.Progress(gtx, tooltipVisible, activeTheme.Motion)
	if !tooltipVisible && progress <= 0 {
		state.tooltipSelection = Selection{}
	}

	opacity := paint.PushOpacity(gtx.Ops, style.opacity)
	if legend.dims.Size.Y > 0 {
		placeBlock(gtx, legend, image.Pt(left, max(gtx.Dp(tokens.PlotPaddingTop), 0)))
	}
	w.drawGrid(gtx, geo, style, tokens)
	w.drawMarkers(ctx, gtx, geo, style, tokens)
	if w.showDependencies {
		drawDependencies(gtx, geo, style, tokens)
	}
	w.drawTasks(ctx, gtx, geo, selection.index, selected, style, tokens)
	w.drawLabels(ctx, gtx, state, geo, style, tokens, enabled)
	w.drawAxisTitles(ctx, gtx, geo, style, tokens)
	if len(tasks) == 0 {
		emptyText := w.emptyText
		if emptyText == "" {
			emptyText = "No visible tasks"
		}
		w.drawEmpty(ctx, gtx, geo.plot, emptyText)
	}
	if tooltipVisible || progress > 0 {
		w.drawTooltip(ctx, gtx, state.tooltipSelection, chart.TooltipAnchor(state.pointer), progress, state.tooltipTransition.Exiting())
	}
	opacity.Pop()
	pointerEnabled := enabled && (w.showTooltip || w.onTaskClick != nil || w.onTimeWindowChange != nil)
	pointerCursor := ganttPointerCursor(state, selected && w.onTaskClick != nil)
	chart.AddPointerInputWithCursor(gtx, plot, pointerEnabled, pointerCursor, &state.pointerTag)
	w.addTaskEditingInput(gtx, plot, enabled && w.taskEditing && w.onTaskChange != nil, state)
	w.addRowScrollInput(gtx, plot, rowsScrollable, state)
	chart.AddClickInput(gtx, size, enabled && (w.onTaskClick != nil || w.onTimeWindowChange != nil), &state.click)
	return layout.Dimensions{Size: size}
}

func visibleRowRange(position layout.Position, taskCount int) (first, last int) {
	if taskCount <= 0 {
		return 0, 0
	}
	count := min(max(position.Count, 0), taskCount)
	maxFirst := max(taskCount-count, 0)
	first = min(max(position.First, 0), maxFirst)
	last = min(first+count, taskCount)
	return first, last
}

func resolvedRowHeight(plotHeight, taskCount, requested, minimum int) int {
	if taskCount <= 0 {
		return max(requested, 1)
	}
	fit := plotHeight / taskCount
	if fit >= max(minimum, 1) {
		return max(fit, 1)
	}
	return max(requested, 1)
}

func (w Widget) updateRowScroll(gtx layout.Context, state *chartState, rowHeight int, enabled bool) {
	if !enabled || rowHeight <= 0 {
		return
	}
	for {
		value, ok := gtx.Event(pointer.Filter{Target: &state.rowScrollTag, Kinds: pointer.Scroll, ScrollY: pointer.ScrollRange{Min: -100000, Max: 100000}})
		if !ok {
			break
		}
		event, ok := value.(pointer.Event)
		if !ok || event.Scroll.Y == 0 {
			continue
		}
		state.rowList.ScrollBy(-event.Scroll.Y / float32(rowHeight))
	}
}

func (w Widget) addRowScrollInput(gtx layout.Context, plot image.Rectangle, enabled bool, state *chartState) {
	if !enabled || plot.Empty() {
		return
	}
	area := clip.Rect(plot).Push(gtx.Ops)
	event.Op(gtx.Ops, &state.rowScrollTag)
	area.Pop()
}

// updateTaskClicks keeps the task captured on press so a release over another
// task cannot retarget the click callback. The click gesture may span frames.
func (w Widget) updateTaskClicks(gtx layout.Context, enabled bool, state *chartState, geo geometry) (activated, reset bool) {
	for {
		value, ok := state.click.Update(gtx.Source)
		if !ok {
			break
		}
		switch value.Kind {
		case gesture.KindPress:
			state.clickPressValid = true
			state.clickPressSelection = Selection{}
			if task, ok := selectionAt(f32.Pt(float32(value.Position.X), float32(value.Position.Y)), true, geo); ok {
				state.clickPressSelection = exportSelection(task.task)
			}
		case gesture.KindClick:
			state.clickPressValid = true
			activated = activated || enabled
			reset = reset || (enabled && value.NumClicks >= 2)
		case gesture.KindCancel:
			state.clickPressValid = false
			state.clickPressSelection = Selection{}
		}
	}
	return activated, reset
}

func ganttPointerCursor(state *chartState, taskClickable bool) pointer.Cursor {
	if state.windowGesture.Dragging() {
		return pointer.CursorGrabbing
	}
	if taskClickable {
		return pointer.CursorPointer
	}
	return pointer.CursorDefault
}

func (w Widget) updateTaskEditing(gtx layout.Context, state *chartState, geo geometry, enabled bool) bool {
	if !enabled || !w.taskEditing || w.onTaskChange == nil {
		state.editMode = taskEditNone
		state.editTaskKey = ""
		state.editDragging = false
		return false
	}
	handled := state.editTaskKey != "" || state.editDragging
	editEdge := float32(max(gtx.Dp(6), 4))
	for {
		value, ok := gtx.Event(pointer.Filter{Target: &state.editTag, Kinds: pointer.Enter | pointer.Leave | pointer.Move | pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			break
		}
		event, ok := value.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Enter, pointer.Move:
			if state.editDragging {
				continue
			}
			_, state.editMode, _ = taskEditHit(event.Position, geo, editEdge)
			if state.editMode == taskEditNone {
				state.editTaskKey = ""
			} else {
				item, _, _ := taskEditHit(event.Position, geo, editEdge)
				state.editTaskKey = item.task.key
				handled = true
			}
		case pointer.Leave:
			if !state.editDragging {
				state.editMode = taskEditNone
				state.editTaskKey = ""
			}
		case pointer.Press:
			if !event.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			item, mode, ok := taskEditHit(event.Position, geo, editEdge)
			if !ok || item.task.milestone {
				continue
			}
			state.editMode = mode
			state.editTaskKey = item.task.key
			handled = true
			state.editPointerID = event.PointerID
			state.editOriginStart = item.task.start
			state.editOriginEnd = item.task.end
			state.editPreviewStart = item.task.start
			state.editPreviewEnd = item.task.end
			state.editAnchorX = event.Position.X
			state.editDragging = false
			gtx.Execute(pointer.GrabCmd{Tag: &state.editTag, ID: event.PointerID})
		case pointer.Drag:
			if event.PointerID != state.editPointerID || state.editTaskKey == "" {
				continue
			}
			handled = true
			state.editDragging = true
			state.editPreviewStart, state.editPreviewEnd = editedInterval(state.editMode, state.editOriginStart, state.editOriginEnd, state.editAnchorX, event.Position.X, geo)
		case pointer.Release, pointer.Cancel:
			if event.PointerID != state.editPointerID {
				continue
			}
			handled = true
			if event.Kind == pointer.Release && state.editDragging {
				start, end := state.editPreviewStart, state.editPreviewEnd
				if !start.Equal(state.editOriginStart) || !end.Equal(state.editOriginEnd) {
					w.onTaskChange(TaskChange{Key: state.editTaskKey, Start: start, End: end, PreviousStart: state.editOriginStart, PreviousEnd: state.editOriginEnd})
				}
			}
			state.editMode = taskEditNone
			state.editTaskKey = ""
			state.editPointerID = 0
			state.editDragging = false
			state.editPreviewStart = time.Time{}
			state.editPreviewEnd = time.Time{}
		}
	}
	return handled || state.editTaskKey != "" || state.editDragging
}

func taskEditHit(position f32.Point, geo geometry, edge float32) (taskGeometry, taskEditMode, bool) {
	point := position.Round()
	for _, item := range geo.tasks {
		if item.task.milestone || item.rect.Empty() || !point.In(item.rect) {
			continue
		}
		if item.unclippedRect.Min.X >= geo.plot.Min.X && float32(point.X-item.unclippedRect.Min.X) <= edge {
			return item, taskEditStart, true
		}
		if item.unclippedRect.Max.X <= geo.plot.Max.X && float32(item.unclippedRect.Max.X-point.X) <= edge {
			return item, taskEditEnd, true
		}
		return item, taskEditMove, true
	}
	return taskGeometry{}, taskEditNone, false
}

func editedInterval(mode taskEditMode, originalStart, originalEnd time.Time, anchorX, currentX float32, geo geometry) (time.Time, time.Time) {
	delta := timeAtPixel(currentX, geo).Sub(timeAtPixel(anchorX, geo))
	start, end := originalStart, originalEnd
	switch mode {
	case taskEditMove:
		minimumStart := geo.start
		if originalStart.Before(minimumStart) {
			minimumStart = originalStart
		}
		maximumEnd := geo.end
		if originalEnd.After(maximumEnd) {
			maximumEnd = originalEnd
		}
		delta = min(max(delta, minimumStart.Sub(originalStart)), maximumEnd.Sub(originalEnd))
		start, end = start.Add(delta), end.Add(delta)
	case taskEditStart:
		start = start.Add(delta)
		if !start.Before(end) {
			start = end.Add(-time.Nanosecond)
		}
	case taskEditEnd:
		end = end.Add(delta)
		if !end.After(start) {
			end = start.Add(time.Nanosecond)
		}
	}
	return start, end
}

func timeAtPixel(x float32, geo geometry) time.Time {
	if geo.plot.Dx() <= 0 || !geo.end.After(geo.start) {
		return geo.start
	}
	ratio := (float64(x) - float64(geo.plot.Min.X)) / float64(geo.plot.Dx())
	ratio = min(max(ratio, 0), 1)
	return geo.start.Add(time.Duration(float64(geo.end.Sub(geo.start)) * ratio))
}

func applyEditPreview(geo geometry, state *chartState) geometry {
	if !state.editDragging || state.editTaskKey == "" {
		return geo
	}
	for index := range geo.tasks {
		item := &geo.tasks[index]
		if item.task.key != state.editTaskKey || item.rect.Empty() {
			continue
		}
		item.task.start, item.task.end = state.editPreviewStart, state.editPreviewEnd
		left := mapTimeUnclamped(item.task.start, geo.start, geo.end, geo.plot)
		right := mapTimeUnclamped(item.task.end, geo.start, geo.end, geo.plot)
		item.unclippedRect = image.Rect(int(math.Round(float64(left))), item.rect.Min.Y, int(math.Round(float64(right))), item.rect.Max.Y)
		item.rect = item.unclippedRect.Intersect(geo.plot)
		item.progressRect = progressGeometry(item.task, item.rect, geo.start, geo.end, geo.plot)
	}
	return geo
}

func (w Widget) addTaskEditingInput(gtx layout.Context, plot image.Rectangle, enabled bool, state *chartState) {
	if !enabled || plot.Empty() {
		return
	}
	area := clip.Rect(plot).Push(gtx.Ops)
	if cursor, ok := taskEditingCursor(state); ok {
		cursor.Add(gtx.Ops)
	}
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &state.editTag)
	pass.Pop()
	area.Pop()
}

func taskEditingCursor(state *chartState) (pointer.Cursor, bool) {
	if state.editMode == taskEditStart || state.editMode == taskEditEnd {
		return pointer.CursorEastWestResize, true
	}
	if state.editDragging {
		return pointer.CursorGrabbing, true
	}
	return pointer.CursorDefault, false
}

func (w Widget) animationProgress(ctx *frame.Context, gtx layout.Context, state *chartState, tasks []resolvedTask, enabled bool) float32 {
	if !w.animation || len(tasks) == 0 {
		return 1
	}
	// The resolved-task revision changes only when the immutable task source or
	// its theme-derived colors change. This keeps animation frames independent
	// of the number of tasks; retain the content signature as a fallback for
	// direct callers that do not go through Layout.
	signature := ""
	if state.resolvedTasksReady {
		signature = fmt.Sprintf("%d/%d", state.resolvedTasksRevision, state.visibilityRevision)
	} else {
		signature = ganttTaskSignature(tasks)
	}
	if !state.animationReady {
		state.animationReady = true
		state.animationRevision = 1
		state.animationSignature = signature
	} else if state.animationSignature != signature {
		state.animationRevision++
		state.animationSignature = signature
	}
	duration, easing := w.animationDuration, w.animationEasing
	if state.animationRevision > 1 {
		duration, easing = w.updateDuration, w.updateEasing
	}
	progress, _ := animation.Tween(w.key+":gantt-entry", 1).
		Initial(0).
		Revision(state.animationRevision).
		Duration(duration).
		Easing(easing).
		Disabled(!enabled).
		Sample(ctx, gtx)
	return progress
}

func ganttTaskSignature(tasks []resolvedTask) string {
	var signature strings.Builder
	for _, task := range tasks {
		fmt.Fprintf(&signature, "%s|%d|%d|%.4f|%s|%s|%t|%t|%d|%d|%d|%d;", task.key, task.start.UnixNano(), task.end.UnixNano(), task.progress, task.group, task.parentKey, task.collapsed, task.milestone, task.color.R, task.color.G, task.color.B, task.color.A)
	}
	return signature.String()
}

func (w Widget) resolveTasksCached(state *chartState, activeTheme *theme.Theme) []resolvedTask {
	inputSignature := ganttTaskInputSignature(w.tasks)
	themeSignature := ganttSeriesSignature(activeTheme.Components.GanttChart.SeriesColors)
	if state.resolvedTasksReady && state.resolvedTaskInputSignature == inputSignature && state.resolvedTasksLen == len(w.tasks) && state.resolvedThemeSignature == themeSignature {
		return state.resolvedTasksCache
	}
	resolved := w.resolveTasks(activeTheme)
	state.resolvedTasksCache = resolved
	state.resolvedTasksReady = true
	state.resolvedTasksLen = len(w.tasks)
	state.resolvedTaskInputSignature = inputSignature
	state.resolvedThemeSignature = themeSignature
	state.resolvedTasksRevision++
	state.visibleTasksCache = nil
	state.visibleTasksReady = false
	state.legendGroupsCache = nil
	state.legendGroupsReady = false
	state.taskLabelWidthReady = false
	return resolved
}

func (w Widget) visibleTasksCached(state *chartState, tasks []resolvedTask) []resolvedTask {
	hiddenSignature := ganttStringSliceSignature(w.hiddenGroups)
	if state.visibleTasksReady && state.visibleTasksResolvedRev == state.resolvedTasksRevision && state.visibleTasksVisibilityRev == state.visibilityRevision && state.visibleTasksHiddenSig == hiddenSignature {
		return state.visibleTasksCache
	}
	visible := w.visibleTasks(state, tasks)
	state.visibleTasksCache = visible
	state.visibleTasksReady = true
	state.visibleTasksResolvedRev = state.resolvedTasksRevision
	state.visibleTasksVisibilityRev = state.visibilityRevision
	state.visibleTasksHiddenSig = hiddenSignature
	return visible
}

func ganttSeriesSignature(colors [9]color.NRGBA) uint64 {
	hash := uint64(1469598103934665603)
	for _, col := range colors {
		hash = ganttHashByte(hash, col.R)
		hash = ganttHashByte(hash, col.G)
		hash = ganttHashByte(hash, col.B)
		hash = ganttHashByte(hash, col.A)
	}
	return hash
}

func ganttTaskInputSignature(tasks []Task) uint64 {
	hash := uint64(1469598103934665603)
	for _, task := range tasks {
		hash = ganttHashString(hash, task.key)
		hash = ganttHashString(hash, task.label)
		hash = ganttHashString(hash, task.group)
		hash = ganttHashUint64(hash, uint64(task.start.UnixNano()))
		hash = ganttHashUint64(hash, uint64(task.end.UnixNano()))
		hash = ganttHashUint64(hash, uint64(task.baselineStart.UnixNano()))
		hash = ganttHashUint64(hash, uint64(task.baselineEnd.UnixNano()))
		hash = ganttHashUint64(hash, uint64(math.Float32bits(task.progress)))
		hash = ganttHashColor(hash, task.color)
		hash = ganttHashBool(hash, task.hasBaseline)
		hash = ganttHashBool(hash, task.hasProgress)
		hash = ganttHashBool(hash, task.hasColor)
		hash = ganttHashBool(hash, task.collapsed)
		hash = ganttHashBool(hash, task.milestone)
		for _, dependency := range task.dependsOn {
			hash = ganttHashString(hash, dependency)
		}
		hash = ganttHashUint64(hash, uint64(len(task.dependsOn)))
		hash = ganttHashString(hash, task.parentKey)
		hash = ganttHashByte(hash, 0xff)
	}
	return hash
}

func ganttHashColor(hash uint64, value color.NRGBA) uint64 {
	hash = ganttHashByte(hash, value.R)
	hash = ganttHashByte(hash, value.G)
	hash = ganttHashByte(hash, value.B)
	return ganttHashByte(hash, value.A)
}

func ganttHashBool(hash uint64, value bool) uint64 {
	if value {
		return ganttHashByte(hash, 1)
	}
	return ganttHashByte(hash, 0)
}

func ganttHashString(hash uint64, value string) uint64 {
	hash = ganttHashUint64(hash, uint64(len(value)))
	for index := 0; index < len(value); index++ {
		hash = ganttHashByte(hash, value[index])
	}
	return hash
}

func ganttHashUint64(hash, value uint64) uint64 {
	for shift := uint(0); shift < 64; shift += 8 {
		hash = ganttHashByte(hash, byte(value>>shift))
	}
	return hash
}

func ganttStringSliceSignature(values []string) uint64 {
	hash := uint64(1469598103934665603)
	for _, value := range values {
		for index := 0; index < len(value); index++ {
			hash = ganttHashByte(hash, value[index])
		}
		hash = ganttHashByte(hash, 0)
	}
	return hash
}

func ganttHashByte(hash uint64, value byte) uint64 {
	return (hash ^ uint64(value)) * 1099511628211
}

func (w Widget) resolveTasks(activeTheme *theme.Theme) []resolvedTask {
	seen := make(map[string]struct{}, len(w.tasks))
	groupColors := make(map[string]color.NRGBA)
	result := make([]resolvedTask, 0, len(w.tasks))
	for index, task := range w.tasks {
		if task.key == "" || task.start.IsZero() || task.end.IsZero() || (!task.milestone && !task.end.After(task.start)) {
			panic("flowui: invalid gantt task")
		}
		if _, exists := seen[task.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate gantt task key %q", task.key))
		}
		seen[task.key] = struct{}{}
		if task.label == "" {
			task.label = task.key
		}
		col := task.color
		if !task.hasColor {
			if task.group != "" {
				col = groupColors[task.group]
				if col.A == 0 {
					col = activeTheme.Components.GanttChart.SeriesColors[len(groupColors)%len(activeTheme.Components.GanttChart.SeriesColors)]
					groupColors[task.group] = col
				}
			} else {
				col = activeTheme.Components.GanttChart.SeriesColors[index%len(activeTheme.Components.GanttChart.SeriesColors)]
			}
		}
		progress := float32(1)
		if task.hasProgress {
			progress = task.progress
		}
		result = append(result, resolvedTask{Task: task, color: col, progress: progress})
	}
	validateDependencies(result)
	validateHierarchy(result)
	byKey := make(map[string]*resolvedTask, len(result))
	indexByKey := make(map[string]int, len(result))
	for index := range result {
		result[index].parentIndex = -1
		byKey[result[index].key] = &result[index]
		indexByKey[result[index].key] = index
	}
	for index := range result {
		if result[index].parentKey == "" {
			continue
		}
		if parent := byKey[result[index].parentKey]; parent != nil {
			parent.hasChildren = true
			result[index].parentIndex = indexByKey[result[index].parentKey]
		}
	}
	return result
}

func validateHierarchy(tasks []resolvedTask) {
	byKey := make(map[string]*resolvedTask, len(tasks))
	for index := range tasks {
		byKey[tasks[index].key] = &tasks[index]
	}
	for _, task := range tasks {
		if task.parentKey != "" {
			if _, ok := byKey[task.parentKey]; !ok {
				panic(fmt.Sprintf("flowui: gantt task %q has unknown parent %q", task.key, task.parentKey))
			}
			if task.parentKey == task.key {
				panic(fmt.Sprintf("flowui: gantt task %q cannot parent itself", task.key))
			}
		}
	}
	visiting := make(map[string]bool, len(tasks))
	resolved := make(map[string]bool, len(tasks))
	var visit func(*resolvedTask) int
	visit = func(task *resolvedTask) int {
		if resolved[task.key] {
			return task.depth
		}
		if visiting[task.key] {
			panic(fmt.Sprintf("flowui: gantt parent cycle includes task %q", task.key))
		}
		visiting[task.key] = true
		if task.parentKey != "" {
			task.depth = visit(byKey[task.parentKey]) + 1
		}
		delete(visiting, task.key)
		resolved[task.key] = true
		return task.depth
	}
	for index := range tasks {
		visit(&tasks[index])
	}
}

func (w Widget) visibleTasks(state *chartState, tasks []resolvedTask) []resolvedTask {
	if len(tasks) == 0 {
		return nil
	}
	visible := make([]resolvedTask, 0, len(tasks))
	for _, task := range tasks {
		if !w.groupHidden(state, task.group) && !w.parentHidden(state, task, tasks) {
			visible = append(visible, task)
		}
	}
	return visible
}

func (w Widget) parentHidden(state *chartState, task resolvedTask, tasks []resolvedTask) bool {
	for task.parentKey != "" {
		if task.parentIndex < 0 || task.parentIndex >= len(tasks) {
			return true
		}
		parent := tasks[task.parentIndex]
		if w.taskCollapsed(state, parent) {
			return true
		}
		task = parent
	}
	return false
}

func (w Widget) taskCollapsed(state *chartState, task resolvedTask) bool {
	if state.localCollapsed != nil {
		if collapsed, ok := state.localCollapsed[task.key]; ok {
			return collapsed
		}
	}
	return task.collapsed
}

func (w Widget) groupHidden(state *chartState, group string) bool {
	if slices.Contains(w.hiddenGroups, group) {
		return true
	}
	return state.localHiddenGroups != nil && state.localHiddenGroups[group]
}

func (w Widget) legendGroups(tasks []resolvedTask) []resolvedTask {
	entries := make([]resolvedTask, 0)
	seen := make(map[string]struct{})
	for _, task := range tasks {
		if task.group == "" {
			continue
		}
		if _, ok := seen[task.group]; ok {
			continue
		}
		seen[task.group] = struct{}{}
		entries = append(entries, task)
	}
	return entries
}

func (w Widget) legendGroupsCached(state *chartState, tasks []resolvedTask) []resolvedTask {
	if state.legendGroupsReady && state.legendGroupsResolvedRev == state.resolvedTasksRevision {
		return state.legendGroupsCache
	}
	entries := w.legendGroups(tasks)
	state.legendGroupsCache = entries
	state.legendGroupsReady = true
	state.legendGroupsResolvedRev = state.resolvedTasksRevision
	return entries
}

func (w Widget) handleLegendClicks(gtx layout.Context, state *chartState, tasks []resolvedTask, enabled bool) bool {
	changed := false
	for _, entry := range w.legendGroupsCached(state, tasks) {
		if !state.legendItem(entry.group).Clicked(gtx) {
			continue
		}
		hidden := !w.groupHidden(state, entry.group)
		if w.onLegendChange != nil {
			w.onLegendChange(entry.group, hidden)
		} else if enabled {
			if state.localHiddenGroups == nil {
				state.localHiddenGroups = make(map[string]bool)
			}
			if hidden {
				state.localHiddenGroups[entry.group] = true
			} else {
				delete(state.localHiddenGroups, entry.group)
			}
		}
		changed = true
	}
	if changed {
		state.visibilityRevision++
	}
	return changed
}

func (w Widget) handleTaskToggleClicks(gtx layout.Context, state *chartState, tasks []resolvedTask, enabled bool) bool {
	if !enabled {
		return false
	}
	changed := false
	for _, task := range tasks {
		if !task.hasChildren || !state.toggleItem(task.key).Clicked(gtx) {
			continue
		}
		collapsed := w.taskCollapsed(state, task)
		next := !collapsed
		if w.onTaskToggle != nil {
			w.onTaskToggle(task.key, next)
		} else {
			if state.localCollapsed == nil {
				state.localCollapsed = make(map[string]bool)
			}
			state.localCollapsed[task.key] = next
		}
		changed = true
	}
	if changed {
		state.visibilityRevision++
	}
	return changed
}

func validateDependencies(tasks []resolvedTask) {
	byKey := make(map[string]resolvedTask, len(tasks))
	for _, task := range tasks {
		byKey[task.key] = task
	}
	for _, task := range tasks {
		seenDependencies := make(map[string]struct{}, len(task.dependsOn))
		for _, dependency := range task.dependsOn {
			if dependency == "" {
				panic(fmt.Sprintf("flowui: gantt task %q has an empty dependency", task.key))
			}
			if _, ok := byKey[dependency]; !ok {
				panic(fmt.Sprintf("flowui: gantt task %q depends on unknown task %q", task.key, dependency))
			}
			if dependency == task.key {
				panic(fmt.Sprintf("flowui: gantt task %q cannot depend on itself", task.key))
			}
			if _, duplicate := seenDependencies[dependency]; duplicate {
				panic(fmt.Sprintf("flowui: gantt task %q lists dependency %q more than once", task.key, dependency))
			}
			seenDependencies[dependency] = struct{}{}
		}
	}
	visiting := make(map[string]bool, len(tasks))
	visited := make(map[string]bool, len(tasks))
	var visit func(string)
	visit = func(key string) {
		if visited[key] {
			return
		}
		if visiting[key] {
			panic(fmt.Sprintf("flowui: gantt dependency cycle includes task %q", key))
		}
		visiting[key] = true
		for _, dependency := range byKey[key].dependsOn {
			visit(dependency)
		}
		delete(visiting, key)
		visited[key] = true
	}
	for _, task := range tasks {
		visit(task.key)
	}
}

func (w Widget) timeExtent(tasks []resolvedTask) (time.Time, time.Time) {
	if w.hasTimeRange {
		return w.timeStart, w.timeEnd
	}
	start, end := tasks[0].start, tasks[0].end
	for _, task := range tasks[1:] {
		if task.start.Before(start) {
			start = task.start
		}
		if task.end.After(end) {
			end = task.end
		}
	}
	padding := max(end.Sub(start)/20, 24*time.Hour)
	return start.Add(-padding), end.Add(padding)
}

func (w Widget) effectiveTimeWindow(state *chartState, fullStart, fullEnd time.Time) chart.DataWindow {
	if w.hasTimeWindow {
		window := normalizeTimeWindow(w.timeWindowStart, w.timeWindowEnd, fullStart, fullEnd)
		if !state.timeWindowReady || !state.timeWindowConfigStart.Equal(w.timeWindowStart) || !state.timeWindowConfigEnd.Equal(w.timeWindowEnd) {
			state.timeWindow, state.timeWindowReady = window, true
			state.timeWindowConfigStart, state.timeWindowConfigEnd = w.timeWindowStart, w.timeWindowEnd
		}
		return state.timeWindow
	}
	if state.timeWindowReady {
		return state.timeWindow
	}
	window := chart.FullDataWindow()
	state.timeWindow, state.timeWindowReady = window, true
	return window
}

func normalizeTimeWindow(start, end, fullStart, fullEnd time.Time) chart.DataWindow {
	span := fullEnd.Sub(fullStart)
	if span <= 0 {
		return chart.FullDataWindow()
	}
	startRatio := float64(start.Sub(fullStart)) / float64(span)
	endRatio := float64(end.Sub(fullStart)) / float64(span)
	if endRatio <= 0 || startRatio >= 1 {
		panic("flowui: gantt time window must overlap the chart time range")
	}
	startRatio = min(max(startRatio, 0), 1)
	endRatio = min(max(endRatio, 0), 1)
	if endRatio <= startRatio {
		panic("flowui: gantt time window must overlap the chart time range")
	}
	return chart.NewDataWindow(startRatio, endRatio)
}

func mapTimeWindow(fullStart, fullEnd time.Time, window chart.DataWindow) (time.Time, time.Time) {
	return unmapTime(window.Start, fullStart, fullEnd), unmapTime(window.End, fullStart, fullEnd)
}

func unmapTime(ratio float64, start, end time.Time) time.Time {
	return start.Add(time.Duration(float64(end.Sub(start)) * ratio))
}

func (w Widget) resolveGeometry(tasks []resolvedTask, plot image.Rectangle, start, end time.Time, requestedBarHeight int, reveal ...float32) geometry {
	rowHeight := 0
	if len(tasks) > 0 {
		rowHeight = plot.Dy() / len(tasks)
	}
	return w.resolveVisibleGeometry(tasks, plot, start, end, requestedBarHeight, rowHeight, 0, 0, len(tasks), reveal...)
}

func (w Widget) resolveVisibleGeometry(tasks []resolvedTask, plot image.Rectangle, start, end time.Time, requestedBarHeight, rowHeight, firstRow, rowOffset, totalRows int, reveal ...float32) geometry {
	animationProgress := float32(1)
	if len(reveal) > 0 {
		animationProgress = min(max(reveal[0], 0), 1)
	}
	if rowHeight <= 0 {
		rowHeight = max(plot.Dy()/max(totalRows, 1), 1)
	}
	rowBand := float32(rowHeight)
	geo := geometry{plot: plot, start: start, end: end, rowBand: rowBand}
	for index := 0; index < w.tickCount; index++ {
		ratio := float64(index) / float64(max(w.tickCount-1, 1))
		at := start.Add(time.Duration(float64(end.Sub(start)) * ratio))
		geo.ticks = append(geo.ticks, timeTick{at: at, label: w.timeLabel(at, end.Sub(start)), pixel: mapTime(at, start, end, plot)})
	}
	for index, task := range tasks {
		rowIndex := firstRow + index
		entryProgress := ganttEntryProgress(animationProgress, rowIndex, max(totalRows, 1))
		left := mapTimeUnclamped(task.start, start, end, plot)
		renderEnd := task.start.Add(time.Duration(float64(task.end.Sub(task.start)) * float64(entryProgress)))
		right := mapTimeUnclamped(renderEnd, start, end, plot)
		fullRight := mapTimeUnclamped(task.end, start, end, plot)
		top := plot.Min.Y + rowOffset + index*rowHeight
		bottom := top + rowHeight
		barHeight := min(max(requestedBarHeight, 1), max(int(geo.rowBand)-6, 1))
		vertical := top + max((bottom-top-barHeight)/2, 0)
		rect := image.Rect(int(math.Round(float64(left))), vertical, int(math.Round(float64(right))), vertical+barHeight)
		unclippedRect := image.Rect(int(math.Round(float64(left))), vertical, int(math.Round(float64(fullRight))), vertical+barHeight)
		if task.milestone {
			if entryProgress < 1 {
				geo.tasks = append(geo.tasks, taskGeometry{task: task, index: rowIndex, rowTop: top})
				continue
			}
			center := int(math.Round(float64(left)))
			size := min(max(barHeight, 8), max(int(geo.rowBand)-4, 8))
			rect = image.Rect(center-size/2, vertical, center+(size+1)/2, vertical+size)
			unclippedRect = rect
		}
		rect = rect.Intersect(plot)
		baseline := image.Rectangle{}
		if task.hasBaseline && !task.milestone {
			baselineLeft := int(math.Round(float64(mapTime(task.baselineStart, start, end, plot))))
			baselineRight := int(math.Round(float64(mapTime(task.baselineEnd, start, end, plot))))
			baselineHeight := min(max(requestedBarHeight/4, 3), max(bottom-top-barHeight-2, 1))
			baselineTop := min(rect.Max.Y+2, bottom-baselineHeight)
			baseline = image.Rect(baselineLeft, baselineTop, baselineRight, baselineTop+baselineHeight).Intersect(plot)
		}
		progressRect := progressGeometry(task, rect, start, end, plot)
		geo.tasks = append(geo.tasks, taskGeometry{task: task, rect: rect, unclippedRect: unclippedRect, progressRect: progressRect, baseline: baseline, index: rowIndex, rowTop: top})
	}
	return geo
}

func progressGeometry(task resolvedTask, visibleRect image.Rectangle, start, end time.Time, plot image.Rectangle) image.Rectangle {
	if task.milestone || visibleRect.Empty() || task.progress <= 0 || !task.end.After(task.start) {
		return image.Rectangle{}
	}
	completedAt := task.start.Add(time.Duration(float64(task.end.Sub(task.start)) * float64(task.progress)))
	completedX := int(math.Round(float64(mapTimeUnclamped(completedAt, start, end, plot))))
	return image.Rect(visibleRect.Min.X, visibleRect.Min.Y, completedX, visibleRect.Max.Y).Intersect(visibleRect)
}

func ganttEntryProgress(total float32, index, count int) float32 {
	if total >= 1 || count <= 1 {
		return min(max(total, 0), 1)
	}
	delay := float32(index) / float32(count-1) * 0.55
	return min(max((total-delay)/0.45, 0), 1)
}

func mapTime(value, start, end time.Time, plot image.Rectangle) float32 {
	ratio := float64(value.Sub(start)) / float64(end.Sub(start))
	ratio = min(max(ratio, 0), 1)
	return float32(plot.Min.X) + float32(ratio)*float32(plot.Dx())
}

func mapTimeUnclamped(value, start, end time.Time, plot image.Rectangle) float32 {
	ratio := float64(value.Sub(start)) / float64(end.Sub(start))
	return float32(plot.Min.X) + float32(ratio)*float32(plot.Dx())
}
func (w Widget) timeLabel(value time.Time, span time.Duration) string {
	if w.formatTime != nil {
		return w.formatTime(value)
	}
	switch {
	case span <= 48*time.Hour:
		return value.Format("Jan 2 15:04")
	case span <= 60*24*time.Hour:
		return value.Format("Jan 2")
	case span <= 2*365*24*time.Hour:
		return value.Format("Jan 2006")
	default:
		return value.Format("2006")
	}
}

func selectionAt(position f32.Point, hovered bool, geo geometry) (taskGeometry, bool) {
	if !hovered {
		return taskGeometry{}, false
	}
	point := position.Round()
	for _, task := range geo.tasks {
		if point.In(task.rect) {
			return task, true
		}
	}
	return taskGeometry{}, false
}
func exportSelection(task resolvedTask) Selection {
	return Selection{Key: task.key, Label: task.label, Group: task.group, Parent: task.parentKey, Depth: task.depth, Start: task.start, End: task.end, Progress: task.progress, Color: task.color, Milestone: task.milestone, Collapsed: task.collapsed}
}

type style struct {
	axis, label, grid, dependency, marker, background, emphasis color.NRGBA
	opacity                                                     float32
}

func ganttStyle(activeTheme *theme.Theme, disabled bool) style {
	grid := activeTheme.Palette.Border
	grid.A = byte(float32(grid.A) * .75)
	dependency := activeTheme.Palette.MutedForeground
	dependency.A = byte(float32(dependency.A) * .7)
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return style{axis: activeTheme.Palette.Border, label: activeTheme.Palette.MutedForeground, grid: grid, dependency: dependency, marker: activeTheme.Palette.Accent, background: activeTheme.Palette.SurfaceSecondary, emphasis: activeTheme.Palette.Foreground, opacity: opacity}
}

func (w Widget) drawGrid(gtx layout.Context, geo geometry, style style, tokens theme.GanttChartTheme) {
	if geo.plot.Empty() {
		return
	}
	width := float32(max(gtx.Dp(tokens.GridWidth), 1))
	if w.showGrid {
		for _, tick := range geo.ticks {
			chart.StrokeLine(gtx, f32.Pt(tick.pixel, float32(geo.plot.Min.Y)), f32.Pt(tick.pixel, float32(geo.plot.Max.Y)), width, style.grid)
		}
		rowClip := clip.Rect(geo.plot).Push(gtx.Ops)
		for _, task := range geo.tasks {
			y := float32(task.rowTop)
			chart.StrokeLine(gtx, f32.Pt(float32(geo.plot.Min.X), y), f32.Pt(float32(geo.plot.Max.X), y), width, style.grid)
		}
		if len(geo.tasks) > 0 {
			last := geo.tasks[len(geo.tasks)-1]
			y := float32(last.rowTop) + geo.rowBand
			chart.StrokeLine(gtx, f32.Pt(float32(geo.plot.Min.X), y), f32.Pt(float32(geo.plot.Max.X), y), width, style.grid)
		}
		rowClip.Pop()
	}
	axisWidth := float32(max(gtx.Dp(tokens.AxisWidth), 1))
	chart.StrokeLine(gtx, f32.Pt(float32(geo.plot.Min.X), float32(geo.plot.Min.Y)), f32.Pt(float32(geo.plot.Min.X), float32(geo.plot.Max.Y)), axisWidth, style.axis)
	chart.StrokeLine(gtx, f32.Pt(float32(geo.plot.Min.X), float32(geo.plot.Max.Y)), f32.Pt(float32(geo.plot.Max.X), float32(geo.plot.Max.Y)), axisWidth, style.axis)
}

func (w Widget) drawTasks(ctx *frame.Context, gtx layout.Context, geo geometry, selectedIndex int, selected bool, style style, tokens theme.GanttChartTheme) {
	for _, item := range geo.tasks {
		if item.rect.Empty() {
			continue
		}
		if !item.baseline.Empty() {
			baselineColor := item.task.color
			baselineColor.A = 0xa0
			paint.FillShape(gtx.Ops, baselineColor, clip.UniformRRect(item.baseline, min(item.baseline.Dy()/2, 2)).Op(gtx.Ops))
		}
		itemOpacity := float32(1)
		if selected && item.index != selectedIndex {
			itemOpacity = .42
		}
		fade := paint.PushOpacity(gtx.Ops, itemOpacity)
		if item.task.milestone {
			drawMilestone(gtx, item.rect, item.task.color, selected && item.index == selectedIndex, style.emphasis)
			fade.Pop()
			continue
		}
		radius := min(max(gtx.Dp(tokens.BarRadius), 0), min(item.rect.Dx(), item.rect.Dy())/2)
		background := item.task.color
		background.A = byte(float32(background.A) * .24)
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(item.rect, radius).Op(gtx.Ops))
		if !item.progressRect.Empty() {
			progress := item.progressRect
			paint.FillShape(gtx.Ops, item.task.color, clip.UniformRRect(progress, min(radius, min(progress.Dx(), progress.Dy())/2)).Op(gtx.Ops))
		}
		if w.showTaskLabels {
			w.drawTaskLabel(ctx, gtx, item, tokens)
		}
		fade.Pop()
	}
}

func drawMilestone(gtx layout.Context, rect image.Rectangle, col color.NRGBA, selected bool, emphasis color.NRGBA) {
	center := f32.Pt(float32(rect.Min.X+rect.Dx()/2), float32(rect.Min.Y+rect.Dy()/2))
	radius := float32(min(rect.Dx(), rect.Dy())) / 2
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X, center.Y-radius))
	path.LineTo(f32.Pt(center.X+radius, center.Y))
	path.LineTo(f32.Pt(center.X, center.Y+radius))
	path.LineTo(f32.Pt(center.X-radius, center.Y))
	path.Close()
	spec := path.End()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: spec}.Op())
	if selected {
		outline := clip.Stroke{Path: spec, Width: 2}
		paint.FillShape(gtx.Ops, emphasis, outline.Op())
	}
}

func (w Widget) drawTaskLabel(ctx *frame.Context, gtx layout.Context, item taskGeometry, tokens theme.GanttChartTheme) {
	padding := max(gtx.Dp(tokens.TaskLabelPaddingX), 0)
	if item.progressRect.Empty() {
		return
	}
	available := item.progressRect.Dx() - padding*2
	if available <= 0 {
		return
	}
	// Measure without the bar constraint first. The task name is already present
	// in the row label, so a partial duplicate inside a narrow bar is less useful.
	measureGtx := gtx
	measureGtx.Constraints.Max.X = 1 << 15
	label := chartText(ctx, measureGtx, item.task.label, tokens.AxisTextSize, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, 1<<15)
	if label.dims.Size.X+4 > available || label.dims.Size.Y > item.rect.Dy() {
		return
	}
	labelArea := item.progressRect
	clipArea := clip.Rect(labelArea).Push(gtx.Ops)
	chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(item.rect.Min.X+padding, item.rect.Min.Y+(item.rect.Dy()-label.dims.Size.Y)/2))
	clipArea.Pop()
}

func drawDependencies(gtx layout.Context, geo geometry, style style, tokens theme.GanttChartTheme) {
	index := make(map[string]taskGeometry, len(geo.tasks))
	for _, task := range geo.tasks {
		index[task.task.key] = task
	}
	width := float32(max(gtx.Dp(tokens.DependencyWidth), 1))
	area := clip.Rect(geo.plot).Push(gtx.Ops)
	defer area.Pop()
	for _, target := range geo.tasks {
		for _, sourceKey := range target.task.dependsOn {
			source, ok := index[sourceKey]
			if !ok || source.rect.Empty() || target.rect.Empty() {
				continue
			}
			from, to, middle, forward := dependencyRoute(source, target, geo.plot)
			strokeDependency(gtx, from, f32.Pt(middle, from.Y), width, style.dependency, float32(max(gtx.Dp(tokens.DependencyDash), 1)))
			strokeDependency(gtx, f32.Pt(middle, from.Y), f32.Pt(middle, to.Y), width, style.dependency, float32(max(gtx.Dp(tokens.DependencyDash), 1)))
			strokeDependency(gtx, f32.Pt(middle, to.Y), to, width, style.dependency, float32(max(gtx.Dp(tokens.DependencyDash), 1)))
			drawArrow(gtx, to, forward, style.dependency)
		}
	}
}

func dependencyRoute(source, target taskGeometry, plot image.Rectangle) (from, to f32.Point, middle float32, forward bool) {
	sourceY := float32(source.rect.Min.Y + source.rect.Dy()/2)
	targetY := float32(target.rect.Min.Y + target.rect.Dy()/2)
	forward = target.rect.Min.X >= source.rect.Max.X
	if forward {
		from = f32.Pt(float32(source.rect.Max.X), sourceY)
		to = f32.Pt(float32(target.rect.Min.X), targetY)
	} else {
		from = f32.Pt(float32(source.rect.Min.X), sourceY)
		to = f32.Pt(float32(target.rect.Max.X), targetY)
	}
	middle = (from.X + to.X) / 2
	margin := float32(4)
	middle = min(max(middle, float32(plot.Min.X)+margin), float32(plot.Max.X)-margin)
	return from, to, middle, forward
}

func strokeDependency(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA, dash float32) {
	dx, dy := to.X-from.X, to.Y-from.Y
	distance := float32(math.Hypot(float64(dx), float64(dy)))
	if distance <= 0 || dash <= 0 {
		return
	}
	for offset := float32(0); offset < distance; offset += dash * 2 {
		end := min(offset+dash, distance)
		chart.StrokeLine(gtx, f32.Pt(from.X+dx*offset/distance, from.Y+dy*offset/distance), f32.Pt(from.X+dx*end/distance, from.Y+dy*end/distance), width, col)
	}
}
func drawArrow(gtx layout.Context, point f32.Point, forward bool, col color.NRGBA) {
	direction := float32(1)
	if !forward {
		direction = -1
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(point.X, point.Y))
	path.LineTo(f32.Pt(point.X-5*direction, point.Y-3))
	path.LineTo(f32.Pt(point.X-5*direction, point.Y+3))
	path.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func (w Widget) drawLabels(ctx *frame.Context, gtx layout.Context, state *chartState, geo geometry, style style, tokens theme.GanttChartTheme, enabled bool) {
	gap := max(gtx.Dp(tokens.TickLabelGap), 0)
	indent := max(gtx.Dp(tokens.TaskIndent), 0)
	toggleSize := max(gtx.Dp(tokens.TaskToggleSize), 0)
	toggleGap := max(gtx.Dp(tokens.TaskToggleGap), 0)
	toggleWidth := toggleSize + toggleGap
	hierarchical := hasHierarchy(geo.tasks)
	labelColumnWidth := 0
	for _, task := range geo.tasks {
		label := chartText(ctx, gtx, task.task.label, tokens.AxisTextSize, style.label, max(geo.plot.Min.X-gap, 1))
		extra := 0
		if hierarchical {
			extra = toggleWidth
		}
		labelColumnWidth = max(labelColumnWidth, label.dims.Size.X+task.task.depth*indent+extra)
	}
	rowClip := clip.Rect(image.Rect(0, geo.plot.Min.Y, max(gtx.Constraints.Max.X, geo.plot.Max.X), geo.plot.Max.Y)).Push(gtx.Ops)
	for _, task := range geo.tasks {
		label := chartText(ctx, gtx, task.task.label, tokens.AxisTextSize, style.label, max(geo.plot.Min.X-gap, 1))
		rowTop := task.rowTop
		rowHeight := max(int(math.Round(float64(geo.rowBand))), 1)
		y := rowTop + max((rowHeight-label.dims.Size.Y)/2, 0)
		baseX := geo.plot.Min.X - gap - labelColumnWidth + task.task.depth*indent
		if hierarchical && task.task.hasChildren && toggleSize > 0 {
			toggle := state.toggleItem(task.task.key)
			itemGtx := gtx
			if !enabled {
				itemGtx = itemGtx.Disabled()
			}
			itemGtx.Constraints = layout.Exact(image.Pt(toggleWidth, rowHeight))
			offset := op.Offset(image.Pt(baseX, rowTop)).Push(gtx.Ops)
			toggle.Layout(itemGtx, func(itemGtx layout.Context) layout.Dimensions {
				pointer.CursorPointer.Add(itemGtx.Ops)
				drawTaskToggle(itemGtx, image.Rect(0, 0, toggleSize, rowHeight), w.taskCollapsed(state, task.task), style.label)
				return layout.Dimensions{Size: image.Pt(toggleWidth, rowHeight)}
			})
			offset.Pop()
		}
		x := baseX
		if hierarchical {
			x += toggleWidth
		}
		chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(max(x, 0), y))
	}
	rowClip.Pop()
	lastRight := math.MinInt
	for _, tick := range geo.ticks {
		label := chartText(ctx, gtx, tick.label, tokens.AxisTextSize, style.label, max(geo.plot.Dx(), 1))
		x := int(math.Round(float64(tick.pixel))) - label.dims.Size.X/2
		x = min(max(x, geo.plot.Min.X), max(geo.plot.Max.X-label.dims.Size.X, geo.plot.Min.X))
		if x < lastRight+8 {
			continue
		}
		chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(x, geo.plot.Max.Y+gap))
		lastRight = x + label.dims.Size.X
	}
}

func drawTaskToggle(gtx layout.Context, area image.Rectangle, collapsed bool, col color.NRGBA) {
	center := f32.Pt(float32(area.Min.X+area.Dx()/2), float32(area.Min.Y+area.Dy()/2))
	size := float32(min(max(area.Dx()/2, 3), max(area.Dy()/3, 3)))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	if collapsed {
		path.MoveTo(f32.Pt(center.X-size/2, center.Y-size))
		path.LineTo(f32.Pt(center.X-size/2, center.Y+size))
		path.LineTo(f32.Pt(center.X+size, center.Y))
	} else {
		path.MoveTo(f32.Pt(center.X-size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X+size, center.Y-size/2))
		path.LineTo(f32.Pt(center.X, center.Y+size))
	}
	path.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func (w Widget) drawMarkers(ctx *frame.Context, gtx layout.Context, geo geometry, style style, tokens theme.GanttChartTheme) {
	for _, marker := range w.markers {
		if marker.at.Before(geo.start) || marker.at.After(geo.end) {
			continue
		}
		col := marker.color
		if col.A == 0 {
			col = style.marker
		}
		x := mapTime(marker.at, geo.start, geo.end, geo.plot)
		chart.StrokeLine(gtx, f32.Pt(x, float32(geo.plot.Min.Y)), f32.Pt(x, float32(geo.plot.Max.Y)), float32(max(gtx.Dp(tokens.MarkerWidth), 1)), col)
		if marker.label == "" {
			continue
		}
		label := chartText(ctx, gtx, marker.label, tokens.AxisTextSize, col, max(geo.plot.Dx(), 1))
		y := geo.plot.Min.Y - label.dims.Size.Y - max(gtx.Dp(tokens.MarkerLabelGap), 0)
		if w.taskAxis != "" {
			y -= gtx.Sp(tokens.AxisTextSize) + max(gtx.Dp(tokens.AxisNameGap), 0)
		}
		position := chart.ClampLabelPosition(image.Pt(int(math.Round(float64(x)))-label.dims.Size.X/2, y), label.dims.Size, image.Rect(geo.plot.Min.X, 0, geo.plot.Max.X, geo.plot.Min.Y))
		chart.PlaceRecorded(gtx, label.call, label.dims, position)
	}
}

func (w Widget) drawAxisTitles(ctx *frame.Context, gtx layout.Context, geo geometry, style style, tokens theme.GanttChartTheme) {
	if w.taskAxis != "" {
		label := chartText(ctx, gtx, w.taskAxis, tokens.AxisTextSize, style.label, max(geo.plot.Min.X, 1))
		chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(max(geo.plot.Min.X-label.dims.Size.X-max(gtx.Dp(tokens.TickLabelGap), 0), 0), max(geo.plot.Min.Y-label.dims.Size.Y-max(gtx.Dp(tokens.AxisNameGap), 0), 0)))
	}
	if w.timeAxis != "" {
		label := chartText(ctx, gtx, w.timeAxis, tokens.AxisTextSize, style.label, max(geo.plot.Dx(), 1))
		chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(max(geo.plot.Max.X-label.dims.Size.X, geo.plot.Min.X), geo.plot.Max.Y+max(gtx.Dp(tokens.TickLabelGap), 0)+label.dims.Size.Y))
	}
}

type recordedBlock struct {
	call op.CallOp
	dims layout.Dimensions
}

func (w Widget) recordLegend(ctx *frame.Context, gtx layout.Context, state *chartState, tasks []resolvedTask, style style, maxWidth int, enabled bool) recordedBlock {
	if maxWidth <= 0 {
		return recordedBlock{}
	}
	tokens := frame.ActiveTheme(ctx).Components.GanttChart
	entries := w.legendGroupsCached(state, tasks)
	if len(entries) < 2 {
		return recordedBlock{}
	}
	markerSize := max(gtx.Dp(tokens.LegendMarkerSize), 2)
	markerGap := max(gtx.Dp(tokens.LegendMarkerGap), 0)
	itemGap := max(gtx.Dp(tokens.LegendItemGap), 0)
	lineGap := max(gtx.Dp(tokens.LegendLineGap), 0)
	macro := op.Record(gtx.Ops)
	x, y, rowHeight, usedWidth := 0, 0, 0, 0
	for _, item := range entries {
		label := chartText(ctx, gtx, item.group, tokens.LegendTextSize, style.label, max(maxWidth-markerSize-markerGap, 1))
		itemWidth := markerSize + markerGap + label.dims.Size.X
		itemHeight := max(markerSize, label.dims.Size.Y)
		if x > 0 && x+itemWidth > maxWidth {
			y += rowHeight + lineGap
			x, rowHeight = 0, 0
		}
		legendItem := state.legendItem(item.group)
		itemGtx := gtx
		if !enabled {
			itemGtx = itemGtx.Disabled()
		}
		itemGtx.Constraints = layout.Exact(image.Pt(itemWidth, itemHeight))
		itemMacro := op.Record(gtx.Ops)
		legendItem.Layout(itemGtx, func(itemGtx layout.Context) layout.Dimensions {
			if legendItem.Hovered() && enabled {
				paint.FillShape(itemGtx.Ops, frame.ActiveTheme(ctx).Palette.SurfaceHover, clip.UniformRRect(image.Rectangle{Max: image.Pt(itemWidth, itemHeight)}, min(itemHeight/2, 4)).Op(itemGtx.Ops))
			}
			opacity := float32(1)
			if w.groupHidden(state, item.group) {
				opacity = .38
			}
			fade := paint.PushOpacity(itemGtx.Ops, opacity)
			pointer.CursorPointer.Add(itemGtx.Ops)
			marker := image.Rect(0, (itemHeight-markerSize)/2, markerSize, (itemHeight-markerSize)/2+markerSize)
			paint.FillShape(itemGtx.Ops, item.color, clip.UniformRRect(marker, min(markerSize/2, 3)).Op(itemGtx.Ops))
			chart.PlaceRecorded(itemGtx, label.call, label.dims, image.Pt(markerSize+markerGap, (itemHeight-label.dims.Size.Y)/2))
			fade.Pop()
			return layout.Dimensions{Size: image.Pt(itemWidth, itemHeight)}
		})
		itemCall := itemMacro.Stop()
		offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
		itemCall.Add(gtx.Ops)
		offset.Pop()
		x += itemWidth + itemGap
		usedWidth = max(usedWidth, min(x-itemGap, maxWidth))
		rowHeight = max(rowHeight, itemHeight)
	}
	return recordedBlock{call: macro.Stop(), dims: layout.Dimensions{Size: image.Pt(usedWidth, y+rowHeight)}}
}

func placeBlock(gtx layout.Context, block recordedBlock, position image.Point) {
	offset := op.Offset(position).Push(gtx.Ops)
	block.call.Add(gtx.Ops)
	offset.Pop()
}

type recordedText struct {
	call op.CallOp
	dims layout.Dimensions
}

func chartText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, col color.NRGBA, maxWidth int) recordedText {
	call, dims := chart.RecordText(ctx, gtx, value, size, font.Normal, col, maxWidth)
	return recordedText{call: call, dims: dims}
}

func (w Widget) taskLabelWidth(ctx *frame.Context, gtx layout.Context, state *chartState, tasks []resolvedTask, maxWidth int) int {
	tokens := frame.ActiveTheme(ctx).Components.GanttChart
	hiddenSignature := ganttStringSliceSignature(w.hiddenGroups)
	textPx := gtx.Sp(tokens.AxisTextSize)
	if state.taskLabelWidthReady && state.taskLabelWidthResolvedRev == state.resolvedTasksRevision && state.taskLabelWidthVisibilityRev == state.visibilityRevision && state.taskLabelWidthHiddenSig == hiddenSignature && state.taskLabelWidthMax == maxWidth && state.taskLabelWidthMetric == gtx.Metric && state.taskLabelWidthTextPx == textPx {
		return state.taskLabelWidthValue
	}
	indent := max(gtx.Dp(tokens.TaskIndent), 0)
	toggleWidth := max(gtx.Dp(tokens.TaskToggleSize), 0) + max(gtx.Dp(tokens.TaskToggleGap), 0)
	hierarchical := false
	for _, task := range tasks {
		if task.hasChildren {
			hierarchical = true
			break
		}
	}
	width := 0
	for _, task := range tasks {
		label := chartText(ctx, gtx, task.label, tokens.AxisTextSize, color.NRGBA{A: 0xff}, maxWidth)
		extra := 0
		if hierarchical {
			extra = toggleWidth
		}
		width = max(width, label.dims.Size.X+task.depth*indent+extra)
	}
	state.taskLabelWidthValue = width
	state.taskLabelWidthReady = true
	state.taskLabelWidthResolvedRev = state.resolvedTasksRevision
	state.taskLabelWidthVisibilityRev = state.visibilityRevision
	state.taskLabelWidthHiddenSig = hiddenSignature
	state.taskLabelWidthMax = maxWidth
	state.taskLabelWidthMetric = gtx.Metric
	state.taskLabelWidthTextPx = textPx
	return width
}

func hasHierarchy(tasks []taskGeometry) bool {
	for _, task := range tasks {
		if task.task.hasChildren {
			return true
		}
	}
	return false
}

func (w Widget) drawEmpty(ctx *frame.Context, gtx layout.Context, area image.Rectangle, emptyText ...string) {
	value := w.emptyText
	if len(emptyText) > 0 {
		value = emptyText[0]
	}
	if value == "" {
		value = "No tasks"
	}
	label := chartText(ctx, gtx, value, frame.ActiveTheme(ctx).Components.GanttChart.AxisTextSize, frame.ActiveTheme(ctx).Palette.MutedForeground, max(area.Dx(), 1))
	chart.PlaceRecorded(gtx, label.call, label.dims, image.Pt(area.Min.X+max((area.Dx()-label.dims.Size.X)/2, 0), area.Min.Y+max((area.Dy()-label.dims.Size.Y)/2, 0)))
}

func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, selection Selection, anchor image.Rectangle, progress float32, exiting bool) {
	if selection.Key == "" {
		return
	}
	var content frame.Widget
	if w.tooltipContent != nil {
		content = w.tooltipContent(selection)
	} else {
		content = frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			rows := []chart.TooltipRow{
				{Text: fmt.Sprintf("%s - %s", w.timeLabel(selection.Start, selection.End.Sub(selection.Start)), w.timeLabel(selection.End, selection.End.Sub(selection.Start))), Color: selection.Color},
				{Text: fmt.Sprintf("Progress  %.0f%%", selection.Progress*100)},
			}
			if selection.Group != "" {
				rows = append(rows, chart.TooltipRow{Text: selection.Group})
			}
			tokens := frame.ActiveTheme(ctx).Components.Tooltip
			return chart.LayoutTooltipRows(ctx, gtx, selection.Label, rows, tokens.TextSize, frame.ActiveTheme(ctx).Palette.OverlayForeground, max(gtx.Dp(7), 2), max(gtx.Dp(4), 0), chart.TooltipMarkerSquare)
		})
	}
	tooltip.NewPopup(content).Placement(overlay.PopoverRightStart).Offset(max(frame.ActiveTheme(ctx).Components.GanttChart.TooltipGap, 0)).TransformMotion(false).Progress(progress).Exiting(exiting).Layout(ctx, gtx, anchor)
}
