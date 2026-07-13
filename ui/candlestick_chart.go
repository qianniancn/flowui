package ui

import "github.com/qianniancn/FlowUI/internal/components/candlestick"

type CandlestickChartWidget = candlestick.Widget
type CandlestickChartData = candlestick.Candle

func CandlestickChart(key string, data []CandlestickChartData) CandlestickChartWidget {
	return candlestick.New(key, data)
}

// Candle creates one OHLC value in ECharts order: open, close, lowest, highest.
func Candle(open, close, lowest, highest float64) CandlestickChartData {
	return candlestick.OHLC(open, close, lowest, highest)
}
