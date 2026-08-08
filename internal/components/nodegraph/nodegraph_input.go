package nodegraph

import (
	"image"
	"io"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
)

type graphState struct {
	viewport            Viewport
	ready               bool
	fitViewActive       bool
	gesture             graphGesture
	pointerID           pointer.ID
	anchor              f32.Point
	origin              Point
	dragNodes           []dragNode
	dragStarted         bool
	dragMoved           bool
	resizeNode          resizeNode
	resizeMoved         bool
	selectionStart      f32.Point
	selectionCurrent    f32.Point
	selectionBase       map[string]bool
	selectionCurrentSet map[string]bool
	selectionAdditive   bool
	connectionSource    portHit
	connectionTarget    portHit
	connectionCurrent   f32.Point
	connectionMode      connectionMode
	connectionEdge      Edge
	hoveredNode         string
	hoveredPort         bool
	hoveredEdge         string
	hoveredResize       ResizeHandle
	hoveredReconnect    bool
	focusedNode         string
	focusVisible        bool
	minimapOffset       Point
	pointerPosition     f32.Point
	pointerPositionSet  bool
	press               graphPress
	lastActivation      graphActivation
	dropTag             byte
	clipboard           Fragment
	pasteCount          int
}

type graphGesture uint8

const (
	gestureNone graphGesture = iota
	gesturePan
	gestureNodeDrag
	gestureSelectionBox
	gestureConnect
	gestureMinimap
	gestureNodeResize
)

type graphTargetKind uint8

const (
	graphTargetNone graphTargetKind = iota
	graphTargetCanvas
	graphTargetNode
	graphTargetEdge
)

type graphPress struct {
	target   graphTargetKind
	id       string
	position f32.Point
	moved    bool
}

type graphActivation struct {
	target   graphTargetKind
	id       string
	position f32.Point
	time     time.Duration
	valid    bool
}

const (
	nodeGraphClickSlop       = float32(3)
	nodeGraphDoubleClickSlop = float32(6)
	nodeGraphDoubleClickWait = 500 * time.Millisecond
	nodeGraphDropDataLimit   = int64(1 << 20)
)

type reconnectEndpoint uint8

const (
	reconnectSource reconnectEndpoint = iota
	reconnectTarget
)

type connectionMode uint8

const (
	connectionNew connectionMode = iota
	connectionReconnectSource
	connectionReconnectTarget
)

type dragNode struct {
	id       string
	position Point
	minimum  Point
	maximum  Point
}

type resizeNode struct {
	id       string
	position Point
	size     Size
	minimum  Size
	maximum  Size
	handle   ResizeHandle
}

var resizeHandleOrder = [...]ResizeHandle{
	ResizeHandleTopLeft,
	ResizeHandleTop,
	ResizeHandleTopRight,
	ResizeHandleRight,
	ResizeHandleBottomRight,
	ResizeHandleBottom,
	ResizeHandleBottomLeft,
	ResizeHandleLeft,
}

