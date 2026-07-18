package linechart

import (
	"image"
	"sort"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/components/tooltip"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotLineChart = "line-chart"

type chartState struct {
	click             gesture.Click
	pointerTag        struct{}
	hovered           bool
	pointer           f32.Point
	animation         chart.DataAnimation[chartData]
	dataCache         chartDataCache
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
	key = frame.ClaimKey(ctx, stateutil.KindLineChart, key)
	return frame.UseState[chartState](ctx, key, stateSlotLineChart)
}

func (s *chartState) updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, window chart.DataWindow, onWindowChange func(chart.DataWindow)) {
	chart.UpdatePointer(gtx, enabled, plot, window, false, onWindowChange, &s.pointerTag, &s.hovered, &s.pointer, &s.windowGesture)
}

func (s *chartState) addPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool) {
	chart.AddPointerInput(gtx, plot, enabled, &s.pointerTag)
}

func (s *chartState) clearSelection() {
	s.hovered = false
	s.tooltipTransition.Reset()
	s.tooltipSelection = chartSelection{}
}

func visibleXValues(values []float64, scale chart.LinearScale) []float64 {
	start := sort.Search(len(values), func(index int) bool { return values[index] >= scale.Minimum })
	end := sort.Search(len(values), func(index int) bool { return values[index] > scale.Maximum })
	return values[start:end]
}

func (s *chartState) selectedX(values []float64, scale chart.LinearScale, plot image.Rectangle) (float64, bool) {
	return chart.NearestX(s.pointer, s.hovered, values, scale.Minimum, scale.Maximum, plot)
}

func (s *chartState) updateClicks(gtx layout.Context, enabled bool) (bool, bool) {
	return chart.UpdateClicks(gtx, enabled, &s.click)
}

func (s *chartState) addClickInput(gtx layout.Context, size image.Point, enabled bool) {
	chart.AddClickInput(gtx, size, enabled, &s.click)
}
