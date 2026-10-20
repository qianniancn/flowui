package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	chartpkg "github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/piechart"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

type DataClicked chartpkg.Selection

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case LegendChanged:
		if model.hidden == nil {
			model.hidden = make(map[string]bool)
		}
		model.hidden[msg.Chart+"/"+msg.Key] = msg.Hidden
	case DataClicked:
		selection := chartpkg.Selection(msg)
		model.selected = fmt.Sprintf("Selected: %s (%.2f%%)", selection.Label, selection.Items[0].Percent)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	standard := piechart.New("traffic-sources", trafficData(model, "pie")).
		Label("Traffic sources").
		OnLegendChange(func(key string, hidden bool) {
			send(LegendChanged{Chart: "pie", Key: key, Hidden: hidden})
		}).
		OnDataClick(func(selection chartpkg.Selection) { send(DataClicked(selection)) }).
		Height(360)

	donut := piechart.New("traffic-donut", trafficData(model, "donut")).
		InnerRadius(.4).
		OuterRadius(.7).
		PadAngle(2).
		Labels(false).
		Label("Traffic source share").
		OnLegendChange(func(key string, hidden bool) {
			send(LegendChanged{Chart: "donut", Key: key, Hidden: hidden})
		}).
		OnDataClick(func(selection chartpkg.Selection) { send(DataClicked(selection)) }).
		TooltipContent(func(selection chartpkg.Selection) ui.Widget {
			item := selection.Items[0]
			return layout.Column(
				text.New(selection.Label).Size(13),
				text.New(fmt.Sprintf("%.0f visits  %.2f%%", item.Y, item.Percent)).Size(12),
			).Gap(4)
		}).
		Height(340)

	rose := piechart.New("nightingale-rose", roseData()).
		RoseType(piechart.RoseArea).
		InnerRadius(.15).
		OuterRadius(.78).
		PadAngle(1).
		Legend(false).
		Label("Nightingale rose").
		OnDataClick(func(selection chartpkg.Selection) { send(DataClicked(selection)) }).
		Height(380)

	selected := model.selected
	if selected == "" {
		selected = "No data selected"
	}
	return layout.Scroll(layout.Box(
		layout.Column(
			text.New("Traffic overview").Size(24),
			text.New(selected).Size(14),
			(surface.New(layout.Box(standard).Style(style.New().Padding(16)))).Style(style.New().Radius(8)),
			(surface.New(layout.Box(donut).Style(style.New().Padding(16)))).Style(style.New().Radius(8)),
			(surface.New(layout.Box(rose).Style(style.New().Padding(16)))).Style(style.New().Radius(8)),
		).Gap(16),
	).Style(style.New().FillWidth().MaxWidth(920).Padding(24)),
	)
}

func roseData() []piechart.Data {
	return []piechart.Data{
		piechart.Slice("rose", "Rose", 40),
		piechart.Slice("lily", "Lily", 38),
		piechart.Slice("tulip", "Tulip", 32),
		piechart.Slice("iris", "Iris", 30),
		piechart.Slice("orchid", "Orchid", 28),
		piechart.Slice("lotus", "Lotus", 26),
		piechart.Slice("daisy", "Daisy", 22),
		piechart.Slice("violet", "Violet", 18),
	}
}

func trafficData(model Model, chart string) []piechart.Data {
	return []piechart.Data{
		piechart.Slice("search", "Search Engine", 1048).Hidden(model.hidden[chart+"/search"]),
		piechart.Slice("direct", "Direct", 735).Hidden(model.hidden[chart+"/direct"]),
		piechart.Slice("email", "Email", 580).Hidden(model.hidden[chart+"/email"]),
		piechart.Slice("union", "Union Ads", 484).Hidden(model.hidden[chart+"/union"]),
		piechart.Slice("video", "Video Ads", 300).Hidden(model.hidden[chart+"/video"]),
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{hidden: make(map[string]bool)},
		Update, View), window.Title("FlowUI Pie Charts"),
		window.Size(980, 820),
		window.Theme(base.Light()),
	)
}
