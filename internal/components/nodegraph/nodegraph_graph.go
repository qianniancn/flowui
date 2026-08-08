package nodegraph

import (
	"fmt"
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
)

type resolvedPort struct {
	port     Port
	point    Point
	output   bool
	position HandlePosition
}

type resolvedNode struct {
	node     Node
	position Point
	size     Size
	inputs   []resolvedPort
	outputs  []resolvedPort
}

type resolvedEdge struct {
	edge   Edge
	source Point
	target Point
}

// graphEdgeCullingPadding covers the horizontal overshoot of a Bezier curve
// when its source is to the right of its target. The curve's control distance
// is half the horizontal delta, and its maximum overshoot is below ten percent.
func graphEdgeCullingPadding(from, to Point) float32 {
	return max(float32(64), abs32(to.X-from.X)*.1)
}

type resolvedGraph struct {
	nodes    []resolvedNode
	byID     map[string]resolvedNode
	edges    []resolvedEdge
	edgeByID map[string]resolvedEdge
	spatial  graphSpatialIndex
}

type portHit struct {
	endpoint Endpoint
	point    Point
	output   bool
	node     Node
	port     Port
}

type reconnectHit struct {
	edge     Edge
	endpoint reconnectEndpoint
	fixed    portHit
}

type graphRect struct {
	min Point
	max Point
}

func resolveGraph(graph Graph) resolvedGraph {
	return resolveGraphMeasured(nil, layout.Context{}, graph)
}

func resolveGraphMeasured(ctx *frame.Context, gtx layout.Context, graph Graph, scope ...string) resolvedGraph {
	result := resolvedGraph{
		nodes:    make([]resolvedNode, 0, len(graph.Nodes)),
		byID:     make(map[string]resolvedNode, len(graph.Nodes)),
		edges:    make([]resolvedEdge, 0, len(graph.Edges)),
		edgeByID: make(map[string]resolvedEdge, len(graph.Edges)),
	}
	nodesByID := make(map[string]Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		validateNode(node)
		if _, exists := nodesByID[node.ID]; exists {
			panic(fmt.Sprintf("flowui: duplicate node graph node ID %q", node.ID))
		}
		nodesByID[node.ID] = node
	}
	for _, node := range graph.Nodes {
		if node.parentID != "" {
			if _, exists := nodesByID[node.parentID]; !exists {
				panic(fmt.Sprintf("flowui: node graph node %q references an unknown parent %q", node.ID, node.parentID))
			}
		}
	}
	visiting := make(map[string]bool, len(graph.Nodes))
	var resolveNode func(string) resolvedNode
	resolveNode = func(id string) resolvedNode {
		if resolved, exists := result.byID[id]; exists {
			return resolved
		}
		if visiting[id] {
			panic(fmt.Sprintf("flowui: node graph parent cycle includes %q", id))
		}
		visiting[id] = true
		node := nodesByID[id]
		position := node.Position
		if node.parentID != "" {
			parent := resolveNode(node.parentID)
			position.X += parent.position.X
			position.Y += parent.position.Y
		}
		resolved := resolvedNode{node: node, position: position, size: resolvedNodeSizeMeasured(ctx, gtx, node, scope...)}
		resolved.inputs, resolved.outputs = resolveNodePorts(resolved)
		result.nodes = append(result.nodes, resolved)
		result.byID[node.ID] = resolved
		visiting[id] = false
		return resolved
	}
	for _, node := range graph.Nodes {
		resolveNode(node.ID)
	}
	seenEdges := make(map[string]struct{}, len(graph.Edges))
	for _, edge := range graph.Edges {
		if edge.ID == "" {
			panic("flowui: empty node graph edge ID")
		}
		if _, exists := seenEdges[edge.ID]; exists {
			panic(fmt.Sprintf("flowui: duplicate node graph edge ID %q", edge.ID))
		}
		seenEdges[edge.ID] = struct{}{}
		source, sourceOK := result.byID[edge.Source.NodeID]
		target, targetOK := result.byID[edge.Target.NodeID]
		if !sourceOK || !targetOK {
			panic(fmt.Sprintf("flowui: edge %q references an unknown node", edge.ID))
		}
		sourceIndex := portIndex(source.node.OutputPorts, edge.Source.PortID)
		targetIndex := portIndex(target.node.InputPorts, edge.Target.PortID)
		if sourceIndex < 0 || targetIndex < 0 {
			panic(fmt.Sprintf("flowui: edge %q references an unknown port", edge.ID))
		}
		resolved := resolvedEdge{
			edge:   edge,
			source: source.portAnchor(true, source.node.OutputPorts[sourceIndex].ID),
			target: target.portAnchor(false, target.node.InputPorts[targetIndex].ID),
		}
		result.edges = append(result.edges, resolved)
		result.edgeByID[edge.ID] = resolved
	}
	result.buildSpatialIndex()
	return result
}

