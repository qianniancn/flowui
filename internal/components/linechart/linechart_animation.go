package linechart

import (
	"math"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type lineChartAnimation struct {
	ready     bool
	revision  uint64
	from      chartData
	target    chartData
	displayed chartData
	duration  time.Duration
	easing    animation.Easing
}

func (w Widget) animatedData(ctx *frame.Context, gtx layout.Context, state *chartState, target chartData) chartData {
	transition := &state.animation
	if !transition.ready {
		transition.ready = true
		transition.revision = 1
		transition.from = lineBaselineData(target, w.animationBaseline(target))
		transition.target = target
		transition.duration = w.animationDuration
		transition.easing = w.animationEasing
	} else if !sameLineTarget(transition.target, target) {
		transition.revision++
		transition.from = lineTransitionFrom(transition.displayed, target, w.animationBaseline(target))
		transition.target = target
		transition.duration = w.updateAnimationDuration
		transition.easing = w.updateAnimationEasing
	} else {
		transition.target = target
	}

	progress, running := animation.Tween("data", 1).
		Initial(0).
		Revision(transition.revision).
		Duration(transition.duration).
		Easing(transition.easing).
		Disabled(!w.animation || !target.yExtent.Valid).
		Sample(ctx, gtx)
	if !running && progress == 1 {
		transition.from = chartData{}
		transition.displayed = target
		return target
	}
	transition.displayed = interpolateLineData(transition.from, target, progress, w.animationBaseline(target))
	return transition.displayed
}

func sameLineTarget(previous, target chartData) bool {
	if previous.generation != 0 || target.generation != 0 {
		return previous.generation == target.generation
	}
	return sameLineGeometry(previous, target)
}

func lineBaselineData(target chartData, value float64) chartData {
	baseline := cloneLineData(target)
	for seriesIndex := range baseline.series {
		for pointIndex := range baseline.series[seriesIndex].points {
			point := &baseline.series[seriesIndex].points[pointIndex]
			if point.valid {
				point.Y = value
				point.stackBase = value
				point.hasStackBase = true
			}
		}
	}
	return baseline
}

func lineTransitionFrom(previous, target chartData, baseline float64) chartData {
	from := lineBaselineData(target, baseline)
	previousSeries := make(map[string]resolvedSeries, len(previous.series))
	for _, series := range previous.series {
		previousSeries[series.key] = series
	}
	for seriesIndex := range from.series {
		old, ok := previousSeries[from.series[seriesIndex].key]
		if !ok {
			continue
		}
		for pointIndex := range from.series[seriesIndex].points {
			if pointIndex >= len(old.points) || !from.series[seriesIndex].points[pointIndex].valid || !old.points[pointIndex].valid {
				continue
			}
			point := &from.series[seriesIndex].points[pointIndex]
			point.Point = old.points[pointIndex].Point
			point.stackBase = old.points[pointIndex].stackBase
			point.hasStackBase = old.points[pointIndex].hasStackBase
		}
	}
	return from
}

func (w Widget) animationBaseline(target chartData) float64 {
	scale := w.resolveYScale(target)
	return min(max(float64(0), scale.Minimum), scale.Maximum)
}

func interpolateLineData(from, target chartData, progress float32, baseline float64) chartData {
	result := cloneLineData(target)
	for seriesIndex := range result.series {
		if seriesIndex >= len(from.series) || from.series[seriesIndex].key != result.series[seriesIndex].key {
			continue
		}
		for pointIndex := range result.series[seriesIndex].points {
			if pointIndex >= len(from.series[seriesIndex].points) {
				continue
			}
			start := from.series[seriesIndex].points[pointIndex]
			end := result.series[seriesIndex].points[pointIndex]
			if !start.valid || !end.valid {
				continue
			}
			point := &result.series[seriesIndex].points[pointIndex]
			point.X = animation.LerpFloat64(start.X, end.X, progress)
			point.Y = animation.LerpFloat64(start.Y, end.Y, progress)
			point.stackBase = animation.LerpFloat64(linePointStackBase(start, baseline), linePointStackBase(end, baseline), progress)
			point.hasStackBase = true
		}
	}
	return result
}

func cloneLineData(source chartData) chartData {
	result := source
	result.series = append([]resolvedSeries(nil), source.series...)
	for index := range result.series {
		result.series[index].points = append([]resolvedPoint(nil), source.series[index].points...)
	}
	return result
}

func sameLineGeometry(first, second chartData) bool {
	if len(first.series) != len(second.series) {
		return false
	}
	for seriesIndex := range first.series {
		left, right := first.series[seriesIndex], second.series[seriesIndex]
		if left.key != right.key || len(left.points) != len(right.points) {
			return false
		}
		for pointIndex := range left.points {
			leftPoint, rightPoint := left.points[pointIndex], right.points[pointIndex]
			if leftPoint.valid != rightPoint.valid {
				return false
			}
			if leftPoint.valid && (leftPoint.X != rightPoint.X || leftPoint.Y != rightPoint.Y || !sameLineStackBase(leftPoint, rightPoint)) {
				return false
			}
		}
	}
	return true
}

func linePointStackBase(point resolvedPoint, fallback float64) float64 {
	if point.hasStackBase && finite(point.stackBase) {
		return point.stackBase
	}
	return fallback
}

func sameLineStackBase(first, second resolvedPoint) bool {
	if first.hasStackBase != second.hasStackBase {
		return false
	}
	return !first.hasStackBase || first.stackBase == second.stackBase || math.IsNaN(first.stackBase) && math.IsNaN(second.stackBase)
}
