package ui

import "github.com/qianniancn/flowui/internal/components/linechart"

type LineChartWidget = linechart.Widget
type LineChartSeries = linechart.Series
type LineChartPoint = linechart.Point
type LineChartStepMode = linechart.StepMode
type LineChartLineStyle = linechart.LineStyle
type LineChartSamplingMode = linechart.SamplingMode
type LineChartStackStrategy = linechart.StackStrategy
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

func LineChart(key string, series []LineChartSeries) LineChartWidget {
	return linechart.New(key, series)
}

func LineSeries(key, label string, values []float64) LineChartSeries {
	return linechart.Values(key, label, values)
}

func LineXYSeries(key, label string, points []LineChartPoint) LineChartSeries {
	return linechart.XY(key, label, points)
}
