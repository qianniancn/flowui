package nodegraph

import (
	"image"

	"gioui.org/f32"
)

// Bounds is an axis-aligned rectangle in graph coordinates.
type Bounds struct {
	Min Point
	Max Point
}

// Width returns the non-negative horizontal extent.
func (b Bounds) Width() float32 {
	return max(b.Max.X-b.Min.X, 0)
}

// Height returns the non-negative vertical extent.
func (b Bounds) Height() float32 {
	return max(b.Max.Y-b.Min.Y, 0)
}

// Contains reports whether point lies inside bounds, including its edges.
func (b Bounds) Contains(point Point) bool {
	return point.X >= b.Min.X && point.X <= b.Max.X && point.Y >= b.Min.Y && point.Y <= b.Max.Y
}

// Snapshot is an application-owned graph and viewport value suitable for
// persistence after custom node content has been mapped to application data.
type Snapshot struct {
	Graph    Graph
	Viewport Viewport
}

// NewSnapshot clones graph and combines it with viewport for persistence.
func NewSnapshot(graph Graph, viewport Viewport) Snapshot {
	return Snapshot{Graph: cloneGraph(graph), Viewport: viewport}
}

// NodeByID returns a copy of the node with id.
func NodeByID(graph Graph, id string) (Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return cloneNode(node), true
		}
	}
	return Node{}, false
}

// EdgeByID returns a copy of the edge with id.
func EdgeByID(graph Graph, id string) (Edge, bool) {
	for _, edge := range graph.Edges {
		if edge.ID == id {
			return edge, true
		}
	}
	return Edge{}, false
}

// NodeWorldPosition returns a node's resolved position, including all parents.
func NodeWorldPosition(graph Graph, id string) (Point, bool) {
	resolved := resolveGraph(graph)
	node, ok := resolved.byID[id]
	return node.position, ok
}

// NodesBounds returns the resolved world-coordinate bounds for ids. With no
// ids it uses every node in graph.
func NodesBounds(graph Graph, ids ...string) (Bounds, bool) {
	resolved := resolveGraph(graph)
	include := make(map[string]bool, len(ids))
	for _, id := range ids {
		include[id] = true
	}
	var result Bounds
	found := false
	for _, node := range resolved.nodes {
		if len(include) > 0 && !include[node.node.ID] {
			continue
		}
		bounds := Bounds{Min: node.position, Max: Point{X: node.position.X + node.size.Width, Y: node.position.Y + node.size.Height}}
		if !found {
			result, found = bounds, true
			continue
		}
		result.Min.X = min(result.Min.X, bounds.Min.X)
		result.Min.Y = min(result.Min.Y, bounds.Min.Y)
		result.Max.X = max(result.Max.X, bounds.Max.X)
		result.Max.Y = max(result.Max.Y, bounds.Max.Y)
	}
	return result, found
}

// IntersectingNodes returns nodes intersecting bounds. Set partial to false to
// require complete containment. Returned nodes retain their local positions.
func IntersectingNodes(graph Graph, bounds Bounds, partial bool) []Node {
	resolved := resolveGraph(graph)
	mode := SelectionBoxFull
	if partial {
		mode = SelectionBoxPartial
	}
	box := graphRect{min: bounds.Min, max: bounds.Max}
	result := make([]Node, 0)
	for _, node := range resolved.nodes {
		if selectionBoxMatches(box, node, mode) {
			result = append(result, cloneNode(node.node))
		}
	}
	return result
}

// NodeConnections returns every edge connected to nodeID.
func NodeConnections(graph Graph, nodeID string) []Edge {
	result := make([]Edge, 0)
	for _, edge := range graph.Edges {
		if edge.Source.NodeID == nodeID || edge.Target.NodeID == nodeID {
			result = append(result, edge)
		}
	}
	return result
}

// PortConnections returns edges connected to endpoint. output selects whether
// endpoint is evaluated as an edge source or target.
func PortConnections(graph Graph, endpoint Endpoint, output bool) []Edge {
	result := make([]Edge, 0)
	for _, edge := range graph.Edges {
		candidate := edge.Target
		if output {
			candidate = edge.Source
		}
		if candidate == endpoint {
			result = append(result, edge)
		}
	}
	return result
}

// WorldToScreen converts a graph point to device pixels for viewport.
func WorldToScreen(viewport Viewport, point Point, pixelsPerDP float32) Point {
	screen := worldToScreen(viewport, point, pixelsPerDP)
	return Point{X: screen.X, Y: screen.Y}
}

// ScreenToWorld converts a device-pixel point to graph coordinates for viewport.
func ScreenToWorld(viewport Viewport, point Point, pixelsPerDP float32) Point {
	return screenToWorld(viewport, pointF32(point), pixelsPerDP)
}

// FitViewport calculates a viewport that contains ids, or every graph node
// when ids is empty. Size is the available canvas size in device pixels.
func FitViewport(graph Graph, ids []string, size image.Point, pixelsPerDP, padding, minZoom, maxZoom float32) (Viewport, bool) {
	if len(ids) == 0 {
		return fitGraphViewport(resolveGraph(graph), size, pixelsPerDP, padding, minZoom, maxZoom)
	}
	bounds, ok := NodesBounds(graph, ids...)
	if !ok {
		return Viewport{}, false
	}
	proxy := resolvedGraph{nodes: []resolvedNode{{position: bounds.Min, size: Size{Width: bounds.Width(), Height: bounds.Height()}}}}
	return fitGraphViewport(proxy, size, pixelsPerDP, padding, minZoom, maxZoom)
}

func pointF32(point Point) f32.Point {
	return f32.Pt(point.X, point.Y)
}
