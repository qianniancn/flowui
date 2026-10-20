package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/gantt"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	selected   string
	changes    map[string]gantt.TaskChange
	collapsed  map[string]bool
	hidden     map[string]bool
	timeWindow gantt.TimeWindow
}
type Msg struct {
	Scenario        string
	Selection       gantt.Selection
	Change          gantt.TaskChange
	ToggleKey       string
	ToggleCollapsed bool
	LegendGroup     string
	LegendHidden    bool
	TimeWindow      gantt.TimeWindow
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch {
	case msg.Change.Key != "":
		if model.changes == nil {
			model.changes = make(map[string]gantt.TaskChange)
		}
		model.changes[msg.Change.Key] = msg.Change
		model.selected = fmt.Sprintf("Edited: %s", msg.Change.Key)
	case msg.ToggleKey != "":
		if model.collapsed == nil {
			model.collapsed = make(map[string]bool)
		}
		model.collapsed[msg.ToggleKey] = msg.ToggleCollapsed
		model.selected = fmt.Sprintf("%s %s", msg.ToggleKey, map[bool]string{true: "collapsed", false: "expanded"}[msg.ToggleCollapsed])
	case msg.LegendGroup != "":
		if model.hidden == nil {
			model.hidden = make(map[string]bool)
		}
		model.hidden[msg.LegendGroup] = msg.LegendHidden
		model.selected = fmt.Sprintf("Legend: %s", msg.LegendGroup)
	case !msg.TimeWindow.Start.IsZero():
		model.timeWindow = msg.TimeWindow
	case msg.Selection.Key != "":
		model.selected = fmt.Sprintf("%s - %s", msg.Scenario, msg.Selection.Label)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	selected := "Select a task"
	if model.selected != "" {
		selected = "Selected: " + model.selected
	}
	return layout.Scroll(layout.Box(layout.Column(
		text.New("Project schedules").Size(24),
		text.New(selected).Size(14),
		ganttSection("G2 interval schedule", g2IntervalSchedule(send)),
		ganttSection("Phased software delivery", phasedSoftwareDelivery(send)),
		ganttSection("Dependency flow", dependencyFlow(send)),
		ganttSection("Event plan", eventPlan(send)),
		ganttSection("Software release", softwareRelease(send)),
		ganttSection("Marketing campaign", marketingCampaign(send)),
		ganttSection("Portfolio delivery", portfolioDelivery(send)),
		ganttSection("Hierarchical product roadmap", hierarchicalRoadmap(send)),
		ganttSection("Controlled interactive schedule", controlledSchedule(send, model)),
	).Gap(16)).Style(style.New().FillWidth().MaxWidth(1040).Padding(24)))
}

func g2IntervalSchedule(send ui.Send[Msg]) gantt.Widget {
	origin := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)
	day := func(value int) time.Time { return origin.AddDate(0, 0, value-1) }
	tasks := []struct {
		key, label string
		start, end int
	}{
		{"activity-planning", "活动策划", 1, 4},
		{"venue-logistics", "场地物流规划", 3, 13},
		{"select-supplier", "选择供应商", 5, 8},
		{"rent-venue", "租赁场地", 9, 13},
		{"book-catering", "预定餐饮服务商", 10, 14},
		{"rent-decoration", "租赁活动装饰团队", 12, 17},
		{"rehearsal", "彩排", 14, 16},
		{"celebration", "活动庆典", 17, 18},
	}
	items := make([]gantt.Task, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, gantt.NewTask(task.key, task.label, day(task.start), day(task.end)).Group(task.label))
	}
	return gantt.New("g2-interval-schedule", items).
		TimeRange(day(0), day(19)).
		Legend(true).TaskLabels(true).TaskAxis("任务").TimeAxis("时间（天）").TimeTicks(7).
		FormatTime(func(value time.Time) string {
			return fmt.Sprintf("%d", int(value.Sub(origin).Hours()/24)+1)
		}).
		Dependencies(false).Animation(false).
		OnTaskClick(selectTask(send, "G2 interval schedule"))
}

func phasedSoftwareDelivery(send ui.Send[Msg]) gantt.Widget {
	origin := time.Date(2025, time.February, 3, 0, 0, 0, 0, time.Local)
	week := func(value int) time.Time { return origin.AddDate(0, 0, (value-1)*7) }
	tasks := []struct {
		key, label, phase string
		start, end        int
	}{
		{"requirements", "需求分析", "规划阶段", 1, 5},
		{"architecture", "系统设计", "设计阶段", 4, 10},
		{"frontend", "前端开发", "开发阶段", 8, 20},
		{"backend", "后端开发", "开发阶段", 10, 22},
		{"integration", "集成测试", "测试阶段", 18, 25},
		{"deployment", "系统部署", "部署阶段", 24, 28},
		{"acceptance", "用户验收", "验收阶段", 26, 30},
	}
	items := make([]gantt.Task, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, gantt.NewTask(task.key, task.label, week(task.start), week(task.end)).Group(task.phase))
	}
	return gantt.New("phased-software-delivery", items).
		TimeRange(week(0), week(32)).
		Legend(true).TaskLabels(true).TaskAxis("项目任务").TimeAxis("时间（周）").TimeTicks(7).
		FormatTime(func(value time.Time) string {
			return fmt.Sprintf("%d", int(value.Sub(origin).Hours()/(24*7))+1)
		}).
		Dependencies(false).Animation(true).AnimationDuration(2200 * time.Millisecond).
		OnTaskClick(selectTask(send, "Phased software delivery"))
}

