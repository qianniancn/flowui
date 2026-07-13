package main

import (
	"math"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	requests := ui.LineChart("request-volume", []ui.LineChartSeries{
		ui.LineSeries("api", "API", []float64{128, 146, 138, 172, 194, 188, 224, 236, 218, 252, 268, 286}),
		ui.LineSeries("worker", "Workers", []float64{82, 94, 106, 98, 126, math.NaN(), 142, 156, 164, 178, 184, 196}),
		ui.LineSeries("cache", "Cache", []float64{56, 62, 68, 74, 72, math.NaN(), 88, 92, 98, 106, 112, 118}).ConnectNulls(true),
	}).
		Categories([]string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00"}).
		XAxis("Time").
		YAxis("Requests / min").
		Label("Request volume by service").
		Height(330)

	latency := ui.LineChart("latency", []ui.LineChartSeries{
		ui.LineXYSeries("p50", "P50", []ui.LineChartPoint{
			{X: 0, Y: 18}, {X: 5, Y: 21}, {X: 10, Y: 19}, {X: 15, Y: 24}, {X: 20, Y: 23}, {X: 25, Y: 28},
		}).Smooth(true).ShowPoints(false),
		ui.LineXYSeries("p95", "P95", []ui.LineChartPoint{
			{X: 0, Y: 42}, {X: 5, Y: 48}, {X: 10, Y: 45}, {X: 15, Y: 54}, {X: 20, Y: 52}, {X: 25, Y: 61},
		}).Smoothness(0.35),
	}).
		XAxis("Minutes").
		YAxis("Latency (ms)").
		IncludeZero(false).
		Label("Request latency percentiles").
		Height(280)

	return ui.Scroll("line-chart-page",
		ui.Box(
			ui.Column(
				ui.Text("Service telemetry").Size(24),
				ui.Text("Traffic and latency across the current deployment window").Size(14),
				ui.Surface(ui.Box(requests).Padding(16)).Radius(8),
				ui.Surface(ui.Box(latency).Padding(16)).Radius(8),
			).Gap(16),
		).FillWidth().MaxWidth(980).Padding(24),
	)
}

func main() {
	ui.Run(
		Model{},
		Update,
		View,
		ui.Title("FlowUI Line Charts"),
		ui.Size(1040, 820),
	)
}
