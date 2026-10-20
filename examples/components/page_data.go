package main

import (
	"fmt"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/barchart"
	candlestickpkg "github.com/qianniancn/flowui/widget/candlestick"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/chip"
	ganttpkg "github.com/qianniancn/flowui/widget/gantt"
	heatmappkg "github.com/qianniancn/flowui/widget/heatmap"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/linechart"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/piechart"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/timeline"
)

var catalogTableColumns = []table.Column{
	{Key: "name", Label: "Component", RowHeader: true, Sortable: true, Resizable: true, MinWidth: 180, Weight: 1.4},
	{Key: "category", Label: "Category", Sortable: true, Resizable: true, MinWidth: 140, Weight: 1},
	{Key: "status", Label: "Status", Sortable: true, Width: 120},
}

var catalogTableRows = []table.Row{
	{Key: "button", Label: "Button", Cells: []table.Cell{{Text: "Button"}, {Text: "Actions"}, {Text: "Stable", Content: chip.New("Stable").Color(chip.ColorSuccess).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "input", Label: "Input", Cells: []table.Cell{{Text: "Input"}, {Text: "Forms"}, {Text: "Stable", Content: chip.New("Stable").Color(chip.ColorSuccess).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "table", Label: "Table", Cells: []table.Cell{{Text: "Table"}, {Text: "Data"}, {Text: "Updated", Content: chip.New("Updated").Color(chip.ColorAccent).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "surface", Label: "Surface", Cells: []table.Cell{{Text: "Surface"}, {Text: "Content"}, {Text: "Updated", Content: chip.New("Updated").Color(chip.ColorAccent).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "tree", Label: "Tree", Cells: []table.Cell{{Text: "Tree"}, {Text: "Navigation"}, {Text: "Preview", Content: chip.New("Preview").Color(chip.ColorWarning).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "menu", Label: "Menu", Cells: []table.Cell{{Text: "Menu"}, {Text: "Navigation"}, {Text: "Stable", Content: chip.New("Stable").Color(chip.ColorSuccess).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "popover", Label: "Popover", Cells: []table.Cell{{Text: "Popover"}, {Text: "Feedback"}, {Text: "Stable", Content: chip.New("Stable").Color(chip.ColorSuccess).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
	{Key: "virtual-table", Label: "VirtualTable", Cells: []table.Cell{{Text: "VirtualTable"}, {Text: "Data"}, {Text: "Updated", Content: chip.New("Updated").Color(chip.ColorAccent).Variant(chip.VariantSoft).Size(chip.SizeSmall)}}},
}

func tablesPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	rows := sortedCatalogTableRows(model.TableSort)
	return demoPage("Tables",
		demoSection{Title: "Table", Content: demoPanel(
			table.New("catalog-table", catalogTableColumns, rows).
				Variant(table.VariantSecondary).
				GridLines(true).
				Border(true).
				Striped(true).
				MinWidth(620).
				SelectionMode(table.SelectionMultiple).
				SelectedKeys(model.TableSelected).
				DisabledKeys([]string{"surface"}).
				AllowEmptySelection().
				ShowSelectionIndicator().
				SortDescriptor(model.TableSort).
				OnSelectionChange(func(keys []string) {
					send(func(model *Model) { model.TableSelected = append([]string(nil), keys...) })
				}).
				OnSortChange(func(descriptor table.SortDescriptor) {
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
				RowContextMenu(func(row table.Row) menu.Widget {
					return catalogTableRowMenu(row, send)
				}).
				Footer(text.New(fmt.Sprintf("%d components, %d selected, %s", len(rows), len(model.TableSelected), tableSortLabel(model.TableSort))).Size(12)),
		)},
		demoSection{Title: "VirtualTable", Content: demoPanel(
			table.NewVirtual("catalog-virtual-table", catalogTableColumns, 1000, catalogVirtualRow).
				Variant(table.VariantSecondary).
				GridLines(true).
				Striped(true).
				MinWidth(620).
				MaxHeight(300).
				RowHeight(44),
		)},
		demoSection{Title: "Empty state", Content: demoPanel(
			table.New("catalog-empty-table", catalogTableColumns, nil).
				Variant(table.VariantSecondary).
				MinWidth(620).
				EmptyContent(layout.Column(
					icon.New(lucide.Inbox).Size(24),
					text.New("No components found").Size(14),
				).Gap(8).AlignMiddle()),
		)},
		demoSection{Title: "Last table action", Content: text.New(tableActionLabel(model)).Size(12)},
	)
}

func catalogTableRowMenu(row table.Row, send ui.Send[Msg]) menu.Widget {
	return menu.New("catalog-table-row-menu", []menu.Item{
		{Key: "open", Label: "Open " + row.Label, Shortcut: "Enter", Leading: icon.New(lucide.ExternalLink).Size(16)},
		{Key: "copy-key", Label: "Copy key", Leading: icon.New(lucide.Copy).Size(16)},
		menu.Separator(),
		{Key: "disable", Label: "Disable component", Disabled: row.Key == "surface", Leading: icon.New(lucide.Ban).Size(16)},
		{Key: "delete", Label: "Remove from catalog", Variant: menu.ItemDanger, Leading: icon.New(lucide.Trash).Size(16)},
	}).OnActionEvent(func(event menu.ActionEvent) {
		send(func(model *Model) { model.LastAction = fmt.Sprintf("%s: %s", row.Label, event.Key) })
	})
}

func tableSortLabel(descriptor table.SortDescriptor) string {
	if descriptor.Column == "" {
		return "not sorted"
	}
	direction := "ascending"
	if descriptor.Direction == table.SortDescending {
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

func catalogVirtualRow(index int) table.Row {
	name := fmt.Sprintf("Component %04d", index+1)
	category := []string{"Content", "Actions", "Forms", "Navigation", "Feedback", "Data"}[index%6]
	return table.Row{
		Key: name, Label: name,
		Cells: []table.Cell{
			{Text: name},
			{Text: category},
			{Content: chip.New("Ready").Color(chip.ColorSuccess).Variant(chip.VariantSoft).Size(chip.SizeSmall)},
		},
	}
}

func chartsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	selected := model.LastAction
	if selected == "" {
		selected = "Select a chart value"
	}
	clicked := func(selection chart.Selection) {
		send(func(model *Model) { model.LastAction = "Selected " + selection.Label })
	}
	line := (linechart.New("catalog-line-chart", []linechart.Series{
		linechart.Values("desktop", "Desktop", []float64{18, 24, 22, 31, 38, 42, 48}).Area(true).Smooth(true),
		linechart.Values("mobile", "Mobile", []float64{12, 16, 20, 19, 26, 30, 34}).LineStyle(linechart.LineDashed),
	}).Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		XAxis("Day").YAxis("Sessions").Label("Sessions by client").OnDataClick(clicked)).Style(style.New().Height(300))
	bar := (barchart.New("catalog-bar-chart", []barchart.Series{
		barchart.Values("stable", "Stable", []float64{24, 36, 42, 51}).Radius(3).ShowLabels(true),
		barchart.Values("preview", "Preview", []float64{8, 12, 15, 10}).Radius(3),
	}).Categories([]string{"Content", "Forms", "Navigation", "Data"}).
		CategoryAxis("Category").ValueAxis("Components").Label("Component maturity").OnDataClick(clicked)).Style(style.New().Height(300))
	pie := (piechart.New("catalog-pie-chart", []piechart.Data{
		piechart.Slice("content", "Content", 12),
		piechart.Slice("forms", "Forms", 21),
		piechart.Slice("navigation", "Navigation", 10),
		piechart.Slice("feedback", "Feedback", 11),
		piechart.Slice("data", "Data", 7),
	}).InnerRadius(.35).PadAngle(2).Label("Components by category").OnDataClick(clicked)).Style(style.New().Height(320))
	candles, times := catalogCandles()
	candlestick := candlestickpkg.New("catalog-candlestick-chart", candles).
		Times(times).
		Label("Package activity").
		YAxis("Changes").
		OnDataClick(clicked).
		Height(320)
	heatmap := catalogHeatmap(clicked)
	gantt := catalogGantt(func(selection ganttpkg.Selection) {
		send(func(model *Model) { model.LastAction = "Selected " + selection.Label })
	})
	graph := nodegraph.New("catalog-node-graph", nodegraph.Graph{
		Nodes: model.GraphNodes,
		Edges: model.GraphEdges,
	}).Height(320).Viewport(model.GraphViewport).OnViewportChange(func(viewport nodegraph.Viewport) {
		send(func(model *Model) { model.GraphViewport = viewport })
	}).OnNodesChange(func(changes []nodegraph.NodeChange) {
		send(func(model *Model) { model.GraphNodes = nodegraph.ApplyNodeChanges(model.GraphNodes, changes) })
	}).OnConnect(func(connection nodegraph.Connection) {
		send(func(model *Model) { model.GraphEdges = appendCatalogNodeGraphConnection(model.GraphEdges, connection) })
	})
	timeLine := timeline.New([]timeline.Item{
		{Title: text.New("Created"), Content: text.New("Order submitted"), Color: timeline.ColorBlue},
		{Title: text.New("Paid"), Content: text.New("Payment received"), Color: timeline.ColorGreen},
		{Title: text.New("Shipped"), Content: text.New("Waiting for carrier"), Color: timeline.ColorGray},
	}).Mode(timeline.ModeAlternate)

	return demoPage("Charts",
		demoSection{Title: "Selection", Content: text.New(selected).Size(13)},
		demoSection{Title: "LineChart", Content: demoPanel(line)},
		demoSection{Title: "BarChart", Content: demoPanel(bar)},
		demoSection{Title: "PieChart", Content: demoPanel(pie)},
		demoSection{Title: "CandlestickChart", Content: demoPanel(candlestick)},
		demoSection{Title: "Heatmap", Content: demoPanel(heatmap)},
		demoSection{Title: "GanttChart", Content: demoPanel(gantt)},
		demoSection{Title: "NodeGraph", Content: graph},
		demoSection{Title: "TimeLine", Content: demoPanel(timeLine)},
	)
}

func catalogNodeGraphNodes() []nodegraph.Node {
	return []nodegraph.Node{
		nodegraph.NewNode("source", "Source", nodegraph.Point{X: 60, Y: 90}).
			Outputs(nodegraph.NewPort("data", "Data")),
		nodegraph.NewNode("transform", "Transform", nodegraph.Point{X: 340, Y: 130}).
			Inputs(nodegraph.NewPort("input", "Input")).
			Outputs(nodegraph.NewPort("output", "Output")),
	}
}

func catalogNodeGraphEdges() []nodegraph.Edge {
	return []nodegraph.Edge{
		nodegraph.NewEdge("source-data", nodegraph.NewEndpoint("source", "data"), nodegraph.NewEndpoint("transform", "input")),
	}
}

func appendCatalogNodeGraphConnection(edges []nodegraph.Edge, connection nodegraph.Connection) []nodegraph.Edge {
	for _, edge := range edges {
		if edge.Source == connection.Source && edge.Target == connection.Target {
			return edges
		}
	}
	id := "connection-" + connection.Source.NodeID + "-" + connection.Source.PortID + "-" + connection.Target.NodeID + "-" + connection.Target.PortID
	return append(edges, nodegraph.NewEdge(id, connection.Source, connection.Target))
}

func catalogCandles() ([]candlestickpkg.Candle, []time.Time) {
	values := [][4]float64{
		{18, 24, 16, 27}, {24, 22, 20, 29}, {22, 31, 21, 34}, {31, 28, 26, 36},
		{28, 38, 27, 41}, {38, 42, 35, 45}, {42, 39, 37, 46}, {39, 48, 38, 51},
	}
	candles := make([]candlestickpkg.Candle, len(values))
	times := make([]time.Time, len(values))
	start := time.Date(2026, time.July, 12, 0, 0, 0, 0, time.Local)
	for index, value := range values {
		candles[index] = candlestickpkg.OHLC(value[0], value[1], value[2], value[3])
		times[index] = start.AddDate(0, 0, index)
	}
	return candles, times
}

func catalogHeatmap(clicked func(chart.Selection)) heatmappkg.Widget {
	start := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 3, -1)
	values := make([]heatmappkg.CalendarValue, 0, 91)
	for day, date := 0, start; !date.After(end); day, date = day+1, date.AddDate(0, 0, 1) {
		values = append(values, heatmappkg.CalendarValue{
			Date:  date,
			Value: float64((day*17 + day/7*11) % 100),
		})
	}
	return heatmappkg.New("catalog-heatmap", start, end, values).
		Height(190).
		Label("Component activity").
		OnDataClick(clicked)
}

func catalogGantt(clicked func(ganttpkg.Selection)) ganttpkg.Widget {
	start := time.Date(2026, time.July, 6, 0, 0, 0, 0, time.Local)
	day := func(offset int) time.Time { return start.AddDate(0, 0, offset) }
	tasks := []ganttpkg.Task{
		ganttpkg.NewTask("catalog-plan", "Plan component", day(0), day(4)).Group("Planning").Progress(1),
		ganttpkg.NewTask("catalog-build", "Build component", day(4), day(11)).Group("Implementation").Progress(.65).DependsOn("catalog-plan"),
		ganttpkg.NewTask("catalog-test", "Test component", day(9), day(14)).Group("Quality").Progress(.35).DependsOn("catalog-build"),
		ganttpkg.NewMilestone("catalog-release", "Release", day(15)).Group("Delivery").DependsOn("catalog-test"),
	}
	return ganttpkg.New("catalog-gantt-chart", tasks).
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
