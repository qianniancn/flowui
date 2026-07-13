package piechart

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type pieChartAnimation struct {
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
		transition.from = pieBaselineData(target)
		transition.target = target
		transition.duration = w.animationDuration
		transition.easing = w.animationEasing
	} else if !samePieGeometry(transition.target, target) {
		transition.revision++
		transition.from = pieTransitionFrom(transition.displayed, target)
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
		Disabled(!w.animation || len(target.slices) == 0).
		Sample(ctx, gtx)
	if !running && progress == 1 {
		transition.from = chartData{}
		transition.displayed = target
		return target
	}
	transition.displayed = interpolatePieData(transition.from, target, progress)
	return transition.displayed
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
		if left.key != right.key || left.value != right.value || left.startAngle != right.startAngle || left.endAngle != right.endAngle {
			return false
		}
	}
	return true
}
