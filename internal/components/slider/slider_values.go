package slider

import (
	"math"
	"strconv"
)

type sliderValues struct {
	minValue  float64
	maxValue  float64
	step      float64
	lower     float64
	upper     float64
	rangeMode bool
}

func (s SliderWidget) resolvedValues() sliderValues {
	values := sliderValues{
		minValue:  s.minValue,
		maxValue:  s.maxValue,
		step:      s.step,
		lower:     s.value,
		upper:     s.upperValue,
		rangeMode: s.rangeMode,
	}
	if !sliderFinite(values.minValue) || !sliderFinite(values.maxValue) || values.maxValue <= values.minValue {
		values.minValue = 0
		values.maxValue = 100
	}
	if !sliderFinite(values.step) || values.step <= 0 {
		values.step = 1
	}
	values.lower = values.snap(values.lower)
	if values.rangeMode {
		values.upper = values.snap(values.upper)
		if values.lower > values.upper {
			values.lower, values.upper = values.upper, values.lower
		}
	} else {
		values.upper = values.lower
	}
	return values
}

func (v sliderValues) snap(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return v.minValue
	}
	value = min(max(value, v.minValue), v.maxValue)
	steps := math.Round((value - v.minValue) / v.step)
	value = v.minValue + steps*v.step
	value = normalizeSliderFloat(value)
	return min(max(value, v.minValue), v.maxValue)
}

func sliderFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func normalizeSliderFloat(value float64) float64 {
	result, err := strconv.ParseFloat(strconv.FormatFloat(value, 'g', 12, 64), 64)
	if err != nil {
		return value
	}
	return result
}

func (v sliderValues) ratio(value float64) float32 {
	return float32((v.snap(value) - v.minValue) / (v.maxValue - v.minValue))
}

func (v sliderValues) value(ratio float32) float64 {
	ratio = min(max(ratio, 0), 1)
	return v.snap(v.minValue + float64(ratio)*(v.maxValue-v.minValue))
}
