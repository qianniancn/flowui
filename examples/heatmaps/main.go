package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	chartpkg "github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/heatmap"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{ selected string }
type Msg struct{ Selection chartpkg.Selection }

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	model.selected = msg.Selection.Label
	return nil
}
func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	selected := "Click a cell"
	if model.selected != "" {
		selected = fmt.Sprintf("Selected: %s", model.selected)
	}
	yearStart := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)
	yearEnd := yearStart.AddDate(1, 0, -1)
	quarterStart := time.Date(2025, time.April, 1, 0, 0, 0, 0, time.Local)
	quarterEnd := quarterStart.AddDate(0, 3, -1)
	monthStart := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, -1)
	recentEnd := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.Local)
	recentStart := recentEnd.AddDate(0, 0, -29)
	rollingYearEnd := dateOnly(time.Now())
	rollingYearStart := rollingYearEnd.AddDate(0, 0, -364)

	click := func(selection chartpkg.Selection) { send(Msg{Selection: selection}) }
	year := heatmap.New("activity-year", yearStart, yearEnd, calendarValues(yearStart, yearEnd, 17)).Height(180).OnDataClick(click)
	quarter := heatmap.New("activity-quarter", quarterStart, quarterEnd, calendarValues(quarterStart, quarterEnd, 29)).Height(180).OnDataClick(click)
	month := heatmap.New("activity-month", monthStart, monthEnd, calendarValues(monthStart, monthEnd, 43)).Height(180).OnDataClick(click)
	recent := heatmap.New("activity-recent", recentStart, recentEnd, calendarValues(recentStart, recentEnd, 61)).Height(180).OnDataClick(click)
	rollingYear := heatmap.New("activity-rolling-year", rollingYearStart, rollingYearEnd, calendarValues(rollingYearStart, rollingYearEnd, 73)).Height(180).OnDataClick(click)

	return layout.Scroll(layout.Box(layout.Column(
		text.New("Activity calendars").Size(24),
		text.New(selected).Size(14),
		calendarSection("Full year · 2025", year),
		calendarSection("Quarter · Q2 2025", quarter),
		calendarSection("Month · March 2025", month),
		calendarSection("Recent 30 days", recent),
		calendarSection("Recent 365 days", rollingYear),
	).Gap(16)).Style(style.New().FillWidth().MaxWidth(900).Padding(24)))
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func calendarValues(start, end time.Time, seed int) []heatmap.CalendarValue {
	start = dateOnly(start)
	end = dateOnly(end)
	values := make([]heatmap.CalendarValue, 0)
	for day, date := 0, start; !date.After(end); day, date = day+1, date.AddDate(0, 0, 1) {
		values = append(values, heatmap.CalendarValue{Date: date, Value: float64((day*seed + day/7*5) % 100)})
	}
	return values
}

func calendarSection(title string, chart heatmap.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(16),
		(surface.New(layout.Box(chart).Style(style.New().Padding(16)))).Style(style.New().Radius(8)),
	).Gap(8)
}
func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Heatmaps"), window.Size(980, 700), window.Theme(base.Light()))
}