func validateNode(node Node) {
	if node.ID == "" {
		panic("flowui: empty node graph node ID")
	}
	if !finite(node.Position.X) || !finite(node.Position.Y) {
		panic(fmt.Sprintf("flowui: node graph node %q has an invalid position", node.ID))
	}
	if node.Size.Width < 0 || node.Size.Height < 0 || !finite(node.Size.Width) || !finite(node.Size.Height) {
		panic(fmt.Sprintf("flowui: node graph node %q has an invalid size", node.ID))
	}
	validatePorts(node.ID, "input", node.InputPorts)
	validatePorts(node.ID, "output", node.OutputPorts)
}

func validatePorts(nodeID, role string, ports []Port) {
	seen := make(map[string]struct{}, len(ports))
	for _, port := range ports {
		if port.ID == "" {
			panic(fmt.Sprintf("flowui: node graph node %q has an empty %s port ID", nodeID, role))
		}
		if _, exists := seen[port.ID]; exists {
			panic(fmt.Sprintf("flowui: node graph node %q has duplicate %s port ID %q", nodeID, role, port.ID))
		}
		seen[port.ID] = struct{}{}
	}
}

func resolvedNodeSize(node Node) Size {
	return resolvedNodeSizeMeasured(nil, layout.Context{}, node)
}

func resolvedNodeSizeMeasured(ctx *frame.Context, gtx layout.Context, node Node, scope ...string) Size {
	rows := max(len(node.InputPorts), len(node.OutputPorts), 1)
	minimumHeight := nodeHeaderHeight + nodeVerticalPad*2 + float32(rows)*nodePortRowHeight
	width := node.Size.Width
	if width <= 0 {
		width = defaultNodeWidth
	}
	height := node.Size.Height
	if height <= 0 {
		height = defaultNodeHeight
	}
	// Explicit dimensions do not depend on content measurement. Avoiding that
	// second layout pass is particularly important while panning custom nodes.
	if node.content != nil && ctx != nil && (node.Size.Width <= 0 || node.Size.Height <= 0) {
		var restoreGraphKey, restoreNodeKey func()
		if len(scope) > 0 && scope[0] != "" {
			restoreGraphKey = frame.PushKey(ctx, scope[0])
			restoreNodeKey = frame.PushKey(ctx, "node:"+node.ID)
		}
		measureGtx := gtx
		measureGtx.Constraints = layout.Constraints{Max: gtx.Constraints.Max}
		dimensions := frame.MeasureWidget(ctx, measureGtx, node.content)
		if restoreNodeKey != nil {
			restoreNodeKey()
			restoreGraphKey()
		}
		pixelsPerDP := graphPixelsPerDP(gtx)
		if pixelsPerDP > 0 {
			contentWidth := float32(dimensions.Size.X)/pixelsPerDP + nodeVerticalPad*2
			contentHeight := float32(dimensions.Size.Y)/pixelsPerDP + nodeVerticalPad*2
			if node.Size.Width <= 0 {
				width = max(width, contentWidth)
			}
			if node.Size.Height <= 0 {
				height = max(height, contentHeight)
			}
		}
	}
	return Size{
		Width:  max(width, minimumNodeWidth(node)),
		Height: max(height, minimumHeight),
	}
}

func minimumNodeWidth(node Node) float32 {
	top, bottom := 0, 0
	for _, port := range node.InputPorts {
		switch port.handlePosition(false) {
		case HandleTop:
			top++
		case HandleBottom:
			bottom++
		}
	}
	for _, port := range node.OutputPorts {
		switch port.handlePosition(true) {
		case HandleTop:
			top++
		case HandleBottom:
			bottom++
		}
	}
	count := max(top, bottom)
	if count == 0 {
		return 0
	}
	return float32(count+1) * nodePortRowHeight * 1.5
}

func resolveNodePorts(node resolvedNode) ([]resolvedPort, []resolvedPort) {
	counts := [4]int{}
	for _, port := range node.node.InputPorts {
		counts[port.handlePosition(false)]++
	}
	for _, port := range node.node.OutputPorts {
		counts[port.handlePosition(true)]++
	}
	indices := [4]int{}
	inputs := resolvePorts(node, node.node.InputPorts, false, counts, &indices)
	outputs := resolvePorts(node, node.node.OutputPorts, true, counts, &indices)
	return inputs, outputs
}

