package ui

import "github.com/qianniancn/flowui/internal/components/slider"

type SliderWidget = slider.SliderWidget

// SliderOrientation controls the direction of a slider.
type SliderOrientation = slider.SliderOrientation

const (
	SliderHorizontal = slider.SliderHorizontal
	SliderVertical   = slider.SliderVertical
)

// Slider creates a single-value slider initialized with value.
func Slider(key string, value float64) SliderWidget {
	return slider.Slider(key, value)
}

// RangeSlider creates a two-handle slider initialized with lowerValue and upperValue.
func RangeSlider(key string, lowerValue, upperValue float64) SliderWidget {
	return slider.RangeSlider(key, lowerValue, upperValue)
}
