package ui

import "github.com/qianniancn/FlowUI/internal/components/piechart"

type PieChartWidget = piechart.Widget
type PieChartData = piechart.Data
type PieChartRoseType = piechart.RoseType

const (
	PieRoseNone   = piechart.RoseNone
	PieRoseRadius = piechart.RoseRadius
	PieRoseArea   = piechart.RoseArea
)

func PieChart(key string, data []PieChartData) PieChartWidget {
	return piechart.New(key, data)
}

func PieData(key, label string, value float64) PieChartData {
	return piechart.Slice(key, label, value)
}