func resolvePorts(node resolvedNode, ports []Port, output bool, counts [4]int, indices *[4]int) []resolvedPort {
	resolved := make([]resolvedPort, len(ports))
	for index, port := range ports {
		position := port.handlePosition(output)
		sideIndex := indices[position]
		indices[position]++
		point := node.position
		switch position {
		case HandleLeft, HandleRight:
			point.Y += nodeHeaderHeight + nodeVerticalPad + float32(sideIndex)*nodePortRowHeight + nodePortRowHeight/2
			if position == HandleRight {
				point.X += node.size.Width
			}
		case HandleTop, HandleBottom:
			point.X += node.size.Width * float32(sideIndex+1) / float32(counts[position]+1)
			if position == HandleBottom {
				point.Y += node.size.Height
			}
		}
		resolved[index] = resolvedPort{port: port, point: point, output: output, position: position}
	}
	return resolved
}

func portIndex(ports []Port, id string) int {
	for index, port := range ports {
		if port.ID == id {
			return index
		}
	}
	return -1
}

func (node resolvedNode) portAnchor(output bool, id string) Point {
	ports := node.inputs
	if output {
		ports = node.outputs
	}
	for _, port := range ports {
		if port.port.ID == id {
			return port.point
		}
	}
	return node.position
}

func (node resolvedNode) port(output bool, id string) (resolvedPort, bool) {
	ports := node.inputs
	if output {
		ports = node.outputs
	}
	for _, port := range ports {
		if port.port.ID == id {
			return port, true
		}
	}
	return resolvedPort{}, false
}

func (g resolvedGraph) nodeAt(point Point) string {
	for _, index := range g.spatial.nodeCandidates(graphRect{min: point, max: point}) {
		node := g.nodes[index]
		if point.X >= node.position.X && point.X <= node.position.X+node.size.Width && point.Y >= node.position.Y && point.Y <= node.position.Y+node.size.Height {
			return node.node.ID
		}
	}
	return ""
}

func (g resolvedGraph) portAt(point Point, radius float32) (portHit, bool) {
	if radius <= 0 || !finite(radius) {
		return portHit{}, false
	}
	radiusSquared := radius * radius
	query := graphRect{min: Point{X: point.X - radius, Y: point.Y - radius}, max: Point{X: point.X + radius, Y: point.Y + radius}}
	for _, index := range g.spatial.nodeCandidates(query) {
		node := g.nodes[index]
		for _, resolvedPort := range node.outputs {
			if squaredDistance(point, resolvedPort.point) <= radiusSquared {
				return portHit{endpoint: Endpoint{NodeID: node.node.ID, PortID: resolvedPort.port.ID}, point: resolvedPort.point, output: true, node: node.node, port: resolvedPort.port}, true
			}
		}
		for _, resolvedPort := range node.inputs {
			if squaredDistance(point, resolvedPort.point) <= radiusSquared {
				return portHit{endpoint: Endpoint{NodeID: node.node.ID, PortID: resolvedPort.port.ID}, point: resolvedPort.point, node: node.node, port: resolvedPort.port}, true
			}
		}
	}
	return portHit{}, false
}

func (g resolvedGraph) nodesInRect(rect graphRect) []resolvedNode {
	indices := g.spatial.nodeCandidates(rect)
	result := make([]resolvedNode, len(indices))
	for index, candidate := range indices {
		result[index] = g.nodes[candidate]
	}
	return result
}

func (g resolvedGraph) portConnectable(hit portHit, defaultValue bool) bool {
	return hit.node.isConnectable(defaultValue) && hit.port.isConnectable(defaultValue)
}

func (g resolvedGraph) connectionAllowed(connection Connection, excludedEdgeID string) bool {
	source, sourceOK := g.byID[connection.Source.NodeID]
	target, targetOK := g.byID[connection.Target.NodeID]
	if !sourceOK || !targetOK {
		return false
	}
	sourcePort, sourceOK := source.port(true, connection.Source.PortID)
	targetPort, targetOK := target.port(false, connection.Target.PortID)
	if !sourceOK || !targetOK {
		return false
	}
	if sourcePort.port.dataType != "" && targetPort.port.dataType != "" && sourcePort.port.dataType != targetPort.port.dataType {
		return false
	}
	if sourcePort.port.maxLinks > 0 && g.connectionCount(connection.Source, excludedEdgeID) >= sourcePort.port.maxLinks {
		return false
	}
	if targetPort.port.maxLinks > 0 && g.connectionCount(connection.Target, excludedEdgeID) >= targetPort.port.maxLinks {
		return false
	}
	return true
}

