package slider

import (
	"math"
	"strconv"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Orientation = theme.SliderOrientation

const (
	Horizontal = theme.SliderHorizontal
	Vertical   = theme.SliderVertical
)

type Widget struct {
	override        *theme.SliderOverride
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
	orientation     Orientation
	formatValue     func(float64) string
	onChange        func(float64)
	onRangeChange   func(float64, float64)
}

func New(key string, value float64) Widget {
	return Widget{
		key:      key,
		value:    value,
		hasValue: true,
		maxValue: 100,
		step:     1,
	}
}

func RangeSlider(key string, lowerValue, upperValue float64) Widget {
	return Widget{
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

func (s Widget) Range(minValue, maxValue float64) Widget {
	if math.IsNaN(minValue) || math.IsInf(minValue, 0) || math.IsNaN(maxValue) || math.IsInf(maxValue, 0) || maxValue <= minValue {
		panic("flowui: slider maximum must be greater than minimum")
	}
	s.minValue = minValue
	s.maxValue = maxValue
	return s
}

func (s Widget) Step(step float64) Widget {
	if math.IsNaN(step) || math.IsInf(step, 0) || step <= 0 {
		panic("flowui: slider step must be positive")
	}
	s.step = step
	return s
}

func (s Widget) Label(label string) Widget {
	s.label = label
	return s
}

func (s Widget) ShowValue() Widget {
	s.showValue = true
	return s
}

func (s Widget) ValueText(text string) Widget {
	s.valueText = text
	s.hasValueText = true
	s.showValue = true
	return s
}

func (s Widget) FormatValue(format func(float64) string) Widget {
	s.formatValue = format
	return s
}

func (s Widget) OnChange(fn func(float64)) Widget {
	s.onChange = fn
	return s
}

func (s Widget) Value(value float64) Widget {
	s.value = value
	s.hasValue = true
	return s
}

func (s Widget) DefaultValue(value float64) Widget {
	s.defaultValue = value
	s.hasDefault = true
	s.hasValue = false
	return s
}

func (s Widget) UpperValue(value float64) Widget {
	s.upperValue = value
	s.hasUpperValue = true
	return s
}

func (s Widget) DefaultUpperValue(value float64) Widget {
	s.defaultUpper = value
	s.hasDefaultUpper = true
	s.hasUpperValue = false
	return s
}

func (s Widget) OnRangeChange(fn func(float64, float64)) Widget {
	s.onRangeChange = fn
	return s
}

func (s Widget) Orientation(orientation Orientation) Widget {
	s.orientation = orientation
	return s
}

func (s Widget) Vertical() Widget {
	s.orientation = Vertical
	return s
}

func (s Widget) Disabled(disabled bool) Widget {
	s.disabled = disabled
	return s
}

func (s Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return s.layoutWithOverride(ctx, gtx, s.override)
}

func (s Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.SliderOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
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
	core.RegisterFieldFocus(ctx, key, &state.lowerThumb.clickable, enabled)
	if !enabled {
		gtx = gtx.Disabled()
	}

	if enabled {
		state.updateThumbPresses(ctx, gtx, values.rangeMode)
		if changedValues, changed, thumb, keyboard := state.update(gtx, values); changed {
			values = changedValues
			state.syncRatios(values)
			core.RequestFocusVisible(ctx, state.thumbTag(thumb), keyboard)
			s.dispatchWithState(state, values)
		}
	}

	hovered := state.lowerThumb.clickable.Hovered() || values.rangeMode && state.upperThumb.clickable.Hovered()
	dragging := state.lower.Dragging() || values.rangeMode && state.upper.Dragging()
	pressed := state.lowerThumb.clickable.Pressed() || dragging || values.rangeMode && state.upperThumb.clickable.Pressed()
	lowerFocused := gtx.Focused(&state.lowerThumb.clickable)
	upperFocused := values.rangeMode && gtx.Focused(&state.upperThumb.clickable)
	focused := lowerFocused || upperFocused
	focusVisible := core.FocusVisible(ctx, &state.lowerThumb.clickable, lowerFocused)
	if values.rangeMode {
		focusVisible = focusVisible || core.FocusVisible(ctx, &state.upperThumb.clickable, upperFocused)
	}
	styleState := style.State{
		Hovered:      hovered,
		Pressed:      pressed,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     !enabled,
		Dragging:     dragging,
	}
	activeTheme := core.ActiveTheme(ctx)
	visual := sliderVisualCached(activeTheme, s.orientation, override, &state.styleCache)
	componentStyle := s.resolveStyle(ctx, gtx, key, styleState, visual)
	layoutSlider := core.WithFieldSemantics(ctx, key, func(gtx layout.Context) layout.Dimensions {
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		semantic.DescriptionOp(s.semanticDescriptionFor(ctx, values)).Add(gtx.Ops)
		return s.layout(ctx, gtx, state, componentStyle, values)
	})
	return layoutui.Styled(ctx, gtx, key, styleState, componentStyle.visual.Override, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutSlider(gtx)
	}))
}

func (s Widget) axis() layout.Axis {
	if s.orientation == Vertical {
		return layout.Vertical
	}
	return layout.Horizontal
}

func (s Widget) dispatchWithState(state *sliderState, values sliderValues) {
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

func (s Widget) outputText(values sliderValues) string {
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

func (s Widget) format(value float64) string {
	if s.formatValue != nil {
		return s.formatValue(value)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (s Widget) semanticDescriptionFor(ctx *core.Context, values sliderValues) string {
	label := s.label
	if label == "" {
		label = ctx.Text(i18n.SliderDefault)
	}
	output := s.outputText(values)
	if output == "" {
		if values.rangeMode {
			output = s.format(values.lower) + " - " + s.format(values.upper)
		} else {
			output = s.format(values.lower)
		}
	}
	return ctx.Textf(i18n.SliderValue, label, output)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.SliderOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
