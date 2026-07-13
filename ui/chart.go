package ui

import "github.com/qianniancn/FlowUI/internal/components/chart"

// ChartDatum describes one visible series value in a chart selection.
type ChartDatum = chart.Datum

// ChartSelection describes the values selected at one X position or category.
type ChartSelection = chart.Selection

// ChartDataWindow is a normalized visible range from the data start to end.
type ChartDataWindow = chart.DataWindow

type ChartAxis = chart.Axis
type ChartMarkLine = chart.MarkLine
type ChartMarkArea = chart.MarkArea
type ChartMarkPoint = chart.MarkPoint

const (
	ChartAxisX = chart.AxisX
	ChartAxisY = chart.AxisY
)

func MarkLine(axis ChartAxis, value float64) ChartMarkLine {
	return chart.NewMarkLine(axis, value)
}

func MarkArea(axis ChartAxis, start, end float64) ChartMarkArea {
	return chart.NewMarkArea(axis, start, end)
}

func MarkPoint(x, y float64) ChartMarkPoint {
	return chart.NewMarkPoint(x, y)
}
