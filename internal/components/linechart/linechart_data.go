package linechart

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type resolvedPoint struct {
	Point
	rawY         float64
	stackBase    float64
	hasStackBase bool
	valid        bool
}

type resolvedSeries struct {
	key             string
	label           string
	hidden          bool
	points          []resolvedPoint
	color           color.NRGBA
	width           float32
	showPoints      bool
	pointsSet       bool
	pointSize       int
	connectNulls    bool
	smooth          float32
	step            StepMode
	lineStyle       LineStyle
	area            bool
	areaColor       color.NRGBA
	sampling        SamplingMode
	stack           string
	stackStrategy   StackStrategy
	stackOrder      StackOrder
	stackedOnSmooth float32
}

type dataExtent struct {
	minimum float64
	maximum float64
	valid   bool
}

func (e *dataExtent) include(value float64) {
	if !finite(value) {
		return
	}
	if !e.valid {
		e.minimum, e.maximum, e.valid = value, value, true
		return
	}
	e.minimum = min(e.minimum, value)
	e.maximum = max(e.maximum, value)
}

type chartData struct {
	series  []resolvedSeries
	legend  []resolvedSeries
	xExtent dataExtent
	yExtent dataExtent
	xValues []float64
}

func resolveChartData(widget Widget, activeTheme *theme.Theme, dp func(unit.Dp) int) chartData {
	seen := make(map[string]struct{}, len(widget.series))
	data := chartData{}
	xSeen := make(map[uint64]struct{})
	for index, source := range widget.series {
		if source.key == "" {
			panic("flowui: empty line chart series key")
		}
		if _, exists := seen[source.key]; exists {
			panic(fmt.Sprintf("flowui: duplicate line chart series key %q", source.key))
		}
		seen[source.key] = struct{}{}
		label := source.label
		if label == "" {
			label = source.key
		}
		lineColor := source.color
		if !source.hasColor {
			colors := activeTheme.Components.LineChart.SeriesColors
			lineColor = colors[index%len(colors)]
		}
		width := float32(dp(activeTheme.Components.LineChart.LineWidth))
		if source.width > 0 {
			width = float32(dp(source.width))
		}
		width = max(width, 1)

		resolved := resolvedSeries{
			key:           source.key,
			label:         label,
			hidden:        source.hidden,
			color:         lineColor,
			width:         width,
			showPoints:    !source.hasShowPoints || source.showPoints,
			pointsSet:     source.hasShowPoints,
			pointSize:     dp(source.pointSize),
			connectNulls:  source.connectNulls,
			smooth:        source.smooth,
			step:          source.step,
			lineStyle:     source.lineStyle,
			area:          source.area,
			areaColor:     source.areaColor,
			sampling:      source.sampling,
			stack:         source.stack,
			stackStrategy: source.stackStrategy,
			stackOrder:    source.stackOrder,
		}
		if source.area && !source.hasAreaColor {
			resolved.areaColor = lineColor
			resolved.areaColor.A = 0x30
		}
		data.legend = append(data.legend, resolved)
		if source.hidden {
			continue
		}

		points := source.resolvedPoints()
		resolved.points = make([]resolvedPoint, len(points))
		for pointIndex, point := range points {
			validX := finite(point.X)
			valid := validX && finite(point.Y)
			resolved.points[pointIndex] = resolvedPoint{Point: point, rawY: point.Y, valid: valid}
			if !valid {
				continue
			}
			data.xExtent.include(point.X)
			bits := math.Float64bits(point.X)
			if point.X == 0 {
				bits = 0
			}
			if _, exists := xSeen[bits]; !exists {
				xSeen[bits] = struct{}{}
				data.xValues = append(data.xValues, point.X)
			}
		}
		data.series = append(data.series, resolved)
	}

	applyLineStacks(&data, len(widget.categories) > 0)
	for _, series := range data.series {
		for _, point := range series.points {
			if point.valid {
				data.yExtent.include(point.Y)
			}
		}
	}

	if len(widget.categories) > 0 {
		data.xExtent.include(0)
		data.xExtent.include(float64(len(widget.categories) - 1))
	}
	sort.Float64s(data.xValues)
	return data
}

func (s Series) resolvedPoints() []Point {
	if s.points != nil {
		return s.points
	}
	points := make([]Point, len(s.values))
	for index, value := range s.values {
		points[index] = Point{X: float64(index), Y: value}
	}
	return points
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
