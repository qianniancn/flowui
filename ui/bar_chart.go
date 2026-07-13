package ui

import "github.com/qianniancn/FlowUI/internal/components/barchart"

type BarChartWidget = barchart.Widget
type BarChartSeries = barchart.Series

func BarChart(key string, series []BarChartSeries) BarChartWidget {
	return barchart.New(key, series)
}

func BarSeries(key, label string, values []float64) BarChartSeries {
	return barchart.Values(key, label, values)
}