func (w Widget) updateGestures(ctx *frame.Context, gtx layout.Context, value *graphState, graph resolvedGraph, current Viewport, selected, selectedEdges map[string]bool, size image.Point, enabled bool) (Viewport, bool) {
	next := current
	changed := false
	pixelsPerDP := graphPixelsPerDP(gtx)
	filter := pointer.Filter{
		Target:  value,
		Kinds:   pointer.Enter | pointer.Leave | pointer.Move | pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel | pointer.Scroll,
		ScrollX: pointer.ScrollRange{Min: -100000, Max: 100000},
		ScrollY: pointer.ScrollRange{Min: -100000, Max: 100000},
	}
	for {
		input, ok := gtx.Event(filter)
		if !ok {
			break
		}
		eventValue, ok := input.(pointer.Event)
		if !ok {
			continue
		}
		if !enabled {
			value.resetGesture()
			value.clearHover()
			continue
		}
		if eventValue.Kind != pointer.Leave {
			value.pointerPosition = eventValue.Position
			value.pointerPositionSet = true
		}
		switch eventValue.Kind {
		case pointer.Enter, pointer.Move:
			if value.gesture == gestureNone {
				w.updateHover(value, graph, selected, next, pixelsPerDP, eventValue.Position, eventValue.Modifiers)
			}
		case pointer.Leave:
			if value.gesture == gestureNone {
				w.emitHoverTransitions(value.hoveredNode, value.hoveredEdge, "", "", screenToWorld(next, eventValue.Position, pixelsPerDP), eventValue.Modifiers)
				value.clearHover()
			}
		case pointer.Scroll:
			if eventValue.Scroll.Y == 0 && eventValue.Scroll.X == 0 {
				continue
			}
			if w.zoomOnScroll && eventValue.Scroll.Y != 0 {
				steps := min(max(-eventValue.Scroll.Y/40, -4), 4)
				zoom := min(max(next.Zoom*float32(math.Pow(1.1, float64(steps))), w.minZoom), w.maxZoom)
				if zoom != next.Zoom {
					next = zoomAt(next, zoom, eventValue.Position, pixelsPerDP)
					changed = true
				}
			} else if w.panOnScroll {
				scale := worldScale(next, pixelsPerDP)
				next.Origin.X += eventValue.Scroll.X / scale
				next.Origin.Y += eventValue.Scroll.Y / scale
				changed = true
			}
		case pointer.Press:
			value.press = graphPress{}
			if control, ok := w.controlAt(eventValue.Position, size); ok {
				if nextViewport, controlChanged := w.applyControl(control, graph, next, size, pixelsPerDP); controlChanged {
					next = nextViewport
					changed = true
				}
				continue
			}
			if w.minimapEnabled {
				if world, ok := minimapWorldAt(graph, eventValue.Position, size); ok {
					visibleWidth := float32(size.X) / worldScale(next, pixelsPerDP)
					visibleHeight := float32(size.Y) / worldScale(next, pixelsPerDP)
					value.gesture = gestureMinimap
					value.pointerID = eventValue.PointerID
					value.minimapOffset = Point{X: visibleWidth / 2, Y: visibleHeight / 2}
					viewportRect := minimapViewportScreen(graph, next, size, pixelsPerDP)
					if eventValue.Position.Round().In(viewportRect) {
						value.minimapOffset = Point{X: world.X - next.Origin.X, Y: world.Y - next.Origin.Y}
					}
					next.Origin = Point{X: world.X - value.minimapOffset.X, Y: world.Y - value.minimapOffset.Y}
					changed = true
					interact.GrabPointer(gtx, value, eventValue)
					continue
				}
			}
			if eventValue.Buttons.Contain(pointer.ButtonSecondary) {
				w.emitContextMenu(graph, next, pixelsPerDP, eventValue)
				continue
			}
			middle := w.panOnMiddle && eventValue.Buttons.Contain(pointer.ButtonTertiary)
			if middle {
				w.startPan(gtx, value, next, eventValue)
				continue
			}
			if !interact.IsPrimaryPointerPress(eventValue) {
				continue
			}
			frame.RequestFocusVisible(ctx, value, false)
			value.focusVisible = false
			world := screenToWorld(next, eventValue.Position, pixelsPerDP)
			if resize, ok := w.resizeAt(graph, selected, eventValue.Position, next, pixelsPerDP); ok {
				w.startNodeResize(gtx, value, resize, eventValue)
				continue
			}
			if reconnect, ok := w.reconnectAt(graph, world, 10/worldScale(next, pixelsPerDP)); ok {
				w.startReconnection(gtx, value, reconnect, eventValue)
				continue
			}
			if port, ok := graph.portAt(world, 10/worldScale(next, pixelsPerDP)); ok && port.output && graph.portConnectable(port, w.nodesConnectable) {
				w.startConnection(gtx, value, port, eventValue)
				continue
			}
			nodeID := graph.nodeAt(world)
			if nodeID == "" {
				value.focusedNode = ""
				if edgeID := w.edgeAt(graph, eventValue.Position, next, pixelsPerDP); edgeID != "" {
					value.press = graphPress{target: graphTargetEdge, id: edgeID, position: eventValue.Position}
					w.selectEdge(graph, selected, selectedEdges, edgeID, eventValue.Modifiers)
					continue
				}
				value.press = graphPress{target: graphTargetCanvas, position: eventValue.Position}
				if w.selectionOnDrag && w.selectionMode != SelectionNone {
					if !selectionModifier(eventValue.Modifiers) {
						w.emitEdgeSelectionChanges(graph, selectedEdges, nil)
					}
					w.startSelectionBox(gtx, value, selected, eventValue)
					continue
				}
				if w.selectionMode != SelectionNone && !selectionModifier(eventValue.Modifiers) {
					w.emitSelectionChanges(graph, selected, nil)
					w.emitEdgeSelectionChanges(graph, selectedEdges, nil)
				}
				w.startPan(gtx, value, next, eventValue)
				continue
			}
			value.press = graphPress{target: graphTargetNode, id: nodeID, position: eventValue.Position}
			value.focusedNode = nodeID
			if w.selectionMode != SelectionNone && graph.byID[nodeID].node.isSelectable(w.nodesSelectable) && !selectionModifier(eventValue.Modifiers) {
				w.emitEdgeSelectionChanges(graph, selectedEdges, nil)
			}
			w.startNodeDrag(gtx, value, graph, selected, nodeID, eventValue)
		case pointer.Drag:
			if eventValue.PointerID != value.pointerID {
				continue
			}
			value.updatePressMovement(eventValue.Position, nodeGraphClickSlop)
			switch value.gesture {
			case gesturePan:
				scale := worldScale(next, pixelsPerDP)
				if scale <= 0 {
					continue
				}
				next.Origin = Point{
					X: value.origin.X - (eventValue.Position.X-value.anchor.X)/scale,
					Y: value.origin.Y - (eventValue.Position.Y-value.anchor.Y)/scale,
				}
				changed = true
			case gestureNodeDrag:
				w.updateNodeDrag(value, eventValue.Position, next, pixelsPerDP)
			case gestureNodeResize:
				w.updateNodeResize(value, eventValue.Position, next, pixelsPerDP)
			case gestureSelectionBox:
				w.updateSelectionBox(value, graph, next, pixelsPerDP, eventValue.Position)
			case gestureConnect:
				w.updateConnection(value, graph, next, pixelsPerDP, eventValue.Position)
			case gestureMinimap:
				if world, ok := minimapWorldAt(graph, eventValue.Position, size); ok {
					next.Origin = Point{X: world.X - value.minimapOffset.X, Y: world.Y - value.minimapOffset.Y}
					next = normalizeViewport(next, w.minZoom, w.maxZoom)
					changed = true
				}
			}
		case pointer.Release, pointer.Cancel:
			if value.gesture == gestureNone {
				if eventValue.Kind == pointer.Release {
					w.finishPress(value, graph, next, pixelsPerDP, eventValue)
					w.updateHover(value, graph, selected, next, pixelsPerDP, eventValue.Position, eventValue.Modifiers)
				} else {
					value.press = graphPress{}
				}
				continue
			}
			if eventValue.PointerID != value.pointerID {
				continue
			}
			switch value.gesture {
			case gestureNodeDrag:
				if value.dragStarted && value.dragMoved {
					w.emitNodeDrag(value, eventValue.Position, next, pixelsPerDP, false)
				}
			case gestureNodeResize:
				if value.resizeMoved {
					w.emitNodeResize(value, eventValue.Position, next, pixelsPerDP, false)
				}
			case gestureSelectionBox:
				if value.dragStarted {
					w.updateSelectionBox(value, graph, next, pixelsPerDP, eventValue.Position)
				} else if !value.selectionAdditive {
					w.emitSelectionChanges(graph, value.selectionCurrentSet, nil)
				}
			case gestureConnect:
				w.completeConnection(value, graph, next, pixelsPerDP, eventValue.Position)
			}
			if eventValue.Kind == pointer.Release {
				w.finishPress(value, graph, next, pixelsPerDP, eventValue)
			} else {
				value.press = graphPress{}
			}
			value.resetGesture()
			if eventValue.Kind == pointer.Release {
				w.updateHover(value, graph, selected, next, pixelsPerDP, eventValue.Position, eventValue.Modifiers)
			} else {
				value.clearHover()
			}
		}
	}
	if !enabled {
		value.resetGesture()
		value.clearHover()
	}
	return normalizeViewport(next, w.minZoom, w.maxZoom), changed
}

func (w Widget) addInput(gtx layout.Context, value *graphState, graph resolvedGraph, enabled bool) {
	if !enabled {
		return
	}
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	nodeGraphCursor(w, value, graph).Add(gtx.Ops)
	event.Op(gtx.Ops, value)
	if w.acceptsDrops() {
		event.Op(gtx.Ops, &value.dropTag)
	}
	area.Pop()
}