func dependencyFlow(send ui.Send[Msg]) gantt.Widget {
	origin := time.Date(2025, time.April, 7, 0, 0, 0, 0, time.Local)
	week := func(value int) time.Time { return origin.AddDate(0, 0, (value-1)*7) }
	tasks := []gantt.Task{
		gantt.NewTask("task-a", "任务A", week(1), week(5)).Group("任务A"),
		gantt.NewTask("task-b", "任务B", week(5), week(10)).Group("任务B").DependsOn("task-a"),
		gantt.NewTask("task-c", "任务C", week(8), week(15)).Group("任务C").DependsOn("task-b"),
		gantt.NewTask("task-d", "任务D", week(15), week(20)).Group("任务D").DependsOn("task-c"),
	}
	return gantt.New("dependency-flow", tasks).
		TimeRange(week(0), week(22)).
		Legend(true).TaskLabels(true).TaskAxis("项目任务").TimeAxis("时间（周）").TimeTicks(6).
		FormatTime(func(value time.Time) string {
			return fmt.Sprintf("%d", int(value.Sub(origin).Hours()/(24*7))+1)
		}).
		Dependencies(true).Animation(false).
		OnTaskClick(selectTask(send, "Dependency flow"))
}

func eventPlan(send ui.Send[Msg]) gantt.Widget {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []gantt.Task{
		gantt.NewTask("plan", "Event planning", day(0), day(4)).Baseline(day(0), day(3)).Group("Planning").Progress(1),
		gantt.NewTask("logistics", "Venue logistics", day(3), day(13)).Baseline(day(3), day(11)).Group("Planning").Progress(.7).DependsOn("plan"),
		gantt.NewTask("vendors", "Select vendors", day(5), day(8)).Baseline(day(4), day(8)).Group("Procurement").Progress(1).DependsOn("plan"),
		gantt.NewTask("venue", "Hire venue", day(9), day(13)).Baseline(day(8), day(12)).Group("Procurement").Progress(.8).DependsOn("vendors"),
		gantt.NewTask("catering", "Book catering", day(10), day(14)).Baseline(day(9), day(13)).Group("Procurement").Progress(.45).DependsOn("vendors"),
		gantt.NewTask("decor", "Event decoration", day(12), day(17)).Baseline(day(11), day(16)).Group("Delivery").Progress(.25).DependsOn("venue"),
		gantt.NewTask("rehearsal", "Rehearsal", day(14), day(16)).Baseline(day(13), day(16)).Group("Delivery").Progress(0).DependsOn("catering", "decor"),
		gantt.NewMilestone("approval", "Launch approval", day(16)).Group("Delivery").DependsOn("rehearsal"),
		gantt.NewTask("event", "Event celebration", day(17), day(18)).Group("Delivery").DependsOn("approval"),
	}
	chart := gantt.New("event-plan", tasks).
		TimeRange(day(-1), day(19)).Legend(true).TaskLabels(true).
		TaskAxis("Workstream").TimeAxis("January 2025").
		Marker(gantt.NewTimeMarker(day(12)).Text("Status date")).
		TimeTicks(6).
		OnTaskClick(selectTask(send, "Event plan"))
	return chart
}

