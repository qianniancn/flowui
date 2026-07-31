package ui

import "github.com/qianniancn/flowui/internal/components/chart"

// ChartDatum describes one visible value in a chart selection.
type ChartDatum = chart.Datum

// ChartSelection describes the current chart selection.
type ChartSelection = chart.Selection

// ChartDataWindow is a normalized visible range from the data start to end.
type ChartDataWindow = chart.DataWindow

// ChartAxis identifies the x or y axis used by an annotation.
type ChartAxis = chart.Axis

// ChartMarkLine describes a reference line on a chart.
type ChartMarkLine = chart.MarkLine

// ChartMarkArea describes a shaded range on a chart.
type ChartMarkArea = chart.MarkArea

// ChartMarkPoint describes a point annotation on a chart.
type ChartMarkPoint = chart.MarkPoint

const (
	ChartAxisX = chart.AxisX
	ChartAxisY = chart.AxisY
)

// MarkLine creates a horizontal or vertical reference line.
func MarkLine(axis ChartAxis, value float64) ChartMarkLine {
	return chart.NewMarkLine(axis, value)
}

// MarkArea creates a shaded reference range on one chart axis.
func MarkArea(axis ChartAxis, start, end float64) ChartMarkArea {
	return chart.NewMarkArea(axis, start, end)
}

// MarkPoint creates a point annotation at x and y.
func MarkPoint(x, y float64) ChartMarkPoint {
	return chart.NewMarkPoint(x, y)
}
