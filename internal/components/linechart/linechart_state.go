package linechart

import (
	"image"
	"sort"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotLineChart = "line-chart"

type chartState struct {
	root          widget.Clickable
	pointerTag    struct{}
	focus         stateutil.FocusAnimation
	hovered       bool
	pointer       f32.Point
	keyboard      bool
	keyboardIndex int
	pointerIndex  int
}

func chartStateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindLineChart, key)
	return frame.UseStateWith(ctx, key, stateSlotLineChart, func() *chartState {
		return &chartState{keyboardIndex: -1, pointerIndex: -1}
	})
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool) {
	for {
		value, ok := gtx.Event(pointer.Filter{
			Target: &s.pointerTag,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Cancel,
		})
		if !ok {
			break
		}
		eventValue, ok := value.(pointer.Event)
		if !ok {
			continue
		}
		if !enabled {
			s.hovered = false
			continue
		}
		switch eventValue.Kind {
		case pointer.Enter, pointer.Move, pointer.Drag, pointer.Press:
			s.hovered = true
			s.pointer = eventValue.Position
			s.keyboard = false
			s.keyboardIndex = -1
			s.pointerIndex = -1
		case pointer.Leave, pointer.Cancel:
			s.hovered = false
			s.pointerIndex = -1
		}
	}
	if !enabled {
		s.hovered = false
	}
}

func (s *chartState) updateKeyboard(ctx *frame.Context, gtx layout.Context, xValues []float64, enabled bool) {
	for {
		value, ok := gtx.Event(
			key.Filter{Focus: &s.root, Name: key.NameLeftArrow},
			key.Filter{Focus: &s.root, Name: key.NameRightArrow},
			key.Filter{Focus: &s.root, Name: key.NameHome},
			key.Filter{Focus: &s.root, Name: key.NameEnd},
			key.Filter{Focus: &s.root, Name: key.NameEscape},
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		if eventValue.Name == key.NameEscape {
			s.keyboard = false
			s.keyboardIndex = -1
			continue
		}
		if len(xValues) == 0 {
			continue
		}
		s.keyboard = true
		s.hovered = false
		s.focus.Prepare(true)
		frame.RequestFocusVisible(ctx, &s.root, true)
		switch eventValue.Name {
		case key.NameHome:
			s.keyboardIndex = 0
		case key.NameEnd:
			s.keyboardIndex = len(xValues) - 1
		case key.NameLeftArrow:
			if s.keyboardIndex < 0 {
				if s.pointerIndex >= 0 {
					s.keyboardIndex = max(s.pointerIndex-1, 0)
				} else {
					s.keyboardIndex = len(xValues) - 1
				}
			} else {
				s.keyboardIndex = max(s.keyboardIndex-1, 0)
			}
		case key.NameRightArrow:
			if s.keyboardIndex < 0 {
				if s.pointerIndex >= 0 {
					s.keyboardIndex = min(s.pointerIndex+1, len(xValues)-1)
				} else {
					s.keyboardIndex = 0
				}
			} else {
				s.keyboardIndex = min(s.keyboardIndex+1, len(xValues)-1)
			}
		}
	}
	if s.keyboardIndex >= len(xValues) {
		s.keyboardIndex = len(xValues) - 1
	}
}

func (s *chartState) addPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool) {
	if !enabled || plot.Empty() {
		return
	}
	area := clip.Rect(plot).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	pointer.CursorCrosshair.Add(gtx.Ops)
	event.Op(gtx.Ops, &s.pointerTag)
	pass.Pop()
	area.Pop()
}

func (s *chartState) clearSelection() {
	s.hovered = false
	s.keyboard = false
	s.keyboardIndex = -1
	s.pointerIndex = -1
}

func visibleXValues(values []float64, scale linearScale) []float64 {
	start := sort.Search(len(values), func(index int) bool { return values[index] >= scale.minimum })
	end := sort.Search(len(values), func(index int) bool { return values[index] > scale.maximum })
	return values[start:end]
}

func (s *chartState) selectedX(values []float64, scale linearScale, plot image.Rectangle, focused bool) (float64, bool) {
	if s.keyboard && focused && s.keyboardIndex >= 0 && s.keyboardIndex < len(values) {
		return values[s.keyboardIndex], true
	}
	if !s.hovered || len(values) == 0 || plot.Empty() {
		return 0, false
	}
	ratio := (float64(s.pointer.X) - float64(plot.Min.X)) / float64(plot.Dx())
	ratio = min(max(ratio, 0), 1)
	target := scale.minimum + ratio*(scale.maximum-scale.minimum)
	index := sort.SearchFloat64s(values, target)
	if index <= 0 {
		s.pointerIndex = 0
		return values[0], true
	}
	if index >= len(values) {
		s.pointerIndex = len(values) - 1
		return values[len(values)-1], true
	}
	if target-values[index-1] <= values[index]-target {
		s.pointerIndex = index - 1
		return values[index-1], true
	}
	s.pointerIndex = index
	return values[index], true
}

func (s *chartState) requestPointerFocus(ctx *frame.Context, gtx layout.Context, enabled bool) {
	presses := stateutil.ActivePresses(s.root.History())
	for s.root.Clicked(gtx) {
	}
	if enabled {
		frame.FocusOnPress(ctx, &s.root, s.root.History(), presses)
	}
}