func softwareRelease(send ui.Send[Msg]) gantt.Widget {
	start := time.Date(2025, time.March, 3, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []gantt.Task{
		gantt.NewTask("research", "Customer research", day(0), day(5)).Baseline(day(0), day(4)).Group("Discovery").Progress(1),
		gantt.NewTask("scope", "Release scope", day(4), day(8)).Baseline(day(3), day(7)).Group("Discovery").Progress(1).DependsOn("research"),
		gantt.NewTask("design", "Design system", day(6), day(14)).Baseline(day(5), day(12)).Group("Experience").Progress(.85).DependsOn("scope"),
		gantt.NewTask("frontend", "Frontend build", day(10), day(22)).Baseline(day(9), day(20)).Group("Engineering").Progress(.6).DependsOn("design"),
		gantt.NewTask("backend", "Platform services", day(10), day(24)).Baseline(day(9), day(21)).Group("Engineering").Progress(.5).DependsOn("scope"),
		gantt.NewTask("quality", "Acceptance testing", day(22), day(29)).Baseline(day(20), day(27)).Group("Quality").Progress(.15).DependsOn("frontend", "backend"),
		gantt.NewMilestone("freeze", "Code freeze", day(25)).Group("Quality").DependsOn("frontend", "backend"),
		gantt.NewTask("rollout", "Progressive rollout", day(30), day(35)).Group("Release").Progress(0).DependsOn("quality", "freeze"),
		gantt.NewMilestone("launch", "Public release", day(35)).Group("Release").DependsOn("rollout"),
	}
	return gantt.New("software-release", tasks).
		TimeRange(day(-1), day(37)).Legend(true).TaskLabels(true).
		TaskAxis("Delivery").TimeAxis("March - April 2025").
		Marker(gantt.NewTimeMarker(day(18)).Text("Sprint review")).
		TimeTicks(7).
		OnTaskClick(selectTask(send, "Software release"))
}

func marketingCampaign(send ui.Send[Msg]) gantt.Widget {
	start := time.Date(2025, time.May, 5, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []gantt.Task{
		gantt.NewTask("positioning", "Audience positioning", day(0), day(6)).Baseline(day(0), day(5)).Group("Strategy").Progress(1),
		gantt.NewTask("message", "Campaign messaging", day(4), day(10)).Baseline(day(4), day(9)).Group("Strategy").Progress(.9).DependsOn("positioning"),
		gantt.NewTask("film", "Hero film", day(8), day(19)).Baseline(day(7), day(17)).Group("Creative").Progress(.55).DependsOn("message"),
		gantt.NewTask("landing", "Landing page", day(9), day(17)).Baseline(day(8), day(16)).Group("Creative").Progress(.7).DependsOn("message"),
		gantt.NewTask("media", "Paid media setup", day(14), day(21)).Baseline(day(13), day(20)).Group("Channels").Progress(.35).DependsOn("film"),
		gantt.NewTask("community", "Community activation", day(16), day(24)).Baseline(day(15), day(23)).Group("Channels").Progress(.2).DependsOn("landing"),
		gantt.NewMilestone("go-live", "Campaign live", day(21)).Group("Launch").DependsOn("film", "landing", "media"),
		gantt.NewTask("optimize", "Daily optimization", day(22), day(34)).Group("Launch").Progress(0).DependsOn("go-live"),
	}
	return gantt.New("marketing-campaign", tasks).
		TimeRange(day(-1), day(36)).Legend(true).TaskLabels(true).
		TaskAxis("Campaign activity").TimeAxis("May - June 2025").
		Marker(gantt.NewTimeMarker(day(14)).Text("Creative review")).
		TimeTicks(7).
		OnTaskClick(selectTask(send, "Marketing campaign"))
}

func portfolioDelivery(send ui.Send[Msg]) gantt.Widget {
	start := time.Date(2025, time.July, 7, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []gantt.Task{
		gantt.NewTask("atlas-design", "Atlas - design", day(0), day(7)).Group("Atlas").Progress(1),
		gantt.NewTask("atlas-build", "Atlas - build", day(6), day(19)).Group("Atlas").Progress(.65).DependsOn("atlas-design"),
		gantt.NewTask("atlas-test", "Atlas - test", day(18), day(24)).Group("Atlas").Progress(.1).DependsOn("atlas-build"),
		gantt.NewTask("nova-design", "Nova - design", day(3), day(10)).Group("Nova").Progress(.8),
		gantt.NewTask("nova-build", "Nova - build", day(9), day(22)).Group("Nova").Progress(.45).DependsOn("nova-design"),
		gantt.NewTask("nova-test", "Nova - test", day(21), day(27)).Group("Nova").Progress(0).DependsOn("nova-build"),
		gantt.NewTask("orbit-design", "Orbit - design", day(7), day(13)).Group("Orbit").Progress(.55),
		gantt.NewTask("orbit-build", "Orbit - build", day(12), day(25)).Group("Orbit").Progress(.25).DependsOn("orbit-design"),
		gantt.NewMilestone("portfolio-review", "Portfolio review", day(20)).Group("Review").DependsOn("atlas-build", "nova-build", "orbit-build"),
	}
	return gantt.New("portfolio-delivery", tasks).
		TimeRange(day(-1), day(29)).Legend(true).TaskLabels(true).
		TaskAxis("Initiative").TimeAxis("July - August 2025").
		Marker(gantt.NewTimeMarker(day(20)).Text("Portfolio review")).
		TimeTicks(7).
		OnTaskClick(selectTask(send, "Portfolio delivery"))
}

func hierarchicalRoadmap(send ui.Send[Msg]) gantt.Widget {
	start := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []gantt.Task{
		gantt.NewTask("platform", "Platform foundation", day(0), day(20)).Baseline(day(0), day(18)).Group("Platform").Progress(.7),
		gantt.NewTask("api", "Public API", day(2), day(10)).Baseline(day(2), day(9)).Group("Platform").Progress(1).Parent("platform"),
		gantt.NewTask("observability", "Observability", day(7), day(16)).Baseline(day(6), day(14)).Group("Platform").Progress(.55).Parent("platform").DependsOn("api"),
		gantt.NewTask("experience", "Product experience", day(8), day(24)).Baseline(day(7), day(21)).Group("Experience").Progress(.4),
		gantt.NewTask("workspace", "Workspace flows", day(9), day(18)).Baseline(day(8), day(16)).Group("Experience").Progress(.65).Parent("experience").DependsOn("api"),
		gantt.NewTask("beta", "Design partner beta", day(18), day(24)).Group("Release").Progress(.15).DependsOn("observability", "workspace"),
		gantt.NewMilestone("launch", "General availability", day(26)).Group("Release").DependsOn("beta"),
	}
	return gantt.New("hierarchical-roadmap", tasks).
		TimeRange(day(-1), day(28)).TimeWindow(day(2), day(27)).
		Legend(true).TaskLabels(true).TaskAxis("Work package").TimeAxis("September 2025").
		Marker(gantt.NewTimeMarker(day(18)).Text("Readiness review")).TimeTicks(7).
		AnimationDuration(550 * time.Millisecond).UpdateAnimationDuration(350 * time.Millisecond).
		OnTaskClick(selectTask(send, "Hierarchical product roadmap"))
}

func controlledSchedule(send ui.Send[Msg], model Model) gantt.Widget {
	start := time.Date(2025, time.November, 3, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	interval := func(key string, from, to time.Time) (time.Time, time.Time) {
		if change, ok := model.changes[key]; ok {
			return change.Start, change.End
		}
		return from, to
	}
	projectStart, projectEnd := interval("project", day(0), day(26))
	apiStart, apiEnd := interval("api", day(2), day(10))
	uiStart, uiEnd := interval("ui", day(6), day(16))
	betaStart, betaEnd := interval("beta", day(17), day(23))
	launchAt := day(26)
	window := model.timeWindow
	if window.Start.IsZero() {
		window = gantt.TimeWindow{Start: day(1), End: day(27)}
	}
	hidden := make([]string, 0, len(model.hidden))
	for group, value := range model.hidden {
		if value {
			hidden = append(hidden, group)
		}
	}
	tasks := []gantt.Task{
		gantt.NewTask("project", "Project delivery", projectStart, projectEnd).Baseline(day(0), day(24)).Group("Program").Progress(.55).Collapsed(model.collapsed["project"]),
		gantt.NewTask("api", "API contract", apiStart, apiEnd).Baseline(day(1), day(9)).Group("Engineering").Progress(.9).Parent("project"),
		gantt.NewTask("ui", "Workspace UI", uiStart, uiEnd).Baseline(day(5), day(15)).Group("Experience").Progress(.65).Parent("project").DependsOn("api"),
		gantt.NewTask("beta", "Design partner beta", betaStart, betaEnd).Group("Validation").Progress(.2).DependsOn("api", "ui"),
		gantt.NewMilestone("release", "General availability", launchAt).Group("Release").DependsOn("beta"),
	}
	return gantt.New("controlled-schedule", tasks).
		TimeRange(day(-1), day(28)).TimeWindow(window.Start, window.End).
		HiddenGroups(hidden...).Legend(true).TaskLabels(true).TaskAxis("Work package").TimeAxis("November 2025").
		TimeTicks(7).AnimationDuration(450 * time.Millisecond).Editable(true).
		OnLegendChange(func(group string, hidden bool) { send(Msg{LegendGroup: group, LegendHidden: hidden}) }).
		OnTaskToggle(func(key string, collapsed bool) { send(Msg{ToggleKey: key, ToggleCollapsed: collapsed}) }).
		OnTaskChange(func(change gantt.TaskChange) { send(Msg{Change: change}) }).
		OnTimeWindowChange(func(start, end time.Time) { send(Msg{TimeWindow: gantt.TimeWindow{Start: start, End: end}}) }).
		OnTaskClick(selectTask(send, "Controlled interactive schedule"))
}

func selectTask(send ui.Send[Msg], scenario string) func(gantt.Selection) {
	return func(selection gantt.Selection) { send(Msg{Scenario: scenario, Selection: selection}) }
}

func ganttSection(title string, chart gantt.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(16),
		(surface.New(layout.Box(chart).Style(style.New().Padding(16)))).Style(style.New().Radius(8)),
	).Gap(8)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Gantt Charts"), window.Size(1120, 700), window.Theme(base.Light()))
}
