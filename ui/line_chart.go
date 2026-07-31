package ui

import "github.com/qianniancn/flowui/internal/components/linechart"

type LineChartWidget = linechart.Widget

// LineChartSeries contains the values and label for one line series.
type LineChartSeries = linechart.Series

// LineChartPoint stores one x/y value in a line series.
type LineChartPoint = linechart.Point

// LineChartStepMode controls how a stepped line changes between points.
type LineChartStepMode = linechart.StepMode

// LineChartLineStyle controls how a line is stroked.
type LineChartLineStyle = linechart.LineStyle

// LineChartSamplingMode controls how dense data is reduced for drawing.
type LineChartSamplingMode = linechart.SamplingMode

// LineChartStackStrategy controls how series are stacked.
type LineChartStackStrategy = linechart.StackStrategy

// LineChartStackOrder controls the order used when stacking series.
type LineChartStackOrder = linechart.StackOrder

const (
	LineStepNone   = linechart.StepNone
	LineStepStart  = linechart.StepStart
	LineStepMiddle = linechart.StepMiddle
	LineStepEnd    = linechart.StepEnd

	LineSolid  = linechart.LineSolid
	LineDashed = linechart.LineDashed
	LineDotted = linechart.LineDotted

	LineSamplingNone   = linechart.SamplingNone
	LineSamplingMinMax = linechart.SamplingMinMax

	LineStackSameSign = linechart.StackSameSign
	LineStackAll      = linechart.StackAll
	LineStackPositive = linechart.StackPositive
	LineStackNegative = linechart.StackNegative

	LineStackSeriesAscending  = linechart.StackSeriesAscending
	LineStackSeriesDescending = linechart.StackSeriesDescending
)

// LineChart creates a line chart for the supplied series.
func LineChart(key string, series []LineChartSeries) LineChartWidget {
	return linechart.New(key, series)
}

// LineSeries creates a series with evenly spaced values.
func LineSeries(key, label string, values []float64) LineChartSeries {
	return linechart.Values(key, label, values)
}

// LineXYSeries creates a series from explicit x/y points.
func LineXYSeries(key, label string, points []LineChartPoint) LineChartSeries {
	return linechart.XY(key, label, points)
}
