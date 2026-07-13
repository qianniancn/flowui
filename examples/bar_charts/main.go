package main

import (
	"image/color"
	"math"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	orders := ui.BarChart("quarterly-orders", []ui.BarChartSeries{
		ui.BarSeries("online", "Online", []float64{182, 214, 238, 276, 264, 312}).Radius(3).MaxWidth(38),
		ui.BarSeries("retail", "Retail", []float64{126, 148, 172, 164, math.NaN(), 198}).Radius(3).MaxWidth(38),
	}).
		Categories([]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}).
		XAxis("Month").
		YAxis("Orders").
		Label("Orders by channel").
		Height(320)

	traffic := ui.BarChart("weekly-traffic", []ui.BarChartSeries{
		ui.BarSeries("organic", "Organic", []float64{42, 48, 46, 54, 58, 62, 68}).Stack("visits").Radius(2).Background(true),
		ui.BarSeries("paid", "Paid", []float64{18, 22, 20, 26, 24, 28, 30}).Stack("visits").Radius(2),
		ui.BarSeries("returns", "Returns", []float64{-4, -5, -3, -6, -4, -5, -7}).Radius(2).MaxWidth(32),
	}).
		Categories([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		XAxis("Day").
		YAxis("Visits (thousands)").
		Label("Traffic composition and returns").
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
			MaxWidth(52),
	}).
		Categories([]string{"North", "East", "South", "West", "Central", "Overseas"}).
		XAxis("Region").
		YAxis("Revenue").
		Legend(false).
		Label("Revenue by region").
		Height(300)

	return ui.Scroll("bar-chart-page",
		ui.Box(
			ui.Column(
				ui.Text("Business overview").Size(24),
				ui.Text("Grouped, stacked, and item-colored category comparisons").Size(14),
				ui.Surface(ui.Box(orders).Padding(16)).Radius(8),
				ui.Surface(ui.Box(traffic).Padding(16)).Radius(8),
				ui.Surface(ui.Box(regions).Padding(16)).Radius(8),
			).Gap(16),
		).FillWidth().MaxWidth(980).Padding(24),
	)
}

func main() {
	ui.Run(
		Model{},
		Update,
		View,
		ui.Title("FlowUI Bar Charts"),
		ui.Size(1040, 820),
	)
}