func (g resolvedGraph) connectionCount(endpoint Endpoint, excludedEdgeID string) int {
	count := 0
	for _, edge := range g.edges {
		if edge.edge.ID == excludedEdgeID {
			continue
		}
		if edge.edge.Source == endpoint || edge.edge.Target == endpoint {
			count++
		}
	}
	return count
}

func (g resolvedGraph) hasSelectedAncestor(node Node, selected map[string]bool) bool {
	for parentID := node.parentID; parentID != ""; {
		if selected[parentID] {
			return true
		}
		parent, exists := g.byID[parentID]
		if !exists {
			return false
		}
		parentID = parent.node.parentID
	}
	return false
}

func (g resolvedGraph) dragNode(node resolvedNode) dragNode {
	result := dragNode{
		id:       node.node.ID,
		position: node.node.Position,
		minimum:  Point{X: -math.MaxFloat32, Y: -math.MaxFloat32},
		maximum:  Point{X: math.MaxFloat32, Y: math.MaxFloat32},
	}
	if !node.node.isConstrainedToParent() || node.node.parentID == "" {
		return result
	}
	parent, exists := g.byID[node.node.parentID]
	if !exists {
		return result
	}
	result.minimum = Point{}
	result.maximum = Point{
		X: max(parent.size.Width-node.size.Width, 0),
		Y: max(parent.size.Height-node.size.Height, 0),
	}
	return result
}

func graphBox(first, second Point) graphRect {
	return graphRect{
		min: Point{X: min(first.X, second.X), Y: min(first.Y, second.Y)},
		max: Point{X: max(first.X, second.X), Y: max(first.Y, second.Y)},
	}
}

func selectionBoxMatches(box graphRect, node resolvedNode, mode SelectionBoxMode) bool {
	nodeBox := graphRect{
		min: node.position,
		max: Point{X: node.position.X + node.size.Width, Y: node.position.Y + node.size.Height},
	}
	if mode == SelectionBoxFull {
		return nodeBox.min.X >= box.min.X && nodeBox.min.Y >= box.min.Y && nodeBox.max.X <= box.max.X && nodeBox.max.Y <= box.max.Y
	}
	return nodeBox.max.X >= box.min.X && nodeBox.min.X <= box.max.X && nodeBox.max.Y >= box.min.Y && nodeBox.min.Y <= box.max.Y
}

func squaredDistance(first, second Point) float32 {
	dx, dy := first.X-second.X, first.Y-second.Y
	return dx*dx + dy*dy
}

func graphBezierDistanceSquared(point, from, to f32.Point, minimumDistance float32) float32 {
	const segments = 32
	previous := from
	distance := float32(math.MaxFloat32)
	for index := 1; index <= segments; index++ {
		current := graphBezierPoint(from, to, minimumDistance, float32(index)/segments)
		distance = min(distance, pointSegmentDistanceSquared(point, previous, current))
		previous = current
	}
	return distance
}

func graphEdgeDistanceSquared(edge Edge, point, from, to f32.Point, minimumDistance float32) float32 {
	if edge.edgeTypeValue() == EdgeBezier {
		return graphBezierDistanceSquared(point, from, to, minimumDistance)
	}
	points := graphEdgePolyline(edge.edgeTypeValue(), from, to, minimumDistance)
	distance := float32(math.MaxFloat32)
	for index := 1; index < len(points); index++ {
		distance = min(distance, pointSegmentDistanceSquared(point, points[index-1], points[index]))
	}
	return distance
}

func graphEdgePolyline(edgeType EdgeType, from, to f32.Point, minimumDistance float32) []f32.Point {
	switch edgeType {
	case EdgeStraight:
		return []f32.Point{from, to}
	case EdgeStep, EdgeSmoothStep:
		midX := (from.X + to.X) / 2
		return []f32.Point{from, f32.Pt(midX, from.Y), f32.Pt(midX, to.Y), to}
	default:
		points := make([]f32.Point, 33)
		for index := range points {
			points[index] = graphBezierPoint(from, to, minimumDistance, float32(index)/float32(len(points)-1))
		}
		return points
	}
}

