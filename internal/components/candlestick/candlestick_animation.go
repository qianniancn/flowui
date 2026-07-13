package candlestick

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type candlestickAnimation struct {
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
		transition.from = candleBaselineData(target)
		transition.target = target
		transition.duration = w.animationDuration
		transition.easing = w.animationEasing
	} else if !sameCandleData(transition.target, target) {
		transition.revision++
		transition.from = candleTransitionFrom(transition.displayed, target)
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
		Disabled(!w.animation || !target.extent.valid).
		Sample(ctx, gtx)
	if !running && progress == 1 {
		transition.from = chartData{}
		transition.displayed = target
		return target
	}
	transition.displayed = interpolateCandleData(transition.from, target, progress)
	return transition.displayed
}

func candleBaselineData(target chartData) chartData {
	result := cloneCandleData(target)
	for index := range result.candles {
		if result.candles[index].valid {
			open := result.candles[index].open
			result.candles[index].close = open
			result.candles[index].low = open
			result.candles[index].high = open
		}
	}
	return result
}

func candleTransitionFrom(previous, target chartData) chartData {
	result := candleBaselineData(target)
	for index := range result.candles {
		if index < len(previous.candles) && result.candles[index].valid && previous.candles[index].valid {
			old := previous.candles[index]
			result.candles[index].open = old.open
			result.candles[index].close = old.close
			result.candles[index].low = old.low
			result.candles[index].high = old.high
		}
	}
	return result
}

func interpolateCandleData(from, target chartData, progress float32) chartData {
	result := cloneCandleData(target)
	for index := range result.candles {
		if index >= len(from.candles) || !result.candles[index].valid || !from.candles[index].valid {
			continue
		}
		start := from.candles[index]
		candle := &result.candles[index]
		candle.open = animation.LerpFloat64(start.open, candle.open, progress)
		candle.close = animation.LerpFloat64(start.close, candle.close, progress)
		candle.low = animation.LerpFloat64(start.low, candle.low, progress)
		candle.high = animation.LerpFloat64(start.high, candle.high, progress)
	}
	return result
}

func cloneCandleData(source chartData) chartData {
	result := source
	result.candles = append([]resolvedCandle(nil), source.candles...)
	return result
}

func sameCandleData(first, second chartData) bool {
	if len(first.candles) != len(second.candles) {
		return false
	}
	for index := range first.candles {
		left, right := first.candles[index], second.candles[index]
		if left.valid != right.valid || left.open != right.open || left.close != right.close || left.low != right.low || left.high != right.high {
			return false
		}
	}
	return true
}
