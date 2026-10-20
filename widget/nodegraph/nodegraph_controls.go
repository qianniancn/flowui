package nodegraph

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type graphControl uint8

const (
	controlZoomOut graphControl = iota
	controlZoomIn
	controlFit
)

func (w Widget) controlAt(position f32.Point, size image.Point, metrics theme.NodeGraphMetrics, pixelsPerDP float32) (graphControl, bool) {
	if !w.controlsEnabled {
		return 0, false
	}
	rect := graphControlsRect(size, metrics, pixelsPerDP)
	if !position.Round().In(rect) {
		return 0, false
	}
	buttonWidth := rect.Dx() / 3
	index := min(max((int(position.X)-rect.Min.X)/buttonWidth, 0), 2)
	return graphControl(index), true
}

func (w Widget) applyControl(control graphControl, graph resolvedGraph, viewport Viewport, size image.Point, pixelsPerDP float32) (Viewport, bool) {
	switch control {
	case controlZoomOut, controlZoomIn:
		factor := float32(1.2)
		if control == controlZoomOut {
			factor = 1 / factor
		}
		nextZoom := min(max(viewport.Zoom*factor, w.minZoom), w.maxZoom)
		if nextZoom == viewport.Zoom {
			return viewport, false
		}
		center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
		return zoomAt(viewport, nextZoom, center, pixelsPerDP), true
	case controlFit:
		return fitGraphViewport(graph, size, pixelsPerDP, w.fitViewPadding, w.minZoom, w.maxZoom)
	default:
		return viewport, false
	}
}

func graphControlsRect(size image.Point, metrics theme.NodeGraphMetrics, pixelsPerDP float32) image.Rectangle {
	width := max(int(float32(metrics.ControlWidth)*pixelsPerDP), 1)
	height := max(int(float32(metrics.ControlHeight)*pixelsPerDP), 1)
	margin := max(int(float32(metrics.ControlMargin)*pixelsPerDP), 0)
	return image.Rect(margin, size.Y-margin-height, margin+width, size.Y-margin)
}

func drawNodeControls(ctx *core.Context, gtx layout.Context, visual theme.NodeGraphVisual, enabled bool) {
	if !enabled {
		return
	}
	rect := graphControlsRect(gtx.Constraints.Max, visual.Metrics, graphPixelsPerDP(gtx))
	if rect.Empty() {
		return
	}
	visual.Controls(gtx, theme.NodeGraphControlsRequest{
		Rect: rect, ButtonWidth: rect.Dx() / 3,
		Radius:       max(gtx.Dp(visual.Metrics.ControlRadius), 1),
		DividerInset: max(gtx.Dp(visual.Metrics.ControlDividerInset), 1),
		TextSize:     visual.Metrics.ControlTextSize, FitSpan: visual.Metrics.ControlFitSpan,
		FitArm: visual.Metrics.ControlFitArm, Background: visual.Colors.ControlBackground,
		Foreground: visual.Colors.ControlForeground, Border: visual.Colors.ControlBorder,
	})
}
