package slider

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotSlider = "slider"

func sliderStateFor(ctx *frame.Context, key string) (string, *sliderState) {
	key = frame.ClaimKey(ctx, state.KindSlider, key)
	return key, frame.UseState[sliderState](ctx, key, stateSlotSlider)
}

type sliderState struct {
	lower      widget.Float
	upper      widget.Float
	lowerThumb sliderThumbState
	upperThumb sliderThumbState
	axis       layout.Axis
	axisReady  bool
}

type sliderThumbState struct {
	clickable           widget.Clickable
	focus               state.FocusAnimation
	scale               float32
	scaleFrom           float32
	scaleTo             float32
	scaleAt             time.Time
	scaleReady          bool
	pointerFocus        bool
	pointerFocusPending bool
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

func (s *sliderState) update(gtx layout.Context, values sliderValues) (sliderValues, bool, int) {
	if keyboardValues, changed, thumb := s.updateKeyboard(gtx, values); changed {
		return keyboardValues, true, thumb
	}

	lowerChanged := s.lower.Update(gtx)
	upperChanged := values.rangeMode && s.upper.Update(gtx)
	if !lowerChanged && !upperChanged {
		return values, false, 0
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
	s.thumb(thumb).pointerFocus = true
	s.thumb(thumb).pointerFocusPending = true
	if !next.rangeMode {
		next.upper = next.lower
	}
	return next, true, thumb
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
			s.thumb(index).pointerFocus = false
			s.thumb(index).pointerFocusPending = false
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

func (t *sliderThumbState) focusOpacity(gtx layout.Context, focused bool) float32 {
	visible := t.focus.Visible(focused, t.clickable.History())
	if focused {
		t.pointerFocusPending = false
	} else if !t.pointerFocusPending {
		t.pointerFocus = false
	}
	visible = visible && !t.pointerFocus
	return t.focus.Opacity(gtx, visible)
}

func (t *sliderThumbState) draggingScale(gtx layout.Context, dragging bool, targetScale float32) float32 {
	if targetScale <= 0 || targetScale > 1 {
		targetScale = 0.9
	}
	target := float32(1)
	if dragging {
		target = targetScale
	}
	if !t.scaleReady {
		t.scale = target
		t.scaleFrom = target
		t.scaleTo = target
		t.scaleAt = gtx.Now
		t.scaleReady = true
		return target
	}
	if target != t.scaleTo {
		t.scaleFrom = t.scale
		t.scaleTo = target
		t.scaleAt = gtx.Now
	}
	if t.scaleFrom == t.scaleTo {
		t.scale = t.scaleTo
		return t.scale
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(t.scaleAt), sliderThumbScaleDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	t.scale = render.Lerp(t.scaleFrom, t.scaleTo, progress)
	return t.scale
}
