package ui

import "github.com/qianniancn/FlowUI/internal/components/piechart"

type PieChartWidget = piechart.Widget
type PieChartData = piechart.Data

func PieChart(key string, data []PieChartData) PieChartWidget {
	return piechart.New(key, data)
}

func PieData(key, label string, value float64) PieChartData {
	return piechart.Slice(key, label, value)
}
