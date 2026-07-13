package barchart

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type barChartAnimation struct {
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
		transition.from = barBaselineData(target)
		transition.target = target
		transition.duration = w.animationDuration
		transition.easing = w.animationEasing
	} else if !sameBarGeometry(transition.target, target) {
		transition.revision++
		transition.from = barTransitionFrom(transition.displayed, target)
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
		Disabled(!w.animation || !target.yExtent.valid).
		Sample(ctx, gtx)
	if !running && progress == 1 {
		transition.from = chartData{}
		transition.displayed = target
		return target
	}
	transition.displayed = interpolateBarData(transition.from, target, progress)
	return transition.displayed
}

func barBaselineData(target chartData) chartData {
	baseline := cloneBarData(target)
	for seriesIndex := range baseline.series {
		for valueIndex := range baseline.series[seriesIndex].values {
			bar := &baseline.series[seriesIndex].values[valueIndex]
			if bar.valid {
				bar.value = 0
				bar.start = 0
				bar.end = 0
			}
		}
	}
	return baseline
}

func barTransitionFrom(previous, target chartData) chartData {
	from := cloneBarData(target)
	for seriesIndex := range from.series {
		for valueIndex := range from.series[seriesIndex].values {
			bar := &from.series[seriesIndex].values[valueIndex]
			if bar.valid {
				bar.value = 0
				bar.end = bar.start
			}
		}
	}
	previousSeries := make(map[string]resolvedSeries, len(previous.series))
	for _, series := range previous.series {
		previousSeries[series.key] = series
	}
	for seriesIndex := range from.series {
		old, ok := previousSeries[from.series[seriesIndex].key]
		if !ok {
			continue
		}
		for valueIndex := range from.series[seriesIndex].values {
			if valueIndex >= len(old.values) || !from.series[seriesIndex].values[valueIndex].valid || !old.values[valueIndex].valid {
				continue
			}
			from.series[seriesIndex].values[valueIndex].value = old.values[valueIndex].value
			from.series[seriesIndex].values[valueIndex].start = old.values[valueIndex].start
			from.series[seriesIndex].values[valueIndex].end = old.values[valueIndex].end
		}
	}
	return from
}

func interpolateBarData(from, target chartData, progress float32) chartData {
	result := cloneBarData(target)
	for seriesIndex := range result.series {
		if seriesIndex >= len(from.series) || from.series[seriesIndex].key != result.series[seriesIndex].key {
			continue
		}
		for valueIndex := range result.series[seriesIndex].values {
			if valueIndex >= len(from.series[seriesIndex].values) {
				continue
			}
			start := from.series[seriesIndex].values[valueIndex]
			end := result.series[seriesIndex].values[valueIndex]
			if !start.valid || !end.valid {
				continue
			}
			bar := &result.series[seriesIndex].values[valueIndex]
			bar.value = animation.LerpFloat64(start.value, end.value, progress)
			bar.start = animation.LerpFloat64(start.start, end.start, progress)
			bar.end = animation.LerpFloat64(start.end, end.end, progress)
		}
	}
	return result
}

func cloneBarData(source chartData) chartData {
	result := source
	result.series = append([]resolvedSeries(nil), source.series...)
	for index := range result.series {
		result.series[index].values = append([]resolvedBar(nil), source.series[index].values...)
	}
	return result
}

func sameBarGeometry(first, second chartData) bool {
	if len(first.series) != len(second.series) || first.categories != second.categories {
		return false
	}
	for seriesIndex := range first.series {
		left, right := first.series[seriesIndex], second.series[seriesIndex]
		if left.key != right.key || len(left.values) != len(right.values) {
			return false
		}
		for valueIndex := range left.values {
			leftBar, rightBar := left.values[valueIndex], right.values[valueIndex]
			if leftBar.valid != rightBar.valid {
				return false
			}
			if leftBar.valid && (leftBar.value != rightBar.value || leftBar.start != rightBar.start || leftBar.end != rightBar.end) {
				return false
			}
		}
	}
	return true
}
