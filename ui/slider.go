package ui

import "github.com/qianniancn/flowui/internal/components/slider"

type SliderWidget = slider.SliderWidget
type SliderOrientation = slider.SliderOrientation

const (
	SliderHorizontal = slider.SliderHorizontal
	SliderVertical   = slider.SliderVertical
)

func Slider(key string, value float64) SliderWidget {
	return slider.Slider(key, value)
}

func RangeSlider(key string, lowerValue, upperValue float64) SliderWidget {
	return slider.RangeSlider(key, lowerValue, upperValue)
}
