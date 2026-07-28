package piechart

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/frame"
)

func (w Widget) animatedData(ctx *frame.Context, gtx layout.Context, state *chartState, target chartData) chartData {
	return state.animation.Update(ctx, gtx, target, w.animation && len(target.slices) > 0,
		w.animationDuration, w.updateAnimationDuration, w.animationEasing, w.updateAnimationEasing,
		samePieTarget, pieBaselineData, pieTransitionFrom, interpolatePieData)
}

func samePieTarget(previous, target chartData) bool {
	if previous.generation != 0 || target.generation != 0 {
		return previous.generation == target.generation
	}
	return samePieGeometry(previous, target)
}

func pieBaselineData(target chartData) chartData {
	result := clonePieData(target)
	for index := range result.slices {
		result.slices[index].startAngle = target.start
		result.slices[index].endAngle = target.start
	}
	return result
}

func pieTransitionFrom(previous, target chartData) chartData {
	result := pieBaselineData(target)
	old := make(map[string]resolvedSlice, len(previous.slices))
	for _, slice := range previous.slices {
		old[slice.key] = slice
	}
	for index := range result.slices {
		if slice, ok := old[result.slices[index].key]; ok {
			result.slices[index].startAngle = slice.startAngle
			result.slices[index].endAngle = slice.endAngle
		}
	}
	return result
}

func interpolatePieData(from, target chartData, progress float32) chartData {
	result := clonePieData(target)
	for index := range result.slices {
		if index >= len(from.slices) || result.slices[index].key != from.slices[index].key {
			continue
		}
		result.slices[index].startAngle = animation.LerpFloat(from.slices[index].startAngle, result.slices[index].startAngle, progress)
		result.slices[index].endAngle = animation.LerpFloat(from.slices[index].endAngle, result.slices[index].endAngle, progress)
		result.slices[index].radiusRatio = animation.LerpFloat(from.slices[index].radiusRatio, result.slices[index].radiusRatio, progress)
	}
	return result
}

func clonePieData(source chartData) chartData {
	result := source
	result.slices = append([]resolvedSlice(nil), source.slices...)
	result.legend = append([]resolvedSlice(nil), source.legend...)
	return result
}

func samePieGeometry(first, second chartData) bool {
	if len(first.slices) != len(second.slices) || first.start != second.start || first.dir != second.dir {
		return false
	}
	for index := range first.slices {
		left, right := first.slices[index], second.slices[index]
		if left.key != right.key || left.value != right.value || left.startAngle != right.startAngle || left.endAngle != right.endAngle || left.radiusRatio != right.radiusRatio {
			return false
		}
	}
	return true
}
