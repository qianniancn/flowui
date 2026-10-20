package nodegraph

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
)

type minimapGeometry struct {
	rect   image.Rectangle
	inner  image.Rectangle
	bounds graphRect
	scale  float32
}

// nodeGraphMinimapScratch owns frame-local request buffers for one widget.
// Provider renderers consume these slices synchronously during the call.
type nodeGraphMinimapScratch struct {
	nodes      []theme.NodeGraphMinimapNode
	edges      []theme.NodeGraphMinimapEdge
	edgePoints [][]f32.Point
}

func resolveMinimapGeometry(graph resolvedGraph, size image.Point, metrics theme.NodeGraphMetrics, pixelsPerDP float32) (minimapGeometry, bool) {
	if len(graph.nodes) == 0 {
		return minimapGeometry{}, false
	}
	rectWidth := max(int(float32(metrics.MinimapWidth)*pixelsPerDP), 1)
	rectHeight := max(int(float32(metrics.MinimapHeight)*pixelsPerDP), 1)
	margin := max(int(float32(metrics.MinimapMargin)*pixelsPerDP), 0)
	padding := max(int(float32(metrics.MinimapPadding)*pixelsPerDP), 0)
	rect := image.Rect(size.X-margin-rectWidth, size.Y-margin-rectHeight, size.X-margin, size.Y-margin)
	if rect.Min.X < 0 || rect.Min.Y < 0 {
		return minimapGeometry{}, false
	}
	bounds := graphBounds(graph)
	inner := rect.Inset(padding)
	graphWidth := max(bounds.max.X-bounds.min.X, 1)
	graphHeight := max(bounds.max.Y-bounds.min.Y, 1)
	return minimapGeometry{rect: rect, inner: inner, bounds: bounds, scale: min(float32(inner.Dx())/graphWidth, float32(inner.Dy())/graphHeight)}, true
}

func (m minimapGeometry) mapPoint(point Point) f32.Point {
	return f32.Pt(float32(m.inner.Min.X)+(point.X-m.bounds.min.X)*m.scale, float32(m.inner.Min.Y)+(point.Y-m.bounds.min.Y)*m.scale)
}

func (m minimapGeometry) worldPoint(position f32.Point) Point {
	return Point{X: m.bounds.min.X + (position.X-float32(m.inner.Min.X))/m.scale, Y: m.bounds.min.Y + (position.Y-float32(m.inner.Min.Y))/m.scale}
}

func minimapWorldAt(graph resolvedGraph, position f32.Point, size image.Point, metrics theme.NodeGraphMetrics, pixelsPerDP float32) (Point, bool) {
	geometry, ok := resolveMinimapGeometry(graph, size, metrics, pixelsPerDP)
	if !ok || !position.Round().In(geometry.inner) {
		return Point{}, false
	}
	return geometry.worldPoint(position), true
}

func minimapViewportScreen(graph resolvedGraph, viewport Viewport, size image.Point, metrics theme.NodeGraphMetrics, pixelsPerDP float32) image.Rectangle {
	geometry, ok := resolveMinimapGeometry(graph, size, metrics, pixelsPerDP)
	if !ok {
		return image.Rectangle{}
	}
	visibleWidth := float32(size.X) / worldScale(viewport, pixelsPerDP)
	visibleHeight := float32(size.Y) / worldScale(viewport, pixelsPerDP)
	first := geometry.mapPoint(viewport.Origin)
	last := geometry.mapPoint(Point{X: viewport.Origin.X + visibleWidth, Y: viewport.Origin.Y + visibleHeight})
	return image.Rect(int(math.Round(float64(first.X))), int(math.Round(float64(first.Y))), int(math.Round(float64(last.X))), int(math.Round(float64(last.Y))))
}

func drawNodeMinimap(gtx layout.Context, graph resolvedGraph, viewport Viewport, selected map[string]bool, visual theme.NodeGraphVisual, enabled bool, scratch *nodeGraphMinimapScratch) {
	if !enabled || len(graph.nodes) == 0 {
		return
	}
	if scratch == nil {
		scratch = &nodeGraphMinimapScratch{}
	}
	canvas := gtx.Constraints.Max
	geometry, ok := resolveMinimapGeometry(graph, canvas, visual.Metrics, graphPixelsPerDP(gtx))
	if !ok {
		return
	}
	rect := geometry.rect
	if cap(scratch.nodes) < len(graph.nodes) {
		scratch.nodes = make([]theme.NodeGraphMinimapNode, 0, len(graph.nodes))
	}
	nodes := scratch.nodes[:0]
	for _, node := range graph.nodes {
		first := geometry.mapPoint(node.position)
		last := geometry.mapPoint(Point{X: node.position.X + node.size.Width, Y: node.position.Y + node.size.Height})
		nodeRect := image.Rect(int(math.Round(float64(first.X))), int(math.Round(float64(first.Y))), int(math.Round(float64(last.X))), int(math.Round(float64(last.Y))))
		if nodeRect.Empty() {
			nodeRect.Max = nodeRect.Min.Add(image.Pt(2, 2))
		}
		nodes = append(nodes, theme.NodeGraphMinimapNode{Rect: nodeRect, Selected: selected[node.node.ID]})
	}
	if cap(scratch.edges) < len(graph.edges) {
		scratch.edges = make([]theme.NodeGraphMinimapEdge, 0, len(graph.edges))
	}
	if cap(scratch.edgePoints) < len(graph.edges) {
		scratch.edgePoints = make([][]f32.Point, len(graph.edges))
	}
	edges := scratch.edges[:0]
	for _, edge := range graph.edges {
		colorValue := edge.edge.Color
		if colorValue.A == 0 {
			colorValue = visual.Colors.EdgeColor
		}
		edgeIndex := len(edges)
		points := scratch.edgePoints[edgeIndex][:0]
		points = graphEdgePolylineInto(points, edge.edge.edgeTypeValue(), f32.Pt(edge.source.X, edge.source.Y), f32.Pt(edge.target.X, edge.target.Y), 16)
		scratch.edgePoints[edgeIndex] = points
		mapped := points[:0]
		for _, point := range points {
			mapped = append(mapped, geometry.mapPoint(Point{X: point.X, Y: point.Y}))
		}
		// Keep world and mapped points in the same reusable buffer: only mapped
		// points escape to the renderer request.
		scratch.edgePoints[edgeIndex] = mapped
		edges = append(edges, theme.NodeGraphMinimapEdge{Points: mapped, Color: colorValue})
	}
	scratch.nodes = nodes
	scratch.edges = edges
	viewportRect := minimapViewportScreen(graph, viewport, canvas, visual.Metrics, graphPixelsPerDP(gtx))
	visual.Minimap(gtx, theme.NodeGraphMinimapRequest{
		Rect: rect, Inner: geometry.inner, Nodes: nodes, Edges: edges, Viewport: viewportRect,
		Radius:        max(gtx.Dp(visual.Metrics.MinimapRadius), 1),
		BorderWidth:   max(gtx.Dp(visual.Metrics.MinimapBorderWidth), 1),
		EdgeWidth:     max(float32(gtx.Dp(visual.Metrics.MinimapEdgeWidth)), .7),
		ViewportWidth: max(float32(gtx.Dp(visual.Metrics.MinimapViewportWidth)), 1),
		Background:    visual.Colors.NodeBackground, Border: visual.Colors.CanvasBorder,
		Selected: visual.Colors.SelectedNodeBorder,
	})
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
