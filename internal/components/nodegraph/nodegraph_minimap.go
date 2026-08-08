package nodegraph

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/theme"
)

const (
	minimapWidth  = 180
	minimapHeight = 120
	minimapMargin = 12
)

type minimapGeometry struct {
	rect   image.Rectangle
	inner  image.Rectangle
	bounds graphRect
	scale  float32
}

func resolveMinimapGeometry(graph resolvedGraph, size image.Point) (minimapGeometry, bool) {
	if len(graph.nodes) == 0 {
		return minimapGeometry{}, false
	}
	rect := image.Rect(size.X-minimapMargin-minimapWidth, size.Y-minimapMargin-minimapHeight, size.X-minimapMargin, size.Y-minimapMargin)
	if rect.Min.X < 0 || rect.Min.Y < 0 {
		return minimapGeometry{}, false
	}
	bounds := graphBounds(graph)
	inner := rect.Inset(8)
	width := max(bounds.max.X-bounds.min.X, 1)
	height := max(bounds.max.Y-bounds.min.Y, 1)
	return minimapGeometry{rect: rect, inner: inner, bounds: bounds, scale: min(float32(inner.Dx())/width, float32(inner.Dy())/height)}, true
}

func (m minimapGeometry) mapPoint(point Point) f32.Point {
	return f32.Pt(float32(m.inner.Min.X)+(point.X-m.bounds.min.X)*m.scale, float32(m.inner.Min.Y)+(point.Y-m.bounds.min.Y)*m.scale)
}

func (m minimapGeometry) worldPoint(position f32.Point) Point {
	return Point{X: m.bounds.min.X + (position.X-float32(m.inner.Min.X))/m.scale, Y: m.bounds.min.Y + (position.Y-float32(m.inner.Min.Y))/m.scale}
}

func minimapWorldAt(graph resolvedGraph, position f32.Point, size image.Point) (Point, bool) {
	geometry, ok := resolveMinimapGeometry(graph, size)
	if !ok || !position.Round().In(geometry.inner) {
		return Point{}, false
	}
	return geometry.worldPoint(position), true
}

func minimapViewportScreen(graph resolvedGraph, viewport Viewport, size image.Point, pixelsPerDP float32) image.Rectangle {
	geometry, ok := resolveMinimapGeometry(graph, size)
	if !ok {
		return image.Rectangle{}
	}
	visibleWidth := float32(size.X) / worldScale(viewport, pixelsPerDP)
	visibleHeight := float32(size.Y) / worldScale(viewport, pixelsPerDP)
	first := geometry.mapPoint(viewport.Origin)
	last := geometry.mapPoint(Point{X: viewport.Origin.X + visibleWidth, Y: viewport.Origin.Y + visibleHeight})
	return image.Rect(int(math.Round(float64(first.X))), int(math.Round(float64(first.Y))), int(math.Round(float64(last.X))), int(math.Round(float64(last.Y))))
}

func drawNodeMinimap(gtx layout.Context, graph resolvedGraph, viewport Viewport, selected map[string]bool, tokens theme.NodeGraphTheme, enabled bool) {
	if !enabled || len(graph.nodes) == 0 {
		return
	}
	canvas := gtx.Constraints.Max
	geometry, ok := resolveMinimapGeometry(graph, canvas)
	if !ok {
		return
	}
	rect := geometry.rect
	paint.FillShape(gtx.Ops, tokens.NodeBackground, clip.UniformRRect(rect, 5).Op(gtx.Ops))
	drawGraphRoundedStroke(gtx, rect, 5, max(gtx.Dp(unit.Dp(1)), 1), tokens.CanvasBorder)
	clipStack := clip.Rect(rect.Inset(1)).Push(gtx.Ops)
	for _, edge := range graph.edges {
		colorValue := edge.edge.Color
		if colorValue.A == 0 {
			colorValue = tokens.EdgeColor
		}
		points := graphEdgePolyline(edge.edge.edgeTypeValue(), geometry.mapPoint(edge.source), geometry.mapPoint(edge.target), 16)
		drawGraphPolyline(gtx, points, max(float32(gtx.Dp(unit.Dp(1)))*.7, .7), colorValue)
	}
	for _, node := range graph.nodes {
		first := geometry.mapPoint(node.position)
		last := geometry.mapPoint(Point{X: node.position.X + node.size.Width, Y: node.position.Y + node.size.Height})
		nodeRect := image.Rect(int(math.Round(float64(first.X))), int(math.Round(float64(first.Y))), int(math.Round(float64(last.X))), int(math.Round(float64(last.Y))))
		if nodeRect.Empty() {
			nodeRect.Max = nodeRect.Min.Add(image.Pt(2, 2))
		}
		colorValue := tokens.NodeBorder
		if selected[node.node.ID] {
			colorValue = tokens.SelectedNodeBorder
		}
		paint.FillShape(gtx.Ops, colorValue, clip.Rect(nodeRect).Op())
	}
	viewportRect := minimapViewportScreen(graph, viewport, canvas, graphPixelsPerDP(gtx))
	if !viewportRect.Empty() {
		drawGraphRoundedStroke(gtx, viewportRect, 1, max(gtx.Dp(unit.Dp(1)), 1), tokens.SelectedNodeBorder)
	}
	clipStack.Pop()
}

func graphBounds(graph resolvedGraph) graphRect {
	first := graph.nodes[0]
	result := graphRect{min: first.position, max: Point{X: first.position.X + first.size.Width, Y: first.position.Y + first.size.Height}}
	for _, node := range graph.nodes[1:] {
		result.min.X = min(result.min.X, node.position.X)
		result.min.Y = min(result.min.Y, node.position.Y)
		result.max.X = max(result.max.X, node.position.X+node.size.Width)
		result.max.Y = max(result.max.Y, node.position.Y+node.size.Height)
	}
	return result
}
