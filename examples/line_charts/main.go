package main

import (
	"fmt"
	"math"

	"github.com/qianniancn/FlowUI/ui"
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

func Update(model *Model, msg Msg) {
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
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	requestWindow := chartWindow(model, "request-volume")
	requests := ui.LineChart("request-volume", []ui.LineChartSeries{
		ui.LineSeries("api", "API", []float64{128, 146, 138, 172, 194, 188, 224, 236, 218, 252, 268, 286}).Area(true).PointSize(7).Sampling(ui.LineSamplingMinMax).Hidden(model.hidden["api"]),
		ui.LineSeries("worker", "Workers", []float64{82, 94, 106, 98, 126, math.NaN(), 142, 156, 164, 178, 184, 196}).LineStyle(ui.LineDashed).Hidden(model.hidden["worker"]),
		ui.LineSeries("cache", "Cache", []float64{56, 62, 68, 74, 72, math.NaN(), 88, 92, 98, 106, 112, 118}).ConnectNulls(true).Step(ui.LineStepEnd).Hidden(model.hidden["cache"]),
	}).
		Categories([]string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00"}).
		XAxis("Time").
		YAxis("Requests / min").
		Label("Request volume by service").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		TooltipContent(chartTooltip).
		MarkLines([]ui.ChartMarkLine{ui.MarkLine(ui.ChartAxisY, 220).Text("Capacity")}).
		MarkAreas([]ui.ChartMarkArea{ui.MarkArea(ui.ChartAxisY, 240, 300).Text("Watch")}).
		MarkPoints([]ui.ChartMarkPoint{ui.MarkPoint(11, 286).Text("Peak")}).
		DataWindow(float32(requestWindow.Start), float32(requestWindow.End)).
		OnDataWindowChange(func(window ui.ChartDataWindow) {
			send(DataWindowChanged{Chart: "request-volume", Window: window})
		}).
		Height(330)

	latencyWindow := chartWindow(model, "latency")
	latency := ui.LineChart("latency", []ui.LineChartSeries{
		ui.LineXYSeries("p50", "P50", []ui.LineChartPoint{
			{X: 0, Y: 18}, {X: 5, Y: 21}, {X: 10, Y: 19}, {X: 15, Y: 24}, {X: 20, Y: 23}, {X: 25, Y: 28},
		}).Smooth(true).ShowPoints(false).Area(true).Hidden(model.hidden["p50"]),
		ui.LineXYSeries("p95", "P95", []ui.LineChartPoint{
			{X: 0, Y: 42}, {X: 5, Y: 48}, {X: 10, Y: 45}, {X: 15, Y: 54}, {X: 20, Y: 52}, {X: 25, Y: 61},
		}).Smoothness(0.35).LineStyle(ui.LineDotted).Hidden(model.hidden["p95"]),
	}).
		XAxis("Minutes").
		YAxis("Latency (ms)").
		IncludeZero(false).
		Label("Request latency percentiles").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		DataWindow(float32(latencyWindow.Start), float32(latencyWindow.End)).
		OnDataWindowChange(func(window ui.ChartDataWindow) {
			send(DataWindowChanged{Chart: "latency", Window: window})
		}).
		Height(280)

	stackedTraffic := ui.LineChart("stacked-traffic", []ui.LineChartSeries{
		ui.LineSeries("stacked-direct", "Direct", []float64{42, 48, 46, 54, 58, 62, 68}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-direct"]),
		ui.LineSeries("stacked-referral", "Referral", []float64{24, 28, 26, 32, 30, 36, 38}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-referral"]),
		ui.LineSeries("stacked-campaign", "Campaign", []float64{16, 18, 22, 20, 24, 26, 30}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-campaign"]),
	}).
		Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		XAxis("Day").
		YAxis("Visits (thousands)").
		Label("Traffic sources, stacked total").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		Height(300)
	selection := "No data selected"
	if model.selected != "" {
		selection = "Selected: " + model.selected
	}

	return ui.Scroll("line-chart-page",
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Text("Service telemetry").Size(24)),
					ui.Button("reset-chart-view", ui.Text("Reset view")).
						Variant(ui.ButtonSecondary).
						OnClick(func() { send(ResetView{}) }),
				).AlignMiddle(),
				ui.Text(selection).Size(14),
				ui.Surface(ui.Box(requests).Style(ui.Padding(16).Radius(8))),
				ui.Surface(ui.Box(latency).Style(ui.Padding(16).Radius(8))),
				ui.Surface(ui.Box(stackedTraffic).Style(ui.Padding(16).Radius(8))),
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
	ui.Run(
		Model{hidden: make(map[string]bool), windows: make(map[string]ui.ChartDataWindow)},
		Update,
		View,
		ui.Title("FlowUI Line Charts"),
		ui.Size(1040, 820),
	)
}
