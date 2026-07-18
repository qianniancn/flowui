package colorpicker

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotColorSlider = "color-slider"

type ColorSliderWidget struct {
	key        string
	value      color.NRGBA
	channel    ColorChannel
	label      string
	onChange   func(color.NRGBA)
	disabled   bool
	showLabel  bool
	showOutput bool
	color      *colorValueState
}

type colorSliderState struct {
	control colorControlState
	color   colorValueState
}

func ColorSlider(key string, value color.NRGBA, channel ColorChannel) ColorSliderWidget {
	return ColorSliderWidget{
		key:        key,
		value:      value,
		channel:    channel,
		showLabel:  true,
		showOutput: true,
	}
}

func (slider ColorSliderWidget) Label(label string) ColorSliderWidget {
	slider.label = label
	slider.showLabel = true
	return slider
}

func (slider ColorSliderWidget) HideLabel() ColorSliderWidget {
	slider.showLabel = false
	return slider
}

func (slider ColorSliderWidget) ShowOutput(show bool) ColorSliderWidget {
	slider.showOutput = show
	return slider
}

func (slider ColorSliderWidget) OnChange(fn func(color.NRGBA)) ColorSliderWidget {
	slider.onChange = fn
	return slider
}

func (slider ColorSliderWidget) Disabled(disabled bool) ColorSliderWidget {
	slider.disabled = disabled
	return slider
}

func (slider ColorSliderWidget) withColorState(state *colorValueState) ColorSliderWidget {
	slider.color = state
	return slider
}

func (slider ColorSliderWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindColorSlider, slider.key)
	sliderState := frame.UseState[colorSliderState](ctx, key, stateSlotColorSlider)
	valueState := &sliderState.color
	if slider.color != nil {
		valueState = slider.color
	} else {
		valueState.sync(slider.value)
	}
	enabled := gtx.Enabled() && !slider.disabled

	var children [2]layout.FlexChild
	count := 0
	if slider.showLabel || slider.showOutput {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return slider.layoutHeader(ctx, gtx, valueState)
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return slider.layoutTrack(ctx, gtx, sliderState, valueState, enabled)
	})
	count++

	opacity := paint.PushOpacity(gtx.Ops, func() float32 {
		if enabled {
			return 1
		}
		return frame.ActiveTheme(ctx).DisabledOpacityValue()
	}())
	dimensions := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.ColorSlider.HeaderGap),
	}.Layout(gtx, children[:count]...)
	opacity.Pop()
	return dimensions
}

func (slider ColorSliderWidget) layoutHeader(ctx *frame.Context, gtx layout.Context, valueState *colorValueState) layout.Dimensions {
	textSize := float32(frame.ActiveTheme(ctx).Components.ColorSlider.TextSize)
	var children [2]layout.FlexChild
	count := 0
	if slider.showLabel {
		children[count] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return text.New(slider.resolvedLabel(ctx)).
				Size(textSize).
				Weight(font.Medium).
				Color(frame.ActiveTheme(ctx).Palette.OverlayForegroundColor()).
				Layout(ctx, gtx)
		})
		count++
	}
	if slider.showOutput {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(slider.output(valueState)).
				Size(textSize).
				Weight(font.Medium).
				Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
				Layout(ctx, gtx)
		})
		count++
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children[:count]...)
}

func (slider ColorSliderWidget) layoutTrack(ctx *frame.Context, gtx layout.Context, sliderState *colorSliderState, valueState *colorValueState, enabled bool) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.ColorSlider
	size := image.Pt(gtx.Constraints.Max.X, max(gtx.Dp(tokens.TrackHeight), 1))
	trackGtx := gtx
	trackGtx.Constraints = layout.Exact(size)
	focus := sliderState.control.focusOpacity(ctx, gtx)
	thumbSize := max(gtx.Dp(tokens.ThumbSize), 1)
	thumbBorder := max(gtx.Dp(tokens.ThumbBorderWidth), 1)

	switch slider.channel {
	case ColorChannelAlpha:
		currentColor := valueState.syncedColor
		current := float64(currentColor.A) / 255
		if next, changed := sliderState.control.updateAxis(ctx, trackGtx, size, current, .01, enabled); changed {
			value := currentColor
			value.A = uint8(next*255 + .5)
			valueState.accept(value, valueState.hsv().h)
			slider.dispatch(value)
			current = next
		}
		value := currentColor
		value.A = uint8(current*255 + .5)
		drawAlphaSlider(gtx, size, value, thumbSize, thumbBorder, focus, frame.ActiveTheme(ctx).Palette.Focus)
	default:
		current := valueState.hsv()
		if next, changed := sliderState.control.updateAxis(ctx, trackGtx, size, current.h, 1.0/360, enabled); changed {
			current.h = next
			value := hsvToNRGBA(current)
			valueState.accept(value, next)
			slider.dispatch(value)
		}
		drawHueSlider(gtx, size, current, thumbSize, thumbBorder, focus, frame.ActiveTheme(ctx).Palette.Focus)
	}
	addColorControlInput(gtx, &sliderState.control, size, enabled, false, slider.resolvedLabel(ctx), slider.output(valueState))
	return layout.Dimensions{Size: size}
}

func (slider ColorSliderWidget) dispatch(value color.NRGBA) {
	if slider.onChange != nil && value != slider.value {
		slider.onChange(value)
	}
}

func (slider ColorSliderWidget) resolvedLabel(ctx *frame.Context) string {
	if slider.label != "" {
		return slider.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		if slider.channel == ColorChannelAlpha {
			return "透明度"
		}
		return "色相"
	}
	if slider.channel == ColorChannelAlpha {
		return "Alpha"
	}
	return "Hue"
}

func (slider ColorSliderWidget) output(valueState *colorValueState) string {
	var output [8]byte
	if slider.channel == ColorChannelAlpha {
		value := int64(math.Round(float64(valueState.syncedColor.A) / 255 * 100))
		formatted := strconv.AppendInt(output[:0], value, 10)
		return string(append(formatted, '%'))
	}
	value := int64(math.Round(valueState.hsv().h*360)) % 360
	formatted := strconv.AppendInt(output[:0], value, 10)
	return string(append(formatted, "°"...))
}
