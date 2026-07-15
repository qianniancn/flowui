package candlestick

import "github.com/qianniancn/FlowUI/internal/components/chart"

type linearScale struct {
	minimum  float64
	maximum  float64
	interval float64
	ticks    []float64
}

func newLinearScale(minimum, maximum float64, targetTicks int, fixed bool) linearScale {
	return linearScaleFrom(chart.NewLinearScale(minimum, maximum, targetTicks, false, fixed))
}

func linearScaleFrom(scale chart.LinearScale) linearScale {
	return linearScale{minimum: scale.Minimum, maximum: scale.Maximum, interval: scale.Interval, ticks: scale.Ticks}
}

func (s linearScale) shared() chart.LinearScale {
	return chart.LinearScale{Minimum: s.minimum, Maximum: s.maximum, Interval: s.interval, Ticks: s.ticks}
}

func (s linearScale) ratio(value float64) float64 {
	return s.shared().Ratio(value)
}

func (s linearScale) valueAt(ratio float64) float64 {
	return s.shared().ValueAt(ratio)
}

func (s linearScale) contains(value float64) bool {
	return s.shared().Contains(value)
}

func formatAxisNumber(value, interval float64) string {
	return chart.FormatAxisNumber(value, interval)
}
