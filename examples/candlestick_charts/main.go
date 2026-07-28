package main

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

//go:embed data.csv
var marketCSV string

var marketTimes, marketCandles, marketSignals = loadMarketData()

type Model struct {
	window   ui.ChartDataWindow
	selected string
}

type Msg any

type DataWindowChanged ui.ChartDataWindow
type DataClicked ui.ChartSelection
type ResetView struct{}

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case DataWindowChanged:
		model.window = ui.ChartDataWindow(msg)
	case DataClicked:
		selection := ui.ChartSelection(msg)
		item := selection.Items[0]
		model.selected = fmt.Sprintf("%s  Open %.2f  High %.2f  Low %.2f  Close %.2f", selection.Label, item.Open, item.High, item.Low, item.Close)
	case ResetView:
		model.window = ui.ChartDataWindow{End: 1}
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	chart := ui.CandlestickChart("btc-usdt", marketCandles).
		Times(marketTimes).
		Label("BTC/USDT 1h").
		YAxis("USDT").
		FormatY(func(value float64) string { return fmt.Sprintf("%.0f", value) }).
		DataWindow(float32(model.window.Start), float32(model.window.End)).
		OnDataWindowChange(func(window ui.ChartDataWindow) { send(DataWindowChanged(window)) }).
		OnDataClick(func(selection ui.ChartSelection) { send(DataClicked(selection)) }).
		MarkPoints(marketSignals).
		Height(440)

	selection := model.selected
	if selection == "" {
		selection = "No candle selected"
	}
	return ui.Scroll("candlestick-chart-page",
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Text("BTC/USDT 1h").Size(24)),
					ui.Button("reset-chart-view", ui.Text("Reset view")).
						Variant(ui.ButtonSecondary).
						OnClick(func() { send(ResetView{}) }),
				).AlignMiddle(),
				ui.Text(selection).Size(14),
				ui.Surface(ui.Box(chart).Style(ui.Padding(16).Radius(8))),
			).Gap(16),
		).Style(ui.FillWidth().MaxWidth(1100).Padding(24)),
	)
}

func loadMarketData() ([]time.Time, []ui.CandlestickChartData, []ui.ChartMarkPoint) {
	records, err := csv.NewReader(strings.NewReader(marketCSV)).ReadAll()
	if err != nil {
		panic(fmt.Errorf("read candlestick example data: %w", err))
	}
	if len(records) < 2 {
		panic("candlestick example data has no rows")
	}
	times := make([]time.Time, 0, len(records)-1)
	candles := make([]ui.CandlestickChartData, 0, len(records)-1)
	lows := make([]float64, 0, len(records)-1)
	highs := make([]float64, 0, len(records)-1)
	for row, record := range records[1:] {
		if len(record) < 5 {
			panic(fmt.Errorf("candlestick example data row %d has %d columns", row+2, len(record)))
		}
		times = append(times, parseMarketTime(record[0], row+2))
		low := mustPrice(record[3], row+2)
		high := mustPrice(record[2], row+2)
		candles = append(candles, ui.Candle(
			mustPrice(record[1], row+2),
			mustPrice(record[4], row+2),
			low,
			high,
		))
		lows = append(lows, low)
		highs = append(highs, high)
	}
	return times, candles, signalMarkers(lows, highs)
}

func mustPrice(value string, row int) float64 {
	price, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(fmt.Errorf("candlestick example data row %d: %w", row, err))
	}
	return price
}

func parseMarketTime(value string, row int) time.Time {
	for _, format := range []string{"2006-01-02 15:04:05", "2006/1/2 15:04"} {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed
		}
	}
	panic(fmt.Errorf("candlestick example data row %d has invalid time %q", row, value))
}

func signalMarkers(lows, highs []float64) []ui.ChartMarkPoint {
	buyIndex := len(lows) * 3 / 4
	sellIndex := len(highs) * 7 / 8
	return []ui.ChartMarkPoint{
		ui.MarkPoint(float64(buyIndex), lows[buyIndex]).
			Text("Buy").
			Color(color.NRGBA{R: 0x47, G: 0xb2, B: 0x62, A: 0xff}).
			Size(18).
			Content(ui.Icon(lucide.ArrowUp).Size(14)),
		ui.MarkPoint(float64(sellIndex), highs[sellIndex]).
			Text("Sell").
			Color(color.NRGBA{R: 0xeb, G: 0x54, B: 0x54, A: 0xff}).
			Size(18).
			Content(ui.Icon(lucide.ArrowDown).Size(14)),
	}
}

func main() {
	ui.Run(
		Model{window: ui.ChartDataWindow{Start: .5, End: 1}},
		Update,
		View,
		ui.Title("FlowUI Candlestick Chart"),
		ui.Size(1160, 720),
	)
}
