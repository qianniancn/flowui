package candlestick

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

const stateSlotCandlestickChart = "candlestick-chart"

type chartState struct {
	click             gesture.Click
	pointerTag        struct{}
	hovered           bool
	pointer           f32.Point
	windowGesture     chart.DataWindowGesture
	animation         candlestickAnimation
	tooltipTransition tooltip.PopupTransition
	tooltipSelection  chartSelection
}

func chartStateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindCandlestickChart, key)
	return frame.UseState[chartState](ctx, key, stateSlotCandlestickChart)
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, window chart.DataWindow, onWindowChange func(chart.DataWindow)) {
	activeWindow := window
	kinds := pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Release | pointer.Cancel
	scrollX, scrollY := pointer.ScrollRange{}, pointer.ScrollRange{}
	if onWindowChange != nil {
		kinds |= pointer.Scroll
		scrollX = pointer.ScrollRange{Min: -100000, Max: 100000}
		scrollY = pointer.ScrollRange{Min: -100000, Max: 100000}
	}
	for {
		value, ok := gtx.Event(pointer.Filter{Target: &s.pointerTag, Kinds: kinds, ScrollX: scrollX, ScrollY: scrollY})
		if !ok {
			break
		}
		eventValue, ok := value.(pointer.Event)
		if !ok {
			continue
		}
		if !enabled {
			s.hovered = false
			s.windowGesture.Cancel()
			continue
		}
		if onWindowChange != nil {
			if next, changed := s.windowGesture.Update(eventValue, plot, activeWindow, false); changed {
				activeWindow = next
				onWindowChange(next)
			}
		} else {
			s.windowGesture.Cancel()
		}
		switch eventValue.Kind {
		case pointer.Enter, pointer.Move, pointer.Drag, pointer.Press, pointer.Scroll:
			s.hovered = true
			s.pointer = eventValue.Position
		case pointer.Leave, pointer.Cancel:
			s.hovered = false
		}
	}
	if !enabled {
		s.hovered = false
		s.windowGesture.Cancel()
	}
}

func (s *chartState) selectedIndex(start, end int, plot image.Rectangle) (int, bool) {
	if !s.hovered || end <= start || plot.Empty() {
		return 0, false
	}
	ratio := (float64(s.pointer.X) - float64(plot.Min.X)) / float64(plot.Dx())
	index := start + int(ratio*float64(end-start))
	return min(max(index, start), end-1), true
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

func (s *chartState) updateClicks(gtx layout.Context, enabled bool) (activated, reset bool) {
	for {
		click, ok := s.click.Update(gtx.Source)
		if !ok {
			break
		}
		if click.Kind == gesture.KindClick {
			activated = true
			reset = reset || click.NumClicks >= 2
		}
	}
	return activated && enabled, reset && enabled
}

func (s *chartState) addClickInput(gtx layout.Context, size image.Point, enabled bool) {
	if !enabled || size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	s.click.Add(gtx.Ops)
	pass.Pop()
	area.Pop()
}

func (s *chartState) clearSelection() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSelection = chartSelection{}
}
