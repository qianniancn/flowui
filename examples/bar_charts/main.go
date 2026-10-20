package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/barchart"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chart"
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
	ordersWindow := chartWindow(model, "quarterly-orders")
	orders := barchart.New("quarterly-orders", []barchart.Series{
		barchart.Values("online", "Online", []float64{182, 214, 238, 276, 264, 312}).Radius(3).MaxWidth(38).ShowLabels(true).Hidden(model.hidden["online"]),
		barchart.Values("retail", "Retail", []float64{126, 148, 172, 164, math.NaN(), 198}).Radius(3).MaxWidth(38).Hidden(model.hidden["retail"]),
	}).
		Categories([]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}).
		CategoryAxis("Month").
		ValueAxis("Orders").
		Label("Orders by channel").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		TooltipContent(chartTooltip).
		MarkLines([]chart.MarkLine{chart.NewMarkLine(chart.AxisY, 250).Text("Target")}).
		MarkAreas([]chart.MarkArea{chart.NewMarkArea(chart.AxisY, 280, 330).Text("Stretch")}).
		MarkPoints([]chart.MarkPoint{chart.NewMarkPoint(5, 312).Text("Peak")}).
		DataWindow(float32(ordersWindow.Start), float32(ordersWindow.End)).
		OnDataWindowChange(func(window chart.DataWindow) {
			send(DataWindowChanged{Chart: "quarterly-orders", Window: window})
		}).
		Height(320)

	trafficWindow := chartWindow(model, "weekly-traffic")
	traffic := barchart.New("weekly-traffic", []barchart.Series{
		barchart.Values("organic", "Organic", []float64{42, 48, 46, 54, 58, 62, 68}).Stack("visits").Radius(2).Background(true).Hidden(model.hidden["organic"]),
		barchart.Values("paid", "Paid", []float64{18, 22, 20, 26, 24, 28, 30}).Stack("visits").Radius(2).Hidden(model.hidden["paid"]),
		barchart.Values("returns", "Returns", []float64{-4, -5, -3, -6, -4, -5, -7}).Radius(2).MaxWidth(32).Hidden(model.hidden["returns"]),
	}).
		Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		CategoryAxis("Day").
		ValueAxis("Visits (thousands)").
		Label("Traffic composition and returns").
		OnLegendChange(func(key string, hidden bool) { send(LegendChanged{Key: key, Hidden: hidden}) }).
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		DataWindow(float32(trafficWindow.Start), float32(trafficWindow.End)).
		OnDataWindowChange(func(window chart.DataWindow) {
			send(DataWindowChanged{Chart: "weekly-traffic", Window: window})
		}).
		Height(320)

	regions := barchart.New("regional-revenue", []barchart.Series{
		barchart.Values("revenue", "Revenue", []float64{86, 112, 98, 134, 121, 146}).
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
		Orientation(barchart.Horizontal).
		Legend(false).
		Label("Revenue by region").
		OnDataClick(func(selection chart.Selection) { send(DataClicked(selection)) }).
		Height(340)
	selection := "No data selected"
	if model.selected != "" {
		selection = "Selected: " + model.selected
	}

	return layout.Scroll(layout.Box(
		layout.Column(
			layout.Row(
				layout.Expanded(text.New("Business overview").Size(24)),
				button.New("Reset view").
					Variant(button.Secondary).
					OnClick(func() { send(ResetView{}) }),
			).AlignMiddle(),
			text.New(selection).Size(14),
			surface.New(layout.Box(orders).Style(style.New().Padding(16).Radius(8))),
			surface.New(layout.Box(traffic).Style(style.New().Padding(16).Radius(8))),
			surface.New(layout.Box(regions).Style(style.New().Padding(16).Radius(8))),
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
		Update, View), window.Title("FlowUI Bar Charts"),
		window.Size(1040, 820),
		window.Theme(base.Light()),
	)
}
