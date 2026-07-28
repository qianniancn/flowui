package barchart

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/frame"
)

func (w Widget) animatedData(ctx *frame.Context, gtx layout.Context, state *chartState, target chartData) chartData {
	return state.animation.Update(ctx, gtx, target, w.animation && target.yExtent.Valid,
		w.animationDuration, w.updateAnimationDuration, w.animationEasing, w.updateAnimationEasing,
		sameBarTarget, barBaselineData, barTransitionFrom, interpolateBarData)
}

func sameBarTarget(previous, target chartData) bool {
	if previous.generation != 0 || target.generation != 0 {
		return previous.generation == target.generation
	}
	return sameBarGeometry(previous, target)
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
