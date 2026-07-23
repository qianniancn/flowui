package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	hidden   map[string]bool
	selected string
}

type Msg any

type LegendChanged struct {
	Chart  string
	Key    string
	Hidden bool
}

type DataClicked ui.ChartSelection

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case LegendChanged:
		if model.hidden == nil {
			model.hidden = make(map[string]bool)
		}
		model.hidden[msg.Chart+"/"+msg.Key] = msg.Hidden
	case DataClicked:
		selection := ui.ChartSelection(msg)
		model.selected = fmt.Sprintf("Selected: %s (%.2f%%)", selection.Label, selection.Items[0].Percent)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	standard := ui.PieChart("traffic-sources", trafficData(model, "pie")).
		Label("Traffic sources").
		OnLegendChange(func(key string, hidden bool) {
			send(LegendChanged{Chart: "pie", Key: key, Hidden: hidden})
		}).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		Height(360)

	donut := ui.PieChart("traffic-donut", trafficData(model, "donut")).
		InnerRadius(.4).
		OuterRadius(.7).
		PadAngle(2).
		Labels(false).
		Label("Traffic source share").
		OnLegendChange(func(key string, hidden bool) {
			send(LegendChanged{Chart: "donut", Key: key, Hidden: hidden})
		}).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		TooltipContent(func(selection ui.ChartSelection) ui.Widget {
			item := selection.Items[0]
			return ui.Column(
				ui.Text(selection.Label).Size(13),
				ui.Text(fmt.Sprintf("%.0f visits  %.2f%%", item.Y, item.Percent)).Size(12),
			).Gap(4)
		}).
		Height(340)

	rose := ui.PieChart("nightingale-rose", roseData()).
		RoseType(ui.PieRoseArea).
		InnerRadius(.15).
		OuterRadius(.78).
		PadAngle(1).
		Legend(false).
		Label("Nightingale rose").
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		Height(380)

	selected := model.selected
	if selected == "" {
		selected = "No data selected"
	}
	return ui.Scroll("pie-chart-page",
		ui.Box(
			ui.Column(
				ui.Text("Traffic overview").Size(24),
				ui.Text(selected).Size(14),
				ui.Surface(ui.Box(standard).Padding(16)).Style(ui.Radius(8)),
				ui.Surface(ui.Box(donut).Padding(16)).Style(ui.Radius(8)),
				ui.Surface(ui.Box(rose).Padding(16)).Style(ui.Radius(8)),
			).Gap(16),
		).FillWidth().MaxWidth(920).Padding(24),
	)
}

func roseData() []ui.PieChartData {
	return []ui.PieChartData{
		ui.PieData("rose", "Rose", 40),
		ui.PieData("lily", "Lily", 38),
		ui.PieData("tulip", "Tulip", 32),
		ui.PieData("iris", "Iris", 30),
		ui.PieData("orchid", "Orchid", 28),
		ui.PieData("lotus", "Lotus", 26),
		ui.PieData("daisy", "Daisy", 22),
		ui.PieData("violet", "Violet", 18),
	}
}

func trafficData(model Model, chart string) []ui.PieChartData {
	return []ui.PieChartData{
		ui.PieData("search", "Search Engine", 1048).Hidden(model.hidden[chart+"/search"]),
		ui.PieData("direct", "Direct", 735).Hidden(model.hidden[chart+"/direct"]),
		ui.PieData("email", "Email", 580).Hidden(model.hidden[chart+"/email"]),
		ui.PieData("union", "Union Ads", 484).Hidden(model.hidden[chart+"/union"]),
		ui.PieData("video", "Video Ads", 300).Hidden(model.hidden[chart+"/video"]),
	}
}

func main() {
	ui.Run(
		Model{hidden: make(map[string]bool)},
		Update,
		View,
		ui.Title("FlowUI Pie Charts"),
		ui.Size(980, 820),
	)
}
