package slider

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
)

const stateSlotSlider = "slider"

func sliderStateFor(ctx *frame.Context, key string) (string, *sliderState) {
	key = frame.ClaimKey(ctx, state.KindSlider, key)
	return key, frame.UseState[sliderState](ctx, key, stateSlotSlider)
}

type sliderState struct {
	lowerDisclosure disclosure.Binding[float64]
	upperDisclosure disclosure.Binding[float64]
	currentLower    float64
	currentUpper    float64
	lower           widget.Float
	upper           widget.Float
	lowerThumb      sliderThumbState
	upperThumb      sliderThumbState
	axis            layout.Axis
	axisReady       bool
}

// sliderLowerDisclosureCfg builds a disclosure.Config for the lower/value.
func sliderLowerDisclosureCfg(widget SliderWidget) disclosure.Config[float64] {
	return disclosure.Config[float64]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

// sliderUpperDisclosureCfg builds a disclosure.Config for the upper value (range mode).
func sliderUpperDisclosureCfg(widget SliderWidget) disclosure.Config[float64] {
	return disclosure.Config[float64]{
		Controlled: widget.hasUpperValue,
		Value:      widget.upperValue,
		HasDefault: widget.hasDefaultUpper,
		Default:    widget.defaultUpper,
		OnChange:   nil, // Upper uses onRangeChange, not onChange
	}
}

func (s *sliderState) bindLower(widget SliderWidget) {
	s.lowerDisclosure.Bind(sliderLowerDisclosureCfg(widget))
}

func (s *sliderState) bindUpper(widget SliderWidget) {
	s.upperDisclosure.Bind(sliderUpperDisclosureCfg(widget))
}

func (s *sliderState) currentLowerValue(widget SliderWidget) float64 {
	s.currentLower = s.lowerDisclosure.Current(sliderLowerDisclosureCfg(widget))
	return s.currentLower
}

func (s *sliderState) currentUpperValue(widget SliderWidget) float64 {
	s.currentUpper = s.upperDisclosure.Current(sliderUpperDisclosureCfg(widget))
	return s.currentUpper
}

func (s *sliderState) requestLower(widget SliderWidget, value float64) float64 {
	s.currentLower, _ = s.lowerDisclosure.Request(sliderLowerDisclosureCfg(widget), value)
	return s.currentLower
}

func (s *sliderState) requestUpper(widget SliderWidget, value float64) float64 {
	s.currentUpper, _ = s.upperDisclosure.Request(sliderUpperDisclosureCfg(widget), value)
	return s.currentUpper
}

type sliderThumbState struct {
	clickable widget.Clickable
	focus     state.FocusAnimation
	scale     animation.FloatTransition
}

const sliderThumbScaleDuration = 150 * time.Millisecond

func (s *sliderState) setAxis(axis layout.Axis) {
	if s.axisReady && s.axis != axis {
		s.lower = widget.Float{Value: s.lower.Value}
		s.upper = widget.Float{Value: s.upper.Value}
	}
	s.axis = axis
	s.axisReady = true
}

func (s *sliderState) sync(values sliderValues) {
	if !s.lower.Dragging() {
		s.lower.Value = values.ratio(values.lower)
	}
	if values.rangeMode && !s.upper.Dragging() {
		s.upper.Value = values.ratio(values.upper)
	}
}

func (s *sliderState) syncRatios(values sliderValues) {
	s.lower.Value = values.ratio(values.lower)
	if values.rangeMode {
		s.upper.Value = values.ratio(values.upper)
	}
}

func (s *sliderState) updateThumbPresses(ctx *frame.Context, gtx layout.Context, rangeMode bool) {
	s.updateThumbPress(ctx, gtx, &s.lowerThumb)
	if rangeMode {
		s.updateThumbPress(ctx, gtx, &s.upperThumb)
	}
}

func (s *sliderState) updateThumbPress(ctx *frame.Context, gtx layout.Context, thumb *sliderThumbState) {
	presses := state.ActivePresses(thumb.clickable.History())
	for thumb.clickable.Clicked(gtx) {
	}
	frame.FocusOnPress(ctx, &thumb.clickable, thumb.clickable.History(), presses)
}

func (s *sliderState) update(gtx layout.Context, values sliderValues) (sliderValues, bool, int, bool) {
	if keyboardValues, changed, thumb := s.updateKeyboard(gtx, values); changed {
		return keyboardValues, true, thumb, true
	}

	lowerChanged := s.lower.Update(gtx)
	upperChanged := values.rangeMode && s.upper.Update(gtx)
	if !lowerChanged && !upperChanged {
		return values, false, 0, false
	}

	next := values
	thumb := 0
	if lowerChanged {
		next.lower = values.value(s.lower.Value)
		if values.rangeMode {
			next.lower = min(next.lower, values.upper)
		}
	}
	if upperChanged {
		next.upper = max(values.value(s.upper.Value), values.lower)
		thumb = 1
	}
	if !next.rangeMode {
		next.upper = next.lower
	}
	return next, true, thumb, false
}

func (s *sliderState) updateKeyboard(gtx layout.Context, values sliderValues) (sliderValues, bool, int) {
	tags := []event.Tag{&s.lowerThumb.clickable}
	if values.rangeMode {
		tags = append(tags, &s.upperThumb.clickable)
	}
	for index, tag := range tags {
		for {
			e, ok := gtx.Event(
				key.Filter{Focus: tag, Name: key.NameLeftArrow},
				key.Filter{Focus: tag, Name: key.NameRightArrow},
				key.Filter{Focus: tag, Name: key.NameDownArrow},
				key.Filter{Focus: tag, Name: key.NameUpArrow},
				key.Filter{Focus: tag, Name: key.NameHome},
				key.Filter{Focus: tag, Name: key.NameEnd},
				key.Filter{Focus: tag, Name: key.NamePageDown},
				key.Filter{Focus: tag, Name: key.NamePageUp},
			)
			if !ok {
				break
			}
			event, ok := e.(key.Event)
			if !ok || event.State != key.Press {
				continue
			}
			next := values
			current := next.lower
			minimum, maximum := next.minValue, next.maxValue
			if index == 1 {
				current = next.upper
				minimum = next.lower
			} else if next.rangeMode {
				maximum = next.upper
			}
			switch event.Name {
			case key.NameLeftArrow, key.NameDownArrow:
				current -= next.step
			case key.NameRightArrow, key.NameUpArrow:
				current += next.step
			case key.NameHome:
				current = minimum
			case key.NameEnd:
				current = maximum
			case key.NamePageDown:
				current -= next.step * 10
			case key.NamePageUp:
				current += next.step * 10
			}
			current = min(max(next.snap(current), minimum), maximum)
			if index == 0 {
				next.lower = current
				if !next.rangeMode {
					next.upper = current
				}
			} else {
				next.upper = current
			}
			if next.lower != values.lower || next.upper != values.upper {
				return next, true, index
			}
		}
	}
	return values, false, 0
}

func (s *sliderState) thumbTag(index int) event.Tag {
	if index == 1 {
		return &s.upperThumb.clickable
	}
	return &s.lowerThumb.clickable
}

func (s *sliderState) thumb(index int) *sliderThumbState {
	if index == 1 {
		return &s.upperThumb
	}
	return &s.lowerThumb
}

func (s *sliderState) dragging(index int) bool {
	if index == 1 {
		return s.upper.Dragging()
	}
	return s.lower.Dragging()
}

func (t *sliderThumbState) focusOpacity(ctx *frame.Context, gtx layout.Context, focused bool) float32 {
	visible := frame.FocusVisible(ctx, &t.clickable, focused)
	return t.focus.Opacity(gtx, visible, frame.ActiveTheme(ctx).Motion)
}

func (t *sliderThumbState) draggingScale(ctx *frame.Context, gtx layout.Context, dragging bool, targetScale float32) float32 {
	if targetScale <= 0 || targetScale > 1 {
		targetScale = 0.9
	}
	target := float32(1)
	if dragging {
		target = targetScale
	}
	return t.scale.Value(gtx, target, sliderThumbScaleDuration, animation.EaseSmoothstep, frame.ActiveTheme(ctx).Motion)
}