func (w Widget) acceptsDrops() bool {
	return w.onDrop != nil && len(w.dropTypes) > 0
}

func (w Widget) updateDropEvents(gtx layout.Context, value *graphState, viewport Viewport, pixelsPerDP float32) {
	if !w.acceptsDrops() || !value.pointerPositionSet {
		return
	}
	for _, mime := range w.dropTypes {
		for {
			raw, ok := gtx.Event(transfer.TargetFilter{Target: &value.dropTag, Type: mime})
			if !ok {
				break
			}
			eventValue, ok := raw.(transfer.DataEvent)
			if !ok || eventValue.Open == nil {
				continue
			}
			reader := eventValue.Open()
			if reader == nil {
				continue
			}
			data, err := io.ReadAll(io.LimitReader(reader, nodeGraphDropDataLimit+1))
			_ = reader.Close()
			if err != nil || len(data) > int(nodeGraphDropDataLimit) {
				continue
			}
			w.onDrop(DropEvent{
				Position: screenToWorld(viewport, value.pointerPosition, pixelsPerDP),
				MIME:     eventValue.Type,
				Data:     append([]byte(nil), data...),
			})
		}
	}
}

func (w Widget) selectedNodeSet(graph resolvedGraph) map[string]bool {
	selected := make(map[string]bool, len(graph.nodes))
	if w.hasSelectedKeys {
		for _, key := range w.selectedKeys {
			if _, exists := graph.byID[key]; exists {
				selected[key] = true
			}
		}
		return selected
	}
	for _, node := range graph.nodes {
		if node.node.Selected {
			selected[node.node.ID] = true
		}
	}
	return selected
}

func (w Widget) selectedEdgeSet(graph resolvedGraph) map[string]bool {
	selected := make(map[string]bool, len(graph.edges))
	if w.hasSelectedEdges {
		for _, key := range w.selectedEdgeKeys {
			if _, exists := graph.edgeByID[key]; exists {
				selected[key] = true
			}
		}
		return selected
	}
	for _, edge := range graph.edges {
		if edge.edge.Selected {
			selected[edge.edge.ID] = true
		}
	}
	return selected
}

func (w Widget) startPan(gtx layout.Context, value *graphState, viewport Viewport, eventValue pointer.Event) {
	value.gesture = gesturePan
	value.pointerID = eventValue.PointerID
	value.anchor = eventValue.Position
	value.origin = viewport.Origin
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) startNodeDrag(gtx layout.Context, value *graphState, graph resolvedGraph, selected map[string]bool, nodeID string, eventValue pointer.Event) {
	node := graph.byID[nodeID]
	nextSelection := selected
	if w.selectionMode != SelectionNone && node.node.isSelectable(w.nodesSelectable) {
		nextSelection = updateSelection(selected, nodeID, w.selectionMode, selectionModifier(eventValue.Modifiers))
		w.emitSelectionChanges(graph, selected, nextSelection)
	}
	if !node.node.isDraggable(w.nodesDraggable) {
		return
	}
	dragged := make([]dragNode, 0, len(nextSelection))
	if nextSelection[nodeID] {
		for _, candidate := range graph.nodes {
			if nextSelection[candidate.node.ID] && candidate.node.isDraggable(w.nodesDraggable) && !graph.hasSelectedAncestor(candidate.node, nextSelection) {
				dragged = append(dragged, graph.dragNode(candidate))
			}
		}
	} else {
		dragged = append(dragged, graph.dragNode(node))
	}
	if len(dragged) == 0 {
		return
	}
	value.gesture = gestureNodeDrag
	value.pointerID = eventValue.PointerID
	value.anchor = eventValue.Position
	value.dragNodes = dragged
	value.dragStarted = w.dragThreshold == 0
	value.dragMoved = false
	value.selectionCurrentSet = nextSelection
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) resizeAt(graph resolvedGraph, selected map[string]bool, position f32.Point, viewport Viewport, pixelsPerDP float32) (resizeNode, bool) {
	for index := len(graph.nodes) - 1; index >= 0; index-- {
		node := graph.nodes[index]
		if !selected[node.node.ID] || !node.node.isResizable(w.nodesResizable) {
			continue
		}
		handle := ResizeHandle(0)
		for _, candidate := range resizeHandleOrder {
			if node.node.resizeHandleSet(w.resizeHandles)&candidate == 0 {
				continue
			}
			anchor := worldToScreen(viewport, resizeHandlePoint(node, candidate), pixelsPerDP)
			if abs32(position.X-anchor.X) <= 10 && abs32(position.Y-anchor.Y) <= 10 {
				handle = candidate
				break
			}
		}
		if handle == 0 {
			continue
		}
		rows := max(len(node.node.InputPorts), len(node.node.OutputPorts), 1)
		minimum := Size{
			Width:  max(float32(80), node.node.minSize.Width),
			Height: max(nodeHeaderHeight+nodeVerticalPad*2+float32(rows)*nodePortRowHeight, max(float32(48), node.node.minSize.Height)),
		}
		maximum := node.node.maxSize
		return resizeNode{id: node.node.ID, position: node.node.Position, size: node.size, minimum: minimum, maximum: maximum, handle: handle}, true
	}
	return resizeNode{}, false
}

func (w Widget) startNodeResize(gtx layout.Context, value *graphState, resize resizeNode, eventValue pointer.Event) {
	value.gesture = gestureNodeResize
	value.pointerID = eventValue.PointerID
	value.anchor = eventValue.Position
	value.resizeNode = resize
	value.resizeMoved = false
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) updateNodeResize(value *graphState, position f32.Point, viewport Viewport, pixelsPerDP float32) {
	w.emitNodeResize(value, position, viewport, pixelsPerDP, true)
	value.resizeMoved = true
}

func (w Widget) emitNodeResize(value *graphState, position f32.Point, viewport Viewport, pixelsPerDP float32, resizing bool) {
	resize := value.resizeNode
	if resize.id == "" {
		return
	}
	scale := worldScale(viewport, pixelsPerDP)
	delta := Point{X: (position.X - value.anchor.X) / scale, Y: (position.Y - value.anchor.Y) / scale}
	nextPosition, nextSize := resizedNodeBounds(resize, delta)
	changes := make([]NodeChange, 0, 2)
	if nextPosition != resize.position {
		changes = append(changes, NodeChange{ID: resize.id, Kind: NodeChangePosition, Position: nextPosition, Resizing: resizing})
	}
	changes = append(changes, NodeChange{ID: resize.id, Kind: NodeChangeSize, Size: nextSize, Resizing: resizing})
	w.emitNodesChange(changes)
}

