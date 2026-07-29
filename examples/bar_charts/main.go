package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	hidden   map[string]bool
	windows  map[string]ui.ChartDataWindow
	selected string
}

type Msg any

type LegendChanged struct {
	Key    string
	Hidden bool
}

type DataClicked ui.ChartSelection

type DataWindowChanged struct {
	Chart  string
	Window ui.ChartDataWindow
}

type ResetView struct{}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case LegendChanged:
		if model.hidden == nil {
			model.hidden = make(map[string]bool)
		}
		model.hidden[msg.Key] = msg.Hidden
	case DataClicked:
		selection := ui.ChartSelection(msg)
		model.selected = selection.Label
	case DataWindowChanged:
		if model.windows == nil {
			model.windows = make(map[string]ui.ChartDataWindow)
		}
		model.windows[msg.Chart] = msg.Window
	case ResetView:
		clear(model.windows)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	ordersWindow := chartWindow(model, "quarterly-orders")
	orders := ui.BarChart("quarterly-orders", []ui.BarChartSeries{
		ui.BarSeries("online", "Online", []float64{182, 214, 238, 276, 264, 312}).Radius(3).MaxWidth(38).ShowLabels(true).Hidden(model.hidden["online"]),
		ui.BarSeries("retail", "Retail", []float64{126, 148, 172, 164, math.NaN(), 198}).Radius(3).MaxWidth(38).Hidden(model.hidden["retail"]),
	}).
		Categories([]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}).
		CategoryAxis("Month").
		ValueAxis("Orders").
		Label("Orders by channel").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		TooltipContent(chartTooltip).
		MarkLines([]ui.ChartMarkLine{ui.MarkLine(ui.ChartAxisY, 250).Text("Target")}).
		MarkAreas([]ui.ChartMarkArea{ui.MarkArea(ui.ChartAxisY, 280, 330).Text("Stretch")}).
		MarkPoints([]ui.ChartMarkPoint{ui.MarkPoint(5, 312).Text("Peak")}).
		DataWindow(float32(ordersWindow.Start), float32(ordersWindow.End)).
		OnDataWindowChange(func(window ui.ChartDataWindow) {
			send(DataWindowChanged{Chart: "quarterly-orders", Window: window})
		}).
		Height(320)

	trafficWindow := chartWindow(model, "weekly-traffic")
	traffic := ui.BarChart("weekly-traffic", []ui.BarChartSeries{
		ui.BarSeries("organic", "Organic", []float64{42, 48, 46, 54, 58, 62, 68}).Stack("visits").Radius(2).Background(true).Hidden(model.hidden["organic"]),
		ui.BarSeries("paid", "Paid", []float64{18, 22, 20, 26, 24, 28, 30}).Stack("visits").Radius(2).Hidden(model.hidden["paid"]),
		ui.BarSeries("returns", "Returns", []float64{-4, -5, -3, -6, -4, -5, -7}).Radius(2).MaxWidth(32).Hidden(model.hidden["returns"]),
	}).
		Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		CategoryAxis("Day").
		ValueAxis("Visits (thousands)").
		Label("Traffic composition and returns").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		DataWindow(float32(trafficWindow.Start), float32(trafficWindow.End)).
		OnDataWindowChange(func(window ui.ChartDataWindow) {
			send(DataWindowChanged{Chart: "weekly-traffic", Window: window})
		}).
		Height(320)

	regions := ui.BarChart("regional-revenue", []ui.BarChartSeries{
		ui.BarSeries("revenue", "Revenue", []float64{86, 112, 98, 134, 121, 146}).
			ItemColors([]color.NRGBA{
				{R: 0x50, G: 0x70, B: 0xdd, A: 0xff},
				{R: 0xb6, G: 0xd6, B: 0x34, A: 0xff},
				{R: 0xff, G: 0x74, B: 0x6c, A: 0xff},
				{R: 0x0c, G: 0xa8, B: 0xdf, A: 0xff},
				{R: 0xff, G: 0xd1, B: 0x0a, A: 0xff},
				{R: 0x78, G: 0x5d, B: 0xb0, A: 0xff},
			}).
			Radius(4).
			MaxWidth(28).
			ShowLabels(true),
	}).
		Categories([]string{"North", "East", "South", "West", "Central", "Overseas"}).
		CategoryAxis("Region").
		ValueAxis("Revenue").
		Orientation(ui.BarHorizontal).
		Legend(false).
		Label("Revenue by region").
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		Height(340)
	selection := "No data selected"
	if model.selected != "" {
		selection = "Selected: " + model.selected
	}

	return ui.Scroll("bar-chart-page",
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Text("Business overview").Size(24)),
					ui.Button("reset-chart-view", ui.Text("Reset view")).
						Variant(ui.ButtonSecondary).
						OnClick(func() { send(ResetView{}) }),
				).AlignMiddle(),
				ui.Text(selection).Size(14),
				ui.Surface(ui.Box(orders).Style(ui.Padding(16).Radius(8))),
				ui.Surface(ui.Box(traffic).Style(ui.Padding(16).Radius(8))),
				ui.Surface(ui.Box(regions).Style(ui.Padding(16).Radius(8))),
			).Gap(16),
		).Style(ui.FillWidth().MaxWidth(980).Padding(24)),
	)
}

func chartWindow(model Model, key string) ui.ChartDataWindow {
	if window, ok := model.windows[key]; ok {
		return window
	}
	return ui.ChartDataWindow{End: 1}
}

func chartTooltip(selection ui.ChartSelection) ui.Widget {
	rows := []ui.Widget{ui.Text(selection.Label).Size(13)}
	for _, item := range selection.Items {
		rows = append(rows, ui.Text(fmt.Sprintf("%s  %.1f", item.SeriesLabel, item.Y)).Size(12))
	}
	return ui.Column(rows...).Gap(4)
}

func main() {
	ui.Run(ui.NewProgram(Model{hidden: make(map[string]bool), windows: make(map[string]ui.ChartDataWindow)},
		Update, View), ui.Title("FlowUI Bar Charts"),
		ui.Size(1040, 820),
	)
}
