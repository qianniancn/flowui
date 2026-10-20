package main

import (
	"fmt"
	"math"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/linechart"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	hidden   map[string]bool
	windows  map[string]chart.DataWindow
	selected string
}

type Msg any

type LegendChanged struct {
	Key    string
	Hidden bool
}

type DataClicked chart.Selection

type DataWindowChanged struct {
	Chart  string
	Window chart.DataWindow
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
		selection := chart.Selection(msg)
		model.selected = selection.Label
	case DataWindowChanged:
		if model.windows == nil {
			model.windows = make(map[string]chart.DataWindow)
		}
		model.windows[msg.Chart] = msg.Window
	case ResetView:
		clear(model.windows)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	requestWindow := chartWindow(model, "request-volume")
	requests := linechart.New("request-volume", []linechart.Series{
		linechart.Values("api", "API", []float64{128, 146, 138, 172, 194, 188, 224, 236, 218, 252, 268, 286}).Area(true).PointSize(7).Sampling(linechart.SamplingMinMax).Hidden(model.hidden["api"]),
		linechart.Values("worker", "Workers", []float64{82, 94, 106, 98, 126, math.NaN(), 142, 156, 164, 178, 184, 196}).LineStyle(linechart.LineDashed).Hidden(model.hidden["worker"]),
		linechart.Values("cache", "Cache", []float64{56, 62, 68, 74, 72, math.NaN(), 88, 92, 98, 106, 112, 118}).ConnectNulls(true).Step(linechart.StepEnd).Hidden(model.hidden["cache"]),
	}).
		Categories([]string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00"}).
		XAxis("Time").
		YAxis("Requests / min").
		Label("Request volume by service").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		TooltipContent(chartTooltip).
		MarkLines([]chart.MarkLine{chart.NewMarkLine(chart.AxisY, 220).Text("Capacity")}).
		MarkAreas([]chart.MarkArea{chart.NewMarkArea(chart.AxisY, 240, 300).Text("Watch")}).
		MarkPoints([]chart.MarkPoint{chart.NewMarkPoint(11, 286).Text("Peak")}).
		DataWindow(float32(requestWindow.Start), float32(requestWindow.End)).
		OnDataWindowChange(func(window chart.DataWindow) {
			send(DataWindowChanged{Chart: "request-volume", Window: window})
		}).
		Height(330)

	latencyWindow := chartWindow(model, "latency")
	latency := linechart.New("latency", []linechart.Series{
		linechart.XY("p50", "P50", []linechart.Point{
			{X: 0, Y: 18}, {X: 5, Y: 21}, {X: 10, Y: 19}, {X: 15, Y: 24}, {X: 20, Y: 23}, {X: 25, Y: 28},
		}).Smooth(true).ShowPoints(false).Area(true).Hidden(model.hidden["p50"]),
		linechart.XY("p95", "P95", []linechart.Point{
			{X: 0, Y: 42}, {X: 5, Y: 48}, {X: 10, Y: 45}, {X: 15, Y: 54}, {X: 20, Y: 52}, {X: 25, Y: 61},
		}).Smoothness(0.35).LineStyle(linechart.LineDotted).Hidden(model.hidden["p95"]),
	}).
		XAxis("Minutes").
		YAxis("Latency (ms)").
		IncludeZero(false).
		Label("Request latency percentiles").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		DataWindow(float32(latencyWindow.Start), float32(latencyWindow.End)).
		OnDataWindowChange(func(window chart.DataWindow) {
			send(DataWindowChanged{Chart: "latency", Window: window})
		}).
		Height(280)

	stackedTraffic := linechart.New("stacked-traffic", []linechart.Series{
		linechart.Values("stacked-direct", "Direct", []float64{42, 48, 46, 54, 58, 62, 68}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-direct"]),
		linechart.Values("stacked-referral", "Referral", []float64{24, 28, 26, 32, 30, 36, 38}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-referral"]),
		linechart.Values("stacked-campaign", "Campaign", []float64{16, 18, 22, 20, 24, 26, 30}).
			Stack("traffic").Area(true).Smooth(true).Hidden(model.hidden["stacked-campaign"]),
	}).
		Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		XAxis("Day").
		YAxis("Visits (thousands)").
		Label("Traffic sources, stacked total").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		Height(300)
	selection := "No data selected"
	if model.selected != "" {
		selection = "Selected: " + model.selected
	}

	return layout.Scroll(layout.Box(
		layout.Column(
			layout.Row(
				layout.Expanded(text.New("Service telemetry").Size(24)),
				button.New("Reset view").
					Variant(button.Secondary).
					OnClick(func() { send(ResetView{}) }),
			).AlignMiddle(),
			text.New(selection).Size(14),
			surface.New(layout.Box(requests).Style(style.New().Padding(16).Radius(8))),
			surface.New(layout.Box(latency).Style(style.New().Padding(16).Radius(8))),
			surface.New(layout.Box(stackedTraffic).Style(style.New().Padding(16).Radius(8))),
		).Gap(16),
	).Style(style.New().FillWidth().MaxWidth(980).Padding(24)),
	)
}

func chartWindow(model Model, key string) chart.DataWindow {
	if window, ok := model.windows[key]; ok {
		return window
	}
	return chart.DataWindow{End: 1}
}

func chartTooltip(selection chart.Selection) ui.Widget {
	rows := []ui.Widget{text.New(selection.Label).Size(13)}
	for _, item := range selection.Items {
		rows = append(rows, text.New(fmt.Sprintf("%s  %.1f", item.SeriesLabel, item.Y)).Size(12))
	}
	return layout.Column(rows...).Gap(4)
}

func main() {
	ui.Run(ui.NewProgram(Model{hidden: make(map[string]bool), windows: make(map[string]chart.DataWindow)},
		Update, View), window.Title("FlowUI Line Charts"),
		window.Size(1040, 820),
		window.Theme(base.Light()),
	)
}
