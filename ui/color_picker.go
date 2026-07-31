package ui

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/components/colorpicker"
)

type ColorPickerWidget = colorpicker.ColorPickerWidget

type ColorWheelWidget = colorpicker.ColorWheelWidget

type ColorAreaWidget = colorpicker.ColorAreaWidget

type ColorFieldWidget = colorpicker.ColorFieldWidget

type ColorSliderWidget = colorpicker.ColorSliderWidget

type ColorSwatchWidget = colorpicker.ColorSwatchWidget

type ColorSwatchPickerWidget = colorpicker.ColorSwatchPickerWidget

// ColorChannel identifies the channel controlled by a ColorSlider.
type ColorChannel = colorpicker.ColorChannel

// ColorSwatchSize controls the size of a color swatch.
type ColorSwatchSize = colorpicker.ColorSwatchSize

// ColorSwatchShape controls the shape of a color swatch.
type ColorSwatchShape = colorpicker.ColorSwatchShape

// ColorSwatchPickerLayout controls how swatches are arranged.
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

// ColorPicker creates a complete color picker initialized with value.
func ColorPicker(key string, value color.NRGBA) ColorPickerWidget {
	return colorpicker.ColorPicker(key, value)
}

// ColorWheel creates a controlled hue and saturation picker. Brightness and
// alpha are preserved from value when the wheel changes.
func ColorWheel(key string, value color.NRGBA) ColorWheelWidget {
	return colorpicker.ColorWheel(key, value)
}

// ColorArea creates a two-dimensional saturation and brightness picker.
func ColorArea(key string, value color.NRGBA) ColorAreaWidget {
	return colorpicker.ColorArea(key, value)
}

// ColorField creates a text field for editing a color value.
func ColorField(key string, value color.NRGBA) ColorFieldWidget {
	return colorpicker.ColorField(key, value)
}

// ColorSlider creates a slider for one color channel.
func ColorSlider(key string, value color.NRGBA, channel ColorChannel) ColorSliderWidget {
	return colorpicker.ColorSlider(key, value, channel)
}

// ColorSwatch creates a single color swatch.
func ColorSwatch(value color.NRGBA) ColorSwatchWidget {
	return colorpicker.ColorSwatch(value)
}

// ColorSwatchPicker creates a selectable palette of color swatches.
func ColorSwatchPicker(key string, value color.NRGBA, colors []color.NRGBA) ColorSwatchPickerWidget {
	return colorpicker.ColorSwatchPicker(key, value, colors)
}