func resizeHandlePoint(node resolvedNode, handle ResizeHandle) Point {
	x := node.position.X + node.size.Width/2
	y := node.position.Y + node.size.Height/2
	if handle == ResizeHandleTopLeft || handle == ResizeHandleLeft || handle == ResizeHandleBottomLeft {
		x = node.position.X
	} else if handle == ResizeHandleTopRight || handle == ResizeHandleRight || handle == ResizeHandleBottomRight {
		x = node.position.X + node.size.Width
	}
	if handle == ResizeHandleTopLeft || handle == ResizeHandleTop || handle == ResizeHandleTopRight {
		y = node.position.Y
	} else if handle == ResizeHandleBottomLeft || handle == ResizeHandleBottom || handle == ResizeHandleBottomRight {
		y = node.position.Y + node.size.Height
	}
	return Point{X: x, Y: y}
}

func resizeHandleScreenPoint(rect image.Rectangle, handle ResizeHandle) image.Point {
	x := (rect.Min.X + rect.Max.X) / 2
	y := (rect.Min.Y + rect.Max.Y) / 2
	if handle == ResizeHandleTopLeft || handle == ResizeHandleLeft || handle == ResizeHandleBottomLeft {
		x = rect.Min.X
	} else if handle == ResizeHandleTopRight || handle == ResizeHandleRight || handle == ResizeHandleBottomRight {
		x = rect.Max.X
	}
	if handle == ResizeHandleTopLeft || handle == ResizeHandleTop || handle == ResizeHandleTopRight {
		y = rect.Min.Y
	} else if handle == ResizeHandleBottomLeft || handle == ResizeHandleBottom || handle == ResizeHandleBottomRight {
		y = rect.Max.Y
	}
	return image.Pt(x, y)
}

func resizedNodeBounds(resize resizeNode, delta Point) (Point, Size) {
	nextPosition, nextSize := resize.position, resize.size
	left := resize.handle == ResizeHandleLeft || resize.handle == ResizeHandleTopLeft || resize.handle == ResizeHandleBottomLeft
	right := resize.handle == ResizeHandleRight || resize.handle == ResizeHandleTopRight || resize.handle == ResizeHandleBottomRight
	top := resize.handle == ResizeHandleTop || resize.handle == ResizeHandleTopLeft || resize.handle == ResizeHandleTopRight
	bottom := resize.handle == ResizeHandleBottom || resize.handle == ResizeHandleBottomLeft || resize.handle == ResizeHandleBottomRight
	if left {
		nextSize.Width = resize.size.Width - delta.X
	} else if right {
		nextSize.Width = resize.size.Width + delta.X
	}
	if top {
		nextSize.Height = resize.size.Height - delta.Y
	} else if bottom {
		nextSize.Height = resize.size.Height + delta.Y
	}
	nextSize.Width = max(nextSize.Width, resize.minimum.Width)
	nextSize.Height = max(nextSize.Height, resize.minimum.Height)
	if resize.maximum.Width > 0 {
		nextSize.Width = min(nextSize.Width, resize.maximum.Width)
	}
	if resize.maximum.Height > 0 {
		nextSize.Height = min(nextSize.Height, resize.maximum.Height)
	}
	if left {
		nextPosition.X += resize.size.Width - nextSize.Width
	}
	if top {
		nextPosition.Y += resize.size.Height - nextSize.Height
	}
	return nextPosition, nextSize
}

func (w Widget) updateNodeDrag(value *graphState, position f32.Point, viewport Viewport, pixelsPerDP float32) {
	if !value.dragStarted {
		delta := position.Sub(value.anchor)
		if delta.X*delta.X+delta.Y*delta.Y < w.dragThreshold*w.dragThreshold {
			return
		}
		value.dragStarted = true
	}
	changes := w.nodeDragChanges(value, position, viewport, pixelsPerDP, true)
	if len(changes) == 0 {
		return
	}
	value.dragMoved = true
	w.emitNodesChange(changes)
}

func (w Widget) emitNodeDrag(value *graphState, position f32.Point, viewport Viewport, pixelsPerDP float32, dragging bool) {
	w.emitNodesChange(w.nodeDragChanges(value, position, viewport, pixelsPerDP, dragging))
}

func (w Widget) nodeDragChanges(value *graphState, position f32.Point, viewport Viewport, pixelsPerDP float32, dragging bool) []NodeChange {
	if len(value.dragNodes) == 0 {
		return nil
	}
	scale := worldScale(viewport, pixelsPerDP)
	delta := Point{X: (position.X - value.anchor.X) / scale, Y: (position.Y - value.anchor.Y) / scale}
	if w.snapToGrid {
		delta = snappedDragDelta(value.dragNodes[0].position, delta, w.snapGrid)
	}
	changes := make([]NodeChange, len(value.dragNodes))
	for index, node := range value.dragNodes {
		next := Point{X: node.position.X + delta.X, Y: node.position.Y + delta.Y}
		next.X = min(max(next.X, node.minimum.X), node.maximum.X)
		next.Y = min(max(next.Y, node.minimum.Y), node.maximum.Y)
		changes[index] = NodeChange{ID: node.id, Kind: NodeChangePosition, Position: next, Dragging: dragging}
	}
	return changes
}

func (w Widget) emitSelectionChanges(graph resolvedGraph, previous, next map[string]bool) {
	if selectionSetsEqual(previous, next) {
		return
	}
	changes := make([]NodeChange, 0, len(graph.nodes))
	for _, node := range graph.nodes {
		before, after := previous[node.node.ID], next[node.node.ID]
		if before != after {
			changes = append(changes, NodeChange{ID: node.node.ID, Kind: NodeChangeSelection, Selected: after})
		}
	}
	w.emitNodesChange(changes)
}

func (w Widget) emitNodesChange(changes []NodeChange) {
	if len(changes) > 0 && w.onNodesChange != nil {
		w.onNodesChange(changes)
	}
}

