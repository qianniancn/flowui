package slider

import (
	"fmt"
	"math"
	"strconv"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type SliderOrientation uint8

const (
	SliderHorizontal SliderOrientation = iota
	SliderVertical
)

type SliderWidget struct {
	key             string
	value           float64
	hasValue        bool
	defaultValue    float64
	hasDefault      bool
	upperValue      float64
	hasUpperValue   bool
	defaultUpper    float64
	hasDefaultUpper bool
	minValue        float64
	maxValue        float64
	step            float64
	label           string
	valueText       string
	hasValueText    bool
	showValue       bool
	rangeMode       bool
	disabled        bool
	orientation     SliderOrientation
	formatValue     func(float64) string
	onChange        func(float64)
	onRangeChange   func(float64, float64)
	customStyle     flowstyle.Style
}

func Slider(key string, value float64) SliderWidget {
	return SliderWidget{
		key:      key,
		value:    value,
		hasValue: true,
		maxValue: 100,
		step:     1,
	}
}

func RangeSlider(key string, lowerValue, upperValue float64) SliderWidget {
	return SliderWidget{
		key:           key,
		value:         lowerValue,
		hasValue:      true,
		upperValue:    upperValue,
		hasUpperValue: true,
		maxValue:      100,
		step:          1,
		rangeMode:     true,
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

func (s SliderWidget) Value(value float64) SliderWidget {
	s.value = value
	s.hasValue = true
	return s
}

func (s SliderWidget) DefaultValue(value float64) SliderWidget {
	s.defaultValue = value
	s.hasDefault = true
	s.hasValue = false
	return s
}

func (s SliderWidget) UpperValue(value float64) SliderWidget {
	s.upperValue = value
	s.hasUpperValue = true
	return s
}

func (s SliderWidget) DefaultUpperValue(value float64) SliderWidget {
	s.defaultUpper = value
	s.hasDefaultUpper = true
	s.hasUpperValue = false
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

func (s SliderWidget) Style(value flowstyle.Style) SliderWidget {
	s.customStyle = value
	return s
}

func (s SliderWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, state := sliderStateFor(ctx, s.key)

	// Bind disclosure state
	state.bindLower(s)
	lowerValue := state.currentLowerValue(s)
	upperValue := s.upperValue
	if s.rangeMode {
		state.bindUpper(s)
		upperValue = state.currentUpperValue(s)
	}

	// Create values with current state
	values := s.resolvedValuesWithCurrent(lowerValue, upperValue)
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
		if changedValues, changed, thumb, keyboard := state.update(gtx, values); changed {
			values = changedValues
			state.syncRatios(values)
			frame.RequestFocusVisible(ctx, state.thumbTag(thumb), keyboard)
			s.dispatchWithState(state, values)
		}
	}

	hovered := state.lowerThumb.clickable.Hovered() || values.rangeMode && state.upperThumb.clickable.Hovered()
	dragging := state.lower.Dragging() || values.rangeMode && state.upper.Dragging()
	pressed := state.lowerThumb.clickable.Pressed() || dragging || values.rangeMode && state.upperThumb.clickable.Pressed()
	lowerFocused := gtx.Focused(&state.lowerThumb.clickable)
	upperFocused := values.rangeMode && gtx.Focused(&state.upperThumb.clickable)
	focused := lowerFocused || upperFocused
	focusVisible := frame.FocusVisible(ctx, &state.lowerThumb.clickable, lowerFocused)
	if values.rangeMode {
		focusVisible = focusVisible || frame.FocusVisible(ctx, &state.upperThumb.clickable, upperFocused)
	}
	styleState := flowstyle.StyleState{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     !enabled,
		Dragging:     dragging,
	}
	componentStyle := s.resolveStyle(ctx, gtx, key, styleState)
	layoutSlider := frame.WithFieldSemantics(ctx, key, func(gtx layout.Context) layout.Dimensions {
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		semantic.DescriptionOp(s.semanticDescription(values)).Add(gtx.Ops)
		return s.layout(ctx, gtx, state, componentStyle, values)
	})
	return layoutui.LayoutStyled(ctx, gtx, key, styleState, s.customStyle, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutSlider(gtx)
	}))
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

func (s SliderWidget) dispatchWithState(state *sliderState, values sliderValues) {
	if values.rangeMode {
		state.requestLower(s, values.lower)
		state.requestUpper(s, values.upper)
		if s.onRangeChange != nil {
			s.onRangeChange(values.lower, values.upper)
		}
		return
	}
	state.requestLower(s, values.lower)
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
