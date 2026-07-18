package candlestick

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	signDown = -1
	signDoji = 0
	signUp   = 1
)

type dataExtent = chart.Extent

type resolvedCandle struct {
	index int
	open  float64
	close float64
	low   float64
	high  float64
	sign  int
	color color.NRGBA
	valid bool
}

type chartData struct {
	candles    []resolvedCandle
	extent     dataExtent
	generation uint64
}

type chartDataCache struct {
	cache chart.DataCache[chartData]
}

func (c *chartDataCache) resolve(widget Widget, activeTheme *theme.Theme) chartData {
	data, generation := c.cache.Resolve(widget.hasDataVersion, widget.dataVersion, activeTheme, unit.Metric{}, func() chartData {
		return resolveChartData(widget, activeTheme)
	})
	data.generation = generation
	return data
}

func resolveChartData(widget Widget, activeTheme *theme.Theme) chartData {
	tokens := activeTheme.Components.CandlestickChart
	upColor := tokens.UpColor
	if widget.hasUpColor {
		upColor = widget.upColor
	}
	downColor := tokens.DownColor
	if widget.hasDownColor {
		downColor = widget.downColor
	}
	dojiColor := tokens.DojiColor
	hasDojiColor := dojiColor.A != 0
	if widget.hasDojiColor {
		dojiColor, hasDojiColor = widget.dojiColor, true
	}

	result := chartData{candles: make([]resolvedCandle, len(widget.data))}
	for index, source := range widget.data {
		candle := resolvedCandle{index: index, open: source.open, close: source.close, low: source.low, high: source.high}
		candle.valid = finite(candle.open) && finite(candle.close) && finite(candle.low) && finite(candle.high)
		if !candle.valid {
			result.candles[index] = candle
			continue
		}
		candle.sign = candleSign(widget.data, index, hasDojiColor)
		switch candle.sign {
		case signDown:
			candle.color = downColor
		case signDoji:
			candle.color = dojiColor
		default:
			candle.color = upColor
		}
		result.extent.Include(candle.open)
		result.extent.Include(candle.close)
		result.extent.Include(candle.low)
		result.extent.Include(candle.high)
		result.candles[index] = candle
	}
	return result
}

func candleSign(values []Candle, index int, hasDojiColor bool) int {
	value := values[index]
	if value.open > value.close {
		return signDown
	}
	if value.open < value.close {
		return signUp
	}
	if hasDojiColor {
		return signDoji
	}
	if index == 0 || values[index-1].close <= value.close {
		return signUp
	}
	return signDown
}
