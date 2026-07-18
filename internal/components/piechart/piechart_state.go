package piechart

import (
	"image"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotPieChart = "pie-chart"

type chartState struct {
	click             gesture.Click
	pointerTag        struct{}
	hovered           bool
	pointer           f32.Point
	animation         chart.DataAnimation[chartData]
	dataCache         chartDataCache
	legendItems       map[string]*chart.LegendItem
	legendFrame       map[string]struct{}
	tooltipTransition tooltip.PopupTransition
	tooltipSlice      resolvedSlice
}

func chartStateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindPieChart, key)
	return frame.UseState[chartState](ctx, key, stateSlotPieChart)
}

func (s *chartState) beginLegendFrame() {
	if s.legendFrame == nil {
		s.legendFrame = make(map[string]struct{})
	} else {
		clear(s.legendFrame)
	}
}

func (s *chartState) endLegendFrame() {
	for key := range s.legendItems {
		if _, ok := s.legendFrame[key]; !ok {
			delete(s.legendItems, key)
		}
	}
}

func (s *chartState) legendItem(key string) *chart.LegendItem {
	if s.legendItems == nil {
		s.legendItems = make(map[string]*chart.LegendItem)
	}
	s.legendFrame[key] = struct{}{}
	if item := s.legendItems[key]; item != nil {
		return item
	}
	item := new(chart.LegendItem)
	s.legendItems[key] = item
	return item
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool) {
	for {
		value, ok := gtx.Event(pointer.Filter{
			Target: &s.pointerTag,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Release | pointer.Cancel,
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
		case pointer.Leave, pointer.Cancel:
			s.hovered = false
		}
	}
	if !enabled {
		s.hovered = false
	}
}

func (s *chartState) addPointerInput(gtx layout.Context, area image.Rectangle, enabled bool) {
	if !enabled || area.Empty() {
		return
	}
	stack := clip.Rect(area).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)
	event.Op(gtx.Ops, &s.pointerTag)
	pass.Pop()
	stack.Pop()
}

func (s *chartState) updateClick(gtx layout.Context, enabled bool) bool {
	activated := false
	for {
		click, ok := s.click.Update(gtx.Source)
		if !ok {
			break
		}
		activated = activated || click.Kind == gesture.KindClick
	}
	return activated && enabled
}

func (s *chartState) addClickInput(gtx layout.Context, area image.Rectangle, enabled bool) {
	if !enabled || area.Empty() {
		return
	}
	stack := clip.Rect(area).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	s.click.Add(gtx.Ops)
	pass.Pop()
	stack.Pop()
}

func (s *chartState) clearTooltip() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSlice = resolvedSlice{}
}