func graphBezierPoint(from, to f32.Point, minimumDistance, progress float32) f32.Point {
	distance := max(abs32(to.X-from.X)*.5, minimumDistance)
	remaining := 1 - progress
	first := remaining * remaining * remaining
	second := 3 * remaining * remaining * progress
	third := 3 * remaining * progress * progress
	fourth := progress * progress * progress
	return f32.Pt(
		first*from.X+second*(from.X+distance)+third*(to.X-distance)+fourth*to.X,
		first*from.Y+second*from.Y+third*to.Y+fourth*to.Y,
	)
}

func pointSegmentDistanceSquared(point, from, to f32.Point) float32 {
	delta := to.Sub(from)
	lengthSquared := delta.X*delta.X + delta.Y*delta.Y
	if lengthSquared == 0 {
		return squaredDistanceScreen(point, from)
	}
	progress := ((point.X-from.X)*delta.X + (point.Y-from.Y)*delta.Y) / lengthSquared
	progress = min(max(progress, 0), 1)
	return squaredDistanceScreen(point, f32.Pt(from.X+delta.X*progress, from.Y+delta.Y*progress))
}

func squaredDistanceScreen(first, second f32.Point) float32 {
	dx, dy := first.X-second.X, first.Y-second.Y
	return dx*dx + dy*dy
}

func normalizeViewport(value Viewport, minimum, maximum float32) Viewport {
	if !finite(value.Origin.X) {
		value.Origin.X = 0
	}
	if !finite(value.Origin.Y) {
		value.Origin.Y = 0
	}
	if !finite(value.Zoom) || value.Zoom <= 0 {
		value.Zoom = 1
	}
	value.Zoom = min(max(value.Zoom, minimum), maximum)
	return value
}

func worldScale(viewport Viewport, pixelsPerDP float32) float32 {
	return max(viewport.Zoom*pixelsPerDP, 0.0001)
}

func graphPixelsPerDP(gtx layout.Context) float32 {
	value := gtx.Metric.PxPerDp
	if !finite(value) || value <= 0 {
		return 1
	}
	return value
}

func worldToScreen(viewport Viewport, point Point, pixelsPerDP float32) f32.Point {
	scale := worldScale(viewport, pixelsPerDP)
	return f32.Pt((point.X-viewport.Origin.X)*scale, (point.Y-viewport.Origin.Y)*scale)
}

func screenToWorld(viewport Viewport, point f32.Point, pixelsPerDP float32) Point {
	scale := worldScale(viewport, pixelsPerDP)
	return Point{X: viewport.Origin.X + point.X/scale, Y: viewport.Origin.Y + point.Y/scale}
}

func zoomAt(viewport Viewport, zoom float32, screen f32.Point, pixelsPerDP float32) Viewport {
	world := screenToWorld(viewport, screen, pixelsPerDP)
	viewport.Zoom = zoom
	scale := worldScale(viewport, pixelsPerDP)
	viewport.Origin = Point{X: world.X - screen.X/scale, Y: world.Y - screen.Y/scale}
	return viewport
}

func fitGraphViewport(graph resolvedGraph, canvas image.Point, pixelsPerDP, padding, minimum, maximum float32) (Viewport, bool) {
	if len(graph.nodes) == 0 || canvas.X <= 0 || canvas.Y <= 0 || pixelsPerDP <= 0 {
		return Viewport{}, false
	}
	bounds := graphBounds(graph)
	minimumPoint, maximumPoint := bounds.min, bounds.max
	availableWidth := float32(canvas.X) * (1 - padding*2)
	availableHeight := float32(canvas.Y) * (1 - padding*2)
	width := maximumPoint.X - minimumPoint.X
	height := maximumPoint.Y - minimumPoint.Y
	if availableWidth <= 0 || availableHeight <= 0 || width <= 0 || height <= 0 {
		return Viewport{}, false
	}
	zoom := min(availableWidth/(width*pixelsPerDP), availableHeight/(height*pixelsPerDP))
	zoom = min(max(zoom, minimum), maximum)
	visibleWidth := float32(canvas.X) / (zoom * pixelsPerDP)
	visibleHeight := float32(canvas.Y) / (zoom * pixelsPerDP)
	center := Point{X: (minimumPoint.X + maximumPoint.X) / 2, Y: (minimumPoint.Y + maximumPoint.Y) / 2}
	return Viewport{Origin: Point{X: center.X - visibleWidth/2, Y: center.Y - visibleHeight/2}, Zoom: zoom}, true
}

func finite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func abs32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}