func (w Widget) selectEdge(graph resolvedGraph, selectedNodes, selectedEdges map[string]bool, edgeID string, modifiers key.Modifiers) {
	edge := graph.edgeByID[edgeID]
	if w.selectionMode == SelectionNone || !edge.edge.isSelectable(w.edgesSelectable) {
		return
	}
	if !selectionModifier(modifiers) {
		w.emitSelectionChanges(graph, selectedNodes, nil)
	}
	next := updateSelection(selectedEdges, edgeID, w.selectionMode, selectionModifier(modifiers))
	w.emitEdgeSelectionChanges(graph, selectedEdges, next)
}

func (w Widget) emitEdgeSelectionChanges(graph resolvedGraph, previous, next map[string]bool) {
	if selectionSetsEqual(previous, next) {
		return
	}
	changes := make([]EdgeChange, 0, len(graph.edges))
	for _, edge := range graph.edges {
		before, after := previous[edge.edge.ID], next[edge.edge.ID]
		if before != after {
			changes = append(changes, EdgeChange{ID: edge.edge.ID, Kind: EdgeChangeSelection, Selected: after})
		}
	}
	w.emitEdgesChange(changes)
}

func (w Widget) emitEdgesChange(changes []EdgeChange) {
	if len(changes) > 0 && w.onEdgesChange != nil {
		w.onEdgesChange(changes)
	}
}

func (w Widget) updateKeyboard(gtx layout.Context, value *graphState, graph resolvedGraph, selected, selectedEdges map[string]bool, enabled bool) {
	if !enabled {
		return
	}
	for {
		_, ok := gtx.Event(key.FocusFilter{Target: value})
		if !ok {
			break
		}
	}
	for _, name := range []key.Name{key.Name("A"), key.Name("C"), key.Name("X"), key.Name("V"), key.Name("Y"), key.Name("Z")} {
		for {
			eventValue, ok := gtx.Event(key.Filter{Focus: value, Name: name, Required: key.ModShortcut, Optional: key.ModShift})
			if !ok {
				break
			}
			keyEvent, ok := eventValue.(key.Event)
			if !ok || keyEvent.State != key.Press || !keyEvent.Modifiers.Contain(key.ModShortcut) {
				continue
			}
			switch keyEvent.Name {
			case key.Name("A"):
				next := make(map[string]bool, len(graph.nodes))
				for _, node := range graph.nodes {
					if node.node.isSelectable(w.nodesSelectable) {
						next[node.node.ID] = true
					}
				}
				w.emitSelectionChanges(graph, selected, next)
				edgeNext := make(map[string]bool, len(graph.edges))
				for _, edge := range graph.edges {
					if edge.edge.isSelectable(w.edgesSelectable) {
						edgeNext[edge.edge.ID] = true
					}
				}
				w.emitEdgeSelectionChanges(graph, selectedEdges, edgeNext)
			case key.Name("C"), key.Name("X"):
				fragment := CopySelection(w.graph, selected)
				if len(fragment.Nodes) == 0 {
					continue
				}
				value.clipboard = fragment
				value.pasteCount = 0
				if w.onCopy != nil {
					w.onCopy(fragment)
				}
				if keyEvent.Name == key.Name("X") {
					if w.onCut != nil {
						w.onCut(fragment)
					}
					w.deleteSelection(graph, selected, selectedEdges)
				}
			case key.Name("V"):
				if len(value.clipboard.Nodes) == 0 || w.onPaste == nil {
					continue
				}
				value.pasteCount++
				offset := float32(value.pasteCount * 24)
				w.onPaste(value.clipboard, Point{X: offset, Y: offset})
			case key.Name("Y"):
				if w.onRedo != nil {
					w.onRedo()
				}
			case key.Name("Z"):
				if keyEvent.Modifiers.Contain(key.ModShift) {
					if w.onRedo != nil {
						w.onRedo()
					}
				} else if w.onUndo != nil {
					w.onUndo()
				}
			}
		}
	}
	for _, name := range []key.Name{key.NameDeleteBackward, key.NameDeleteForward} {
		for {
			eventValue, ok := gtx.Event(key.Filter{Focus: value, Name: name})
			if !ok {
				break
			}
			keyEvent, ok := eventValue.(key.Event)
			if ok && keyEvent.State == key.Press {
				w.deleteSelection(graph, selected, selectedEdges)
			}
		}
	}
	for _, navigation := range []struct {
		name key.Name
		dx   float32
		dy   float32
	}{
		{name: key.NameLeftArrow, dx: -1},
		{name: key.NameRightArrow, dx: 1},
		{name: key.NameUpArrow, dy: -1},
		{name: key.NameDownArrow, dy: 1},
	} {
		for {
			eventValue, ok := gtx.Event(key.Filter{Focus: value, Name: navigation.name})
			if !ok {
				break
			}
			keyEvent, ok := eventValue.(key.Event)
			if !ok || keyEvent.State != key.Press {
				continue
			}
			value.focusVisible = true
			value.focusedNode = nextFocusedNode(graph, value.focusedNode, navigation.dx, navigation.dy)
		}
	}
	for _, navigation := range []struct {
		name  key.Name
		first bool
	}{
		{name: key.NameHome, first: true},
		{name: key.NameEnd},
	} {
		for {
			eventValue, ok := gtx.Event(key.Filter{Focus: value, Name: navigation.name})
			if !ok {
				break
			}
			keyEvent, ok := eventValue.(key.Event)
			if !ok || keyEvent.State != key.Press || len(graph.nodes) == 0 {
				continue
			}
			value.focusVisible = true
			if navigation.first {
				value.focusedNode = graph.nodes[0].node.ID
			} else {
				value.focusedNode = graph.nodes[len(graph.nodes)-1].node.ID
			}
		}
	}
	for _, name := range []key.Name{key.NameEnter, key.NameSpace} {
		for {
			eventValue, ok := gtx.Event(key.Filter{Focus: value, Name: name})
			if !ok {
				break
			}
			keyEvent, ok := eventValue.(key.Event)
			if !ok || keyEvent.State != key.Press || value.focusedNode == "" {
				continue
			}
			if node, exists := graph.byID[value.focusedNode]; exists && node.node.isSelectable(w.nodesSelectable) && w.selectionMode != SelectionNone {
				next := updateSelection(selected, value.focusedNode, w.selectionMode, false)
				w.emitSelectionChanges(graph, selected, next)
			}
		}
	}
}

