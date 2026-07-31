package ui

import "github.com/qianniancn/flowui/internal/components/piechart"

type PieChartWidget = piechart.Widget

// PieChartData describes one labeled slice.
type PieChartData = piechart.Data

// PieChartRoseType controls how a pie chart varies slice radius.
type PieChartRoseType = piechart.RoseType

const (
	PieRoseNone   = piechart.RoseNone
	PieRoseRadius = piechart.RoseRadius
	PieRoseArea   = piechart.RoseArea
)

// PieChart creates a pie or rose chart from data slices.
func PieChart(key string, data []PieChartData) PieChartWidget {
	return piechart.New(key, data)
}

// PieData creates one labeled pie slice.
func PieData(key, label string, value float64) PieChartData {
	return piechart.Slice(key, label, value)
}
