package ui

import "github.com/qianniancn/flowui/internal/components/barchart"

type BarChartWidget = barchart.Widget

// BarChartSeries contains the values and label for one bar series.
type BarChartSeries = barchart.Series

// BarChartLabelPosition controls where value labels are drawn.
type BarChartLabelPosition = barchart.LabelPosition

// BarChartOrientation selects vertical or horizontal bars.
type BarChartOrientation = barchart.Orientation

const (
	BarLabelAuto    = barchart.LabelAuto
	BarLabelInside  = barchart.LabelInside
	BarLabelOutside = barchart.LabelOutside

	BarVertical   = barchart.Vertical
	BarHorizontal = barchart.Horizontal
)

// BarChart creates a chart for the supplied keyed series.
func BarChart(key string, series []BarChartSeries) BarChartWidget {
	return barchart.New(key, series)
}

// BarSeries creates a bar series from one value per category.
func BarSeries(key, label string, values []float64) BarChartSeries {
	return barchart.Values(key, label, values)
}
