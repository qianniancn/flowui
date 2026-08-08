package nodegraph

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/theme"
)

type graphControl uint8

const (
	controlZoomOut graphControl = iota
	controlZoomIn
	controlFit
)

func (w Widget) controlAt(position f32.Point, size image.Point) (graphControl, bool) {
	if !w.controlsEnabled {
		return 0, false
	}
	rect := graphControlsRect(size)
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

func graphControlsRect(size image.Point) image.Rectangle {
	const width, height, margin = 108, 36, 12
	return image.Rect(margin, size.Y-margin-height, margin+width, size.Y-margin)
}

func drawNodeControls(ctx *frame.Context, gtx layout.Context, tokens theme.NodeGraphTheme, enabled bool) {
	if !enabled {
		return
	}
	rect := graphControlsRect(gtx.Constraints.Max)
	if rect.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, tokens.NodeBackground, clip.UniformRRect(rect, 5).Op(gtx.Ops))
	drawGraphRoundedStroke(gtx, rect, 5, max(gtx.Dp(unit.Dp(1)), 1), tokens.CanvasBorder)
	buttonWidth := rect.Dx() / 3
	for index, label := range []string{"-", "+", ""} {
		button := image.Rect(rect.Min.X+index*buttonWidth, rect.Min.Y, rect.Min.X+(index+1)*buttonWidth, rect.Max.Y)
		if index > 0 {
			paint.FillShape(gtx.Ops, tokens.CanvasBorder, clip.Rect(image.Rect(button.Min.X, button.Min.Y+8, button.Min.X+1, button.Max.Y-8)).Op())
		}
		if label != "" {
			drawGraphCenteredText(ctx, gtx, label, unit.Sp(16), font.Normal, tokens.NodeForeground, button)
		} else {
			drawFitIcon(gtx, button, tokens.NodeForeground)
		}
	}
}

func drawFitIcon(gtx layout.Context, rect image.Rectangle, colorValue color.NRGBA) {
	center := f32.Pt(float32(rect.Min.X+rect.Max.X)/2, float32(rect.Min.Y+rect.Max.Y)/2)
	span := float32(min(rect.Dx(), rect.Dy())) * .18
	arm := span * .8
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(center.Add(f32.Pt(-span, -arm)))
	path.LineTo(center.Add(f32.Pt(-span, -span)))
	path.LineTo(center.Add(f32.Pt(-arm, -span)))
	path.MoveTo(center.Add(f32.Pt(span, -arm)))
	path.LineTo(center.Add(f32.Pt(span, -span)))
	path.LineTo(center.Add(f32.Pt(arm, -span)))
	path.MoveTo(center.Add(f32.Pt(-span, arm)))
	path.LineTo(center.Add(f32.Pt(-span, span)))
	path.LineTo(center.Add(f32.Pt(-arm, span)))
	path.MoveTo(center.Add(f32.Pt(span, arm)))
	path.LineTo(center.Add(f32.Pt(span, span)))
	path.LineTo(center.Add(f32.Pt(arm, span)))
	stroke := clip.Stroke{Path: path.End(), Width: float32(max(gtx.Dp(unit.Dp(1)), 1))}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}
