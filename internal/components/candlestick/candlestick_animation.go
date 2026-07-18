package candlestick

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (w Widget) animatedData(ctx *frame.Context, gtx layout.Context, state *chartState, target chartData) chartData {
	return state.animation.Update(ctx, gtx, target, w.animation && target.extent.Valid,
		w.animationDuration, w.updateAnimationDuration, w.animationEasing, w.updateAnimationEasing,
		sameCandleTarget, candleBaselineData, candleTransitionFrom, interpolateCandleData)
}

func sameCandleTarget(previous, target chartData) bool {
	if previous.generation != 0 || target.generation != 0 {
		return previous.generation == target.generation
	}
	return sameCandleData(previous, target)
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
