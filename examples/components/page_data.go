package main

import (
	"fmt"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

var catalogTableColumns = []ui.TableColumn{
	{Key: "name", Label: "Component", RowHeader: true, Sortable: true, Resizable: true, MinWidth: 180, Weight: 1.4},
	{Key: "category", Label: "Category", Sortable: true, Resizable: true, MinWidth: 140, Weight: 1},
	{Key: "status", Label: "Status", Sortable: true, Width: 120},
}

var catalogTableRows = []ui.TableRow{
	{Key: "button", Label: "Button", Cells: []ui.TableCell{{Text: "Button"}, {Text: "Actions"}, {Text: "Stable", Content: ui.Chip("Stable").Color(ui.ChipSuccess).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "input", Label: "Input", Cells: []ui.TableCell{{Text: "Input"}, {Text: "Forms"}, {Text: "Stable", Content: ui.Chip("Stable").Color(ui.ChipSuccess).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "table", Label: "Table", Cells: []ui.TableCell{{Text: "Table"}, {Text: "Data"}, {Text: "Updated", Content: ui.Chip("Updated").Color(ui.ChipAccent).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "surface", Label: "Surface", Cells: []ui.TableCell{{Text: "Surface"}, {Text: "Content"}, {Text: "Updated", Content: ui.Chip("Updated").Color(ui.ChipAccent).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "tree", Label: "Tree", Cells: []ui.TableCell{{Text: "Tree"}, {Text: "Navigation"}, {Text: "Preview", Content: ui.Chip("Preview").Color(ui.ChipWarning).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "menu", Label: "Menu", Cells: []ui.TableCell{{Text: "Menu"}, {Text: "Navigation"}, {Text: "Stable", Content: ui.Chip("Stable").Color(ui.ChipSuccess).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "popover", Label: "Popover", Cells: []ui.TableCell{{Text: "Popover"}, {Text: "Feedback"}, {Text: "Stable", Content: ui.Chip("Stable").Color(ui.ChipSuccess).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
	{Key: "virtual-table", Label: "VirtualTable", Cells: []ui.TableCell{{Text: "VirtualTable"}, {Text: "Data"}, {Text: "Updated", Content: ui.Chip("Updated").Color(ui.ChipAccent).Variant(ui.ChipSoft).Size(ui.ChipSmall)}}},
}

func tablesPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	rows := sortedCatalogTableRows(model.TableSort)
	return demoPage("Tables",
		demoSection{Title: "Table", Content: demoPanel(
			ui.Table("catalog-table", catalogTableColumns, rows).
				Variant(ui.TableSecondary).
				GridLines(true).
				Border(true).
				Striped(true).
				MinWidth(620).
				SelectionMode(ui.TableSelectionMultiple).
				SelectedKeys(model.TableSelected).
				DisabledKeys([]string{"surface"}).
				AllowEmptySelection().
				ShowSelectionIndicator().
				SortDescriptor(model.TableSort).
				OnSelectionChange(func(keys []string) {
					send(func(model *Model) { model.TableSelected = append([]string(nil), keys...) })
				}).
				OnSortChange(func(descriptor ui.TableSortDescriptor) {
					send(func(model *Model) { model.TableSort = descriptor })
				}).
				OnAction(func(key string) {
					send(func(model *Model) { model.LastAction = "Selected " + key })
				}).
				OnRowActivate(func(key string) {
					send(func(model *Model) { model.LastAction = "Activated " + key })
				}).
				OnColumnResize(func(key string, width int) {
					send(func(model *Model) { model.LastAction = fmt.Sprintf("Resized %s to %d dp", key, width) })
				}).
				RowContextMenu(func(row ui.TableRow) ui.MenuWidget {
					return catalogTableRowMenu(row, send)
				}).
				Footer(ui.Text(fmt.Sprintf("%d components, %d selected, %s", len(rows), len(model.TableSelected), tableSortLabel(model.TableSort))).Size(12)),
		)},
		demoSection{Title: "VirtualTable", Content: demoPanel(
			ui.VirtualTable("catalog-virtual-table", catalogTableColumns, 1000, catalogVirtualRow).
				Variant(ui.TableSecondary).
				GridLines(true).
				Striped(true).
				MinWidth(620).
				MaxHeight(300).
				RowHeight(44),
		)},
		demoSection{Title: "Empty state", Content: demoPanel(
			ui.Table("catalog-empty-table", catalogTableColumns, nil).
				Variant(ui.TableSecondary).
				MinWidth(620).
				EmptyContent(ui.Column(
					ui.Icon(lucide.Inbox).Size(24),
					ui.Text("No components found").Size(14),
				).Gap(8).AlignMiddle()),
		)},
		demoSection{Title: "Last table action", Content: ui.Text(tableActionLabel(model)).Size(12)},
	)
}

func catalogTableRowMenu(row ui.TableRow, send ui.Send[Msg]) ui.MenuWidget {
	return ui.Menu("catalog-table-row-menu", []ui.MenuItem{
		{Key: "open", Label: "Open " + row.Label, Shortcut: "Enter", Leading: ui.Icon(lucide.ExternalLink).Size(16)},
		{Key: "copy-key", Label: "Copy key", Leading: ui.Icon(lucide.Copy).Size(16)},
		ui.MenuSeparator(),
		{Key: "disable", Label: "Disable component", Disabled: row.Key == "surface", Leading: ui.Icon(lucide.Ban).Size(16)},
		{Key: "delete", Label: "Remove from catalog", Variant: ui.MenuItemDanger, Leading: ui.Icon(lucide.Trash2).Size(16)},
	}).OnAction(func(action string) {
		send(func(model *Model) { model.LastAction = fmt.Sprintf("%s: %s", row.Label, action) })
	})
}

func tableSortLabel(descriptor ui.TableSortDescriptor) string {
	if descriptor.Column == "" {
		return "not sorted"
	}
	direction := "ascending"
	if descriptor.Direction == ui.TableSortDescending {
		direction = "descending"
	}
	return fmt.Sprintf("sorted by %s (%s)", descriptor.Column, direction)
}

func tableActionLabel(model Model) string {
	if model.LastAction == "" {
		return "No table action"
	}
	return model.LastAction
}

func catalogVirtualRow(index int) ui.TableRow {
	name := fmt.Sprintf("Component %04d", index+1)
	category := []string{"Content", "Actions", "Forms", "Navigation", "Feedback", "Data"}[index%6]
	return ui.TableRow{
		Key: name, Label: name,
		Cells: []ui.TableCell{
			{Text: name},
			{Text: category},
			{Content: ui.Chip("Ready").Color(ui.ChipSuccess).Variant(ui.ChipSoft).Size(ui.ChipSmall)},
		},
	}
}

func chartsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	selected := model.LastAction
	if selected == "" {
		selected = "Select a chart value"
	}
	clicked := func(selection ui.ChartSelection) {
		send(func(model *Model) { model.LastAction = "Selected " + selection.Label })
	}
	line := ui.LineChart("catalog-line-chart", []ui.LineChartSeries{
		ui.LineSeries("desktop", "Desktop", []float64{18, 24, 22, 31, 38, 42, 48}).Area(true).Smooth(true),
		ui.LineSeries("mobile", "Mobile", []float64{12, 16, 20, 19, 26, 30, 34}).LineStyle(ui.LineDashed),
	}).Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		XAxis("Day").YAxis("Sessions").Label("Sessions by client").OnDataClick(clicked).Style(ui.Height(300))
	bar := ui.BarChart("catalog-bar-chart", []ui.BarChartSeries{
		ui.BarSeries("stable", "Stable", []float64{24, 36, 42, 51}).Radius(3).ShowLabels(true),
		ui.BarSeries("preview", "Preview", []float64{8, 12, 15, 10}).Radius(3),
	}).Categories([]string{"Content", "Forms", "Navigation", "Data"}).
		CategoryAxis("Category").ValueAxis("Components").Label("Component maturity").OnDataClick(clicked).Style(ui.Height(300))
	pie := ui.PieChart("catalog-pie-chart", []ui.PieChartData{
		ui.PieData("content", "Content", 12),
		ui.PieData("forms", "Forms", 21),
		ui.PieData("navigation", "Navigation", 10),
		ui.PieData("feedback", "Feedback", 11),
		ui.PieData("data", "Data", 7),
	}).InnerRadius(.35).PadAngle(2).Label("Components by category").OnDataClick(clicked).Style(ui.Height(320))
	candles, times := catalogCandles()
	candlestick := ui.CandlestickChart("catalog-candlestick-chart", candles).
		Times(times).
		Label("Package activity").
		YAxis("Changes").
		OnDataClick(clicked).
		Height(320)
	heatmap := catalogHeatmap(clicked)
	gantt := catalogGantt(func(selection ui.GanttSelection) {
		send(func(model *Model) { model.LastAction = "Selected " + selection.Label })
	})

	return demoPage("Charts",
		demoSection{Title: "Selection", Content: ui.Text(selected).Size(13)},
		demoSection{Title: "LineChart", Content: demoPanel(line)},
		demoSection{Title: "BarChart", Content: demoPanel(bar)},
		demoSection{Title: "PieChart", Content: demoPanel(pie)},
		demoSection{Title: "CandlestickChart", Content: demoPanel(candlestick)},
		demoSection{Title: "Heatmap", Content: demoPanel(heatmap)},
		demoSection{Title: "GanttChart", Content: demoPanel(gantt)},
	)
}

func catalogCandles() ([]ui.CandlestickChartData, []time.Time) {
	values := [][4]float64{
		{18, 24, 16, 27}, {24, 22, 20, 29}, {22, 31, 21, 34}, {31, 28, 26, 36},
		{28, 38, 27, 41}, {38, 42, 35, 45}, {42, 39, 37, 46}, {39, 48, 38, 51},
	}
	candles := make([]ui.CandlestickChartData, len(values))
	times := make([]time.Time, len(values))
	start := time.Date(2026, time.July, 12, 0, 0, 0, 0, time.Local)
	for index, value := range values {
		candles[index] = ui.Candle(value[0], value[1], value[2], value[3])
		times[index] = start.AddDate(0, 0, index)
	}
	return candles, times
}

func catalogHeatmap(clicked func(ui.ChartSelection)) ui.HeatmapWidget {
	start := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 3, -1)
	values := make([]ui.HeatmapValue, 0, 91)
	for day, date := 0, start; !date.After(end); day, date = day+1, date.AddDate(0, 0, 1) {
		values = append(values, ui.HeatmapValue{
			Date:  date,
			Value: float64((day*17 + day/7*11) % 100),
		})
	}
	return ui.Heatmap("catalog-heatmap", start, end, values).
		Height(190).
		Label("Component activity").
		OnDataClick(clicked)
}

func catalogGantt(clicked func(ui.GanttSelection)) ui.GanttChartWidget {
	start := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []ui.GanttTask{
		ui.NewGanttTask("catalog-plan", "Plan component", day(0), day(4)).Group("Planning").Progress(1),
		ui.NewGanttTask("catalog-build", "Build component", day(4), day(11)).Group("Implementation").Progress(.65).DependsOn("catalog-plan"),
		ui.NewGanttTask("catalog-test", "Test component", day(9), day(14)).Group("Quality").Progress(.35).DependsOn("catalog-build"),
		ui.NewGanttMilestone("catalog-release", "Release", day(15)).Group("Delivery").DependsOn("catalog-test"),
	}
	return ui.GanttChart("catalog-gantt-chart", tasks).
		TimeRange(day(-1), day(17)).
		Height(330).
		Legend(true).
		TaskLabels(true).
		Dependencies(true).
		TaskAxis("Work").
		TimeAxis("Schedule").
		Label("Component delivery schedule").
		OnTaskClick(clicked)
}
