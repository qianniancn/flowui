package ui

import (
	"image"

	"github.com/qianniancn/flowui/internal/components/nodegraph"
)

// NodeGraphWidget renders an interactive, node-based graph editor.
type NodeGraphWidget = nodegraph.Widget

// NodeGraphData is the declarative graph data owned by the application model.
type NodeGraphData = nodegraph.Graph

// NodeGraphFragment is a copied selection of nodes and internal edges.
type NodeGraphFragment = nodegraph.Fragment

// NodeGraphHistory stores application-owned snapshots for undo and redo.
type NodeGraphHistory = nodegraph.History

// NodeGraphNode describes one graph node.
type NodeGraphNode = nodegraph.Node

// NodeGraphPort describes one node input or output.
type NodeGraphPort = nodegraph.Port

// NodeGraphEdge describes one directed graph connection.
type NodeGraphEdge = nodegraph.Edge

// NodeGraphEndpoint identifies one port on one node.
type NodeGraphEndpoint = nodegraph.Endpoint

// NodeGraphPoint is a density-independent graph coordinate.
type NodeGraphPoint = nodegraph.Point

// NodeGraphSize is a density-independent graph size.
type NodeGraphSize = nodegraph.Size

// NodeGraphViewport identifies the graph origin and zoom factor.
type NodeGraphViewport = nodegraph.Viewport

// NodeGraphCanvasEvent describes an interaction on empty graph canvas space.
type NodeGraphCanvasEvent = nodegraph.CanvasEvent

// NodeGraphNodeEvent describes an interaction with a node.
type NodeGraphNodeEvent = nodegraph.NodeEvent

// NodeGraphEdgeEvent describes an interaction with an edge.
type NodeGraphEdgeEvent = nodegraph.EdgeEvent

// NodeGraphDropEvent describes externally transferred data dropped on a graph.
type NodeGraphDropEvent = nodegraph.DropEvent

// NodeGraphBounds is an axis-aligned rectangle in graph coordinates.
type NodeGraphBounds = nodegraph.Bounds

// NodeGraphSnapshot is a graph plus viewport value for application persistence.
type NodeGraphSnapshot = nodegraph.Snapshot

// NodeGraphPanelPosition selects a fixed canvas panel placement.
type NodeGraphPanelPosition = nodegraph.PanelPosition

// NodeGraphPanel describes fixed content rendered above the graph canvas.
type NodeGraphPanel = nodegraph.Panel

// NodeGraphViewportOverlay describes content anchored in graph coordinates.
type NodeGraphViewportOverlay = nodegraph.ViewportOverlay

// NodeGraphSelectionMode controls how many nodes may be selected by gestures.
type NodeGraphSelectionMode = nodegraph.SelectionMode

// NodeGraphNodeChangeKind identifies a requested node update.
type NodeGraphNodeChangeKind = nodegraph.NodeChangeKind

// NodeGraphNodeChange describes a controlled node position, selection, or removal request.
type NodeGraphNodeChange = nodegraph.NodeChange

// NodeGraphEdgeChangeKind identifies a requested edge update.
type NodeGraphEdgeChangeKind = nodegraph.EdgeChangeKind

// NodeGraphEdgeChange describes a controlled edge selection or removal request.
type NodeGraphEdgeChange = nodegraph.EdgeChange

// NodeGraphConnection is a requested connection from an output to an input port.
type NodeGraphConnection = nodegraph.Connection

// NodeGraphReconnectMode controls which endpoints of an edge may be reconnected.
type NodeGraphReconnectMode = nodegraph.ReconnectMode

// NodeGraphSelectionBoxMode controls how a selection box matches nodes.
type NodeGraphSelectionBoxMode = nodegraph.SelectionBoxMode

// NodeGraphGridPattern controls the visual form used by an enabled graph grid.
type NodeGraphGridPattern = nodegraph.GridPattern

// NodeGraphHandlePosition identifies the side of a node where a port is rendered.
type NodeGraphHandlePosition = nodegraph.HandlePosition

// NodeGraphResizeHandle selects one or more node resize handles.
type NodeGraphResizeHandle = nodegraph.ResizeHandle

// NodeGraphEdgeType controls edge geometry.
type NodeGraphEdgeType = nodegraph.EdgeType

// NodeGraphEdgeMarker identifies an optional edge endpoint marker.
type NodeGraphEdgeMarker = nodegraph.EdgeMarker