func nextFocusedNode(graph resolvedGraph, focused string, dx, dy float32) string {
	if len(graph.nodes) == 0 {
		return ""
	}
	if focused == "" {
		return graph.nodes[0].node.ID
	}
	current, ok := graph.byID[focused]
	if !ok {
		return graph.nodes[0].node.ID
	}
	center := Point{X: current.position.X + current.size.Width/2, Y: current.position.Y + current.size.Height/2}
	bestID := ""
	bestPrimary, bestSecondary := float32(1e30), float32(1e30)
	for _, candidate := range graph.nodes {
		if candidate.node.ID == focused {
			continue
		}
		candidateCenter := Point{X: candidate.position.X + candidate.size.Width/2, Y: candidate.position.Y + candidate.size.Height/2}
		primary := (candidateCenter.X-center.X)*dx + (candidateCenter.Y-center.Y)*dy
		if primary <= 0 {
			continue
		}
		secondary := abs32((candidateCenter.X-center.X)*dy - (candidateCenter.Y-center.Y)*dx)
		if primary < bestPrimary || (primary == bestPrimary && secondary < bestSecondary) {
			bestID, bestPrimary, bestSecondary = candidate.node.ID, primary, secondary
		}
	}
	if bestID == "" {
		return focused
	}
	return bestID
}

func (w Widget) deleteSelection(graph resolvedGraph, selected, selectedEdges map[string]bool) {
	removedNodes := make(map[string]bool)
	nodeChanges := make([]NodeChange, 0, len(selected))
	for _, node := range graph.nodes {
		if selected[node.node.ID] && node.node.isDeletable(w.nodesDeletable) {
			removedNodes[node.node.ID] = true
			nodeChanges = append(nodeChanges, NodeChange{ID: node.node.ID, Kind: NodeChangeRemove})
		}
	}
	edgeChanges := make([]EdgeChange, 0, len(selectedEdges)+len(removedNodes))
	for _, edge := range graph.edges {
		connectedToRemovedNode := removedNodes[edge.edge.Source.NodeID] || removedNodes[edge.edge.Target.NodeID]
		if connectedToRemovedNode || (selectedEdges[edge.edge.ID] && edge.edge.isDeletable(w.edgesDeletable)) {
			edgeChanges = append(edgeChanges, EdgeChange{ID: edge.edge.ID, Kind: EdgeChangeRemove})
		}
	}
	w.emitNodesChange(nodeChanges)
	w.emitEdgesChange(edgeChanges)
}

func (w Widget) startSelectionBox(gtx layout.Context, value *graphState, selected map[string]bool, eventValue pointer.Event) {
	value.gesture = gestureSelectionBox
	value.pointerID = eventValue.PointerID
	value.anchor = eventValue.Position
	value.selectionStart = eventValue.Position
	value.selectionCurrent = eventValue.Position
	value.selectionBase = cloneSelection(selected)
	value.selectionCurrentSet = cloneSelection(selected)
	value.selectionAdditive = selectionModifier(eventValue.Modifiers)
	value.dragStarted = false
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) updateSelectionBox(value *graphState, graph resolvedGraph, viewport Viewport, pixelsPerDP float32, position f32.Point) {
	if !value.dragStarted {
		delta := position.Sub(value.selectionStart)
		if delta.X*delta.X+delta.Y*delta.Y < w.dragThreshold*w.dragThreshold {
			return
		}
		value.dragStarted = true
	}
	value.selectionCurrent = position
	box := graphBox(screenToWorld(viewport, value.selectionStart, pixelsPerDP), screenToWorld(viewport, position, pixelsPerDP))
	next := make(map[string]bool, len(value.selectionBase)+len(graph.nodes))
	if value.selectionAdditive {
		for key := range value.selectionBase {
			next[key] = true
		}
	}
	for _, node := range graph.nodesInRect(box) {
		if node.node.isSelectable(w.nodesSelectable) && selectionBoxMatches(box, node, w.selectionBoxMode) {
			next[node.node.ID] = true
		}
	}
	w.emitSelectionChanges(graph, value.selectionCurrentSet, next)
	value.selectionCurrentSet = next
}

func (w Widget) startConnection(gtx layout.Context, value *graphState, source portHit, eventValue pointer.Event) {
	value.gesture = gestureConnect
	value.pointerID = eventValue.PointerID
	value.connectionSource = source
	value.connectionTarget = portHit{}
	value.connectionCurrent = eventValue.Position
	value.connectionMode = connectionNew
	value.connectionEdge = Edge{}
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) startReconnection(gtx layout.Context, value *graphState, reconnect reconnectHit, eventValue pointer.Event) {
	value.gesture = gestureConnect
	value.pointerID = eventValue.PointerID
	value.connectionSource = reconnect.fixed
	value.connectionTarget = portHit{}
	value.connectionCurrent = eventValue.Position
	value.connectionEdge = reconnect.edge
	if reconnect.endpoint == reconnectSource {
		value.connectionMode = connectionReconnectSource
	} else {
		value.connectionMode = connectionReconnectTarget
	}
	interact.GrabPointer(gtx, value, eventValue)
}

func (w Widget) updateConnection(value *graphState, graph resolvedGraph, viewport Viewport, pixelsPerDP float32, position f32.Point) {
	value.connectionCurrent = position
	world := screenToWorld(viewport, position, pixelsPerDP)
	target, ok := graph.portAt(world, 10/worldScale(viewport, pixelsPerDP))
	if !ok || !graph.portConnectable(target, w.nodesConnectable) {
		value.connectionTarget = portHit{}
		return
	}
	connection := Connection{}
	if value.connectionMode == connectionReconnectSource {
		if !target.output {
			value.connectionTarget = portHit{}
			return
		}
		connection = Connection{Source: target.endpoint, Target: value.connectionSource.endpoint}
	} else {
		if target.output {
			value.connectionTarget = portHit{}
			return
		}
		connection = Connection{Source: value.connectionSource.endpoint, Target: target.endpoint}
	}
	excludedEdgeID := ""
	if value.connectionMode != connectionNew {
		excludedEdgeID = value.connectionEdge.ID
	}
	if !graph.connectionAllowed(connection, excludedEdgeID) || !w.validConnection(connection) {
		value.connectionTarget = portHit{}
		return
	}
	value.connectionTarget = target
}

