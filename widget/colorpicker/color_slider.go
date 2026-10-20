package colorpicker

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

const stateSlotColorSlider = "color-slider"

type ColorSliderWidget struct {
	override   *theme.ColorPickerOverride
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
	control     colorControlState
	color       colorValueState
	visualCache colorPickerVisualCache
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

func (slider ColorSliderWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return slider.layoutWithOverride(ctx, gtx, slider.override)
}

func (slider ColorSliderWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ColorPickerOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorSlider, slider.key)
	sliderState := core.UseState[colorSliderState](ctx, key, stateSlotColorSlider)
	visual := sliderState.visualCache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
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
			return slider.layoutHeader(ctx, gtx, valueState, visual)
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return slider.layoutTrack(ctx, gtx, sliderState, valueState, visual, enabled)
	})
	count++

	focused := gtx.Focused(&sliderState.control)
	return layoutui.Styled(ctx, gtx, key, style.State{
		Pressed:      sliderState.control.dragging,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &sliderState.control, focused),
		Disabled:     !enabled,
		Dragging:     sliderState.control.dragging,
	}, visual.Override, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, func() float32 {
			if enabled {
				return 1
			}
			return visual.DisabledOpacity
		}())
		defer opacity.Pop()
		dimensions := layout.Flex{
			Axis: layout.Vertical,
			Gap:  gtx.Dp(visual.Slider.HeaderGap),
		}.Layout(gtx, children[:count]...)
		return dimensions
	}))
}

func (slider ColorSliderWidget) layoutHeader(ctx *core.Context, gtx layout.Context, valueState *colorValueState, visual theme.ColorPickerVisual) layout.Dimensions {
	textSize := float32(visual.Slider.TextSize)
	var children [2]layout.FlexChild
	count := 0
	if slider.showLabel {
		children[count] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return text.New(slider.resolvedLabel(ctx)).
				Size(textSize).
				Weight(font.Medium).
				Color(visual.PanelForeground).
				Layout(ctx, gtx)
		})
		count++
	}
	if slider.showOutput {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(slider.output(valueState)).
				Size(textSize).
				Weight(font.Medium).
				Color(visual.MutedForeground).
				Layout(ctx, gtx)
		})
		count++
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children[:count]...)
}

func (slider ColorSliderWidget) layoutTrack(ctx *core.Context, gtx layout.Context, sliderState *colorSliderState, valueState *colorValueState, visual theme.ColorPickerVisual, enabled bool) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, max(gtx.Dp(visual.Slider.TrackHeight), 1))
	trackGtx := gtx
	trackGtx.Constraints = layout.Exact(size)
	focus := sliderState.control.focusOpacity(ctx, gtx, visual.Motion)
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
		visual.Slider.Render(gtx, theme.ColorSliderRenderRequest{
			Size: size, Channel: theme.ColorSliderAlpha, Value: colorPickerHSV(valueState.hsv()), Color: value, Focus: focus,
		})
	default:
		current := valueState.hsv()
		if next, changed := sliderState.control.updateAxis(ctx, trackGtx, size, current.h, 1.0/360, enabled); changed {
			current.h = next
			value := hsvToNRGBA(current)
			valueState.accept(value, next)
			slider.dispatch(value)
		}
		visual.Slider.Render(gtx, theme.ColorSliderRenderRequest{
			Size: size, Channel: theme.ColorSliderHue, Value: colorPickerHSV(current), Color: hsvToNRGBA(current), Focus: focus,
		})
	}
	addColorControlInput(gtx, &sliderState.control, size, enabled, false, slider.resolvedLabel(ctx), slider.output(valueState))
	return layout.Dimensions{Size: size}
}

func (slider ColorSliderWidget) dispatch(value color.NRGBA) {
	if slider.onChange != nil && value != slider.value {
		slider.onChange(value)
	}
}

func (slider ColorSliderWidget) resolvedLabel(ctx *core.Context) string {
	if slider.label != "" {
		return slider.label
	}
	if slider.channel == ColorChannelAlpha {
		return ctx.Text(i18n.ColorPickerAlpha)
	}
	return ctx.Text(i18n.ColorPickerHue)
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

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ColorSliderWidget) Style(values ...style.Style) ColorSliderWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ColorPickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