const (
	NodeGraphSelectionSingle   = nodegraph.SelectionSingle
	NodeGraphSelectionMultiple = nodegraph.SelectionMultiple
	NodeGraphSelectionNone     = nodegraph.SelectionNone

	NodeGraphNodeChangePosition  = nodegraph.NodeChangePosition
	NodeGraphNodeChangeSelection = nodegraph.NodeChangeSelection
	NodeGraphNodeChangeRemove    = nodegraph.NodeChangeRemove
	NodeGraphNodeChangeSize      = nodegraph.NodeChangeSize

	NodeGraphEdgeChangeSelection = nodegraph.EdgeChangeSelection
	NodeGraphEdgeChangeRemove    = nodegraph.EdgeChangeRemove

	NodeGraphSelectionBoxPartial = nodegraph.SelectionBoxPartial
	NodeGraphSelectionBoxFull    = nodegraph.SelectionBoxFull

	NodeGraphGridLines = nodegraph.GridLines
	NodeGraphGridDots  = nodegraph.GridDots

	NodeGraphHandleLeft   = nodegraph.HandleLeft
	NodeGraphHandleRight  = nodegraph.HandleRight
	NodeGraphHandleTop    = nodegraph.HandleTop
	NodeGraphHandleBottom = nodegraph.HandleBottom

	NodeGraphResizeHandleTop         = nodegraph.ResizeHandleTop
	NodeGraphResizeHandleRight       = nodegraph.ResizeHandleRight
	NodeGraphResizeHandleBottom      = nodegraph.ResizeHandleBottom
	NodeGraphResizeHandleLeft        = nodegraph.ResizeHandleLeft
	NodeGraphResizeHandleTopLeft     = nodegraph.ResizeHandleTopLeft
	NodeGraphResizeHandleTopRight    = nodegraph.ResizeHandleTopRight
	NodeGraphResizeHandleBottomRight = nodegraph.ResizeHandleBottomRight
	NodeGraphResizeHandleBottomLeft  = nodegraph.ResizeHandleBottomLeft
	NodeGraphResizeHandleAll         = nodegraph.ResizeHandleAll

	NodeGraphEdgeBezier     = nodegraph.EdgeBezier
	NodeGraphEdgeStraight   = nodegraph.EdgeStraight
	NodeGraphEdgeStep       = nodegraph.EdgeStep
	NodeGraphEdgeSmoothStep = nodegraph.EdgeSmoothStep
	NodeGraphMarkerNone     = nodegraph.MarkerNone
	NodeGraphMarkerArrow    = nodegraph.MarkerArrow

	NodeGraphReconnectBoth   = nodegraph.ReconnectBoth
	NodeGraphReconnectSource = nodegraph.ReconnectSource
	NodeGraphReconnectTarget = nodegraph.ReconnectTarget
	NodeGraphReconnectNone   = nodegraph.ReconnectNone

	NodeGraphPanelTopLeft      = nodegraph.PanelTopLeft
	NodeGraphPanelTopCenter    = nodegraph.PanelTopCenter
	NodeGraphPanelTopRight     = nodegraph.PanelTopRight
	NodeGraphPanelBottomLeft   = nodegraph.PanelBottomLeft
	NodeGraphPanelBottomCenter = nodegraph.PanelBottomCenter
	NodeGraphPanelBottomRight  = nodegraph.PanelBottomRight
)

// NodeGraph creates a node graph widget for graph.
func NodeGraph(key string, graph NodeGraphData) NodeGraphWidget {
	return nodegraph.New(key, graph)
}

// NewNodeGraphNode creates one node at position.
func NewNodeGraphNode(id, title string, position NodeGraphPoint) NodeGraphNode {
	return nodegraph.NewNode(id, title, position)
}

// NewNodeGraphPort creates one named input or output port.
func NewNodeGraphPort(id, label string) NodeGraphPort {
	return nodegraph.NewPort(id, label)
}

// NewNodeGraphEndpoint identifies a node port for an edge endpoint.
func NewNodeGraphEndpoint(nodeID, portID string) NodeGraphEndpoint {
	return nodegraph.NewEndpoint(nodeID, portID)
}

// NewNodeGraphEdge creates a directed edge from source to target.
func NewNodeGraphEdge(id string, source, target NodeGraphEndpoint) NodeGraphEdge {
	return nodegraph.NewEdge(id, source, target)
}

// ApplyNodeGraphChanges applies node position, size, selection, and removal
// requests to a copy of nodes. It is useful in an application's Update function.
func ApplyNodeGraphChanges(nodes []NodeGraphNode, changes []NodeGraphNodeChange) []NodeGraphNode {
	return nodegraph.ApplyNodeChanges(nodes, changes)
}