func (w Widget) completeConnection(value *graphState, graph resolvedGraph, viewport Viewport, pixelsPerDP float32, position f32.Point) {
	w.updateConnection(value, graph, viewport, pixelsPerDP, position)
	if value.connectionTarget.endpoint.NodeID == "" {
		return
	}
	connection := Connection{Source: value.connectionSource.endpoint, Target: value.connectionTarget.endpoint}
	if value.connectionMode == connectionReconnectSource {
		connection = Connection{Source: value.connectionTarget.endpoint, Target: value.connectionSource.endpoint}
	}
	if value.connectionMode == connectionNew {
		if w.onConnect != nil {
			w.onConnect(connection)
		}
		return
	}
	if w.onReconnect != nil {
		w.onReconnect(value.connectionEdge, connection)
	}
}

func (w Widget) updateHover(value *graphState, graph resolvedGraph, selected map[string]bool, viewport Viewport, pixelsPerDP float32, position f32.Point, modifiers key.Modifiers) {
	previousNode := value.hoveredNode
	previousEdge := value.hoveredEdge
	world := screenToWorld(viewport, position, pixelsPerDP)
	value.hoveredNode = graph.nodeAt(world)
	value.hoveredPort = false
	value.hoveredEdge = ""
	value.hoveredResize = 0
	value.hoveredReconnect = false
	if resize, ok := w.resizeAt(graph, selected, position, viewport, pixelsPerDP); ok {
		value.hoveredResize = resize.handle
		w.emitHoverTransitions(previousNode, previousEdge, value.hoveredNode, value.hoveredEdge, world, modifiers)
		return
	}
	if _, ok := w.reconnectAt(graph, world, 10/worldScale(viewport, pixelsPerDP)); ok {
		value.hoveredReconnect = true
		w.emitHoverTransitions(previousNode, previousEdge, value.hoveredNode, value.hoveredEdge, world, modifiers)
		return
	}
	if port, ok := graph.portAt(world, 10/worldScale(viewport, pixelsPerDP)); ok {
		value.hoveredPort = graph.portConnectable(port, w.nodesConnectable)
	}
	if value.hoveredNode == "" && !value.hoveredPort {
		value.hoveredEdge = w.edgeAt(graph, position, viewport, pixelsPerDP)
	}
	w.emitHoverTransitions(previousNode, previousEdge, value.hoveredNode, value.hoveredEdge, world, modifiers)
}

func (w Widget) emitHoverTransitions(previousNode, previousEdge, currentNode, currentEdge string, position Point, modifiers key.Modifiers) {
	if previousNode != currentNode {
		if previousNode != "" && w.onNodeLeave != nil {
			w.onNodeLeave(NodeEvent{NodeID: previousNode, Position: position, Modifiers: modifiers})
		}
		if currentNode != "" && w.onNodeHover != nil {
			w.onNodeHover(NodeEvent{NodeID: currentNode, Position: position, Modifiers: modifiers})
		}
	}
	if previousEdge != currentEdge {
		if previousEdge != "" && w.onEdgeLeave != nil {
			w.onEdgeLeave(EdgeEvent{EdgeID: previousEdge, Position: position, Modifiers: modifiers})
		}
		if currentEdge != "" && w.onEdgeHover != nil {
			w.onEdgeHover(EdgeEvent{EdgeID: currentEdge, Position: position, Modifiers: modifiers})
		}
	}
}

func (value *graphState) updatePressMovement(position f32.Point, slop float32) {
	if value.press.target == graphTargetNone || value.press.moved {
		return
	}
	dx := position.X - value.press.position.X
	dy := position.Y - value.press.position.Y
	value.press.moved = dx*dx+dy*dy > slop*slop
}

func (w Widget) targetAt(graph resolvedGraph, position f32.Point, viewport Viewport, pixelsPerDP float32) (graphTargetKind, string) {
	world := screenToWorld(viewport, position, pixelsPerDP)
	if nodeID := graph.nodeAt(world); nodeID != "" {
		return graphTargetNode, nodeID
	}
	if edgeID := w.edgeAt(graph, position, viewport, pixelsPerDP); edgeID != "" {
		return graphTargetEdge, edgeID
	}
	return graphTargetCanvas, ""
}

