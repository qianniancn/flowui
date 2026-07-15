package candlestick

import (
	"image"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/layout"
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
	dataCache         chartDataCache
	tooltipTransition tooltip.PopupTransition
	tooltipSelection  chartSelection
}

func chartStateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindCandlestickChart, key)
	return frame.UseState[chartState](ctx, key, stateSlotCandlestickChart)
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, window chart.DataWindow, onWindowChange func(chart.DataWindow)) {
	chart.UpdatePointer(gtx, enabled, plot, window, false, onWindowChange, &s.pointerTag, &s.hovered, &s.pointer, &s.windowGesture)
}

func (s *chartState) selectedIndex(start, end int, plot image.Rectangle) (int, bool) {
	return chart.SelectedIndex(s.pointer, s.hovered, start, end, plot, false)
}

func (s *chartState) addPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool) {
	chart.AddPointerInput(gtx, plot, enabled, &s.pointerTag)
}

func (s *chartState) updateClicks(gtx layout.Context, enabled bool) (activated, reset bool) {
	return chart.UpdateClicks(gtx, enabled, &s.click)
}

func (s *chartState) addClickInput(gtx layout.Context, size image.Point, enabled bool) {
	chart.AddClickInput(gtx, size, enabled, &s.click)
}

func (s *chartState) clearSelection() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSelection = chartSelection{}
}
