package ui

import "github.com/qianniancn/FlowUI/internal/components/linechart"

type LineChartWidget = linechart.Widget
type LineChartSeries = linechart.Series
type LineChartPoint = linechart.Point

func LineChart(key string, series []LineChartSeries) LineChartWidget {
	return linechart.New(key, series)
}

func LineSeries(key, label string, values []float64) LineChartSeries {
	return linechart.Values(key, label, values)
}

func LineXYSeries(key, label string, points []LineChartPoint) LineChartSeries {
	return linechart.XY(key, label, points)
}
