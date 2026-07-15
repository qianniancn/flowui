package chart

import (
	"math"
	"strconv"
	"strings"
)

type Extent struct {
	Minimum float64
	Maximum float64
	Valid   bool
}

func (e *Extent) Include(value float64) {
	if !finite(value) {
		return
	}
	if !e.Valid {
		e.Minimum, e.Maximum, e.Valid = value, value, true
		return
	}
	e.Minimum = min(e.Minimum, value)
	e.Maximum = max(e.Maximum, value)
}

type LinearScale struct {
	Minimum  float64
	Maximum  float64
	Interval float64
	Ticks    []float64
}

func NewLinearScale(minimum, maximum float64, targetTicks int, includeZero, fixed bool) LinearScale {
	if !finite(minimum) || !finite(maximum) {
		minimum, maximum = 0, 1
	}
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}
	if includeZero {
		minimum = min(minimum, 0)
		maximum = max(maximum, 0)
	}
	if minimum == maximum {
		padding := math.Abs(minimum) * 0.5
		if padding == 0 || !finite(padding) {
			padding = 1
		}
		minimum -= padding
		maximum += padding
		if includeZero {
			minimum = min(minimum, 0)
			maximum = max(maximum, 0)
		}
	}
	targetTicks = max(targetTicks, 2)
	span := maximum - minimum
	if !finite(span) || span <= 0 {
		span = math.Max(math.Abs(minimum), math.Abs(maximum)) / float64(targetTicks) * 2
	} else {
		span /= float64(targetTicks)
	}
	interval := niceNumber(span)
	if !finite(interval) || interval <= 0 {
		interval = 1
	}

	scaleMinimum, scaleMaximum := minimum, maximum
	if !fixed {
		scaleMinimum = roundScale(math.Floor(minimum/interval)*interval, interval)
		scaleMaximum = roundScale(math.Ceil(maximum/interval)*interval, interval)
	}
	if scaleMinimum == scaleMaximum {
		scaleMaximum = scaleMinimum + interval
	}
	return LinearScale{
		Minimum:  scaleMinimum,
		Maximum:  scaleMaximum,
		Interval: interval,
		Ticks:    scaleTicks(scaleMinimum, scaleMaximum, interval, fixed),
	}
}

func (s LinearScale) Ratio(value float64) float64 {
	if s.Maximum == s.Minimum {
		return 0.5
	}
	span := s.Maximum - s.Minimum
	if finite(span) {
		return (value - s.Minimum) / span
	}
	return (value/2 - s.Minimum/2) / (s.Maximum/2 - s.Minimum/2)
}

func (s LinearScale) ValueAt(ratio float64) float64 {
	ratio = min(max(ratio, 0), 1)
	return (1-ratio)*s.Minimum + ratio*s.Maximum
}

func (s LinearScale) Contains(value float64) bool {
	return value >= s.Minimum && value <= s.Maximum
}

func FormatAxisNumber(value, interval float64) string {
	if math.Abs(value) < math.Abs(interval)*1e-10 {
		value = 0
	}
	precision := axisPrecision(interval)
	absolute := math.Abs(value)
	if absolute != 0 && (absolute >= 1e12 || absolute < 1e-6 && precision > 6) {
		return strconv.FormatFloat(value, 'g', 6, 64)
	}
	text := strconv.FormatFloat(value, 'f', precision, 64)
	if dot := strings.IndexByte(text, '.'); dot >= 0 {
		text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
	}
	return addNumberCommas(text)
}

func niceNumber(value float64) float64 {
	if value <= 0 || !finite(value) {
		return 1
	}
	exponent := math.Floor(math.Log10(value))
	power := math.Pow(10, exponent)
	fraction := value / power
	nice := float64(10)
	switch {
	case fraction < 1.5:
		nice = 1
	case fraction < 2.5:
		nice = 2
	case fraction < 4:
		nice = 3
	case fraction < 7:
		nice = 5
	}
	return nice * power
}

func scaleTicks(minimum, maximum, interval float64, fixed bool) []float64 {
	if interval <= 0 || !finite(interval) {
		return nil
	}
	ticks := make([]float64, 0, 12)
	start := minimum
	if fixed {
		start = math.Ceil(minimum/interval) * interval
		if start-minimum > interval*1e-9 {
			ticks = append(ticks, minimum)
		}
	}
	for value, count := start, 0; value <= maximum+interval*1e-9 && count < 1000; count++ {
		ticks = append(ticks, roundScale(value, interval))
		next := roundScale(value+interval, interval)
		if next == value || !finite(next) {
			break
		}
		value = next
	}
	if fixed && (len(ticks) == 0 || math.Abs(ticks[len(ticks)-1]-maximum) > interval*1e-9) {
		ticks = append(ticks, maximum)
	}
	return ticks
}

func roundScale(value, interval float64) float64 {
	precision := axisPrecision(interval) + 2
	if precision > 15 {
		return math.Round(value/interval) * interval
	}
	if precision <= 0 {
		return math.Round(value)
	}
	power := math.Pow10(precision)
	if math.Abs(value) > math.MaxFloat64/power {
		return math.Round(value/interval) * interval
	}
	return math.Round(value*power) / power
}

func axisPrecision(interval float64) int {
	if interval <= 0 || !finite(interval) {
		return 0
	}
	exponent := math.Floor(math.Log10(math.Abs(interval)))
	precision := 0
	if exponent < 0 {
		precision = int(-exponent)
	}
	if precision > 15 {
		return precision
	}
	scaled := interval * math.Pow10(precision)
	for precision < 10 && math.Abs(scaled-math.Round(scaled)) > 1e-8 {
		precision++
		scaled *= 10
	}
	return precision
}

func addNumberCommas(value string) string {
	sign := ""
	if strings.HasPrefix(value, "-") {
		sign, value = "-", value[1:]
	}
	integer, fraction := value, ""
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		integer, fraction = value[:dot], value[dot:]
	}
	for index := len(integer) - 3; index > 0; index -= 3 {
		integer = integer[:index] + "," + integer[index:]
	}
	return sign + integer + fraction
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
