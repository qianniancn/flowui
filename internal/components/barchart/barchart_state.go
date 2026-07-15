package barchart

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

const stateSlotBarChart = "bar-chart"

type chartState struct {
	click             gesture.Click
	pointerTag        struct{}
	hovered           bool
	pointer           f32.Point
	animation         barChartAnimation
	legendItems       map[string]*chart.LegendItem
	legendFrame       map[string]struct{}
	windowGesture     chart.DataWindowGesture
	tooltipTransition tooltip.PopupTransition
	tooltipSelection  chartSelection
}

func (s *chartState) beginLegendFrame() {
	stateutil.BeginFrameMap(&s.legendFrame)
}

func (s *chartState) endLegendFrame() {
	stateutil.SweepFrameMap(s.legendItems, s.legendFrame)
}

func (s *chartState) legendItem(key string) *chart.LegendItem {
	return stateutil.UseFrameMap(&s.legendItems, &s.legendFrame, key)
}

func chartStateFor(ctx *frame.Context, key string) *chartState {
	key = frame.ClaimKey(ctx, stateutil.KindBarChart, key)
	return frame.UseState[chartState](ctx, key, stateSlotBarChart)
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, window chart.DataWindow, verticalWindow bool, onWindowChange func(chart.DataWindow)) {
	chart.UpdatePointer(gtx, enabled, plot, window, verticalWindow, onWindowChange, &s.pointerTag, &s.hovered, &s.pointer, &s.windowGesture)
}

func (s *chartState) selectedIndex(start, end int, plot image.Rectangle, horizontal bool) (int, bool) {
	return chart.SelectedIndex(s.pointer, s.hovered, start, end, plot, horizontal)
}

func (s *chartState) addPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool) {
	chart.AddPointerInput(gtx, plot, enabled, &s.pointerTag)
}

func (s *chartState) clearSelection() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSelection = chartSelection{}
}

func (s *chartState) updateClicks(gtx layout.Context, enabled bool) (bool, bool) {
	return chart.UpdateClicks(gtx, enabled, &s.click)
}

func (s *chartState) addClickInput(gtx layout.Context, size image.Point, enabled bool) {
	chart.AddClickInput(gtx, size, enabled, &s.click)
}