// ApplyNodeGraphEdgeChanges applies edge selection and removal requests to a
// copy of edges. It is useful in an application's Update function.
func ApplyNodeGraphEdgeChanges(edges []NodeGraphEdge, changes []NodeGraphEdgeChange) []NodeGraphEdge {
	return nodegraph.ApplyEdgeChanges(edges, changes)
}

// ReconnectNodeGraphEdge applies connection endpoints to a copy of the
// matching edge while preserving its application-owned ID.
func ReconnectNodeGraphEdge(old NodeGraphEdge, connection NodeGraphConnection, edges []NodeGraphEdge) []NodeGraphEdge {
	return nodegraph.ReconnectEdge(old, connection, edges)
}

// CopyNodeGraphSelection returns selected nodes, their descendants, and their
// internal edges for application-controlled copy and cut operations.
func CopyNodeGraphSelection(graph NodeGraphData, selected map[string]bool) NodeGraphFragment {
	return nodegraph.CopySelection(graph, selected)
}

// PasteNodeGraphFragment appends a copied selection with application-owned
// IDs. Offset applies to pasted roots while preserving group-local children.
func PasteNodeGraphFragment(graph NodeGraphData, fragment NodeGraphFragment, nodeID func(NodeGraphNode) string, edgeID func(NodeGraphEdge) string, offset NodeGraphPoint) NodeGraphData {
	return nodegraph.PasteFragment(graph, fragment, nodeID, edgeID, offset)
}

// NewNodeGraphHistory initializes a snapshot history for application-owned
// undo and redo handling.
func NewNodeGraphHistory(graph NodeGraphData) *NodeGraphHistory {
	return nodegraph.NewHistory(graph)
}

// NewNodeGraphSnapshot clones graph together with its viewport.
func NewNodeGraphSnapshot(graph NodeGraphData, viewport NodeGraphViewport) NodeGraphSnapshot {
	return nodegraph.NewSnapshot(graph, viewport)
}

// NodeGraphNodeByID returns a copy of the matching node.
func NodeGraphNodeByID(graph NodeGraphData, id string) (NodeGraphNode, bool) {
	return nodegraph.NodeByID(graph, id)
}

// NodeGraphEdgeByID returns a copy of the matching edge.
func NodeGraphEdgeByID(graph NodeGraphData, id string) (NodeGraphEdge, bool) {
	return nodegraph.EdgeByID(graph, id)
}

// NodeGraphNodeWorldPosition resolves a node's world coordinate.
func NodeGraphNodeWorldPosition(graph NodeGraphData, id string) (NodeGraphPoint, bool) {
	return nodegraph.NodeWorldPosition(graph, id)
}

// NodeGraphNodesBounds returns world-coordinate bounds for ids, or all nodes.
func NodeGraphNodesBounds(graph NodeGraphData, ids ...string) (NodeGraphBounds, bool) {
	return nodegraph.NodesBounds(graph, ids...)
}

// NodeGraphIntersectingNodes returns nodes that partially or fully intersect bounds.
func NodeGraphIntersectingNodes(graph NodeGraphData, bounds NodeGraphBounds, partial bool) []NodeGraphNode {
	return nodegraph.IntersectingNodes(graph, bounds, partial)
}

// NodeGraphNodeConnections returns all edges connected to nodeID.
func NodeGraphNodeConnections(graph NodeGraphData, nodeID string) []NodeGraphEdge {
	return nodegraph.NodeConnections(graph, nodeID)
}

// NodeGraphPortConnections returns source or target connections for endpoint.
func NodeGraphPortConnections(graph NodeGraphData, endpoint NodeGraphEndpoint, output bool) []NodeGraphEdge {
	return nodegraph.PortConnections(graph, endpoint, output)
}

// NodeGraphWorldToScreen converts graph coordinates to device pixels.
func NodeGraphWorldToScreen(viewport NodeGraphViewport, point NodeGraphPoint, pixelsPerDP float32) NodeGraphPoint {
	return nodegraph.WorldToScreen(viewport, point, pixelsPerDP)
}

// NodeGraphScreenToWorld converts device pixels to graph coordinates.
func NodeGraphScreenToWorld(viewport NodeGraphViewport, point NodeGraphPoint, pixelsPerDP float32) NodeGraphPoint {
	return nodegraph.ScreenToWorld(viewport, point, pixelsPerDP)
}

// NodeGraphFitViewport calculates a viewport that contains ids, or all nodes
// when ids is empty. Size is the canvas size in device pixels.
func NodeGraphFitViewport(graph NodeGraphData, ids []string, size image.Point, pixelsPerDP, padding, minZoom, maxZoom float32) (NodeGraphViewport, bool) {
	return nodegraph.FitViewport(graph, ids, size, pixelsPerDP, padding, minZoom, maxZoom)
}