func (w Widget) emitContextMenu(graph resolvedGraph, viewport Viewport, pixelsPerDP float32, eventValue pointer.Event) {
	target, id := w.targetAt(graph, eventValue.Position, viewport, pixelsPerDP)
	position := screenToWorld(viewport, eventValue.Position, pixelsPerDP)
	switch target {
	case graphTargetNode:
		if w.onNodeContextMenu != nil {
			w.onNodeContextMenu(NodeEvent{NodeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
	case graphTargetEdge:
		if w.onEdgeContextMenu != nil {
			w.onEdgeContextMenu(EdgeEvent{EdgeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
	case graphTargetCanvas:
		if w.onCanvasContextMenu != nil {
			w.onCanvasContextMenu(CanvasEvent{Position: position, Modifiers: eventValue.Modifiers})
		}
	}
}

func (w Widget) finishPress(value *graphState, graph resolvedGraph, viewport Viewport, pixelsPerDP float32, eventValue pointer.Event) {
	press := value.press
	value.press = graphPress{}
	if press.target == graphTargetNone || press.moved {
		return
	}
	target, id := w.targetAt(graph, eventValue.Position, viewport, pixelsPerDP)
	if target != press.target || id != press.id {
		return
	}
	position := screenToWorld(viewport, eventValue.Position, pixelsPerDP)
	activation := graphActivation{target: target, id: id, position: eventValue.Position, time: eventValue.Time, valid: true}
	double := value.lastActivation.valid && value.lastActivation.target == target && value.lastActivation.id == id &&
		eventValue.Time >= value.lastActivation.time &&
		eventValue.Time-value.lastActivation.time <= nodeGraphDoubleClickWait &&
		activationDistance(value.lastActivation.position, eventValue.Position) <= nodeGraphDoubleClickSlop
	value.lastActivation = activation
	switch target {
	case graphTargetCanvas:
		if w.onCanvasClick != nil {
			w.onCanvasClick(CanvasEvent{Position: position, Modifiers: eventValue.Modifiers})
		}
		if double && w.onCanvasDoubleClick != nil {
			w.onCanvasDoubleClick(CanvasEvent{Position: position, Modifiers: eventValue.Modifiers})
		}
	case graphTargetNode:
		if w.onNodeClick != nil {
			w.onNodeClick(NodeEvent{NodeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
		if double && w.onNodeDoubleClick != nil {
			w.onNodeDoubleClick(NodeEvent{NodeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
	case graphTargetEdge:
		if w.onEdgeClick != nil {
			w.onEdgeClick(EdgeEvent{EdgeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
		if double && w.onEdgeDoubleClick != nil {
			w.onEdgeDoubleClick(EdgeEvent{EdgeID: id, Position: position, Modifiers: eventValue.Modifiers})
		}
	}
}

func activationDistance(first, second f32.Point) float32 {
	dx := second.X - first.X
	dy := second.Y - first.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func (w Widget) edgeAt(graph resolvedGraph, position f32.Point, viewport Viewport, pixelsPerDP float32) string {
	world := screenToWorld(viewport, position, pixelsPerDP)
	margin := 14 / worldScale(viewport, pixelsPerDP)
	for _, index := range graph.spatial.edgeCandidates(graphRect{min: Point{X: world.X - margin, Y: world.Y - margin}, max: Point{X: world.X + margin, Y: world.Y + margin}}) {
		edge := graph.edges[index]
		if graphEdgeDistanceSquared(edge.edge, position, worldToScreen(viewport, edge.source, pixelsPerDP), worldToScreen(viewport, edge.target, pixelsPerDP), 32*worldScale(viewport, pixelsPerDP)) <= 100 {
			return edge.edge.ID
		}
	}
	return ""
}

func (w Widget) reconnectAt(graph resolvedGraph, point Point, radius float32) (reconnectHit, bool) {
	if radius <= 0 || !finite(radius) {
		return reconnectHit{}, false
	}
	radiusSquared := radius * radius
	for index := len(graph.edges) - 1; index >= 0; index-- {
		edge := graph.edges[index]
		if edge.edge.isReconnectable(reconnectSource, w.edgesReconnectable) && squaredDistance(point, edge.source) <= radiusSquared {
			node := graph.byID[edge.edge.Target.NodeID]
			port, ok := node.port(false, edge.edge.Target.PortID)
			if !ok {
				continue
			}
			return reconnectHit{edge: edge.edge, endpoint: reconnectSource, fixed: portHit{endpoint: edge.edge.Target, point: port.point, node: node.node, port: port.port}}, true
		}
		if edge.edge.isReconnectable(reconnectTarget, w.edgesReconnectable) && squaredDistance(point, edge.target) <= radiusSquared {
			node := graph.byID[edge.edge.Source.NodeID]
			port, ok := node.port(true, edge.edge.Source.PortID)
			if !ok {
				continue
			}
			return reconnectHit{edge: edge.edge, endpoint: reconnectTarget, fixed: portHit{endpoint: edge.edge.Source, point: port.point, output: true, node: node.node, port: port.port}}, true
		}
	}
	return reconnectHit{}, false
}

func (w Widget) validConnection(connection Connection) bool {
	return w.isValidConnection == nil || w.isValidConnection(connection)
}

func (state *graphState) resetGesture() {
	state.gesture = gestureNone
	state.dragNodes = nil
	state.dragStarted = false
	state.dragMoved = false
	state.resizeNode = resizeNode{}
	state.resizeMoved = false
	state.selectionBase = nil
	state.selectionCurrentSet = nil
	state.connectionSource = portHit{}
	state.connectionTarget = portHit{}
	state.connectionMode = connectionNew
	state.connectionEdge = Edge{}
	state.minimapOffset = Point{}
}

func (state *graphState) clearHover() {
	state.hoveredNode = ""
	state.hoveredPort = false
	state.hoveredEdge = ""
	state.hoveredResize = 0
	state.hoveredReconnect = false
}

func nodeGraphCursor(w Widget, state *graphState, graph resolvedGraph) pointer.Cursor {
	switch state.gesture {
	case gestureNodeResize:
		return resizeHandleCursor(state.resizeNode.handle)
	case gestureConnect:
		return pointer.CursorCrosshair
	case gestureSelectionBox:
		return pointer.CursorCrosshair
	case gesturePan, gestureMinimap, gestureNodeDrag:
		return pointer.CursorGrabbing
	}
	if state.hoveredResize != 0 {
		return resizeHandleCursor(state.hoveredResize)
	}
	if state.hoveredPort || state.hoveredReconnect {
		return pointer.CursorCrosshair
	}
	if state.hoveredNode != "" {
		node := graph.byID[state.hoveredNode]
		if node.node.isDraggable(w.nodesDraggable) {
			return pointer.CursorGrab
		}
		if node.node.isSelectable(w.nodesSelectable) {
			return pointer.CursorPointer
		}
	}
	if state.hoveredEdge != "" {
		return pointer.CursorPointer
	}
	return pointer.CursorDefault
}

func resizeHandleCursor(handle ResizeHandle) pointer.Cursor {
	switch handle {
	case ResizeHandleTop, ResizeHandleBottom:
		return pointer.CursorNorthSouthResize
	case ResizeHandleLeft, ResizeHandleRight:
		return pointer.CursorEastWestResize
	case ResizeHandleTopLeft, ResizeHandleBottomRight:
		return pointer.CursorNorthWestSouthEastResize
	case ResizeHandleTopRight, ResizeHandleBottomLeft:
		return pointer.CursorNorthEastSouthWestResize
	default:
		return pointer.CursorDefault
	}
}

func selectionModifier(modifiers key.Modifiers) bool {
	return modifiers.Contain(key.ModShift) || modifiers.Contain(key.ModCtrl) || modifiers.Contain(key.ModCommand)
}

func updateSelection(current map[string]bool, nodeID string, mode SelectionMode, toggle bool) map[string]bool {
	next := make(map[string]bool, len(current)+1)
	if mode == SelectionSingle {
		next[nodeID] = true
		return next
	}
	if mode == SelectionMultiple && toggle {
		for key := range current {
			next[key] = true
		}
		if next[nodeID] {
			delete(next, nodeID)
		} else {
			next[nodeID] = true
		}
		return next
	}
	if mode == SelectionMultiple && current[nodeID] {
		for key := range current {
			next[key] = true
		}
		return next
	}
	next[nodeID] = true
	return next
}

func selectionSetsEqual(first, second map[string]bool) bool {
	if len(first) != len(second) {
		return false
	}
	for key := range first {
		if !second[key] {
			return false
		}
	}
	return true
}

func cloneSelection(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source))
	for key := range source {
		result[key] = true
	}
	return result
}

func snappedDragDelta(origin, delta, grid Point) Point {
	return Point{
		X: float32(math.Round(float64((origin.X+delta.X)/grid.X)))*grid.X - origin.X,
		Y: float32(math.Round(float64((origin.Y+delta.Y)/grid.Y)))*grid.Y - origin.Y,
	}
}
