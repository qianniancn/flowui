package slider

import (
	"fmt"
	"math"
	"strconv"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type SliderOrientation uint8

const (
	SliderHorizontal SliderOrientation = iota
	SliderVertical
)

type SliderWidget struct {
	key           string
	value         float64
	upperValue    float64
	minValue      float64
	maxValue      float64
	step          float64
	label         string
	valueText     string
	hasValueText  bool
	showValue     bool
	rangeMode     bool
	disabled      bool
	orientation   SliderOrientation
	formatValue   func(float64) string
	onChange      func(float64)
	onRangeChange func(float64, float64)
}

func Slider(key string, value float64) SliderWidget {
	return SliderWidget{
		key:      key,
		value:    value,
		maxValue: 100,
		step:     1,
	}
}

func RangeSlider(key string, lowerValue, upperValue float64) SliderWidget {
	return SliderWidget{
		key:        key,
		value:      lowerValue,
		upperValue: upperValue,
		maxValue:   100,
		step:       1,
		rangeMode:  true,
	}
}

func (s SliderWidget) Range(minValue, maxValue float64) SliderWidget {
	if math.IsNaN(minValue) || math.IsInf(minValue, 0) || math.IsNaN(maxValue) || math.IsInf(maxValue, 0) || maxValue <= minValue {
		panic("flowui: slider maximum must be greater than minimum")
	}
	s.minValue = minValue
	s.maxValue = maxValue
	return s
}

func (s SliderWidget) Step(step float64) SliderWidget {
	if math.IsNaN(step) || math.IsInf(step, 0) || step <= 0 {
		panic("flowui: slider step must be positive")
	}
	s.step = step
	return s
}

func (s SliderWidget) Label(label string) SliderWidget {
	s.label = label
	return s
}

func (s SliderWidget) ShowValue() SliderWidget {
	s.showValue = true
	return s
}

func (s SliderWidget) ValueText(text string) SliderWidget {
	s.valueText = text
	s.hasValueText = true
	s.showValue = true
	return s
}

func (s SliderWidget) FormatValue(format func(float64) string) SliderWidget {
	s.formatValue = format
	return s
}

func (s SliderWidget) OnChange(fn func(float64)) SliderWidget {
	s.onChange = fn
	return s
}

func (s SliderWidget) OnRangeChange(fn func(float64, float64)) SliderWidget {
	s.onRangeChange = fn
	return s
}

func (s SliderWidget) Orientation(orientation SliderOrientation) SliderWidget {
	s.orientation = orientation
	return s
}

func (s SliderWidget) Vertical() SliderWidget {
	s.orientation = SliderVertical
	return s
}

func (s SliderWidget) Disabled(disabled bool) SliderWidget {
	s.disabled = disabled
	return s
}

func (s SliderWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, state := sliderStateFor(ctx, s.key)
	values := s.resolvedValues()
	axis := s.axis()
	state.setAxis(axis)
	state.sync(values)

	enabled := gtx.Enabled() && !s.disabled
	frame.RegisterFieldFocus(ctx, key, &state.lowerThumb.clickable, enabled)
	if !enabled {
		gtx = gtx.Disabled()
	}

	if enabled {
		state.updateThumbPresses(ctx, gtx, values.rangeMode)
		if changedValues, changed, thumb := state.update(gtx, values); changed {
			values = changedValues
			state.syncRatios(values)
			frame.RequestFocus(ctx, state.thumbTag(thumb))
			s.dispatch(values)
		}
	}

	style := sliderStyleFor(frame.ActiveTheme(ctx), !enabled)
	layoutSlider := frame.WithFieldSemantics(ctx, key, func(gtx layout.Context) layout.Dimensions {
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		semantic.DescriptionOp(s.semanticDescription(values)).Add(gtx.Ops)
		return s.layout(ctx, gtx, state, style, values)
	})
	return layoutSlider(gtx)
}

func (s SliderWidget) axis() layout.Axis {
	if s.orientation == SliderVertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

func (s SliderWidget) dispatch(values sliderValues) {
	if values.rangeMode {
		if s.onRangeChange != nil && (values.lower != s.resolvedValues().lower || values.upper != s.resolvedValues().upper) {
			s.onRangeChange(values.lower, values.upper)
		}
		return
	}
	if s.onChange != nil && values.lower != s.resolvedValues().lower {
		s.onChange(values.lower)
	}
}

func (s SliderWidget) outputText(values sliderValues) string {
	if s.hasValueText {
		return s.valueText
	}
	if !s.showValue {
		return ""
	}
	if values.rangeMode {
		return s.format(values.lower) + " - " + s.format(values.upper)
	}
	return s.format(values.lower)
}

func (s SliderWidget) format(value float64) string {
	if s.formatValue != nil {
		return s.formatValue(value)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (s SliderWidget) semanticDescription(values sliderValues) string {
	label := s.label
	if label == "" {
		label = "Slider"
	}
	output := s.outputText(values)
	if output == "" {
		if values.rangeMode {
			output = s.format(values.lower) + " - " + s.format(values.upper)
		} else {
			output = s.format(values.lower)
		}
	}
	return fmt.Sprintf("%s %s", label, output)
}
