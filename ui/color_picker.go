package ui

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/components/colorpicker"
)

type ColorPickerWidget = colorpicker.ColorPickerWidget
type ColorAreaWidget = colorpicker.ColorAreaWidget
type ColorFieldWidget = colorpicker.ColorFieldWidget
type ColorSliderWidget = colorpicker.ColorSliderWidget
type ColorSwatchWidget = colorpicker.ColorSwatchWidget
type ColorSwatchPickerWidget = colorpicker.ColorSwatchPickerWidget
type ColorChannel = colorpicker.ColorChannel
type ColorSwatchSize = colorpicker.ColorSwatchSize
type ColorSwatchShape = colorpicker.ColorSwatchShape
type ColorSwatchPickerLayout = colorpicker.ColorSwatchPickerLayout

const (
	ColorChannelHue   = colorpicker.ColorChannelHue
	ColorChannelAlpha = colorpicker.ColorChannelAlpha

	ColorSwatchExtraSmall = colorpicker.ColorSwatchExtraSmall
	ColorSwatchSmall      = colorpicker.ColorSwatchSmall
	ColorSwatchMedium     = colorpicker.ColorSwatchMedium
	ColorSwatchLarge      = colorpicker.ColorSwatchLarge
	ColorSwatchExtraLarge = colorpicker.ColorSwatchExtraLarge

	ColorSwatchCircle = colorpicker.ColorSwatchCircle
	ColorSwatchSquare = colorpicker.ColorSwatchSquare

	ColorSwatchPickerGrid  = colorpicker.ColorSwatchPickerGrid
	ColorSwatchPickerStack = colorpicker.ColorSwatchPickerStack
)

func ColorPicker(key string, value color.NRGBA) ColorPickerWidget {
	return colorpicker.ColorPicker(key, value)
}

func ColorArea(key string, value color.NRGBA) ColorAreaWidget {
	return colorpicker.ColorArea(key, value)
}

func ColorField(key string, value color.NRGBA) ColorFieldWidget {
	return colorpicker.ColorField(key, value)
}

func ColorSlider(key string, value color.NRGBA, channel ColorChannel) ColorSliderWidget {
	return colorpicker.ColorSlider(key, value, channel)
}

func ColorSwatch(value color.NRGBA) ColorSwatchWidget {
	return colorpicker.ColorSwatch(value)
}

func ColorSwatchPicker(key string, value color.NRGBA, colors []color.NRGBA) ColorSwatchPickerWidget {
	return colorpicker.ColorSwatchPicker(key, value, colors)
}
